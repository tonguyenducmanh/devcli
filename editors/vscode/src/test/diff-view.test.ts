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
const runDiff = TM_BIN ? test : test.skip;

/** Kho có sẵn một commit, dùng làm điểm xuất phát cho các kiểm thử. */
function seededRepo(prefix: string, body: string): string {
	const root = makeRepo(TM_BIN!, prefix);
	writeRepoFile(root, 'a.txt', body);
	runTm(TM_BIN!, root, ['add', '.']);
	runTm(TM_BIN!, root, ['commit', '-m', 'c1']);
	return root;
}

/** Kích hoạt tiện ích trên kho và chờ trạng thái được đọc xong. */
async function activate(root: string): Promise<FakeState> {
	const { vscode, state } = fakeVscode(TM_BIN!);
	const workspace = (vscode.workspace as { workspaceFolders: unknown });
	workspace.workspaceFolders = [{ uri: state.uri(root), name: 'kho', index: 0 }];
	loadExtension(vscode).activate({ subscriptions: [] as { dispose(): void }[] });
	await state.settle();
	return state;
}

/** Ghi tệp rồi báo cho tiện ích biết, đúng như lúc người dùng lưu tệp. */
async function save(state: FakeState, root: string, relative: string, content: string): Promise<void> {
	writeRepoFile(root, relative, content);
	state.fireFileEvent('change', path.join(root, relative));
	await state.settle();
}

/** Hai phía của khung so sánh mà lệnh mở thay đổi vừa yêu cầu. */
async function sidesOfChange(state: FakeState, root: string, relative: string): Promise<{ original: unknown; modified: unknown }> {
	await state.registeredCommands.get('tm.openChange')?.(undefined, state.uri(path.join(root, relative)));
	const call = state.lastCall('vscode.diff');
	if (!call) {
		throw new Error('lệnh mở thay đổi phải ra lệnh vscode.diff');
	}
	const [original, modified] = call.args as [unknown, unknown];
	return { original, modified };
}

/** Nội dung khung Source Control đang hiện, gộp cả bốn nhóm. */
function listedPaths(state: FakeState): string[] {
	return state.resourceGroups.flatMap(group =>
		(group.resourceStates as { resourceUri: { fsPath: string } }[]).map(item => item.resourceUri.fsPath));
}

/**
 * Xoá một dòng, xem khung so sánh, lưu rồi xoá thêm một dòng.
 *
 * Đây là thao tác lặp lại thường xuyên nhất, và mỗi lần khung so sánh phải phản
 * ánh đúng trạng thái hiện tại chứ không phải trạng thái của lần mở trước.
 */
