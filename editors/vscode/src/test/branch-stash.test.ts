import { execFileSync } from 'node:child_process';
import { readFileSync, rmSync } from 'node:fs';
import * as path from 'node:path';
import { test } from 'node:test';

import {
	fakeVscode,
	FakeState,
	findTm,
	loadExtension,
	makeRepo,
	runTm,
	writeRepoFile
} from './harness';

const TM_BIN = findTm();
const runSmoke = TM_BIN ? test : test.skip;

/** Đọc trạng thái kho bằng chính lệnh tm mà người dùng gõ trên terminal. */
function statusOf(root: string): string {
	return execFileSync(TM_BIN!, ['vcs', '-C', root, 'status'], { encoding: 'utf8' });
}

/** Tên nhánh đang đứng, đọc bằng lệnh CLI để so với kết quả phía tiện ích. */
function currentBranch(root: string): string {
	return execFileSync(TM_BIN!, ['vcs', '-C', root, 'branch', '--show-current'], { encoding: 'utf8' }).trim();
}

/**
 * Danh sách bản lưu tạm, mỗi dòng là thông điệp của một bản.
 *
 * Lệnh in một dòng thông báo riêng khi danh sách trống, nên phần đó bị bỏ qua
 * để phần đối chiếu chỉ còn những bản lưu tạm thật sự tồn tại.
 */
function stashList(root: string): string {
	const raw = execFileSync(TM_BIN!, ['vcs', '-C', root, 'stash', '--list'], { encoding: 'utf8' });
	return raw.split('\n').filter(line => line.trim().startsWith('stash@{')).join('\n').trim();
}

/** Đọc nguyên văn một tệp trong kho, để so nội dung trên đĩa trước và sau. */
function readIn(root: string, relative: string): string {
	return readFileSync(path.join(root, relative), 'utf8');
}

/**
 * Dựng một kho đã có commit rồi kích hoạt tiện ích trên kho đó.
 *
 * Trả về kho cùng trạng thái của bản giả để kiểm thử điều khiển hộp thoại và
 * soi những lệnh VS Code đã được yêu cầu chạy.
 */
async function activateRepo(root: string): Promise<FakeState> {
	const { vscode, state } = fakeVscode(TM_BIN!);
	const workspace = (vscode.workspace as { workspaceFolders: unknown });
	workspace.workspaceFolders = [{ uri: state.uri(root), name: 'kho', index: 0 }];
	const extension = loadExtension(vscode);
	extension.activate({ subscriptions: [] as { dispose(): void }[] });
	await state.settle();
	return state;
}

/** Kho có một commit, dùng làm điểm xuất phát cho các kiểm thử. */
function seededRepo(prefix: string): string {
	const root = makeRepo(TM_BIN!, prefix);
	writeRepoFile(root, 'a.txt', 'dòng 1\ndòng 2\n');
	runTm(TM_BIN!, root, ['add', '.']);
	runTm(TM_BIN!, root, ['commit', '-m', 'c1']);
	return root;
}

/** Chạy một lệnh của tiện ích và chờ cho nó xong. */
async function run(state: FakeState, id: string, ...args: unknown[]): Promise<void> {
	await state.registeredCommands.get(id)?.(...args);
	await state.settle(300);
}

/**
 * Ghi một tệp rồi báo cho tiện ích biết, đúng như lúc người dùng lưu tệp.
 *
 * Bản giả không tự phát hiện thay đổi trên đĩa nên phải bắn sự kiện thủ công,
 * nếu không thì tiện ích vẫn giữ trạng thái cũ và các lệnh như lưu tạm sẽ tưởng
 * là không có thay đổi gì.
 */
async function writeAndNotify(state: FakeState, root: string, relative: string, content: string): Promise<void> {
	writeRepoFile(root, relative, content);
	state.fireFileEvent('change', path.join(root, relative));
	await state.settle();
}

/**
 * Tạo nhánh mới rồi chuyển sang đó, đúng như người dùng bấm *Create Branch*.
 *
 * Lệnh hỏi tên nhánh qua hộp nhập, nên kiểm thử xếp sẵn câu trả lời.
 */
