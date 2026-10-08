import { existsSync } from 'node:fs';
import * as path from 'node:path';
import {
	commands,
	Disposable,
	OutputChannel,
	QuickPickOptions,
	Uri,
	window,
	workspace
} from 'vscode';

import { Model } from './model';
import { Repository } from './repository';
import { Ref, Side } from './uri';
import { TdBranch, TdLogEntry, TdStash, TdTag } from './parse';

/** Thông tin mà hộp chọn nhanh hiển thị được cho một dòng dữ liệu. */
interface Picked<T> {
	/** Nhãn chính. */
	label: string;
	/** Mô tả phụ, ví dụ mã băm ngắn. */
	description?: string;
	/** Dòng chi tiết, ví dụ các tham chiếu trỏ tới commit. */
	detail?: string;
	/** Giá trị thật trả về khi người dùng chọn. */
	value: T;
}

/**
 * Đăng ký toàn bộ lệnh của extension.
 *
 * Tên lệnh và cách xuất hiện được khai báo trong package.json, tệp này chỉ lo
 * phần thực thi. Lệnh nhận kho qua tham số đầu tiên khi VS Code gọi từ khung
 * Source Control, còn gọi từ bảng lệnh thì lấy kho đang hoạt động.
 */
export function registerCommands(model: Model, log: OutputChannel): Disposable[] {
	const td = model.cli;

	// ─── Tiện ích nội bộ ─────────────────────────────────────────

	/**
	 * Kho cần thao tác.
	 *
	 * Chấp nhận mọi dạng mà VS Code và các view trong tiện ích truyền vào:
	 * Repository, SourceControl, một hay nhiều SourceControlResourceState, Uri,
	 * đường dẫn. Không nhận ra dạng nào thì lấy kho đang hoạt động.
	 */
	function repositoryOf(candidate: unknown): Repository | undefined {
		if (candidate instanceof Repository) {
			return candidate;
		}
		if (candidate instanceof Uri) {
			return model.find(candidate.fsPath) ?? model.active;
		}
		if (typeof candidate === 'string') {
			return model.byRoot(candidate) ?? model.active;
		}
		const uri = uriOf(candidate);
		if (uri) {
			return model.find(uri.fsPath) ?? model.active;
		}
		if (candidate && typeof candidate === 'object' && 'rootUri' in candidate) {
			return model.fromScm(candidate as { rootUri?: Uri });
		}
		if (Array.isArray(candidate)) {
			return repositoryOf(candidate[0]);
		}
		return model.active;
	}

	/** Bóc ra tất cả các tệp trong một đối số, kể cả khi người dùng chọn nhiều. */
	function urisOf(value: unknown): Uri[] {
		const out: Uri[] = [];
		const push = (item: unknown) => {
			const uri = uriOf(item);
			if (uri && !out.includes(uri)) {
				out.push(uri);
			}
		};
		if (Array.isArray(value)) {
			value.forEach(push);
		} else {
			push(value);
		}
		return out;
	}

	/** Bóc tệp ra khỏi mọi dạng đối tượng mà khung Source Control truyền vào. */
	function uriOf(value: unknown): Uri | undefined {
		if (value instanceof Uri) {
			return value;
		}
		if (Array.isArray(value)) {
			return value.length > 0 ? uriOf(value[0]) : undefined;
		}
		if (value && typeof value === 'object' && 'resourceUri' in value) {
			const inner = (value as { resourceUri?: unknown }).resourceUri;
			return inner instanceof Uri ? inner : undefined;
		}
		return undefined;
	}

	/** Bắt buộc có kho, nếu không thì báo cho người dùng biết. */
	function requireRepository(candidate: unknown, action: string): Repository {
		const repository = repositoryOf(candidate);
		if (!repository) {
			void window.showErrorMessage(`Không có kho td nào để ${action}.`);
			throw new MissingRepositoryError();
		}
		return repository;
	}

	/**
	 * Chạy một thao tác của td rồi báo lỗi nếu hỏng.
	 *
	 * Repository.run đã tự đọc lại trạng thái và hiện thông báo lỗi, nên ở đây
	 * chỉ ghi vào kênh log để người dùng xem lại được.
	 */
	async function attempt(description: string, action: () => Promise<void>): Promise<void> {
		try {
			await action();
			log.appendLine(`✓ ${description}`);
		} catch (error) {
			if (error instanceof MissingRepositoryError) {
				return;
			}
			log.appendLine(`✗ ${description}: ${errorText(error)}`);
		}
	}

	/**
	 * Bước đầu của mọi thao tác thay đổi: kiểm tra kho rồi chạy lệnh td.
	 *
	 * Trong lúc chờ thì báo trên thanh trạng thái, vì một lệnh td trên kho lớn có
	 * thể mất hơn một giây và người dùng cần biết là đang chờ gì.
	 */
	async function mutate(candidate: unknown, action: string, description: string, operation: (repository: Repository) => Promise<void>): Promise<void> {
		let repository: Repository;
		try {
			repository = requireRepository(candidate, action);
		} catch {
			return;
		}
		const busy = window.setStatusBarMessage(`$(sync~spin) ${description}`, 10000);
		try {
			await attempt(description, () => repository.run(action, () => operation(repository)));
		} finally {
			busy.dispose();
		}
	}

	/**
	 * Lấy danh sách đường dẫn tương đối từ những gì người dùng chọn.
	 *
	 * Khung Source Control truyền vào một hoặc nhiều SourceControlResourceState,
	 * còn các view của tiện ích truyền vào Uri. Cả hai đều phải hiểu được, và
	 * chọn nhiều tệp một lúc thì đường dẫn nằm trong mảng.
	 */
	function pathsOf(repository: Repository, values: unknown[]): string[] {
		const out: string[] = [];
		const push = (value: unknown) => {
			const uri = uriOf(value);
			const relative = uri
				? repository.toRelativePath(uri)
				: typeof value === 'string' && value
					? (path.isAbsolute(value) ? repository.toRelativePath(Uri.file(value)) : value)
					: '';
			if (relative && !relative.startsWith('..') && !out.includes(relative)) {
				out.push(relative);
			}
		};
		for (const value of values) {
			if (Array.isArray(value)) {
				value.forEach(push);
			} else {
				push(value);
			}
		}
		return out;
	}

	/** Hiện một hộp chọn nhanh và trả về giá trị người dùng chọn. */
	async function pick<T>(
		items: T[],
		toPick: (item: T) => Picked<T>,
		options: QuickPickOptions & { empty?: string } = {}
	): Promise<T | undefined> {
		if (items.length === 0) {
			void window.showInformationMessage(options.empty ?? 'Không có gì để chọn.');
			return undefined;
		}
		const { empty, ...quickOptions } = options;
		void empty;
		const picked = await window.showQuickPick(items.map(toPick), {
			matchOnDescription: true,
			matchOnDetail: true,
			...quickOptions
		});
		return picked?.value;
	}

	/**
	 * Hộp chọn nhánh.
	 *
	 * `excludeCurrent` loại nhánh đang đứng, dùng cho merge và rebase vì gộp
	 * nhánh vào chính nó là việc vô nghĩa.
	 */
	async function pickBranch(repository: Repository, placeHolder: string, excludeCurrent = false): Promise<TdBranch | undefined> {
		const branches = await td.branches(repository.root);
		const usable = excludeCurrent ? branches.filter(b => !b.current) : branches;
		return pick(usable, branchPick, { placeHolder, empty: 'Kho chưa có nhánh phù hợp.' });
	}

	/** Hộp chọn commit trong lịch sử. */
	async function pickCommitEntry(repository: Repository, placeHolder: string): Promise<TdLogEntry | undefined> {
		return pick(await td.log(repository.root, 200), commitPick, { placeHolder, empty: 'Kho chưa có commit nào.' });
	}

	/** Hộp chọn bản lưu tạm. */
	async function pickStashEntry(repository: Repository, placeHolder: string): Promise<TdStash | undefined> {
		return pick(await td.stashes(repository.root), stashPick, { placeHolder, empty: 'Kho chưa có bản lưu tạm nào.' });
	}

	/** Hộp chọn tag. */
	async function pickTagEntry(repository: Repository, placeHolder: string): Promise<TdTag | undefined> {
		return pick(await td.tags(repository.root), tagPick, { placeHolder, empty: 'Kho chưa có tag nào.' });
	}

	// ─── Làm mới và khởi tạo ────────────────────────────────────

	async function refresh(): Promise<void> {
		await model.refresh();
	}

	async function init(): Promise<void> {
		const folders = workspace.workspaceFolders;
		if (!folders || folders.length === 0) {
			void window.showErrorMessage('Hãy mở một thư mục trước khi khởi tạo kho td.');
			return;
		}
		const folder = folders.length === 1 ? folders[0] : await window.showWorkspaceFolderPick();
		if (!folder) {
			return;
		}
		const branch = await window.showInputBox({ prompt: 'Tên nhánh khởi tạo', value: 'main' });
		if (branch === undefined) {
			return;
		}
		await attempt(`khởi tạo kho tại ${folder.uri.fsPath}`, async () => {
			await td.init(folder.uri.fsPath, branch.trim() || undefined);
			await model.discover();
		});
	}

	// ─── Stage và unstage ───────────────────────────────────────

	/** Đưa một tệp hoặc cả thư mục vào vùng chuẩn bị. */
	async function stage(candidate: unknown, ...values: unknown[]): Promise<void> {
		let repository: Repository;
		try {
			repository = requireRepository(candidate, 'đưa thay đổi vào vùng chuẩn bị');
		} catch {
			return;
		}
		// Đối số đầu tiên cũng có thể là tệp: khung Source Control truyền thẳng
		// resource state vào lệnh được gọi từ menu ngữ cảnh.
		const list = pathsOf(repository, [candidate, ...values]);
		if (list.length === 0) {
			return;
		}
		await mutate(repository, 'Không đưa được thay đổi vào vùng chuẩn bị', `stage ${list.join(', ')}`,
			r => td.stage(r.root, list));
	}

	/** Gỡ một tệp khỏi vùng chuẩn bị, đưa nó về đúng nội dung của HEAD. */
	async function unstage(candidate: unknown, ...values: unknown[]): Promise<void> {
		let repository: Repository;
		try {
			repository = requireRepository(candidate, 'gỡ thay đổi khỏi vùng chuẩn bị');
		} catch {
			return;
		}
		const list = pathsOf(repository, [candidate, ...values]);
		if (list.length === 0) {
			return;
		}
		await mutate(repository, 'Không gỡ được thay đổi khỏi vùng chuẩn bị', `unstage ${list.join(', ')}`,
			r => td.unstage(r.root, list));
	}

	/** Đưa mọi thay đổi vào vùng chuẩn bị. */
	async function stageAll(candidate: unknown, trackedOnly = false): Promise<void> {
		await mutate(candidate, 'đưa mọi thay đổi vào vùng chuẩn bị', 'stage tất cả', async repository => {
			if (trackedOnly) {
				await td.stageAllTracked(repository.root);
			} else {
				await td.stageAll(repository.root);
			}
		});
	}

	/** Đưa riêng các tệp chưa được theo dõi vào vùng chuẩn bị. */
	async function stageAllUntracked(candidate: unknown): Promise<void> {
		let repository: Repository;
		try {
			repository = requireRepository(candidate, 'đưa tệp chưa theo dõi vào vùng chuẩn bị');
		} catch {
			return;
		}
		const list = repository.current.untracked.map(e => e.path);
		if (list.length === 0) {
			return;
		}
		await mutate(repository, 'Không đưa được tệp chưa theo dõi vào vùng chuẩn bị', `stage ${list.length} tệp chưa theo dõi`,
			r => td.stage(r.root, list));
	}

	/**
	 * Đánh dấu xung đột đã giải quyết.
	 *
	 * Xung đột của td nằm ngay trong vùng chuẩn bị nên đưa tất cả vào đó là xong.
	 */
	async function stageAllMerge(candidate: unknown): Promise<void> {
		await stageAll(candidate);
	}

	async function unstageAll(candidate: unknown): Promise<void> {
		let repository: Repository;
		try {
			repository = requireRepository(candidate, 'gỡ mọi thay đổi khỏi vùng chuẩn bị');
		} catch {
			return;
		}
		const list = repository.current.staged.map(e => e.path);
		if (list.length === 0) {
			return;
		}
		await mutate(repository, 'Không gỡ được thay đổi khỏi vùng chuẩn bị', `unstage ${list.length} tệp`,
			r => td.unstage(r.root, list));
	}

	// ─── Huỷ thay đổi ───────────────────────────────────────────

	/** Huỷ sửa đổi trên đĩa, lấy lại nội dung đang ở vùng chuẩn bị. */
	async function discard(repository: Repository, list: string[], description: string): Promise<void> {
		if (list.length === 0) {
			return;
		}
		const answer = await window.showWarningMessage(
			`${description} ${list.length} tệp? Sửa đổi chưa commit sẽ mất.`,
			{ modal: true },
			'Huỷ thay đổi'
		);
		if (answer !== 'Huỷ thay đổi') {
			return;
		}
		await mutate(repository, 'Không huỷ được thay đổi trên đĩa', `${description} ${list.join(', ')}`,
			r => td.discard(r.root, list));
	}

	async function revertChange(candidate: unknown, ...values: unknown[]): Promise<void> {
		let repository: Repository;
		try {
			repository = requireRepository(candidate, 'huỷ thay đổi');
		} catch {
			return;
		}
		const list = pathsOf(repository, [candidate, ...values]);
		await discard(repository, list, 'Huỷ thay đổi của');
	}

	async function cleanAllTracked(candidate: unknown): Promise<void> {
		let repository: Repository;
		try {
			repository = requireRepository(candidate, 'huỷ mọi thay đổi');
		} catch {
			return;
		}
		const list = repository.current.unstaged.filter(e => e.status !== '?').map(e => e.path);
		await discard(repository, list, 'Huỷ thay đổi của');
	}

	/**
	 * Xoá tệp chưa được theo dõi.
	 *
	 * td không hỏi lại nên phần xác nhận do giao diện đảm nhiệm, đúng như các
	 * thao tác huỷ thay đổi khác.
	 */
	async function cleanAllUntracked(candidate: unknown, ...values: unknown[]): Promise<void> {
		let repository: Repository;
		try {
			repository = requireRepository(candidate, 'xoá tệp chưa được theo dõi');
		} catch {
			return;
		}
		const list = pathsOf(repository, [candidate, ...values]);
		const targets = list.length > 0
			? list
			: repository.current.untracked.map(e => e.path);
		if (targets.length === 0) {
			void window.showInformationMessage('Không có tệp chưa được theo dõi nào.');
			return;
		}
		const answer = await window.showWarningMessage(
			`Xoá ${targets.length} tệp chưa được td theo dõi? Nội dung của chúng sẽ mất và không cứu được trong kho.`,
			{ modal: true },
			'Xoá'
		);
		if (answer !== 'Xoá') {
			return;
		}
		await mutate(repository, 'Không xoá được tệp chưa được theo dõi', `xoá ${targets.length} tệp chưa theo dõi`,
			r => td.clean(r.root, targets));
	}

	// ─── Commit ─────────────────────────────────────────────────

	async function commit(candidate: unknown, all = false, amend = false, allowEmpty = false): Promise<void> {
		let repository: Repository;
		try {
			repository = requireRepository(candidate, 'ghi lại thay đổi');
		} catch {
			return;
		}
		let message = repository.inputValue.trim();
		if (message.length === 0) {
			if (!all && !amend && !allowEmpty && !repository.hasChanges()) {
				void window.showWarningMessage('Không có thay đổi nào để commit.');
				return;
			}
			const typed = await window.showInputBox({
				prompt: amend ? 'Nội dung commit sau khi sửa lại' : 'Nội dung commit',
				placeHolder: 'td đòi nội dung commit, không nhận nội dung rỗng',
				ignoreFocusOut: true,
				validateInput: value => value.trim().length === 0 ? 'Commit cần có nội dung' : undefined
			});
			if (typed === undefined || typed.trim().length === 0) {
				return;
			}
			message = typed.trim();
		}
		const body = message.trim();
		const verb = amend ? 'sửa lại commit' : allowEmpty ? 'tạo commit rỗng' : 'commit';
		await mutate(repository, 'Không ghi được commit', `${verb}: ${headLine(message)}`, async r => {
			if (all) {
				await td.stageAll(r.root);
			}
			if (allowEmpty) {
				await td.commitEmpty(r.root, body);
			} else {
				await td.commit(r.root, body, amend);
			}
		});
		repository.clearInput();
		void window.setStatusBarMessage('$(check) Đã ghi commit.', 3000);
	}

	async function commitStaged(candidate: unknown): Promise<void> {
		await commit(candidate, false);
	}

	async function commitAll(candidate: unknown): Promise<void> {
		await commit(candidate, true);
	}

	async function commitAmend(candidate: unknown): Promise<void> {
		await commit(candidate, true, true);
	}

	async function commitStagedAmend(candidate: unknown): Promise<void> {
		await commit(candidate, false, true);
	}

	async function commitEmpty(candidate: unknown): Promise<void> {
		if (workspace.getConfiguration('td').get<boolean>('confirmEmptyCommits', true)) {
			const answer = await window.showWarningMessage(
				'Tạo commit rỗng? Không tệp nào được ghi vào mốc này.',
				{ modal: true },
				'Tạo'
			);
			if (answer !== 'Tạo') {
				return;
			}
		}
		await commit(candidate, false, false, true);
	}

	// ─── Mở tệp và khung so sánh ───────────────────────────────

	/** Mở tệp, hoặc mở khung so sánh nếu cấu hình yêu cầu. */
	async function openFile(candidate: unknown, second?: unknown): Promise<void> {
		const uri = uriOf(second) ?? uriOf(candidate);
		if (!uri) {
			return;
		}
		const repository = repositoryOf(uri);
		if (repository && workspace.getConfiguration('td', uri).get<boolean>('openDiffOnClick', false)) {
			await openChanges(repository, [uri]);
			return;
		}
		await window.showTextDocument(uri);
	}

	/** Mở khung so sánh cho các tệp đang chọn trong khung Source Control. */
	async function openChange(candidate: unknown, second?: unknown): Promise<void> {
		const uris = urisOf(second ?? candidate);
		let repository: Repository;
		try {
			repository = requireRepository(candidate, 'xem thay đổi');
		} catch {
			return;
		}
		if (uris.length > 0) {
			await openChanges(repository, uris);
		}
	}

	/** Mở tệp như nó nằm trong HEAD. */
	async function openHEADFile(candidate: unknown, second?: unknown): Promise<void> {
		const uri = uriOf(second) ?? uriOf(candidate);
		let repository: Repository;
		try {
			repository = requireRepository(candidate, 'xem nội dung trong HEAD');
		} catch {
			return;
		}
		if (!uri) {
			return;
		}
		const relative = repository.toRelativePath(uri);
		const exists = await td.diff(repository.root, { revision: 'HEAD', paths: [relative] });
		if (exists.length === 0) {
			void window.showWarningMessage(`${relative} không có trong HEAD.`);
			return;
		}
		await window.showTextDocument(repository.uriFor(relative, Ref.Head, Side.Old));
	}

	/** Mở khung so sánh của một commit trong view Commits. */
	async function openCommitChanges(candidate: unknown, hash: string | undefined): Promise<void> {
		let repository: Repository;
		try {
			repository = requireRepository(candidate, 'xem thay đổi của commit');
		} catch {
			return;
		}
		if (!hash) {
			return;
		}
		const patches = await td.diff(repository.root, { revision: hash });
		if (patches.length === 0) {
			void window.showInformationMessage('Commit này không thay đổi tệp nào.');
			return;
		}
		if (patches.length === 1) {
			const only = patches[0];
			await commands.executeCommand(
				'vscode.diff',
				repository.uriFor(only.path, hash, Side.Old),
				repository.uriFor(only.path, hash, Side.New),
				repository.toAbsolutePath(only.path),
				only.path
			);
			return;
		}
		await commands.executeCommand(
			'vscode.changes',
			repository.root,
			patches.map(p => repository.toAbsolutePath(p.path)),
			repository.root
		);
	}

	/** Mở khung so sánh, một tệp thì mở cặp, nhiều tệp thì mở giao diện nhiều tệp. */
	async function openChanges(repository: Repository, uris: Uri[]): Promise<void> {
		if (uris.length === 0) {
			return;
		}
		if (uris.length > 1) {
			await commands.executeCommand('vscode.changes', repository.root, uris, repository.root);
			return;
		}
		const uri = uris[0];
		const relative = repository.toRelativePath(uri);
		const state = repository.current;
		const inIndex = state.staged.some(e => e.path === relative);
		const inWorktree = state.unstaged.some(e => e.path === relative)
			|| state.untracked.some(e => e.path === relative);

		// Chọn hai phía theo đúng ý nghĩa, giống git:
		// đã stage hết thì so HEAD với vùng chuẩn bị; còn sửa trên đĩa thì so
		// vùng chuẩn bị với chính tệp trên đĩa.
		let left: Uri;
		let right: Uri;
		let title: string;
		if (inIndex && !inWorktree) {
			left = repository.uriFor(relative, Ref.Head, Side.Old);
			right = repository.uriFor(relative, Ref.Index, Side.New);
			title = `${relative} (Staged Changes)`;
		} else {
			left = repository.uriFor(relative, Ref.Index, Side.Old);
			// Tệp đã xoá khỏi đĩa thì phía phải là bản rỗng, cũng lấy từ vùng
			// chuẩn bị cho nhất quán.
			right = existsSync(uri.fsPath)
				? uri
				: repository.uriFor(relative, Ref.Index, Side.New);
			title = `${relative} (Working Tree)`;
		}
		await commands.executeCommand('vscode.diff', left, right, uri, title);
	}

	// ─── Nhánh ──────────────────────────────────────────────────

	async function checkout(candidate: unknown, name?: string): Promise<void> {
		let repository: Repository;
		try {
			repository = requireRepository(candidate, 'chuyển nhánh');
		} catch {
			return;
		}
		if (!name) {
			name = (await pickBranch(repository, 'Chọn nhánh để chuyển sang'))?.name;
		}
		if (!name) {
			return;
		}
		await mutate(repository, 'Không chuyển được nhánh', `checkout ${name}`,
			r => td.switchTo(r.root, name as string));
	}

	async function createBranch(candidate: unknown, from?: string): Promise<void> {
		let repository: Repository;
		try {
			repository = requireRepository(candidate, 'tạo nhánh');
		} catch {
			return;
		}
		const name = await window.showInputBox({
			prompt: from ? `Tên nhánh mới, tạo từ ${from}` : 'Tên nhánh mới',
			value: from ?? '',
			validateInput: value => value.trim().length === 0 ? 'Cần tên nhánh' : undefined
		});
		if (!name || !name.trim()) {
			return;
		}
		await mutate(repository, 'Không tạo được nhánh', `tạo nhánh ${name.trim()}`,
			r => td.createBranch(r.root, name.trim(), from));
	}

	/** Tạo nhánh từ một điểm xuất phát được chọn, tương đương `td vcs branch ten <điểm>`. */
	async function branchFrom(candidate: unknown): Promise<void> {
		let repository: Repository;
		try {
			repository = requireRepository(candidate, 'tạo nhánh từ một điểm khác');
		} catch {
			return;
		}
		const entries = await td.log(repository.root, 50);
		if (entries.length === 0) {
			void window.showInformationMessage('Kho chưa có commit nào để tạo nhánh.');
			return;
		}
		const chosen = await pick(entries, commitPick, { placeHolder: 'Chọn điểm xuất phát cho nhánh mới' });
		if (!chosen) {
			return;
		}
		await createBranch(repository, chosen.hash);
	}

	async function deleteBranch(candidate: unknown, name?: string): Promise<void> {
		let repository: Repository;
		try {
			repository = requireRepository(candidate, 'xoá nhánh');
		} catch {
			return;
		}
		let target = name;
		if (!target) {
			const branch = await pickBranch(repository, 'Chọn nhánh cần xoá', true);
			target = branch?.name;
		}
		if (!target) {
			return;
		}
		const answer = await window.showWarningMessage(`Xoá nhánh ${target}?`, { modal: true }, 'Xoá');
		if (answer !== 'Xoá') {
			return;
		}
		await mutate(repository, 'Không xoá được nhánh', `xoá nhánh ${target}`,
			r => td.deleteBranch(r.root, target as string));
	}

	async function renameBranch(candidate: unknown, name?: string): Promise<void> {
		let repository: Repository;
		try {
			repository = requireRepository(candidate, 'đổi tên nhánh');
		} catch {
			return;
		}
		let from = name;
		if (!from) {
			from = (await pickBranch(repository, 'Chọn nhánh cần đổi tên'))?.name;
		}
		if (!from) {
			return;
		}
		const to = await window.showInputBox({ prompt: `Tên mới cho nhánh ${from}`, value: from });
		if (!to || !to.trim() || to.trim() === from) {
			return;
		}
		await mutate(repository, 'Không đổi được tên nhánh', `đổi tên ${from} thành ${to.trim()}`,
			r => td.renameBranch(r.root, from as string, to.trim()));
	}

	async function merge(candidate: unknown, name?: string): Promise<void> {
		let repository: Repository;
		try {
			repository = requireRepository(candidate, 'hợp nhất');
		} catch {
			return;
		}
		if (!name) {
			name = (await pickBranch(repository, 'Chọn nhánh cần hợp nhất vào nhánh hiện tại', true))?.name;
		}
		if (!name) {
			return;
		}
		await mutate(repository, 'Không hợp nhất được', `merge ${name}`,
			r => td.merge(r.root, name as string));
	}

	async function rebase(candidate: unknown, name?: string): Promise<void> {
		let repository: Repository;
		try {
			repository = requireRepository(candidate, 'đặt lại lịch sử');
		} catch {
			return;
		}
		if (!name) {
			name = (await pickBranch(repository, 'Chọn nhánh làm gốc để rebase', true))?.name;
		}
		if (!name) {
			return;
		}
		await mutate(repository, 'Không rebase được', `rebase ${name}`,
			r => td.rebase(r.root, name as string));
	}

	// ─── Thao tác trên một commit cụ thể ────────────────────────

	async function cherryPick(candidate: unknown, hash?: string): Promise<void> {
		let repository: Repository;
		try {
			repository = requireRepository(candidate, 'cherry pick');
		} catch {
			return;
		}
		const target = hash ?? (await pickCommitEntry(repository, 'Chọn commit cần cherry pick'))?.hash;
		if (!target) {
			return;
		}
		await mutate(repository, 'Không cherry pick được', `cherry pick ${target}`,
			r => td.cherryPick(r.root, target));
	}

	async function revert(candidate: unknown, hash?: string): Promise<void> {
		let repository: Repository;
		try {
			repository = requireRepository(candidate, 'hoàn tác commit');
		} catch {
			return;
		}
		const target = hash ?? (await pickCommitEntry(repository, 'Chọn commit cần hoàn tác'))?.hash;
		if (!target) {
			return;
		}
		await mutate(repository, 'Không hoàn tác được', `revert ${target}`,
			r => td.revert(r.root, target));
	}

	// ─── Lưu tạm ────────────────────────────────────────────────

	async function stash(candidate: unknown, includeUntracked = false): Promise<void> {
		let repository: Repository;
		try {
			repository = requireRepository(candidate, 'lưu tạm thay đổi');
		} catch {
			return;
		}
		if (!repository.hasChanges()) {
			void window.showInformationMessage('Không có thay đổi nào để lưu tạm.');
			return;
		}
		const message = await window.showInputBox({
			prompt: 'Mô tả cho bản lưu tạm (bỏ trống nếu không cần)',
			ignoreFocusOut: true
		});
		if (message === undefined) {
			return;
		}
		await mutate(repository, 'Không lưu tạm được thay đổi', `stash${includeUntracked ? ' kèm tệp chưa theo dõi' : ''}`,
			r => td.stash(r.root, message.trim() || undefined, includeUntracked));
	}

	/** Áp dụng hoặc lấy lại một bản lưu tạm. */
	async function applyStash(candidate: unknown, pop = false, given?: number): Promise<void> {
		let repository: Repository;
		try {
			repository = requireRepository(candidate, pop ? 'lấy lại bản lưu tạm' : 'áp dụng bản lưu tạm');
		} catch {
			return;
		}
		let index = given;
		if (index === undefined) {
			const stash = await pickStashEntry(repository, pop ? 'Chọn bản lưu tạm để lấy lại' : 'Chọn bản lưu tạm để áp dụng');
			index = stash?.index ?? 0;
		}
		await mutate(repository, 'Không áp dụng được bản lưu tạm', `stash ${pop ? 'pop' : 'apply'} ${index}`,
			r => td.applyStash(r.root, index as number, pop));
	}

	async function dropStash(candidate: unknown, given?: number): Promise<void> {
		let repository: Repository;
		try {
			repository = requireRepository(candidate, 'xoá bản lưu tạm');
		} catch {
			return;
		}
		let index = given;
		if (index === undefined) {
			const stash = await pickStashEntry(repository, 'Chọn bản lưu tạm cần xoá');
			index = stash?.index ?? 0;
		}
		await mutate(repository, 'Không xoá được bản lưu tạm', `stash drop ${index}`,
			r => td.dropStash(r.root, index as number));
	}

	async function dropAllStashes(candidate: unknown): Promise<void> {
		let repository: Repository;
		try {
			repository = requireRepository(candidate, 'xoá toàn bộ bản lưu tạm');
		} catch {
			return;
		}
		const answer = await window.showWarningMessage('Xoá toàn bộ bản lưu tạm?', { modal: true }, 'Xoá');
		if (answer !== 'Xoá') {
			return;
		}
		await mutate(repository, 'Không xoá được bản lưu tạm', 'xoá toàn bộ bản lưu tạm',
			r => td.clearStashes(r.root));
	}

	// ─── Tag ────────────────────────────────────────────────────

	async function createTag(candidate: unknown, hash?: string): Promise<void> {
		let repository: Repository;
		try {
			repository = requireRepository(candidate, 'tạo tag');
		} catch {
			return;
		}
		const at = hash ?? (await pickCommitEntry(repository, 'Chọn commit cần gắn tag'))?.hash;
		if (!at) {
			return;
		}
		const name = await window.showInputBox({ prompt: `Tên tag tại ${at}`, value: 'v1.0.0' });
		if (!name || !name.trim()) {
			return;
		}
		const note = await window.showInputBox({
			prompt: 'Chú thích cho tag (bỏ trống để tạo tag nhẹ)',
			ignoreFocusOut: true
		});
		if (note === undefined) {
			return;
		}
		await mutate(repository, 'Không tạo được tag', `tạo tag ${name.trim()}`, async r => {
			const status = r.current;
			const needDetach = !status.detached && !!status.branch && !status.branch.startsWith(at) && !at.startsWith(status.head);
			if (!needDetach) {
				await td.createTag(r.root, name.trim(), note.trim() || undefined);
				return;
			}
			// td chỉ tạo tag tại HEAD, nên muốn gắn tag ở commit khác thì phải
			// rời nhánh trong lúc tạo rồi quay lại.
			await td.detach(r.root, at);
			try {
				await td.createTag(r.root, name.trim(), note.trim() || undefined);
			} finally {
				await td.switchTo(r.root, status.branch);
			}
		});
	}

	async function deleteTag(candidate: unknown, name?: string): Promise<void> {
		let repository: Repository;
		try {
			repository = requireRepository(candidate, 'xoá tag');
		} catch {
			return;
		}
		const target = name ?? (await pickTagEntry(repository, 'Chọn tag cần xoá'))?.name;
		if (!target) {
			return;
		}
		await mutate(repository, 'Không xoá được tag', `xoá tag ${target}`,
			r => td.deleteTag(r.root, target));
	}

	// ─── Lệnh phụ ───────────────────────────────────────────────

	async function revealInExplorer(candidate: unknown, second?: unknown): Promise<void> {
		const repository = repositoryOf(second ?? candidate);
		if (!repository) {
			return;
		}
		const uri = uriOf(second) ?? uriOf(candidate) ?? repository.toAbsolutePath('.tdx');
		await commands.executeCommand('revealInExplorer', uri);
	}

	async function pickBranchCommand(): Promise<void> {
		const repository = repositoryOf(undefined);
		if (repository) {
			await pickBranch(repository, 'Chọn nhánh');
		}
	}

	async function pickCommitCommand(): Promise<void> {
		const repository = repositoryOf(undefined);
		if (repository) {
			await pickCommitEntry(repository, 'Chọn một commit');
		}
	}

	async function pickTagCommand(): Promise<void> {
		const repository = repositoryOf(undefined);
		if (repository) {
			await pickTagEntry(repository, 'Chọn một tag');
		}
	}

	// ─── Đăng ký ────────────────────────────────────────────────

	const registry: [string, (...args: any[]) => unknown][] = [
		['td.refresh', refresh],
		['td.init', init],
		['td.commit', commit],
		['td.commitAll', commitAll],
		['td.commitStaged', commitStaged],
		['td.commitAmend', commitAmend],
		['td.commitStagedAmend', commitStagedAmend],
		['td.commitEmpty', commitEmpty],
		['td.stage', stage],
		['td.stageAll', () => stageAll(undefined)],
		['td.stageAllTracked', () => stageAll(undefined, true)],
		['td.stageAllUntracked', stageAllUntracked],
		['td.stageAllMerge', stageAllMerge],
		['td.unstage', unstage],
		['td.unstageAll', unstageAll],
		['td.cleanAll', cleanAllTracked],
		['td.cleanAllTracked', cleanAllTracked],
		['td.cleanAllUntracked', cleanAllUntracked],
		['td.revertChange', revertChange],
		['td.openFile', openFile],
		['td.openFile2', openFile],
		['td.openChange', openChange],
		['td.openHEADFile', openHEADFile],
		['td.openCommitChanges', openCommitChanges],
		['td.checkout', checkout],
		['td.branch', () => createBranch(undefined)],
		['td.branchFrom', branchFrom],
		['td.deleteBranch', deleteBranch],
		['td.renameBranch', renameBranch],
		['td.merge', merge],
		['td.rebase', rebase],
		['td.cherryPick', cherryPick],
		['td.revert', revert],
		['td.stash', () => stash(undefined, false)],
		['td.stashIncludeUntracked', () => stash(undefined, true)],
		['td.stashApply', () => applyStash(undefined, false)],
		['td.stashApplyLatest', () => applyStash(undefined, false, 0)],
		['td.stashPop', () => applyStash(undefined, true)],
		['td.stashPopLatest', () => applyStash(undefined, true, 0)],
		['td.stashDrop', dropStash],
		['td.stashDropAll', dropAllStashes],
		['td.tag', createTag],
		['td.deleteTag', deleteTag],
		['td.revealInExplorer', revealInExplorer],
		['td.showOutput', () => log.show(true)],
		['td.pickBranch', pickBranchCommand],
		['td.pickCommit', pickCommitCommand],
		['td.pickTag', pickTagCommand]
	];

	const disposables = registry.map(([id, handler]) => commands.registerCommand(id, handler));
	disposables.push(commands.registerCommand('td.pickStash', () => applyStash(undefined, true)));
	return disposables;
}

