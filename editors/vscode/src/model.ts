import { readdirSync } from 'node:fs';
import * as path from 'node:path';
import {
	commands,
	Disposable,
	EventEmitter,
	OutputChannel,
	Uri,
	window,
	workspace
} from 'vscode';

import { TmDecorations } from './decorations';
import { Repository } from './repository';
import { findRepositoryRoot, resolveExecutable, Tm } from './tm';
import { TmBranchesProvider, TmCommitsProvider, TmStashesProvider, TmTagsProvider } from './views';

/** Thư mục out/ của kho tm, nơi build_all.sh đặt tệp thực thi. */
const REPO_OUT = path.join(__dirname, '..', '..', '..', 'out');

/**
 * Quản lý các kho tm trong workspace.
 *
 * Model là nơi duy nhất biết kho nào đang mở, khi nào cần đọc lại trạng thái và
 * các view dùng chung sẽ lấy dữ liệu từ đâu. Mỗi kho là một `Repository`.
 */
export class Model implements Disposable {
	private readonly repositories = new Map<string, Repository>();
	private readonly disposables: Disposable[] = [];
	private readonly onDidChangeRepositoryEmitter = new EventEmitter<Repository>();

	/** Bắn khi kho nào đó có trạng thái mới. */
	readonly onDidChangeRepository = this.onDidChangeRepositoryEmitter.event;

	private readonly decorations: TmDecorations;
	private readonly branches: TmBranchesProvider;
	private readonly commits: TmCommitsProvider;
	private readonly stashes: TmStashesProvider;
	private readonly tags: TmTagsProvider;

	private readonly tm: Tm;
	private refreshing = false;
	private refreshQueued = false;

	/** Kho đang được dùng cho các lệnh không nhận kho cụ thể. */
	private activeRoot: string | undefined;

	/** Số kho đang mở, phục vụ điều kiện when trong package.json. */
	private openRepositoryCount = 0;

	/** Lệnh tm không tìm thấy trên máy hay không. */
	private missing = false;

	constructor(
		/** Kênh log dùng chung, để người dùng thấy lệnh tm đã chạy. */
		private readonly log: OutputChannel
	) {
		const config = workspace.getConfiguration('tm');
		this.tm = new Tm(
			resolveExecutable(config.get<string>('path', ''), [path.join(REPO_OUT, '..')]),
			line => this.log.appendLine(line)
		);

		this.decorations = new TmDecorations(() => this.all);
		this.disposables.push(this.decorations);
		this.decorations.setEnabled(config.get<boolean>('decorations.enabled', true));

		this.branches = new TmBranchesProvider(this);
		this.commits = new TmCommitsProvider(this);
		this.stashes = new TmStashesProvider(this);
		this.tags = new TmTagsProvider(this);
		this.disposables.push(
			window.registerTreeDataProvider('tmBranches', this.branches),
			window.registerTreeDataProvider('tmCommits', this.commits),
			window.registerTreeDataProvider('tmStashes', this.stashes),
			window.registerTreeDataProvider('tmTags', this.tags)
		);

		// Mọi kho đều báo về cùng một Model, nên các view chỉ cần theo dõi kho
		// đang hoạt động là đủ dữ liệu.
		this.disposables.push(
			workspace.onDidChangeWorkspaceFolders(() => void this.discover()),
			workspace.onDidChangeConfiguration(event => {
				this.applyDecorationSetting();
				// Đổi cấu hình tm thì dò lại kho: vừa để áp dụng cấu hình mới,
				// vừa làm mới trạng thái ngay một lần cho khỏi phải chờ đổi tệp.
				if (event.affectsConfiguration('tm')) {
					void this.discover(true);
				}
			})
		);

		// Bộ theo dõi luôn có sẵn, cờ autoRefresh kiểm tra lúc hẹn giờ. Nhờ vậy
		// bật hay tắt giữa chừng không cần dựng lại bộ theo dõi.
		this.disposables.push(this.watchWorkspace());
	}

	/** Danh sách kho đang mở. */
	get all(): readonly Repository[] {
		return [...this.repositories.values()];
	}

	/** Lệnh tm dùng cho mọi kho. */
	get cli(): Tm {
		return this.tm;
	}

	/** Tệp thực thi tm đang dùng, hiện ra trong kênh log. */
	get command(): string {
		return this.tm.command;
	}

	/** Kho đang hoạt động: kho chứa tệp đang mở, nếu không thì kho đầu tiên. */
	get active(): Repository | undefined {
		const uri = window.activeTextEditor?.document.uri;
		if (uri && uri.scheme === 'file') {
			const found = this.find(uri.fsPath);
			if (found) {
				return found;
			}
		}
		if (this.activeRoot) {
			const repository = this.repositories.get(this.activeRoot);
			if (repository) {
				return repository;
			}
		}
		return this.all[0];
	}