runSmoke('tạo nhánh: nhập tên rồi phải tạo và chuyển sang đó', async () => {
	const root = seededRepo('tm-branch-');
	try {
		const state = await activateRepo(root);
		state.answers.input.push('tinh-moi');

		await run(state, 'tm.branch');

		if (currentBranch(root) !== 'tinh-moi') {
			throw new Error(`phải chuyển sang nhánh vừa tạo, nhận ${currentBranch(root)}`);
		}
		const branches = execFileSync(TM_BIN!, ['vcs', '-C', root, 'branch', '-l'], { encoding: 'utf8' });
		if (!branches.includes('tinh-moi') || !branches.includes('main')) {
			throw new Error(`phải có cả hai nhánh, nhận ${JSON.stringify(branches)}`);
		}
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});

/**
 * Bỏ trống tên nhánh thì không được tạo gì cả.
 */
runSmoke('tạo nhánh: bỏ trống tên thì không tạo nhánh nào', async () => {
	const root = seededRepo('tm-branch-');
	try {
		const state = await activateRepo(root);
		state.answers.input.push('   ');

		await run(state, 'tm.branch');

		if (currentBranch(root) !== 'main') {
			throw new Error(`phải ở lại nhánh cũ, nhận ${currentBranch(root)}`);
		}
		const branches = execFileSync(TM_BIN!, ['vcs', '-C', root, 'branch', '-l'], { encoding: 'utf8' });
		if (branches.trim().split('\n').length !== 1) {
			throw new Error(`chỉ được có một nhánh, nhận ${JSON.stringify(branches)}`);
		}
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});

/**
 * Tạo nhánh từ một commit đã chọn, tương đương `tm vcs branch ten <commit>`.
 */
runSmoke('tạo nhánh từ một commit đã chọn trong danh sách', async () => {
	const root = seededRepo('tm-branchfrom-');
	writeRepoFile(root, 'b.txt', 'thêm\n');
	runTm(TM_BIN!, root, ['add', '.']);
	runTm(TM_BIN!, root, ['commit', '-m', 'c2']);
	try {
		const state = await activateRepo(root);

		// Chọn commit c1 (nghĩa là chọn mục thứ hai trong danh sách log).
		const entries = execFileSync(TM_BIN!, ['vcs', '-C', root, 'log', '--oneline', '-n2'], { encoding: 'utf8' })
			.split('\n').filter(Boolean);
		const older = entries[1].split(' ')[0];
		state.answers.pick.push({ label: 'c1', description: older, refs: [], value: { hash: older, summary: 'c1', refs: [] } });
		state.answers.input.push('tu-commit');

		await run(state, 'tm.branchFrom');

		// Nhánh mới phải trỏ về commit c1, không phải HEAD hiện tại. `branch -v` in mã
		// băm rút gọn nên so phần đầu, và dấu * đứng trước nhánh đang đứng.
		const branches = execFileSync(TM_BIN!, ['vcs', '-C', root, 'branch', '-v'], { encoding: 'utf8' });
		const line = branches.split('\n').find(l => l.includes('tu-commit')) ?? '';
		const short = line.replace(/^\*?\s*/, '').split(' ')[1] ?? '';
		if (!older.startsWith(short) || short.length < 7) {
			throw new Error(`nhánh mới phải trỏ về ${older}, nhận ${JSON.stringify(line)}`);
		}
		// Và nhánh đó khác HEAD, tức là đã tạo đúng từ commit được chọn.
		const head = (execFileSync(TM_BIN!, ['vcs', '-C', root, 'branch', '-v'], { encoding: 'utf8' })
			.split('\n').find(l => l.includes('main')) ?? '').replace(/^\*?\s*/, '').split(' ')[1] ?? '';
		if (head === short) {
			throw new Error(`nhánh mới phải khác nhánh đang đứng, cả hai đều ở ${head}`);
		}
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});

/**
 * Chuyển nhánh rồi chuyển lại, kiểm tra thanh trạng thái đổi theo.
 */
runSmoke('chuyển nhánh: qua lại giữa hai nhánh', async () => {
	const root = seededRepo('tm-switch-');
	try {
		const state = await activateRepo(root);
		runTm(TM_BIN!, root, ['switch', '-c', 'nhánh-phụ']);

		await run(state, 'tm.checkout', undefined, 'main');
		if (currentBranch(root) !== 'main') {
			throw new Error(`phải về nhánh main, nhận ${currentBranch(root)}`);
		}

		await run(state, 'tm.checkout', undefined, 'nhánh-phụ');
		if (currentBranch(root) !== 'nhánh-phụ') {
			throw new Error(`phải sang nhánh phụ, nhận ${currentBranch(root)}`);
		}

		// Thanh trạng thái phải hiện tên nhánh đang đứng.
		const source = state.sourceControls[0] as { statusBarCommands: { title: string }[] };
		const titles = source.statusBarCommands.map(c => c.title).join(' ');
		if (!titles.includes('nhánh-phụ')) {
			throw new Error(`thanh trạng thái phải hiện nhánh đang đứng, nhận ${JSON.stringify(titles)}`);
		}
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});

/**
 * Lưu tạm thay đổi rồi lấy lại, cây làm việc phải trở về đúng nội dung cũ.
 */
runSmoke('lưu tạm rồi lấy lại: nội dung trên đĩa phải khớp trước khi lưu', async () => {
	const root = seededRepo('tm-stash-');
	try {
		const state = await activateRepo(root);
		const original = readIn(root, 'a.txt');
		await writeAndNotify(state, root, 'a.txt', 'đã sửa\n');
		state.answers.input.push('việc dở dang');

		await run(state, 'tm.stash');

		if (!stashList(root).includes('việc dở dang')) {
			throw new Error(`danh sách lưu tạm phải có mô tả vừa nhập, nhận ${JSON.stringify(stashList(root))}`);
		}
		if (readIn(root, 'a.txt') !== original) {
			throw new Error('sau khi lưu tạm thì cây làm việc phải sạch');
		}

		await run(state, 'tm.stashPopLatest');

		if (readIn(root, 'a.txt') !== 'đã sửa\n') {
			throw new Error(`lấy lại phải trả nội dung đã sửa, nhận ${JSON.stringify(readIn(root, 'a.txt'))}`);
		}
		if (stashList(root) !== '') {
			throw new Error(`lấy lại xong thì danh sách phải trống, nhận ${JSON.stringify(stashList(root))}`);
		}
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});

/**
 * Áp dụng bản lưu tạm thì giữ lại bản đó trong danh sách, khác với lấy lại.
 */
runSmoke('áp dụng bản lưu tạm thì vẫn giữ bản đó trong danh sách', async () => {
	const root = seededRepo('tm-stashapply-');
	try {
		const state = await activateRepo(root);
		await writeAndNotify(state, root, 'a.txt', 'đã sửa\n');
		state.answers.input.push('giữ lại');
		await run(state, 'tm.stash');

		await run(state, 'tm.stashApplyLatest');

		if (readIn(root, 'a.txt') !== 'đã sửa\n') {
			throw new Error('áp dụng phải trả lại nội dung đã sửa');
		}
		if (!stashList(root).includes('giữ lại')) {
			throw new Error(`áp dụng thì bản lưu tạm phải còn lại, nhận ${JSON.stringify(stashList(root))}`);
		}
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});

/**
 * Lưu tạm kèm tệp chưa theo dõi thì tệp đó phải biến mất khỏi đĩa rồi trở lại.
 */
runSmoke('lưu tạm kèm tệp chưa theo dõi thì tệp phải trở lại khi lấy lại', async () => {
	const root = seededRepo('tm-stashu-');
	try {
		const state = await activateRepo(root);
		await writeAndNotify(state, root, 'a.txt', 'đã sửa\n');
		await writeAndNotify(state, root, 'chua-theo-doi.txt', 'tệp mới\n');
		state.answers.input.push('kèm tệp mới');

		await run(state, 'tm.stashIncludeUntracked');

		if (!stashList(root).includes('kèm tệp mới')) {
			throw new Error(`phải lưu tạm được, nhận ${JSON.stringify(stashList(root))}`);
		}
		if (statusOf(root).includes('chua-theo-doi.txt')) {
			throw new Error('tệp chưa theo dõi phải được gỡ khỏi đĩa');
		}

		await run(state, 'tm.stashPopLatest');

		if (!statusOf(root).includes('chua-theo-doi.txt')) {
			throw new Error(`lấy lại thì tệp chưa theo dõi phải trở lại, nhận ${statusOf(root)}`);
		}
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});

/**
 * Không có thay đổi nào thì lệnh lưu tạm phải báo chứ không tạo bản rỗng.
 */
runSmoke('lưu tạm khi không có thay đổi thì phải báo, không tạo bản lưu tạm', async () => {
	const root = seededRepo('tm-stashempty-');
	try {
		const state = await activateRepo(root);
		state.answers.input.push('không có gì');

		await run(state, 'tm.stash');

		if (stashList(root) !== '') {
			throw new Error(`không được tạo bản lưu tạm, nhận ${JSON.stringify(stashList(root))}`);
		}
		if (!state.shown.some(m => m.includes('Không có thay đổi'))) {
			throw new Error(`phải báo là không có thay đổi, nhận ${JSON.stringify(state.shown)}`);
		}
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});

/**
 * Xoá nhánh cần người dùng xác nhận, bấm *Huỷ* thì nhánh phải còn nguyên.
 */
runSmoke('xoá nhánh: bấm huỷ thì nhánh phải còn lại', async () => {
	const root = seededRepo('tm-branchdel-');
	try {
		const state = await activateRepo(root);
		runTm(TM_BIN!, root, ['switch', '-c', 'nhánh-phụ']);
		state.answers.warning.push('Huỷ');

		await run(state, 'tm.deleteBranch', undefined, 'nhánh-phụ');

		const branches = execFileSync(TM_BIN!, ['vcs', '-C', root, 'branch', '-l'], { encoding: 'utf8' });
		if (!branches.includes('nhánh-phụ')) {
			throw new Error(`bấm huỷ thì phải giữ nhánh, nhận ${JSON.stringify(branches)}`);
		}
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});

/**
 * Xoá nhánh khi đồng ý thì nhánh phải biến mất khỏi danh sách.
 */
runSmoke('xoá nhánh: bấm xoá thì nhánh phải biến mất', async () => {
	const root = seededRepo('tm-branchdel2-');
	try {
		const state = await activateRepo(root);
		runTm(TM_BIN!, root, ['switch', '-c', 'nhánh-phụ']);
		runTm(TM_BIN!, root, ['switch', 'main']);
		state.answers.warning.push('Xoá');

		await run(state, 'tm.deleteBranch', undefined, 'nhánh-phụ');

		const branches = execFileSync(TM_BIN!, ['vcs', '-C', root, 'branch', '-l'], { encoding: 'utf8' });
		if (branches.includes('nhánh-phụ')) {
			throw new Error(`nhánh đã xoá phải biến mất, nhận ${JSON.stringify(branches)}`);
		}
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});