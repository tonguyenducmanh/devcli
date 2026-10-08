import * as assert from 'node:assert/strict';
import { execFileSync } from 'node:child_process';
import { chmodSync, existsSync, mkdirSync, mkdtempSync, readdirSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import * as path from 'node:path';
import { test } from 'node:test';

import { resolveExecutable, Td, TdError } from '../td';

/** Gốc kho của td, thư mục cha của thư mục extension. */
const REPO_ROOT = path.join(__dirname, '..', '..', '..', '..');

/** Tìm lệnh td, giống hệt cách kiểm thử phần phân tích tìm. */
function findTd(): string | undefined {
	const fromEnv = process.env.TD_BIN;
	if (fromEnv && existsSync(fromEnv)) {
		return fromEnv;
	}
	try {
		// Tên tệp do build_all.sh đặt theo cấu hình trong scripts/, nên thử cả
		// hai tiền tố td lẫn devcli cho chắc.
		const built = readdirSync(path.join(REPO_ROOT, 'out'))
			.filter(name => /^(td|devcli)(-|\.)/.test(name))
			.map(name => path.join(REPO_ROOT, 'out', name))
			.find(candidate => existsSync(candidate));
		if (built) {
			return built;
		}
	} catch {
		// Chưa build thì thử PATH.
	}
	try {
		execFileSync('td', ['version'], { stdio: 'ignore' });
		return 'td';
	} catch {
		return undefined;
	}
}

const TD_BIN = findTd();
const runTd = TD_BIN ? test : test.skip;

function write(root: string, rel: string, body: string): void {
	const full = path.join(root, rel);
	mkdirSync(path.dirname(full), { recursive: true });
	writeFileSync(full, body, 'utf8');
}

/** Tạo kho td mới với một commit đầu tiên. */
async function repo(td: Td, files: Record<string, string> = { 'a.txt': 'một\nhai\n' }): Promise<string> {
	const root = mkdtempSync(path.join(tmpdir(), 'td-cli-'));
	await td.init(root);
	for (const [rel, body] of Object.entries(files)) {
		write(root, rel, body);
	}
	await td.stageAll(root);
	await td.commit(root, 'c1');
	return root;
}

runTd('Td: đọc trạng thái sau khi sửa, thêm và xoá tệp', async () => {
	const td = new Td(TD_BIN!);
	const root = await repo(td, { 'sua.txt': 'ban đầu\n', 'xoa.txt': 'còn\n' });
	try {
		// Sửa rồi stage một tệp.
		write(root, 'sua.txt', 'đã stage\n');
		await td.stage(root, ['sua.txt']);
		// Sửa trên đĩa một tệp đã stage, thay đổi đó chưa được stage.
		write(root, 'sua.txt', 'đã sửa thêm\n');
		// Xoá một tệp đã có trong kho.
		rmSync(path.join(root, 'xoa.txt'));
		// Tạo tệp mới chưa được theo dõi.
		write(root, 'moi.txt', 'tệp mới\n');

		const status = await td.status(root);
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

runTd('Td: dựng lại hai phía của tệp để mở khung so sánh', async () => {
	const td = new Td(TD_BIN!);
	const root = await repo(td);
	try {
		write(root, 'a.txt', 'một\nhai sửa\nba\n');

		const patch = (await td.diff(root, { paths: ['a.txt'] }))[0];
		assert.equal(patch.old, 'một\nhai\n');
		assert.equal(patch.new, 'một\nhai sửa\nba\n');

		// Phía đã stage dựng từ diff HEAD so với vùng chuẩn bị.
		await td.stage(root, ['a.txt']);
		const staged = (await td.diff(root, { staged: true, paths: ['a.txt'] }))[0];
		assert.equal(staged.old, 'một\nhai\n');
		assert.equal(staged.new, 'một\nhai sửa\nba\n');
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});

runTd('Td: commit rồi đọc lại nội dung từ HEAD', async () => {
	const td = new Td(TD_BIN!);
	const root = await repo(td, { 'a.txt': 'ban đầu\n' });
	try {
		write(root, 'a.txt', 'sau\n');
		await td.stage(root, ['a.txt']);
		await td.commit(root, 'c2');

		// diff của một commit là so với phụ huynh của nó.
		const head = (await td.diff(root, { revision: 'HEAD', paths: ['a.txt'] }))[0];
		assert.equal(head.old, 'ban đầu\n');
		assert.equal(head.new, 'sau\n');
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});

runTd('Td: gỡ khỏi vùng chuẩn bị rồi huỷ sửa đổi trên đĩa', async () => {
	const td = new Td(TD_BIN!);
	const root = await repo(td, { 'a.txt': 'ban đầu\n' });
	try {
		write(root, 'a.txt', 'sửa trên đĩa\n');
		await td.stage(root, ['a.txt']);
		let status = await td.status(root);
		assert.deepEqual(status.staged.map(e => e.path), ['a.txt']);

		await td.unstage(root, ['a.txt']);
		status = await td.status(root);
		assert.deepEqual(status.staged, []);
		assert.deepEqual(status.unstaged.map(e => e.path), ['a.txt']);

		await td.discard(root, ['a.txt']);
		status = await td.status(root);
		assert.deepEqual(status.unstaged, [], 'huỷ sửa đổi thì cây làm việc sạch');
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});

runTd('Td: xoá tệp chưa theo dõi, giữ nguyên tệp đã theo dõi', async () => {
	const td = new Td(TD_BIN!);
	const root = await repo(td, { 'theo-doi.txt': 'còn trong kho\n' });
	try {
		write(root, 'rac.txt', 'tạm\n');
		write(root, 'thu-muc/cong.txt', 'tạm\n');

		await td.clean(root, ['rac.txt']);

		assert.ok(!existsSync(path.join(root, 'rac.txt')), 'rac.txt phải bị xoá');
		assert.ok(existsSync(path.join(root, 'thu-muc/cong.txt')), 'tệp ngoài danh sách phải còn lại');
		assert.ok(existsSync(path.join(root, 'theo-doi.txt')), 'tệp đã theo dõi không được xoá');

		const status = await td.status(root);
		assert.deepEqual(status.untracked.map(e => e.path), ['thu-muc/cong.txt']);
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});

test('resolveExecutable: ưu tiên đường dẫn người dùng đặt trong cấu hình', () => {
	assert.equal(resolveExecutable('/opt/my-td/td'), '/opt/my-td/td');
});

test('resolveExecutable: không có cấu hình thì dùng lệnh td trong PATH', () => {
	assert.equal(resolveExecutable(''), 'td');
});

runTd('Td: nhánh, commit, tag và bản lưu tạm', async () => {
	const td = new Td(TD_BIN!);
	const root = await repo(td);
	try {
		await td.createBranch(root, 'nhanh-phu');
		let branches = await td.branches(root);
		// Tạo nhánh cũng chuyển sang nhánh đó, và nhánh đang đứng đứng đầu.
		assert.deepEqual(branches.map(b => b.name), ['nhanh-phu', 'main']);
		assert.equal(branches[0].current, true);

		await td.switchTo(root, 'main');
		branches = await td.branches(root);
		assert.equal(branches[0].name, 'main');
		write(root, 'việc.txt', 'dở dang\n');
		await td.stage(root, ['việc.txt']);
		await td.stash(root, 'lưu lại');
		const stashes = await td.stashes(root);
		assert.equal(stashes.length, 1);
		assert.equal(stashes[0].message, 'lưu lại');

		// Áp dụng lại bản lưu tạm thì tệp trở lại vùng chuẩn bị như lúc lưu.
		await td.applyStash(root, 0, false);
		const status = await td.status(root);
		assert.ok(status.staged.some(e => e.path === 'việc.txt'));

		await td.createTag(root, 'v1.0', 'ghi chú');
		const tags = await td.tags(root);
		assert.equal(tags[0].name, 'v1.0');
		assert.equal(tags[0].message, 'ghi chú');

		const entries = await td.log(root, 10);
		assert.equal(entries[0].summary, 'c1');

		await td.deleteBranch(root, 'nhanh-phu');
		branches = await td.branches(root);
		assert.deepEqual(branches.map(b => b.name), ['main']);
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});

runTd('Td: báo lỗi khi lệnh thất bại, giữ nguyên thông điệp của td', async () => {
	const td = new Td(TD_BIN!);
	const root = await repo(td);
	try {
		await assert.rejects(
			() => td.merge(root, 'khong-ton-tai'),
			(error: unknown) => {
				assert.ok(error instanceof TdError);
				assert.match(error.message, /không phân giải được tham chiếu/);
				return true;
			}
		);
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});

runTd('Td: commit thiếu nội dung hoặc không có gì để commit thì td từ chối', async () => {
	const td = new Td(TD_BIN!);
	const root = await repo(td);
	try {
		// Có thay đổi đã stage nhưng nội dung commit rỗng.
		write(root, 'a.txt', 'sửa\n');
		await td.stage(root, ['a.txt']);
		await assert.rejects(() => td.commit(root, ''), /thiếu nội dung commit/);

		// Gỡ khỏi vùng chuẩn bị rồi huỷ sửa đổi thì kho sạch, lúc đó td báo lỗi
		// khác: không có gì để commit.
		await td.unstage(root, ['a.txt']);
		await td.discard(root, ['a.txt']);
		await assert.rejects(() => td.commit(root, 'không có gì'), /không có gì để commit/);
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});

runTd('Td: show-file đọc nội dung ở một điểm lịch sử', async () => {
	const td = new Td(TD_BIN!);
	const root = await repo(td, { 'a.txt': 'một\nhai\n', 'khac.txt': 'giữ nguyên\n' });
	try {
		write(root, 'a.txt', 'một\nhai sửa\n');
		await td.stage(root, ['a.txt']);
		await td.commit(root, 'c2');

		const head = await td.showFile(root, 'HEAD', 'a.txt');
		assert.equal(head.found, true);
		assert.equal(head.unsupported, false);
		assert.equal(head.content, 'một\nhai sửa\n');

		const old = await td.showFile(root, 'HEAD~1', 'a.txt');
		assert.equal(old.content, 'một\nhai\n');

		// Tệp mà commit cuối không đụng tới vẫn đọc được.
		const untouched = await td.showFile(root, 'HEAD', 'khac.txt');
		assert.equal(untouched.content, 'giữ nguyên\n');

		// Tệp chưa có ở điểm đó là câu trả lời hợp lệ chứ không phải lỗi.
		const missing = await td.showFile(root, 'HEAD', 'khong-co.txt');
		assert.equal(missing.found, false);
		assert.equal(missing.unsupported, false);
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});

// Lớp vỏ báo như td cũ: chạy được mọi lệnh trừ show-file.
const runOldTd = process.platform === 'win32' ? test.skip : runTd;
runOldTd('Td: show-file báo lệnh còn thiếu khi td trên máy cũ hơn', async () => {
	const dir = mkdtempSync(path.join(tmpdir(), 'td-old-'));
	const real = new Td(TD_BIN!);
	const root = await repo(real);
	const wrapper = path.join(dir, 'td');
	writeFileSync(wrapper, [
		'#!/bin/sh',
		'for a in "$@"; do',
		'\tif [ "$a" = "show-file" ]; then',
		'\t\techo \'không có lệnh nào tên "show-file", xem danh sách: td vcs --help\' >&2',
		'\t\texit 1',
		'\tfi',
		'done',
		`exec ${JSON.stringify(TD_BIN!)} "$@"`,
		''
	].join('\n'), 'utf8');
	chmodSync(wrapper, 0o755);
	const old = new Td(wrapper);
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

runTd('Td: changedFiles chỉ liệt kê tên tệp, không lấy nội dung', async () => {
	const td = new Td(TD_BIN!);
	const root = await repo(td, { 'a.txt': 'một\n', 'b.txt': 'hai\n' });
	try {
		write(root, 'a.txt', 'một sửa\n');
		write(root, 'moi.txt', 'mới\n');

		assert.deepEqual((await td.changedFiles(root, { revision: 'HEAD' })).sort(), ['a.txt', 'b.txt']);

		// Lọc theo danh sách tệp thì chỉ còn tệp được nêu.
		assert.deepEqual(await td.changedFiles(root, { revision: 'HEAD', paths: ['a.txt'] }), ['a.txt']);

		// Không lọc thì lấy cả tệp đã sửa lẫn tệp chưa theo dõi trên đĩa.
		const unstaged = (await td.changedFiles(root)).sort();
		assert.deepEqual(unstaged, ['a.txt', 'moi.txt']);
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});