	/** Kho chứa một đường dẫn, tìm kho gần nhất phía trên. */
	find(fsPath: string): Repository | undefined {
		for (const repository of this.all) {
			const relative = path.relative(repository.root, fsPath);
			if (!relative.startsWith('..') && !path.isAbsolute(relative)) {
				return repository;
			}
		}
		return undefined;
	}

	/** Kho ứng với một SourceControl do VS Code truyền vào, nếu có. */
	fromScm(scm: { rootUri?: Uri } | undefined): Repository | undefined {
		if (!scm?.rootUri) {
			return this.active;
		}
		return this.repositories.get(scm.rootUri.fsPath) ?? this.active;
	}

	/** Kho theo thư mục gốc. */
	byRoot(root: string): Repository | undefined {
		return this.repositories.get(root);
	}

	/**
	 * Dò kho trong workspace rồi mở hoặc đóng cho khớp.
	 *
	 * Dùng cách đi lên từng thư mục con thay vì gọi `tm vcs status` ở mọi nơi
	 * nên việc dò kho không tốn tiến trình nào và trạng thái Source Control có
	 * ngay khi cửa sổ mở ra.
	 */
	async discover(force = false): Promise<void> {
		if (!workspace.getConfiguration('tm').get<boolean>('enabled', true)) {
			return;
		}
		const folders = workspace.workspaceFolders ?? [];
		const found = new Set<string>();

		for (const folder of folders) {
			if (folder.uri.scheme !== 'file') {
				continue;
			}
			const root = findRepositoryRoot(folder.uri.fsPath);
			if (root) {
				found.add(root);
				continue;
			}
			// Thư mục mở không phải gốc kho thì tìm trong các thư mục con một
			// tầng, đủ để nhận ra kho lồng nhau.
			for (const child of firstLevelChildren(folder.uri.fsPath)) {
				const childRoot = findRepositoryRoot(child);
				if (childRoot) {
					found.add(childRoot);
				}
			}
		}

		for (const root of found) {
			if (!this.repositories.has(root)) {
				await this.open(root);
			} else if (force) {
				await this.refresh(root);
			}
		}
		for (const repository of this.all) {
			if (!found.has(repository.root)) {
				await this.close(repository);
			}
		}
		this.setRepositoryCount();
	}

	/** Mở một kho và bắt đầu theo dõi trạng thái của nó. */
	private async open(root: string): Promise<void> {
		const repository = new Repository(root, this.tm);
		this.repositories.set(root, repository);
		this.activeRoot = this.activeRoot ?? root;
		// Mọi lần trạng thái kho đổi đều phải vẽ lại chữ viết tắt và thanh trạng
		// thái, kể cả lần đầu tiên sau khi mở kho.
		repository.onDidChangeState(() => {
			this.decorations.refresh();
			this.updateStatusBar();
			// Bốn khung phía bên đều nghe sự kiện này để nạp lại dữ liệu.
			this.onDidChangeRepositoryEmitter.fire(repository);
		});
		await repository.refresh();
	}

	/** Đóng một kho, thường là vì thư mục của nó đã bị xoá khỏi workspace. */
	private async close(repository: Repository): Promise<void> {
		this.repositories.delete(repository.root);
		this.setRepositoryCount();
		if (this.activeRoot === repository.root) {
			this.activeRoot = this.all[0]?.root;
		}
		repository.dispose();
		this.decorations.refresh();
		this.updateStatusBar();
	}

	/** Cập nhật số kho đang mở vào context key cho các điều kiện when. */
	private setRepositoryCount(): void {
		if (this.openRepositoryCount === this.repositories.size) {
			return;
		}
		this.openRepositoryCount = this.repositories.size;
		void commands.executeCommand('setContext', 'tm.openRepositoryCount', this.openRepositoryCount);
	}

	/** Đọc lại trạng thái một kho. */
	async refresh(root?: string): Promise<void> {
		if (this.refreshing) {
			// Nhiều thay đổi tệp xảy ra liên tiếp thì gộp lại làm một lần.
			this.refreshQueued = true;
			return;
		}
		this.refreshing = true;
		try {
			const targets = root ? [this.repositories.get(root)] : this.all;
			for (const repository of targets) {
				await repository?.refresh();
			}
		} finally {
			this.refreshing = false;
			if (this.refreshQueued) {
				this.refreshQueued = false;
				void this.refresh();
			}
		}
	}

	/** Nối cờ `tm.decorations.enabled` vào phần vẽ chữ viết tắt. */
	private applyDecorationSetting(): void {
		this.decorations.setEnabled(workspace.getConfiguration('tm').get<boolean>('decorations.enabled', true));
		this.decorations.refresh();
	}

