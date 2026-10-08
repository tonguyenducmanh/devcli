import * as assert from 'node:assert/strict';
import { execFileSync } from 'node:child_process';
import { chmodSync, existsSync, mkdirSync, mkdtempSync, readdirSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import * as path from 'node:path';
import { test } from 'node:test';

import { resolveExecutable, Tm, TmError } from '../tm';

/** Gốc kho của tm, thư mục cha của thư mục extension. */
const REPO_ROOT = path.join(__dirname, '..', '..', '..', '..');

/** Tìm lệnh tm, giống hệt cách kiểm thử phần phân tích tìm. */
function findTm(): string | undefined {
	const fromEnv = process.env.TM_BIN;
	if (fromEnv && existsSync(fromEnv)) {
		return fromEnv;
	}
	try {
		// Tên tệp do build_all.sh đặt theo cấu hình trong scripts/, nên thử cả
		// ba tiền tố tên tệp build trong scripts/ cho chắc.
		const built = readdirSync(path.join(REPO_ROOT, 'out'))
			.filter(name => /^(td-devcli|devcli|tm)(-|\.)/.test(name))
			.map(name => path.join(REPO_ROOT, 'out', name))
			.find(candidate => existsSync(candidate));
		if (built) {
			return built;
		}
	} catch {
		// Chưa build thì thử PATH.
	}
	try {
		execFileSync('tm', ['version'], { stdio: 'ignore' });
		return 'tm';
	} catch {
		return undefined;
	}
}

const TM_BIN = findTm();
const runTm = TM_BIN ? test : test.skip;

function write(root: string, rel: string, body: string): void {
	const full = path.join(root, rel);
	mkdirSync(path.dirname(full), { recursive: true });
	writeFileSync(full, body, 'utf8');
}

/** Tạo kho tm mới với một commit đầu tiên. */
async function repo(tm: Tm, files: Record<string, string> = { 'a.txt': 'một\nhai\n' }): Promise<string> {
	const root = mkdtempSync(path.join(tmpdir(), 'tm-cli-'));
	await tm.init(root);
	for (const [rel, body] of Object.entries(files)) {
		write(root, rel, body);
	}
	await tm.stageAll(root);
	await tm.commit(root, 'c1');
	return root;
}

runTm('Tm: đọc trạng thái sau khi sửa, thêm và xoá tệp', async () => {
	const tm = new Tm(TM_BIN!);
	const root = await repo(tm, { 'sua.txt': 'ban đầu\n', 'xoa.txt': 'còn\n' });
	try {
		// Sửa rồi stage một tệp.
		write(root, 'sua.txt', 'đã stage\n');
		await tm.stage(root, ['sua.txt']);
		// Sửa trên đĩa một tệp đã stage, thay đổi đó chưa được stage.
		write(root, 'sua.txt', 'đã sửa thêm\n');
		// Xoá một tệp đã có trong kho.
		rmSync(path.join(root, 'xoa.txt'));
		// Tạo tệp mới chưa được theo dõi.
		write(root, 'moi.txt', 'tệp mới\n');

		const status = await tm.status(root);
		assert.equal(status.branch, 'main');
		assert.equal(status.detached, false);
		assert.deepEqual(status.staged, [{ path: 'sua.txt', status: 'M' }]);
		// Tệp đã stage rồi sửa tiếp thì xuất hiện ở cả hai nhóm.
		assert.deepEqual(status.unstaged, [
			{ path: 'sua.txt', status: 'M' },
			{ path: 'xoa.txt', status: 'D' }
		]);
		assert.deepEqual(status.untracked, [{ path: 'moi.txt', status: '?' }]);
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});

runTm('Tm: dựng lại hai phía của tệp để mở khung so sánh', async () => {
	const tm = new Tm(TM_BIN!);
	const root = await repo(tm);
	try {
		write(root, 'a.txt', 'một\nhai sửa\nba\n');

		const patch = (await tm.diff(root, { paths: ['a.txt'] }))[0];
		assert.equal(patch.old, 'một\nhai\n');
		assert.equal(patch.new, 'một\nhai sửa\nba\n');

		// Phía đã stage dựng từ diff HEAD so với vùng chuẩn bị.
		await tm.stage(root, ['a.txt']);
		const staged = (await tm.diff(root, { staged: true, paths: ['a.txt'] }))[0];
		assert.equal(staged.old, 'một\nhai\n');
		assert.equal(staged.new, 'một\nhai sửa\nba\n');
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});

runTm('Tm: commit rồi đọc lại nội dung từ HEAD', async () => {
	const tm = new Tm(TM_BIN!);
	const root = await repo(tm, { 'a.txt': 'ban đầu\n' });
	try {
		write(root, 'a.txt', 'sau\n');
		await tm.stage(root, ['a.txt']);
		await tm.commit(root, 'c2');

		// diff của một commit là so với phụ huynh của nó.
		const head = (await tm.diff(root, { revision: 'HEAD', paths: ['a.txt'] }))[0];
		assert.equal(head.old, 'ban đầu\n');
		assert.equal(head.new, 'sau\n');
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});

runTm('Tm: gỡ khỏi vùng chuẩn bị rồi huỷ sửa đổi trên đĩa', async () => {
	const tm = new Tm(TM_BIN!);
	const root = await repo(tm, { 'a.txt': 'ban đầu\n' });
	try {
		write(root, 'a.txt', 'sửa trên đĩa\n');
		await tm.stage(root, ['a.txt']);
		let status = await tm.status(root);
		assert.deepEqual(status.staged.map(e => e.path), ['a.txt']);

		await tm.unstage(root, ['a.txt']);
		status = await tm.status(root);
		assert.deepEqual(status.staged, []);
		assert.deepEqual(status.unstaged.map(e => e.path), ['a.txt']);

		await tm.discard(root, ['a.txt']);
		status = await tm.status(root);
		assert.deepEqual(status.unstaged, [], 'huỷ sửa đổi thì cây làm việc sạch');
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});

runTm('Tm: xoá tệp chưa theo dõi, giữ nguyên tệp đã theo dõi', async () => {
	const tm = new Tm(TM_BIN!);
	const root = await repo(tm, { 'theo-doi.txt': 'còn trong kho\n' });
	try {
		write(root, 'rac.txt', 'tạm\n');
		write(root, 'thu-muc/cong.txt', 'tạm\n');

		await tm.clean(root, ['rac.txt']);

		assert.ok(!existsSync(path.join(root, 'rac.txt')), 'rac.txt phải bị xoá');
		assert.ok(existsSync(path.join(root, 'thu-muc/cong.txt')), 'tệp ngoài danh sách phải còn lại');
		assert.ok(existsSync(path.join(root, 'theo-doi.txt')), 'tệp đã theo dõi không được xoá');

		const status = await tm.status(root);
		assert.deepEqual(status.untracked.map(e => e.path), ['thu-muc/cong.txt']);
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});

test('resolveExecutable: ưu tiên đường dẫn người dùng đặt trong cấu hình', () => {
	assert.equal(resolveExecutable('/opt/my-tm/tm'), '/opt/my-tm/tm');
});

test('resolveExecutable: không có cấu hình thì dùng lệnh tm trong PATH', () => {
	assert.equal(resolveExecutable(''), 'tm');
});

runTm('Tm: nhánh, commit, tag và bản lưu tạm', async () => {
	const tm = new Tm(TM_BIN!);
	const root = await repo(tm);
	try {
		await tm.createBranch(root, 'nhanh-phu');
		let branches = await tm.branches(root);
		// Tạo nhánh cũng chuyển sang nhánh đó, và nhánh đang đứng đứng đầu.
		assert.deepEqual(branches.map(b => b.name), ['nhanh-phu', 'main']);
		assert.equal(branches[0].current, true);

		await tm.switchTo(root, 'main');
		branches = await tm.branches(root);
		assert.equal(branches[0].name, 'main');
		write(root, 'việc.txt', 'dở dang\n');
		await tm.stage(root, ['việc.txt']);
		await tm.stash(root, 'lưu lại');
		const stashes = await tm.stashes(root);
		assert.equal(stashes.length, 1);
		assert.equal(stashes[0].message, 'lưu lại');

		// Áp dụng lại bản lưu tạm thì tệp trở lại vùng chuẩn bị như lúc lưu.
		await tm.applyStash(root, 0, false);
		const status = await tm.status(root);
		assert.ok(status.staged.some(e => e.path === 'việc.txt'));

		await tm.createTag(root, 'v1.0', 'ghi chú');
		const tags = await tm.tags(root);
		assert.equal(tags[0].name, 'v1.0');
		assert.equal(tags[0].message, 'ghi chú');

		const entries = await tm.log(root, 10);
		assert.equal(entries[0].summary, 'c1');

		await tm.deleteBranch(root, 'nhanh-phu');
		branches = await tm.branches(root);
		assert.deepEqual(branches.map(b => b.name), ['main']);
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});

runTm('Tm: báo lỗi khi lệnh thất bại, giữ nguyên thông điệp của tm', async () => {
	const tm = new Tm(TM_BIN!);
	const root = await repo(tm);
	try {
		await assert.rejects(
			() => tm.merge(root, 'khong-ton-tai'),
			(error: unknown) => {
				assert.ok(error instanceof TmError);
				assert.match(error.message, /không phân giải được tham chiếu/);
				return true;
			}
		);
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});

runTm('Tm: commit thiếu nội dung hoặc không có gì để commit thì tm từ chối', async () => {
	const tm = new Tm(TM_BIN!);
	const root = await repo(tm);
	try {
		// Có thay đổi đã stage nhưng nội dung commit rỗng.
		write(root, 'a.txt', 'sửa\n');
		await tm.stage(root, ['a.txt']);
		await assert.rejects(() => tm.commit(root, ''), /thiếu nội dung commit/);

		// Gỡ khỏi vùng chuẩn bị rồi huỷ sửa đổi thì kho sạch, lúc đó tm báo lỗi
		// khác: không có gì để commit.
		await tm.unstage(root, ['a.txt']);
		await tm.discard(root, ['a.txt']);
		await assert.rejects(() => tm.commit(root, 'không có gì'), /không có gì để commit/);
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});

runTm('Tm: show-file đọc nội dung ở một điểm lịch sử', async () => {
	const tm = new Tm(TM_BIN!);
	const root = await repo(tm, { 'a.txt': 'một\nhai\n', 'khac.txt': 'giữ nguyên\n' });
	try {
		write(root, 'a.txt', 'một\nhai sửa\n');
		await tm.stage(root, ['a.txt']);
		await tm.commit(root, 'c2');

		const head = await tm.showFile(root, 'HEAD', 'a.txt');
		assert.equal(head.found, true);
		assert.equal(head.unsupported, false);
		assert.equal(head.content, 'một\nhai sửa\n');

		const old = await tm.showFile(root, 'HEAD~1', 'a.txt');
		assert.equal(old.content, 'một\nhai\n');

		// Tệp mà commit cuối không đụng tới vẫn đọc được.
		const untouched = await tm.showFile(root, 'HEAD', 'khac.txt');
		assert.equal(untouched.content, 'giữ nguyên\n');

		// Tệp chưa có ở điểm đó là câu trả lời hợp lệ chứ không phải lỗi.
		const missing = await tm.showFile(root, 'HEAD', 'khong-co.txt');
		assert.equal(missing.found, false);
		assert.equal(missing.unsupported, false);
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});

// Lớp vỏ báo như tm cũ: chạy được mọi lệnh trừ show-file.
const runOldTd = process.platform === 'win32' ? test.skip : runTm;
runOldTd('Tm: show-file báo lệnh còn thiếu khi tm trên máy cũ hơn', async () => {
	const dir = mkdtempSync(path.join(tmpdir(), 'tm-old-'));
	const real = new Tm(TM_BIN!);
	const root = await repo(real);
	const wrapper = path.join(dir, 'tm');
	writeFileSync(wrapper, [
		'#!/bin/sh',
		'for a in "$@"; do',
		'\tif [ "$a" = "show-file" ]; then',
		'\t\techo \'không có lệnh nào tên "show-file", xem danh sách: tm vcs --help\' >&2',
		'\t\texit 1',
		'\tfi',
		'done',
		`exec ${JSON.stringify(TM_BIN!)} "$@"`,
		''
	].join('\n'), 'utf8');
	chmodSync(wrapper, 0o755);
	const old = new Tm(wrapper);
	try {
		const result = await old.showFile(root, 'HEAD', 'a.txt');
		assert.equal(result.unsupported, true);
		assert.equal(result.found, false);
		// Lệnh khác vẫn chạy được qua lớp vỏ.
		const status = await old.status(root);
		assert.equal(status.branch, 'main');
	} finally {
		rmSync(root, { recursive: true, force: true });
		rmSync(dir, { recursive: true, force: true });
	}
});

runTm('Tm: changedFiles chỉ liệt kê tên tệp, không lấy nội dung', async () => {
	const tm = new Tm(TM_BIN!);
	const root = await repo(tm, { 'a.txt': 'một\n', 'b.txt': 'hai\n' });
	try {
		write(root, 'a.txt', 'một sửa\n');
		write(root, 'moi.txt', 'mới\n');

		assert.deepEqual((await tm.changedFiles(root, { revision: 'HEAD' })).sort(), ['a.txt', 'b.txt']);

		// Lọc theo danh sách tệp thì chỉ còn tệp được nêu.
		assert.deepEqual(await tm.changedFiles(root, { revision: 'HEAD', paths: ['a.txt'] }), ['a.txt']);

		// Không lọc thì lấy cả tệp đã sửa lẫn tệp chưa theo dõi trên đĩa.
		const unstaged = (await tm.changedFiles(root)).sort();
		assert.deepEqual(unstaged, ['a.txt', 'moi.txt']);
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});
