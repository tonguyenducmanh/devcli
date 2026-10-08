import { execFileSync } from 'node:child_process';
import { mkdtempSync, rmSync } from 'node:fs';
import { tmpdir } from 'node:os';
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
const runHistory = TM_BIN ? test : test.skip;

/**
 * Một mục của hộp chọn nhanh mà tiện ích dựng, đọc để kiểm thử soi.
 */
interface Picked {
	label: string;
	description?: string;
	detail?: string;
	value: { hash: string; summary: string; refs: string[]; author?: string; when?: string };
}

/** Kho có hai commit để lịch sử có nhiều mục, rồi kích hoạt tiện ích trên đó. */
async function activateWithHistory(): Promise<{ root: string; state: FakeState }> {
	const root = makeRepo(TM_BIN!, 'tm-lich-su-');
	writeRepoFile(root, 'cmd/a.go', 'package main\n');
	writeRepoFile(root, 'docs/b.md', 'bản đầu\n');
	runTm(TM_BIN!, root, ['add', '.']);
	runTm(TM_BIN!, root, ['commit', '-m', 'c1 thêm tệp']);

	// Sửa tệp rồi commit lần nữa để lịch sử có hơn một mục.
	writeRepoFile(root, 'docs/b.md', 'sửa lần hai\n');
	runTm(TM_BIN!, root, ['add', '.']);
	runTm(TM_BIN!, root, ['commit', '-m', 'c2 sửa b.md']);

	// Tệp riêng không bị commit c2 đụng tới.
	writeRepoFile(root, 'cmd/a.go', 'package main // sửa\n');
	runTm(TM_BIN!, root, ['add', '.']);
	runTm(TM_BIN!, root, ['commit', '-m', 'c3 sửa a.go']);

	const { vscode, state } = fakeVscode(TM_BIN!);
	const workspace = (vscode.workspace as { workspaceFolders: unknown });
	workspace.workspaceFolders = [{ uri: state.uri(root), name: 'kho', index: 0 }];
	loadExtension(vscode).activate({ subscriptions: [] as { dispose(): void }[] });
	await state.settle();
	return { root, state };
}

/** Danh sách mục mà hộp chọn nhanh nhận lần gọi gần nhất. */
function lastPick(state: FakeState): Picked[] {
	return (state.picks.at(-1) ?? []) as Picked[];
}

/**
 * Lịch sử của một tệp phải chỉ ra đúng các commit từng sửa tệp đó.
 *
 * Đây là phần cốt lõi của tính năng: lọc nhầm thì danh sách lẫn commit của tệp
 * khác và người dùng bấm nhầm rồi thấy khung so sánh rỗng.
 */
