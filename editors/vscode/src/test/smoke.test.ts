import { execFileSync } from 'node:child_process';
import { rmSync } from 'node:fs';
import { test } from 'node:test';

import {
	fakeVscode,
	findTm,
	loadExtension,
	makeRepo,
	manifest,
	runTm,
	writeRepoFile
} from './harness';

const TM_BIN = findTm();
const runSmoke = TM_BIN ? test : test.skip;

/**
 * Một tệp trong SourceControl mà khung so sánh sẽ mở khi bấm vào.
 *
 * Kiểm thử đi qua cả ba tầng: tệp trong nhóm thay đổi, lệnh mà VS Code gọi khi
 * bấm vào tệp đó, rồi tham số truyền cho `vscode.diff`. Sai ở tầng nào thì
 * người dùng đều thấy cùng một kết quả: bấm tệp không ra khung so sánh.
 */
interface ResourceState {
	resourceUri: { fsPath: string };
	command: { command: string; arguments: unknown[] };
}

/**
 * Chạy phần kích hoạt của tiện ích với API VS Code giả.
 *
 * Bản giả đủ để phần kích hoạt chạy hết: dò kho, đọc trạng thái, đăng ký lệnh,
 * đăng ký bốn khung. Kiểm thử này bắt được lỗi chỉ xuất hiện lúc chạy, ví dụ
 * gọi sai API hoặc quên đăng ký lệnh nào đó.
 */