/** Báo ra chỗ cần kho nhưng không có kho nào, để lệnh dừng im lặng. */
class MissingRepositoryError extends Error {}

/** Chuyển một nhánh thành mục của hộp chọn. */
function branchPick(branch: TdBranch): Picked<TdBranch> {
	const parts = [branch.subject || branch.hash];
	if (branch.upstream) {
		parts.push(`đi trước ${branch.ahead}, đi sau ${branch.behind}`);
	}
	return {
		label: branch.name,
		description: parts[0],
		detail: parts[1],
		value: branch
	};
}

/** Chuyển một commit thành mục của hộp chọn. */
function commitPick(entry: TdLogEntry): Picked<TdLogEntry> {
	return {
		label: entry.summary || '(không có tiêu đề)',
		description: entry.hash,
		detail: entry.refs.length > 0 ? entry.refs.join(', ') : undefined,
		value: entry
	};
}

/** Chuyển một bản lưu tạm thành mục của hộp chọn. */
function stashPick(stash: TdStash): Picked<TdStash> {
	return {
		label: stash.message || `stash@{${stash.index}}`,
		description: `stash@{${stash.index}}`,
		value: stash
	};
}

/** Chuyển một tag thành mục của hộp chọn. */
function tagPick(tag: TdTag): Picked<TdTag> {
	return {
		label: tag.name,
		description: tag.message || tag.hash,
		value: tag
	};
}

/** Lấy dòng đầu tiên của nội dung commit, dùng cho thông báo ngắn. */
function headLine(message: string): string {
	const trimmed = message.trim();
	const index = trimmed.indexOf('\n');
	return index < 0 ? trimmed : trimmed.slice(0, index);
}

/** Lấy thông điệp của một lỗi bất kỳ. */
function errorText(error: unknown): string {
	return error instanceof Error ? error.message : String(error);
}