runHistory('lịch sử tệp chỉ liệt kê commit sửa tệp đó', async () => {
	const { root, state } = await activateWithHistory();
	try {
		state.answers.pick.push(undefined);
		await state.registeredCommands.get('tm.fileHistory')?.(undefined, state.uri(path.join(root, 'docs/b.md')));

		const items = lastPick(state);
		if (items.length !== 2) {
			throw new Error(`b.md chỉ bị hai commit sửa, nhận ${items.length}: ${JSON.stringify(items.map(i => i.label))}`);
		}
		// Mới nhất trước, đúng như lịch sử của git.
		if (items[0].label !== 'c2 sửa b.md' || items[1].label !== 'c1 thêm tệp') {
			throw new Error(`thứ tự phải mới trước cũ, nhận ${JSON.stringify(items.map(i => i.label))}`);
		}
		// Commit sửa cmd/a.go không được lẫn vào.
		if (items.some(item => item.label.includes('a.go'))) {
			throw new Error(`lịch sử của b.md không được chứa commit của a.go: ${JSON.stringify(items.map(i => i.label))}`);
		}
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});

/**
 * Mỗi mục của lịch sử phải có ngày và tác giả.
 *
 * Danh sách chỉ có mã băm thì người dùng không biết mình đang ở đâu khi hai
 * commit gần đây giống nhau.
 */
runHistory('mỗi mục lịch sử có ngày, tác giả và mã băm', async () => {
	const { root, state } = await activateWithHistory();
	try {
		state.answers.pick.push(undefined);
		await state.registeredCommands.get('tm.fileHistory')?.(undefined, state.uri(path.join(root, 'docs/b.md')));

		for (const item of lastPick(state)) {
			if (!/^[0-9a-f]{8,}$/.test(item.description ?? '')) {
				throw new Error(`mỗi mục phải có mã băm, nhận ${JSON.stringify(item)}`);
			}
			// Dòng phụ gồm thời điểm rồi tới tên tác giả, phân cách bằng dấu chấm.
			if (!/^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2} · \S+/.test(item.detail ?? '')) {
				throw new Error(`mỗi mục phải có thời điểm và tác giả, nhận ${JSON.stringify(item)}`);
			}
		}
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});

/**
 * Chọn một commit thì mở khung so sánh đúng thay đổi của tệp tại commit đó.
 *
 * Phải là hai phía ảo của tiện ích: phía gốc là nội dung trước commit, phía phải
 * là nội dung sau commit. Mở tệp thật thì người dùng thấy trạng thái hiện tại chứ
 * không phải lịch sử.
 */
runHistory('chọn commit thì mở khung so sánh đúng thay đổi của tệp tại commit đó', async () => {
	const { root, state } = await activateWithHistory();
	try {
		// Biết trước hai mã băm từ lịch sử để đối chiếu.
		const all = runTmText(root, ['log', '--oneline']);
		const c2 = all.find(l => l.includes('c2 sửa b.md'))!.split(' ')[0];

		state.answers.pick.push({
			label: 'c2 sửa b.md',
			description: c2,
			value: { hash: c2, summary: 'c2 sửa b.md', refs: ['main'] }
		});
		await state.registeredCommands.get('tm.fileHistory')?.(undefined, state.uri(path.join(root, 'docs/b.md')));

		const call = state.lastCall('vscode.diff');
		if (!call) {
			throw new Error('phải mở khung so sánh');
		}
		const [original, modified, title] = call.args as [{ scheme: string; query: string }, { scheme: string; query: string }, string];
		if (original.scheme !== 'tm' || modified.scheme !== 'tm') {
			throw new Error(`cả hai phía phải là tài liệu ảo của tiện ích, nhận ${original.scheme} và ${modified.scheme}`);
		}
		if (!original.query.includes(`ref=${c2}`) || !modified.query.includes(`ref=${c2}`)) {
			throw new Error(`cả hai phía phải trỏ về commit đã chọn, nhận ${original.query}`);
		}
		if (!title.includes('docs/b.md')) {
			throw new Error(`tiêu đề phải nêu tên tệp, nhận ${title}`);
		}

		// Nội dung hai phía phải là trước và sau commit đó, không phải bản hiện tại.
		const before = await state.contentOf(original);
		const after = await state.contentOf(modified);
		if (before !== 'bản đầu\n' || after !== 'sửa lần hai\n') {
			throw new Error(`hai phía phải là nội dung trước và sau commit, nhận ${JSON.stringify({ before, after })}`);
		}
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});

/**
 * Chọn commit đầu tiên thì phía gốc phải rỗng, vì tệp chưa tồn tại trước đó.
 */
runHistory('chọn commit đầu tiên thì phía gốc là tài liệu rỗng', async () => {
	const { root, state } = await activateWithHistory();
	try {
		const all = runTmText(root, ['log', '--oneline']);
		const c1 = all.find(l => l.includes('c1 thêm tệp'))!.split(' ')[0];
		state.answers.pick.push({
			label: 'c1 thêm tệp',
			description: c1,
			value: { hash: c1, summary: 'c1 thêm tệp', refs: ['main'] }
		});
		await state.registeredCommands.get('tm.fileHistory')?.(undefined, state.uri(path.join(root, 'docs/b.md')));

		const [original, modified] = state.lastCall('vscode.diff')?.args as [unknown, unknown];
		if (await state.contentOf(original) !== '') {
			throw new Error(`trước commit đầu tiên thì tệp chưa có, phía gốc phải rỗng, nhận ${JSON.stringify(await state.contentOf(original))}`);
		}
		if (await state.contentOf(modified) !== 'bản đầu\n') {
			throw new Error(`phía phải phải là nội dung sau commit đầu tiên, nhận ${JSON.stringify(await state.contentOf(modified))}`);
		}
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});

/**
 * Lịch sử của một thư mục là lịch sử của mọi tệp bên trong nó.
 */
runHistory('lịch sử của thư mục gộp các tệp bên trong', async () => {
	const { root, state } = await activateWithHistory();
	try {
		state.answers.pick.push(undefined);
		await state.registeredCommands.get('tm.fileHistory')?.(undefined, state.uri(path.join(root, 'cmd')));

		const items = lastPick(state);
		// cmd/a.go bị c1 thêm vào và c3 sửa, còn c2 không đụng tới nên không có mặt.
		const labels = items.map(item => item.label);
		if (labels.length !== 2 || labels[0] !== 'c3 sửa a.go' || labels[1] !== 'c1 thêm tệp') {
			throw new Error(`lịch sử của thư mục cmd phải gồm đúng hai commit sửa a.go, nhận ${JSON.stringify(labels)}`);
		}
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});

/**
 * Tệp chưa từng được commit thì phải báo rõ chứ không mở hộp chọn trống.
 */
runHistory('tệp chưa được commit thì báo chứ không mở danh sách trống', async () => {
	const { root, state } = await activateWithHistory();
	try {
		writeRepoFile(root, 'moi.txt', 'tệp mới\n');
		state.fireFileEvent('create', path.join(root, 'moi.txt'));
		await state.settle();

		await state.registeredCommands.get('tm.fileHistory')?.(undefined, state.uri(path.join(root, 'moi.txt')));

		if (state.picks.length !== 0) {
			throw new Error('tệp chưa có lịch sử thì không được mở hộp chọn');
		}
		if (!state.shown.some(m => m.includes('moi.txt') && m.includes('chưa'))) {
			throw new Error(`phải báo rõ là tệp chưa có lịch sử, nhận ${JSON.stringify(state.shown)}`);
		}
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});

/**
 * Tệp ngoài kho thì phải báo, không được im lặng.
 */
runHistory('tệp ngoài kho tm thì phải báo', async () => {
	const { state } = await activateWithHistory();
	const outside = mkdtempSync(path.join(tmpdir(), 'tm-ngoai-kho-'));
	try {
		await state.registeredCommands.get('tm.fileHistory')?.(undefined, state.uri(path.join(outside, 'khong-o-kho.txt')));
		if (!state.shown.some(m => m.includes('không nằm trong kho'))) {
			throw new Error(`tệp ngoài kho phải báo rõ, nhận ${JSON.stringify(state.shown)}`);
		}
	} finally {
		rmSync(outside, { recursive: true, force: true });
	}
});

/**
 * Bấm Esc ở hộp chọn thì không mở gì cả.
 */
runHistory('huỷ hộp chọn thì không mở khung so sánh', async () => {
	const { root, state } = await activateWithHistory();
	try {
		state.answers.pick.push(undefined);
		await state.registeredCommands.get('tm.fileHistory')?.(undefined, state.uri(path.join(root, 'docs/b.md')));

		if (state.lastCall('vscode.diff')) {
			throw new Error('huỷ hộp chọn thì không được mở khung so sánh');
		}
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});

/** Chạy lệnh tm trong kho và trả về từng dòng. */
function runTmText(root: string, args: string[]): string[] {
	return execFileSync(TM_BIN!, ['vcs', '-C', root, ...args], { encoding: 'utf8' })
		.split('\n')
		.filter(line => line.trim().length > 0);
}