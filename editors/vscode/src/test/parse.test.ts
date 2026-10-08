import * as assert from 'node:assert/strict';
import { execFileSync } from 'node:child_process';
import { existsSync, mkdirSync, mkdtempSync, readdirSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import * as path from 'node:path';
import { test } from 'node:test';

import {
	parseBranches,
	parseLogOneline,
	parsePatches,
	parseStashes,
	parseStatus,
	parseTags
} from '../parse';

/** Gốc kho của td, thư mục cha của thư mục extension. */
const REPO_ROOT = path.join(__dirname, '..', '..', '..', '..');

/**
 * Tìm lệnh td để chạy thật trong kiểm thử.
 *
 * Ưu tiên biến môi trường TD_BIN, sau đó tới tệp đã build trong out/ của kho td,
 * cuối cùng mới tới lệnh `td` có sẵn trong PATH. Không tìm thấy thì kiểm thử
 * phần phân tích vẫn chạy được, chỉ bỏ qua các kiểm thử cần lệnh thật.
 */
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

/** Kiểm thử nào cần lệnh td thật thì bỏ qua khi máy chưa có td. */
const runTd = TD_BIN ? test : test.skip;

/**
 * Chạy `td vcs` trong một kho và trả về stdout.
 *
 * `-C` đặt trước mọi thứ khác, giống hệt cách extension gọi lệnh: đặt cuối dòng
 * sẽ bị td hiểu là đường dẫn khi lệnh có tệp sau dấu `--`.
 */
function td(root: string, ...args: string[]): string {
	return execFileSync(TD_BIN!, ['vcs', '-C', root, ...args], { encoding: 'utf8' });
}

function write(root: string, rel: string, body: string): void {
	const full = path.join(root, rel);
	writeFileSync(full, body, 'utf8');
}

function init(): string {
	const root = mkdtempSync(path.join(tmpdir(), 'td-parse-'));
	td(root, 'init');
	return root;
}

// ─── status ─────────────────────────────────────────────────────

runTd('status: đọc ba nhóm tệp và nhánh đang đứng', () => {
	const root = init();
	try {
		write(root, 'giai.txt', 'a\n');
		write(root, 'xoa.txt', 'b\n');
		td(root, 'add', '.');
		td(root, 'commit', '-m', 'c1');

		write(root, 'giai.txt', 'a\nsửa\n');
		td(root, 'add', 'giai.txt');
		rmSync(path.join(root, 'xoa.txt'));
		write(root, 'moi.txt', 'mới\n');

		const status = parseStatus(td(root, 'status'));
		assert.equal(status.branch, 'main');
		assert.equal(status.detached, false);
		assert.deepEqual(status.staged, [{ path: 'giai.txt', status: 'M' }]);
		assert.deepEqual(status.unstaged, [{ path: 'xoa.txt', status: 'D' }]);
		// Tệp chưa theo dõi chỉ xuất hiện một lần, ở nhóm riêng của nó.
		assert.deepEqual(status.untracked, [{ path: 'moi.txt', status: '?' }]);
		assert.deepEqual(status.conflicts, []);
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});

runTd('status: giữ nguyên đường dẫn có khoảng trắng và dấu tiếng Việt', () => {
	const root = init();
	try {
		mkdirSync(path.join(root, 'thu muc'), { recursive: true });
		write(root, 'thu muc/tên có dấu.txt', 'x\n');
		const status = parseStatus(td(root, 'status'));
		assert.deepEqual(status.untracked, [{ path: 'thu muc/tên có dấu.txt', status: '?' }]);
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});

runTd('status: HEAD tách rời không có tên nhánh', () => {
	const root = init();
	try {
		write(root, 'a.txt', 'a\n');
		td(root, 'add', '.');
		td(root, 'commit', '-m', 'c1');
		td(root, 'checkout', '--detach', 'HEAD');

		const status = parseStatus(td(root, 'status'));
		assert.equal(status.detached, true);
		assert.equal(status.branch, '');
		assert.match(status.head, /^[0-9a-f]{8}$/);
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});

runTd('status: đọc số commit đi trước và đi sau', () => {
	const root = init();
	try {
		write(root, 'a.txt', 'a\n');
		td(root, 'add', '.');
		td(root, 'commit', '-m', 'c1');
		// Nhánh theo dõi phải trỏ tới một nhánh khác thì mới đo được đi trước.
		td(root, 'branch', 'theo-doi');
		td(root, 'switch', 'main');
		// `config` là lệnh gốc của td chứ không nằm dưới `vcs`.
		execFileSync(TD_BIN!, ['config', '-C', root, 'branch.main.merge', 'refs/heads/theo-doi'], { encoding: 'utf8' });
		execFileSync(TD_BIN!, ['config', '-C', root, 'branch.main.remote', 'gia'], { encoding: 'utf8' });
		write(root, 'a.txt', 'b\n');
		td(root, 'add', '.');
		td(root, 'commit', '-m', 'c2');

		const status = parseStatus(td(root, 'status'));
		assert.equal(status.hasUpstream, true);
		assert.equal(status.ahead, 1);
		assert.equal(status.behind, 0);
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});

// ─── diff ───────────────────────────────────────────────────────

runTd('diff: dựng lại cả hai phía cho tệp đã sửa', () => {
	const root = init();
	try {
		write(root, 'a.txt', 'một\nhai\nba\n');
		td(root, 'add', '.');
		td(root, 'commit', '-m', 'c1');
		write(root, 'a.txt', 'một\nhai sửa\nba\nbốn\n');

		const patches = parsePatches(td(root, 'diff', '-U1000000'));
		assert.equal(patches.length, 1);
		assert.equal(patches[0].path, 'a.txt');
		assert.equal(patches[0].status, 'M');
		assert.equal(patches[0].old, 'một\nhai\nba\n');
		assert.equal(patches[0].new, 'một\nhai sửa\nba\nbốn\n');
		assert.equal(patches[0].added, 2);
		assert.equal(patches[0].deleted, 1);
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});

runTd('diff: tệp thêm mới chỉ có phía mới', () => {
	const root = init();
	try {
		write(root, 'a.txt', 'x\n');
		td(root, 'add', '.');
		td(root, 'commit', '-m', 'c1');
		write(root, 'b.txt', 'mới\n');

		const patches = parsePatches(td(root, 'diff', '-U1000000'));
		assert.equal(patches.length, 1);
		assert.equal(patches[0].status, 'A');
		assert.equal(patches[0].old, '');
		assert.equal(patches[0].new, 'mới\n');
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});

runTd('diff: tệp bị xoá chỉ có phía cũ', () => {
	const root = init();
	try {
		write(root, 'a.txt', 'x\n');
		td(root, 'add', '.');
		td(root, 'commit', '-m', 'c1');
		rmSync(path.join(root, 'a.txt'));

		const patches = parsePatches(td(root, 'diff', '-U1000000'));
		assert.equal(patches.length, 1);
		assert.equal(patches[0].status, 'D');
		assert.equal(patches[0].old, 'x\n');
		assert.equal(patches[0].new, '');
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});

runTd('diff: lọc theo đường dẫn sau dấu hai gạch ngang', () => {
	const root = init();
	try {
		write(root, 'muon.txt', 'a\n');
		write(root, 'khong.txt', 'b\n');
		td(root, 'add', '.');
		td(root, 'commit', '-m', 'c1');
		write(root, 'muon.txt', 'sửa\n');
		write(root, 'khong.txt', 'sửa\n');

		const patches = parsePatches(td(root, 'diff', '-U1000000', '--', 'muon.txt'));
		assert.equal(patches.length, 1);
		assert.equal(patches[0].path, 'muon.txt');
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});

runTd('diff: so với một commit cho trước', () => {
	const root = init();
	try {
		write(root, 'a.txt', 'một\n');
		td(root, 'add', '.');
		td(root, 'commit', '-m', 'c1');
		write(root, 'a.txt', 'một\nhai\n');
		td(root, 'add', '.');
		td(root, 'commit', '-m', 'c2');
		const hash = parseLogOneline(td(root, 'log', '--oneline', '-n1'))[0].hash;

		const patches = parsePatches(td(root, 'diff', hash, '-U1000000', '--', 'a.txt'));
		assert.equal(patches.length, 1);
		assert.equal(patches[0].old, 'một\n');
		assert.equal(patches[0].new, 'một\nhai\n');
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});

// ─── branch, tag, stash, log ────────────────────────────────────

runTd('branch: đánh dấu nhánh đang đứng và nhánh phụ', () => {
	const root = init();
	try {
		write(root, 'a.txt', 'a\n');
		td(root, 'add', '.');
		td(root, 'commit', '-m', 'c1');
		td(root, 'branch', 'nhanh-phu');

		const branches = parseBranches(td(root, 'branch', '-vv'));
		assert.equal(branches.length, 2);
		// Tạo nhánh bằng `td vcs branch <tên>` cũng chuyển sang nhánh đó, và nhánh
		// đang đứng luôn đứng đầu, y hệt git.
		assert.equal(branches[0].name, 'nhanh-phu');
		assert.equal(branches[0].current, true);
		assert.equal(branches[0].subject, 'c1');
		assert.equal(branches[1].name, 'main');
		assert.equal(branches[1].current, false);
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});

runTd('tag: đọc tag nhẹ và tag có chú thích', () => {
	const root = init();
	try {
		write(root, 'a.txt', 'a\n');
		td(root, 'add', '.');
		td(root, 'commit', '-m', 'c1');
		td(root, 'tag', 'nhe');
		td(root, 'tag', '-a', 'co-chu-thich', '-m', 'ghi chú');

		const tags = parseTags(td(root, 'tag', '-l'));
		assert.equal(tags.length, 2);
		const light = tags.find(t => t.name === 'nhe')!;
		const annotated = tags.find(t => t.name === 'co-chu-thich')!;
		assert.equal(light.annotated, false);
		assert.match(light.hash, /^[0-9a-f]{8}$/);
		assert.equal(annotated.annotated, true);
		assert.equal(annotated.message, 'ghi chú');
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});

runTd('stash: đọc danh sách bản lưu tạm', () => {
	const root = init();
	try {
		write(root, 'a.txt', 'a\n');
		td(root, 'add', '.');
		td(root, 'commit', '-m', 'c1');
		write(root, 'a.txt', 'việc dở\n');
		td(root, 'stash', '-m', 'việc dở');

		const stashes = parseStashes(td(root, 'stash', '--list'));
		assert.equal(stashes.length, 1);
		assert.equal(stashes[0].index, 0);
		assert.equal(stashes[0].message, 'việc dở');
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});

runTd('log: tách mã băm, tiêu đề và tham chiếu', () => {
	const root = init();
	try {
		write(root, 'a.txt', 'a\n');
		td(root, 'add', '.');
		td(root, 'commit', '-m', 'tiêu đề có dấu');

		const entries = parseLogOneline(td(root, 'log', '--oneline', '-n5'));
		assert.equal(entries.length, 1);
		assert.equal(entries[0].summary, 'tiêu đề có dấu');
		assert.deepEqual(entries[0].refs, ['HEAD -> main']);
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});

// ─── Kiểm thử bằng dữ liệu mẫu, không cần lệnh td ───────────────

test('parseStatus: bỏ qua dòng không thuộc nhóm nào', () => {
	const status = parseStatus([
		'Trên nhánh main',
		'',
		'Thay đổi đã stage:',
		'  sửa  a.txt',
		''
	].join('\n'));
	assert.equal(status.branch, 'main');
	assert.deepEqual(status.staged, [{ path: 'a.txt', status: 'M' }]);
});

test('parsePatches: bỏ qua dòng thông báo khi không có khác biệt', () => {
	assert.deepEqual(parsePatches('Không có khác biệt nào.\n'), []);
});

test('parseLogOneline: bỏ qua dòng không phải commit', () => {
	assert.deepEqual(parseLogOneline('Chưa có commit nào.\n'), []);
});