runSmoke('kích hoạt: dò kho, đọc trạng thái và đăng ký lệnh', async () => {
	const { vscode, state } = fakeVscode(TM_BIN!);

	const root = makeRepo(TM_BIN!, 'tm-smoke-');
	writeRepoFile(root, 'a.txt', 'một\nhai\n');
	writeRepoFile(root, 'thu-muc/b.txt', 'x\n');
	runTm(TM_BIN!, root, ['add', '.']);
	runTm(TM_BIN!, root, ['commit', '-m', 'c1']);
	// Rời kho ở trạng thái có thay đổi: một tệp đã stage, một tệp sửa trên đĩa,
	// một tệp mới chưa theo dõi.
	writeRepoFile(root, 'a.txt', 'một\nhai sửa\n');
	runTm(TM_BIN!, root, ['add', 'a.txt']);
	writeRepoFile(root, 'thu-muc/b.txt', 'x\ny\n');
	writeRepoFile(root, 'moi.txt', 'mới\n');

	const workspace = (vscode.workspace as { workspaceFolders: unknown });
	workspace.workspaceFolders = [{ uri: state.uri(root), name: 'kho', index: 0 }];

	const extension = loadExtension(vscode);
	const context = { subscriptions: [] as { dispose(): void }[] };
	try {
		extension.activate(context);

		// activate() gọi việc dò kho theo lời hứa, chờ nốt thì đó.
		await new Promise(resolve => setTimeout(resolve, 1500));

		const pkg = manifest();

		const missing = pkg.contributes.commands
			.map(c => c.command)
			.filter(id => !state.registeredCommands.has(id));
		if (missing.length > 0) {
			throw new Error(`các lệnh khai báo mà không đăng ký: ${missing.join(', ')}`);
		}

		if (state.sourceControls.length !== 1) {
			throw new Error(`phải có đúng một SourceControl, nhận ${state.sourceControls.length}`);
		}
		const source = state.sourceControls[0] as {
			label: string;
			count: number;
			inputBox: { placeholder: string };
			acceptInputCommand: { command: string } | undefined;
		};
		if (source.acceptInputCommand?.command !== 'tm.commit') {
			throw new Error('ô nhập commit phải có nút commit');
		}
		if (!source.inputBox.placeholder.includes("'main'")) {
			throw new Error(`ô nhập phải nhắc nhánh đang đứng, nhận "${source.inputBox.placeholder}"`);
		}

		// Bốn nhóm thay đổi phải đúng tên và đúng số tệp.
		const byId = new Map(state.resourceGroups.map(g => [g.id, g]));
		if (byId.get('index')?.resourceStates.length !== 1) {
			throw new Error('nhóm Staged Changes phải có đúng một tệp');
		}
		if (byId.get('workingTree')?.resourceStates.length !== 1) {
			throw new Error('nhóm Changes phải có đúng một tệp');
		}
		if (byId.get('untracked')?.resourceStates.length !== 1) {
			throw new Error('nhóm Untracked Changes phải có đúng một tệp');
		}
		if (source.count !== 3) {
			throw new Error(`số đếm phải là 3, nhận ${source.count}`);
		}

		// Bấm vào tệp trong khung Source Control phải mở khung so sánh. Đây là
		// hành vi mặc định của extension git và của tiện ích này.
		const working = byId.get('workingTree')?.resourceStates as ResourceState[] | undefined;
		const row = working?.[0];
		if (row?.command.command !== 'tm.openChange') {
			throw new Error(`bấm tệp phải mở khung so sánh, nhận "${row?.command.command}"`);
		}

		const leaks = context.subscriptions.filter(d => typeof (d as { dispose?: unknown }).dispose !== 'function');
		if (leaks.length > 0) {
			throw new Error(`có ${leaks.length} tài nguyên không huỷ được: ${leaks.map(d => d?.constructor?.name).join(', ')}`);
		}
		for (const d of context.subscriptions) {
			try {
				d.dispose();
			} catch (error) {
				throw new Error(`không huỷ được ${(d as object)?.constructor?.name}: ${String(error)}`);
			}
		}
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});

/**
 * Mở khung so sánh của một tệp phải gọi `vscode.diff` đúng ba đối số.
 *
 * Lệnh này kiểm tra tham số theo tên: đối số thứ ba là tiêu đề và phải là chuỗi.
 * Truyền thừa một địa chỉ vào vị trí đó thì VS Code báo *Invalid argument
 * 'title'* và không mở khung nào.
 */
runSmoke('mở thay đổi của một tệp: vscode.diff nhận (trái, phải, tiêu đề)', async () => {
	const { vscode, state } = fakeVscode(TM_BIN!);
	const root = makeRepo(TM_BIN!, 'tm-diff1-');
	writeRepoFile(root, 'a.txt', 'một\nhai\n');
	runTm(TM_BIN!, root, ['add', '.']);
	runTm(TM_BIN!, root, ['commit', '-m', 'c1']);
	writeRepoFile(root, 'a.txt', 'một\nhai sửa\n');
	runTm(TM_BIN!, root, ['add', 'a.txt']);

	const workspace = (vscode.workspace as { workspaceFolders: unknown });
	workspace.workspaceFolders = [{ uri: state.uri(root), name: 'kho', index: 0 }];

	const extension = loadExtension(vscode);
	const context = { subscriptions: [] as { dispose(): void }[] };
	try {
		extension.activate(context);
		await new Promise(resolve => setTimeout(resolve, 1500));

		await state.registeredCommands.get('tm.openChange')?.(undefined, state.uri(`${root}/a.txt`));

		const call = state.lastCall('vscode.diff');
		if (!call) {
			throw new Error(`phải mở khung so sánh bằng vscode.diff, thấy: ${state.calls.map(c => c.id).join(', ')}`);
		}
		if (call.args.length !== 3) {
			throw new Error(`vscode.diff nhận đúng ba đối số, nhận ${call.args.length}: ${JSON.stringify(call.args)}`);
		}
		const [left, right, title] = call.args as [{ scheme: string }, { scheme: string }, unknown];
		if (typeof title !== 'string' || !title.includes('a.txt')) {
			throw new Error(`đối số thứ ba phải là tiêu đề có tên tệp, nhận ${JSON.stringify(title)}`);
		}
		if (left.scheme !== 'tm' || right.scheme !== 'tm') {
			throw new Error(`hai phía phải là địa chỉ ảo của tiện ích, nhận ${left.scheme} và ${right.scheme}`);
		}
		if (state.lastCall('vscode.changes')) {
			throw new Error('một tệp thì không được mở khung so sánh nhiều tệp');
		}
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});

/**
 * Chọn nhiều tệp phải mở khung so sánh nhiều tệp với đúng hình dạng đối số.
 *
 * `vscode.changes` nhận danh sách bộ ba gồm địa chỉ hiển thị, phía gốc và phía
 * đã sửa. Truyền danh sách địa chỉ thuần thì VS Code báo *Invalid argument
 * 'resourceList'* và không mở gì cả.
 */
runSmoke('mở nhiều tệp: vscode.changes nhận (tiêu đề, [địa chỉ, gốc, đã sửa])', async () => {
	const { vscode, state } = fakeVscode(TM_BIN!);
	const root = makeRepo(TM_BIN!, 'tm-diff2-');
	writeRepoFile(root, 'a.txt', 'một\n');
	writeRepoFile(root, 'b.txt', 'hai\n');
	runTm(TM_BIN!, root, ['add', '.']);
	runTm(TM_BIN!, root, ['commit', '-m', 'c1']);
	writeRepoFile(root, 'a.txt', 'một sửa\n');
	writeRepoFile(root, 'b.txt', 'hai sửa\n');

	const workspace = (vscode.workspace as { workspaceFolders: unknown });
	workspace.workspaceFolders = [{ uri: state.uri(root), name: 'kho', index: 0 }];

	const extension = loadExtension(vscode);
	const context = { subscriptions: [] as { dispose(): void }[] };
	try {
		extension.activate(context);
		await new Promise(resolve => setTimeout(resolve, 1500));

		await state.registeredCommands.get('tm.openChange')?.(undefined, [state.uri(`${root}/a.txt`), state.uri(`${root}/b.txt`)]);

		const call = state.lastCall('vscode.changes');
		if (!call) {
			throw new Error(`phải mở khung so sánh nhiều tệp bằng vscode.changes, thấy: ${state.calls.map(c => c.id).join(', ')}`);
		}
		const [title, list] = call.args as [unknown, unknown];
		if (typeof title !== 'string') {
			throw new Error(`đối số đầu phải là tiêu đề, nhận ${JSON.stringify(title)}`);
		}
		if (!Array.isArray(list) || list.length !== 2) {
			throw new Error(`phải có hai mục, nhận ${JSON.stringify(list)}`);
		}
		for (const entry of list) {
			if (!Array.isArray(entry) || entry.length !== 3) {
				throw new Error(`mỗi mục phải là bộ ba địa chỉ, nhận ${JSON.stringify(entry)}`);
			}
			for (const uri of entry as { scheme?: string }[]) {
				if (uri?.scheme !== 'file' && uri?.scheme !== 'tm') {
					throw new Error(`mỗi địa chỉ phải là file hoặc tm, nhận ${JSON.stringify(uri)}`);
				}
			}
		}
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});

/**
 * Bấm vào một commit phải mở khung so sánh với đúng các tệp commit đó sửa.
 *
 * Trước đây lệnh này truyền vào `vscode.changes` danh sách địa chỉ thuần nên
 * VS Code từ chối và báo lỗi, người dùng không thấy tệp nào được đổi.
 */
runSmoke('bấm vào một commit: mở đúng các tệp commit đó thay đổi', async () => {
	const { vscode, state } = fakeVscode(TM_BIN!);
	const root = makeRepo(TM_BIN!, 'tm-commit-');
	writeRepoFile(root, 'a.txt', 'một\n');
	writeRepoFile(root, 'b.txt', 'hai\n');
	runTm(TM_BIN!, root, ['add', '.']);
	runTm(TM_BIN!, root, ['commit', '-m', 'c1']);
	writeRepoFile(root, 'a.txt', 'một sửa\n');
	writeRepoFile(root, 'moi.txt', 'mới\n');
	runTm(TM_BIN!, root, ['add', '.']);
	runTm(TM_BIN!, root, ['commit', '-m', 'c2']);
	writeRepoFile(root, 'b.txt', 'hai sửa\n');

	const workspace = (vscode.workspace as { workspaceFolders: unknown });
	workspace.workspaceFolders = [{ uri: state.uri(root), name: 'kho', index: 0 }];

	const extension = loadExtension(vscode);
	const context = { subscriptions: [] as { dispose(): void }[] };
	try {
		extension.activate(context);
		await new Promise(resolve => setTimeout(resolve, 1500));

		// Lấy mã băm của commit c2 bằng lệnh tm, tránh phụ thuộc thứ tự hiển thị.
		const entries = execFileSync(TM_BIN!, ['vcs', '-C', root, 'log', '--oneline', '-n1', 'HEAD~1'])
			.toString()
			.trim()
			.split(/\s+/)[0];

		await state.registeredCommands.get('tm.openCommitChanges')?.(undefined, entries);

		const call = state.lastCall('vscode.changes');
		if (!call) {
			throw new Error(`phải mở khung so sánh bằng vscode.changes, thấy: ${state.calls.map(c => c.id).join(', ')}`);
		}
				const list = call.args[1] as { fsPath: string }[][];
		const paths = list.map(entry => entry[0].fsPath.replace(/\\/g, '/')).map(p => p.slice(p.lastIndexOf('/'))).sort();
		if (paths.length !== 2 || paths[0] !== '/a.txt' || paths[1] !== '/moi.txt') {
			throw new Error(`phải đúng hai tệp của commit c2, nhận ${JSON.stringify(paths)}`);
		}
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});

/**
 * Phía nằm trong HEAD phải là nội dung thật của tệp ở HEAD.
 *
 * Cách dựng lại từ khác biệt chỉ đúng với tệp đang chờ commit. Tệp sạch thì
 * không nằm trong khác biệt nào, nên phía này bị bỏ trống và khung so sánh hiện
 * sai. Lệnh `tm vcs show-file` đọc thẳng trong kho nên đúng với mọi tệp.
 */
runSmoke('nội dung ở HEAD đọc bằng show-file, kể cả tệp không nằm trong khác biệt', async () => {
	const { vscode, state } = fakeVscode(TM_BIN!);
	const root = makeRepo(TM_BIN!, 'tm-head-');
	writeRepoFile(root, 'sach.txt', 'không đổi\n');
	runTm(TM_BIN!, root, ['add', '.']);
	runTm(TM_BIN!, root, ['commit', '-m', 'c1']);

	const workspace = (vscode.workspace as { workspaceFolders: unknown });
	workspace.workspaceFolders = [{ uri: state.uri(root), name: 'kho', index: 0 }];

	const extension = loadExtension(vscode);
	const context = { subscriptions: [] as { dispose(): void }[] };
	try {
		extension.activate(context);
		await new Promise(resolve => setTimeout(resolve, 1500));

		// Lệnh mở tệp như nó nằm trong HEAD phải ra tài liệu ảo của HEAD.
		await state.registeredCommands.get('tm.openHEADFile')?.(undefined, state.uri(`${root}/sach.txt`));
		const opened = state.opened[state.opened.length - 1] as { scheme: string; path: string };
		if (opened?.scheme !== 'tm') {
			throw new Error(`phải mở tài liệu ảo của HEAD, nhận ${JSON.stringify(opened)}`);
		}
		const content = await state.contentOf(opened);
		if (content !== 'không đổi\n') {
			throw new Error(`phía HEAD phải là nội dung thật của tệp, nhận ${JSON.stringify(content)}`);
		}
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});

/**
 * Hai phía của tệp bị xoá và tệp chưa theo dõi phải đúng là rỗng và không rỗng.
 *
 * Tệp bị xoá khỏi đĩa thì phía phải phải là tài liệu rỗng, đọc tệp thật sẽ báo
 * lỗi và khung so sánh hiện thông báo lỗi thay vì hiện tệp đã xoá đi.
 */
runSmoke('khung so sánh của tệp bị xoá và tệp chưa theo dõi lấy đúng hai phía', async () => {
	const { vscode, state } = fakeVscode(TM_BIN!);
	const root = makeRepo(TM_BIN!, 'tm-hai-pha-');
	writeRepoFile(root, 'xoa.txt', 'còn trong kho\n');
	runTm(TM_BIN!, root, ['add', '.']);
	runTm(TM_BIN!, root, ['commit', '-m', 'c1']);
	rmSync(`${root}/xoa.txt`);
	writeRepoFile(root, 'moi.txt', 'tệp mới\n');

	const workspace = (vscode.workspace as { workspaceFolders: unknown });
	workspace.workspaceFolders = [{ uri: state.uri(root), name: 'kho', index: 0 }];

	const extension = loadExtension(vscode);
	const context = { subscriptions: [] as { dispose(): void }[] };
	try {
		extension.activate(context);
		await new Promise(resolve => setTimeout(resolve, 1500));

		// Tệp bị xoá: phía gốc là nội dung đã stage, phía phải là rỗng.
		await state.registeredCommands.get('tm.openChange')?.(undefined, state.uri(`${root}/xoa.txt`));
		const deleted = state.lastCall('vscode.diff');
		const [oldSide, newSide] = deleted?.args as [{ scheme: string }, { scheme: string }];
		if (oldSide.scheme !== 'tm' || newSide.scheme !== 'tm') {
			throw new Error('tệp bị xoá vẫn phải so hai phía ảo trong kho');
		}
		if (await state.contentOf(oldSide) !== 'còn trong kho\n') {
			throw new Error('phía gốc của tệp bị xoá phải là nội dung đã stage');
		}
		if (await state.contentOf(newSide) !== '') {
			throw new Error('phía phải của tệp bị xoá phải rỗng');
		}

		// Tệp chưa theo dõi: phía gốc rỗng, phía phải là chính tệp trên đĩa.
		await state.registeredCommands.get('tm.openChange')?.(undefined, state.uri(`${root}/moi.txt`));
		const added = state.lastCall('vscode.diff');
		const [emptySide, fileSide] = added?.args as [{ scheme: string }, { scheme: string; fsPath: string }];
		if (fileSide.scheme !== 'file') {
			throw new Error(`phía phải của tệp chưa theo dõi phải là tệp trên đĩa, nhận ${fileSide.scheme}`);
		}
		if (await state.contentOf(emptySide) !== '') {
			throw new Error('phía gốc của tệp chưa theo dõi phải rỗng');
		}
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});

/**
 * Nút *Open File* trên khung so sánh phải mở tệp thật trên đĩa.
 *
 * Nút được gọi với địa chỉ ảo của tiện ích; mở đúng địa chỉ đó thì chỉ mở lại
 * nội dung đang xem, người dùng tưởng lệnh không chạy.
 */
runSmoke('mở tệp từ khung so sánh thì ra tệp thật trên đĩa', async () => {
	const { vscode, state } = fakeVscode(TM_BIN!);
	const root = makeRepo(TM_BIN!, 'tm-openfile-');
	writeRepoFile(root, 'a.txt', 'một\n');
	runTm(TM_BIN!, root, ['add', '.']);
	runTm(TM_BIN!, root, ['commit', '-m', 'c1']);
	writeRepoFile(root, 'a.txt', 'một sửa\n');

	const workspace = (vscode.workspace as { workspaceFolders: unknown });
	workspace.workspaceFolders = [{ uri: state.uri(root), name: 'kho', index: 0 }];

	const extension = loadExtension(vscode);
	const context = { subscriptions: [] as { dispose(): void }[] };
	try {
		extension.activate(context);
		await new Promise(resolve => setTimeout(resolve, 1500));

		await state.registeredCommands.get('tm.openChange')?.(undefined, state.uri(`${root}/a.txt`));
		const [, right] = state.lastCall('vscode.diff')?.args as [{ scheme: string }, { scheme: string }];

		await state.registeredCommands.get('tm.openFile')?.(undefined, right);
		const opened = state.opened[state.opened.length - 1] as { scheme: string; fsPath: string };
		if (opened?.scheme !== 'file') {
			throw new Error(`phải mở tệp trên đĩa, nhận ${JSON.stringify(opened)}`);
		}
		if (opened.fsPath !== `${root}/a.txt`) {
			throw new Error(`sai đường dẫn tệp, nhận ${opened.fsPath}`);
		}
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});