	/** Đặt các nút ở thanh trạng thái cho mọi kho, giống git. */
	private updateStatusBar(): void {
		for (const repository of this.all) {
			const commands: { command: string; title: string; tooltip: string }[] = [];
			const head = repository.headLabel;
			commands.push({
				command: 'tm.checkout',
				title: `${repository.isBusy ? '$(loading~spin) ' : ''}$(git-branch) ${head}`,
				tooltip: `${head}, ${repository.isBusy ? 'đang xử lý' : 'chuyển nhánh hoặc tag'}`
			});
			const status = repository.current;
			if (status.hasUpstream && (status.ahead > 0 || status.behind > 0)) {
				commands.push({
					command: 'tm.showOutput',
					title: `$(arrow-down) ${status.behind} $(arrow-up) ${status.ahead}`,
					tooltip: 'Nhánh đang đi trước hoặc đi sau nhánh theo dõi'
				});
			}
			if (repository.changeCount > 0) {
				commands.push({
					command: 'tm.commit',
					title: `$(check) ${repository.changeCount}`,
					tooltip: `${repository.changeCount} tệp đang chờ commit`
				});
			}
			if (repository.error) {
				commands.push({
					command: 'tm.showOutput',
					title: '$(error)',
					tooltip: repository.error
				});
			}
			repository.statusBarCommands = commands;
		}
	}

	/**
	 * Theo dõi mọi thay đổi tệp trong workspace để làm mới trạng thái.
	 *
	 * Bộ theo dõi đặt một lần cho cả workspace chứ không đặt theo từng kho, vì
	 * kho có thể mới xuất hiện sau đó (người dùng chạy `tm vcs init` trong
	 * terminal). Mỗi lần có thay đổi, hàm dò lại kho rồi đọc lại trạng thái, vừa
	 * mở kho mới vừa làm mới kho cũ.
	 *
	 * Các lần ghi liên tiếp được gộp lại thành một, nên lưu một tệp không tạo
	 * ra nhiều lần gọi tm.
	 */
	private watchWorkspace(): Disposable {
		let timer: NodeJS.Timeout | undefined;
		const delay = () => workspace.getConfiguration('tm').get<number>('autorefreshDelay', 1000);
		const schedule = () => {
			if (!workspace.getConfiguration('tm').get<boolean>('autoRefresh', true)) {
				return;
			}
			if (timer) {
				clearTimeout(timer);
			}
			timer = setTimeout(() => {
				timer = undefined;
				void this.discover(true);
			}, delay());
		};

		const watchers: Disposable[] = [];
		// Tệp trên đĩa: thêm, sửa, xoá đều làm thay đổi trạng thái kho.
		const worktree = workspace.createFileSystemWatcher('**/*');
		watchers.push(worktree);
		worktree.onDidCreate(schedule);
		worktree.onDidChange(schedule);
		worktree.onDidDelete(schedule);
		// Dữ liệu trong .tmx cũng vậy, ví dụ khi có tiến trình khác ghi vào kho.
		const data = workspace.createFileSystemWatcher('**/.tmx/**');
		watchers.push(data);
		data.onDidCreate(schedule);
		data.onDidChange(schedule);
		data.onDidDelete(schedule);

		const reconfigure = workspace.onDidChangeConfiguration(event => {
			if (event.affectsConfiguration('tm.autoRefresh') || event.affectsConfiguration('tm.autorefreshDelay')) {
				void this.discover(true);
			}
		});

		return new Disposable(() => {
			if (timer) {
				clearTimeout(timer);
			}
			for (const d of watchers) {
				d.dispose();
			}
			reconfigure.dispose();
		});
	}

	/**
	 * Kiểm tra xem máy đã có lệnh tm chưa.
	 *
	 * Khi không có, khung Source Control hiện lời nhắc cài đặt thay vì trống
	 * trơn, đúng như extension git làm khi thiếu git.
	 */
	async checkMissing(): Promise<void> {
		if (this.repositories.size > 0) {
			this.setMissing(false);
			return;
		}
		const missing = !(await this.tm.available());
		this.setMissing(missing);
		if (missing) {
			this.log.appendLine(`không chạy được lệnh "${this.tm.command}", xem tm.path trong cấu hình`);
			void window.showWarningMessage(
				`Không tìm thấy lệnh tm (đang thử "${this.tm.command}"). Cài tm hoặc đặt đường dẫn ở cấu hình tm.path.`
			);
		}
	}

	/** Ghi cờ tm.missing cho các điều kiện when trong package.json. */
	private setMissing(value: boolean): void {
		if (this.missing === value) {
			return;
		}
		this.missing = value;
		void commands.executeCommand('setContext', 'tm.missing', value);
	}

	dispose(): void {
		for (const repository of this.all) {
			repository.dispose();
		}
		this.repositories.clear();
		for (const d of this.disposables.splice(0)) {
			d.dispose();
		}
		this.onDidChangeRepositoryEmitter.dispose();
	}
}

/** Liệt kê các thư mục con trực tiếp, dùng khi dò kho lồng nhau. */
function firstLevelChildren(dir: string): string[] {
	try {
		return readdirSync(dir, { withFileTypes: true })
			.filter(entry => entry.isDirectory() && !entry.name.startsWith('.'))
			.map(entry => path.join(dir, entry.name));
	} catch {
		return [];
	}
}