runDiff('xoá dòng, xem khung so sánh, lưu rồi xoá thêm dòng nữa', async () => {
	const root = seededRepo('tm-xoa-dong-', 'dòng 1\ndòng 2\ndòng 3\ndòng 4\n');
	try {
		const state = await activate(root);

		// Xoá dòng 2 rồi xem khung so sánh.
		await save(state, root, 'a.txt', 'dòng 1\ndòng 3\ndòng 4\n');
		let sides = await sidesOfChange(state, root, 'a.txt');
		if (await state.contentOf(sides.original) !== 'dòng 1\ndòng 2\ndòng 3\ndòng 4\n') {
			throw new Error('phía gốc phải là nội dung đã stage');
		}

		// Lưu rồi xoá thêm một dòng nữa: khung so sánh phải theo kịp.
		await save(state, root, 'a.txt', 'dòng 1\ndòng 4\n');
		sides = await sidesOfChange(state, root, 'a.txt');
		if (await state.contentOf(sides.original) !== 'dòng 1\ndòng 2\ndòng 3\ndòng 4\n') {
			throw new Error('phía gốc vẫn phải là nội dung đã stage, tức là còn dòng 2 và 3');
		}

		// Phía phải là chính tệp trên đĩa, nên phải bám theo lần sửa mới nhất.
		if ((sides.modified as { scheme: string }).scheme !== 'file') {
			throw new Error('phía phải phải là tệp thật trên đĩa để sửa thẳng được');
		}
		const onDisk = readFileSync(path.join(root, 'a.txt'), 'utf8');
		if (onDisk !== 'dòng 1\ndòng 4\n') {
			throw new Error(`tệp trên đĩa phải là nội dung mới, nhận ${JSON.stringify(onDisk)}`);
		}
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});

/**
 * Hoàn tác hết thay đổi thì phía gốc không được bỏ trống.
 *
 * Người dùng bấm hoàn tác từng dòng rồi bấm hoàn tác cả tệp trên khung so sánh,
 * làm cho tệp khớp lại vùng chuẩn bị. Lúc đó không còn khác biệt nào để dựng
 * nội dung vùng chuẩn bị, và nếu vẫn dựng từ khác biệt thì phía gốc rơi về
 * chuỗi rỗng trong khi phía phải vẫn có nội dung. Khung so sánh lúc đó báo thay
 * đổi từ tệp trống sang tệp mới có nội dung, tức là báo sai.
 */
runDiff('hoàn tác hết thay đổi thì phía gốc vẫn là nội dung vùng chuẩn bị', async () => {
	const root = seededRepo('tm-hoan-tac-', 'dòng 1\ndòng 2\ndòng 3\n');
	try {
		const state = await activate(root);
		await save(state, root, 'a.txt', 'dòng 1\ndòng 3\n');

		const { original } = await sidesOfChange(state, root, 'a.txt');
		if (await state.contentOf(original) !== 'dòng 1\ndòng 2\ndòng 3\n') {
			throw new Error('trước khi hoàn tác thì phía gốc phải là nội dung đã stage');
		}

		// Hoàn tác cả tệp: nội dung trên đĩa khớp lại vùng chuẩn bị.
		runTm(TM_BIN!, root, ['restore', 'a.txt']);
		state.fireFileEvent('change', path.join(root, 'a.txt'));
		await state.settle();

		const after = await state.contentOf(original);
		if (after !== 'dòng 1\ndòng 2\ndòng 3\n') {
			throw new Error(`sau khi hoàn tác hết, phía gốc phải vẫn là nội dung đã stage, nhận ${JSON.stringify(after)}`);
		}

		// Và tệp đã khớp thì không còn là thay đổi nữa, không được ghi nhận.
		if (listedPaths(state).includes(path.join(root, 'a.txt'))) {
			throw new Error(`tệp đã hoàn tác hết thì không được còn trong khung Source Control, nhận ${JSON.stringify(listedPaths(state))}`);
		}
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});

/**
 * Sửa, hoàn tác, rồi sửa lại: mỗi vòng đều phải cho ra phía gốc đúng.
 *
 * Lặp lại nhiều vòng để bắt được chỗ nào đó giữ lại nội dung của vòng trước, vì
 * bản giảo nhớ khác biệt thì vòng sau vẫn ra kết quả đúng một cách tình cờ.
 */
runDiff('sửa, hoàn tác rồi sửa lại qua nhiều vòng', async () => {
	const root = seededRepo('tm-nhieu-vong-', 'a\nb\nc\n');
	try {
		const state = await activate(root);
		const staged = 'a\nb\nc\n';

		for (const body of ['a\nc\n', 'a\n', 'x\ny\nz\n']) {
			await save(state, root, 'a.txt', body);
			const { original } = await sidesOfChange(state, root, 'a.txt');
			if (await state.contentOf(original) !== staged) {
				throw new Error(`với nội dung ${JSON.stringify(body)}, phía gốc phải là nội dung đã stage`);
			}
			runTm(TM_BIN!, root, ['restore', 'a.txt']);
			state.fireFileEvent('change', path.join(root, 'a.txt'));
			await state.settle();
			if (await state.contentOf(original) !== staged) {
				throw new Error(`sau khi hoàn tác với ${JSON.stringify(body)}, phía gốc phải là nội dung đã stage`);
			}
		}
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});

/**
 * Đã stage hết thì so HEAD với vùng chuẩn bị, không phải với tệp trên đĩa.
 */
runDiff('thay đổi đã stage thì so HEAD với vùng chuẩn bị', async () => {
	const root = seededRepo('tm-da-stage-', 'bản đầu\n');
	try {
		const state = await activate(root);
		await save(state, root, 'a.txt', 'đã stage\n');
		runTm(TM_BIN!, root, ['add', 'a.txt']);
		state.fireFileEvent('change', path.join(root, '.tmx/index'));
		await state.settle();

		const { original, modified } = await sidesOfChange(state, root, 'a.txt');
		if (await state.contentOf(original) !== 'bản đầu\n') {
			throw new Error('phía gốc phải là nội dung trong HEAD');
		}
		if (await state.contentOf(modified) !== 'đã stage\n') {
			throw new Error('phía phải phải là nội dung đã stage');
		}
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});

/**
 * Gỡ khỏi vùng chuẩn bị rồi sửa tiếp trên đĩa thì phải quay về so với HEAD.
 */
runDiff('gỡ khỏi vùng chuẩn bị xong sửa tiếp trên đĩa thì so với HEAD', async () => {
	const root = seededRepo('tm-unstage-', 'bản đầu\n');
	try {
		const state = await activate(root);
		await save(state, root, 'a.txt', 'đã stage\n');
		runTm(TM_BIN!, root, ['add', 'a.txt']);
		// Gỡ khỏi vùng chuẩn bị rồi sửa tiếp trên đĩa.
		runTm(TM_BIN!, root, ['restore', '--staged', 'a.txt']);
		await save(state, root, 'a.txt', 'đã stage\nrồi sửa thêm\n');

		const { original } = await sidesOfChange(state, root, 'a.txt');
		if (await state.contentOf(original) !== 'bản đầu\n') {
			throw new Error('tệp không còn gì trong vùng chuẩn bị thì phía gốc phải là nội dung đã stage, tức là rỗng');
		}
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});

/**
 * Tệp chưa theo dõi thì phía gốc rỗng, phía phải là chính tệp trên đĩa.
 */
runDiff('tệp chưa theo dõi thì phía gốc rỗng', async () => {
	const root = seededRepo('tm-chua-theo-doi-', 'x\n');
	try {
		const state = await activate(root);
		await save(state, root, 'moi.txt', 'tệp mới\n');

		const { original, modified } = await sidesOfChange(state, root, 'moi.txt');
		if (await state.contentOf(original) !== '') {
			throw new Error('phía gốc của tệp chưa theo dõi phải rỗng');
		}
		if ((modified as { scheme: string }).scheme !== 'file') {
			throw new Error('phía phải phải là tệp thật trên đĩa');
		}
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});

/**
 * Tệp bị xoá khỏi đĩa thì phía gốc là nội dung đã stage, phía phải rỗng.
 */
runDiff('tệp bị xoá khỏi đĩa thì phía phải rỗng chứ không phải nội dung đã stage', async () => {
	const root = seededRepo('tm-xoa-tap-', 'còn trong kho\n');
	try {
		const state = await activate(root);
		rmSync(path.join(root, 'a.txt'));
		state.fireFileEvent('delete', path.join(root, 'a.txt'));
		await state.settle();

		const { original, modified } = await sidesOfChange(state, root, 'a.txt');
		if (await state.contentOf(original) !== 'còn trong kho\n') {
			throw new Error('phía gốc của tệp bị xoá phải là nội dung đã stage');
		}
		if (await state.contentOf(modified) !== '') {
			throw new Error('phía phải của tệp bị xoá phải rỗng');
		}
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});

/**
 * Gỡ tệp đã xoá khỏi đĩa ra khỏi vùng chuẩn bị rồi xem khung so sánh.
 *
 * `tm vcs restore --staged` đưa vùng chuẩn bị về đúng HEAD, nên tệp đã xoá trên
 * đĩa vẫn phải còn trong vùng chuẩn bị và là xoá chưa stage. Khung so sánh lúc
 * đó là vùng chuẩn bị với tệp trên đĩa, nên phía gốc có nội dung và phía phải
 * rỗng, chứ không phải so với HEAD.
 */
runDiff('gỡ tệp đã xoá khỏi vùng chuẩn bị thì phía gốc vẫn có nội dung', async () => {
	const root = seededRepo('tm-unstage-xoa-', 'còn trong kho\n');
	try {
		const state = await activate(root);
		rmSync(path.join(root, 'a.txt'));
		runTm(TM_BIN!, root, ['restore', '--staged', 'a.txt']);
		state.fireFileEvent('delete', path.join(root, 'a.txt'));
		state.fireFileEvent('change', path.join(root, '.tmx/index'));
		await state.settle();

		const { original, modified } = await sidesOfChange(state, root, 'a.txt');
		if (await state.contentOf(original) !== 'còn trong kho\n') {
			throw new Error('phía gốc phải là nội dung vẫn còn trong vùng chuẩn bị');
		}
		if (await state.contentOf(modified) !== '') {
			throw new Error('phía phải phải rỗng vì tệp đã bị xoá khỏi đĩa');
		}
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});

/**
 * Tên tệp có dấu và khoảng trắng phải đọc được ở cả hai phía.
 */
runDiff('tệp tên có dấu và khoảng trắng thì đọc được cả hai phía', async () => {
	const root = makeRepo(TM_BIN!, 'tm-ten-co-dau-');
	writeRepoFile(root, 'thu muc/tên có dấu.txt', 'gốc\n');
	runTm(TM_BIN!, root, ['add', '.']);
	runTm(TM_BIN!, root, ['commit', '-m', 'c1']);
	try {
		const state = await activate(root);
		const relative = 'thu muc/tên có dấu.txt';
		await save(state, root, relative, 'đã sửa\n');

		const { original, modified } = await sidesOfChange(state, root, relative);
		if (await state.contentOf(original) !== 'gốc\n') {
			throw new Error('phía gốc phải đúng nội dung đã stage');
		}
		if ((modified as { fsPath: string }).fsPath !== path.join(root, relative)) {
			throw new Error('phía phải phải trỏ đúng tệp trên đĩa');
		}
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});

/**
 * Mở cùng một tệp nhiều lần phải cho ra cùng một kết quả.
 *
 * Khung so sánh của VS Code gọi lại phần cấp nội dung mỗi khi tệp trên đĩa đổi,
 * nên kết quả phải ổn định chứ không được phụ thuộc vào thứ tự các lần gọi.
 */
runDiff('mở lại khung so sánh nhiều lần thì kết quả không đổi', async () => {
	const root = seededRepo('tm-mo-lai-', 'một\nhai\n');
	try {
		const state = await activate(root);
		await save(state, root, 'a.txt', 'một\nhai\nba\n');

		const { original } = await sidesOfChange(state, root, 'a.txt');
		for (let i = 0; i < 3; i++) {
			const got = await state.contentOf(original);
			if (got !== 'một\nhai\n') {
				throw new Error(`lần mở thứ ${i + 1} phải cho cùng nội dung đã stage, nhận ${JSON.stringify(got)}`);
			}
		}
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});