import * as assert from 'node:assert/strict';
import { execFileSync } from 'node:child_process';
import { mkdirSync, mkdtempSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import * as path from 'node:path';
import { test } from 'node:test';

import { findTm } from './harness';

import {
	parseBranches,
	parseLogOneline,
	parsePatches,
	parseStashes,
	parseStatus,
	parseTags
} from '../parse';

/**
 * Lệnh tm dùng để chạy thật trong kiểm thử.
 *
 * Cách tìm nằm trong harness để mọi kiểm thử dùng chung một cách, và để cách đó
 * loại được tệp build của nền tảng khác nằm cùng thư mục out/.
 */
const TM_BIN = findTm();

/** Kiểm thử nào cần lệnh tm thật thì bỏ qua khi máy chưa có tm. */
const runTm = TM_BIN ? test : test.skip;

/**
 * Chạy `tm vcs` trong một kho và trả về stdout.
 *
 * `-C` đặt trước mọi thứ khác, giống hệt cách extension gọi lệnh: đặt cuối dòng
 * sẽ bị tm hiểu là đường dẫn khi lệnh có tệp sau dấu `--`.
 */
function tm(root: string, ...args: string[]): string {
	return execFileSync(TM_BIN!, ['vcs', '-C', root, ...args], { encoding: 'utf8' });
}

function write(root: string, rel: string, body: string): void {
	const full = path.join(root, rel);
	writeFileSync(full, body, 'utf8');
}

function init(): string {
	const root = mkdtempSync(path.join(tmpdir(), 'tm-parse-'));
	tm(root, 'init');
	return root;
}

// ─── status ─────────────────────────────────────────────────────

runTm('status: đọc ba nhóm tệp và nhánh đang đứng', () => {
	const root = init();
	try {
		write(root, 'giai.txt', 'a\n');
		write(root, 'xoa.txt', 'b\n');
		tm(root, 'add', '.');
		tm(root, 'commit', '-m', 'c1');

		write(root, 'giai.txt', 'a\nsửa\n');
		tm(root, 'add', 'giai.txt');
		rmSync(path.join(root, 'xoa.txt'));
		write(root, 'moi.txt', 'mới\n');

		const status = parseStatus(tm(root, 'status'));
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

runTm('status: giữ nguyên đường dẫn có khoảng trắng và dấu tiếng Việt', () => {
	const root = init();
	try {
		mkdirSync(path.join(root, 'thu muc'), { recursive: true });
		write(root, 'thu muc/tên có dấu.txt', 'x\n');
		const status = parseStatus(tm(root, 'status'));
		// Tệp .tmxignore do `tm vcs init` tạo sẵn cũng chưa được theo dõi, và
		// tên của nó không có dấu hay khoảng trắng nên phải đọc đúng như mọi
		// tệp khác thay vì bị lọc nhầm.
		assert.deepEqual(status.untracked, [
			{ path: '.tmxignore', status: '?' },
			{ path: 'thu muc/tên có dấu.txt', status: '?' }
		]);
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});

runTm('status: HEAD tách rời không có tên nhánh', () => {
	const root = init();
	try {
		write(root, 'a.txt', 'a\n');
		tm(root, 'add', '.');
		tm(root, 'commit', '-m', 'c1');
		tm(root, 'checkout', '--detach', 'HEAD');

		const status = parseStatus(tm(root, 'status'));
		assert.equal(status.detached, true);
		assert.equal(status.branch, '');
		assert.match(status.head, /^[0-9a-f]{8}$/);
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});

runTm('status: đọc số commit đi trước và đi sau', () => {
	const root = init();
	try {
		write(root, 'a.txt', 'a\n');
		tm(root, 'add', '.');
		tm(root, 'commit', '-m', 'c1');
		// Nhánh theo dõi phải trỏ tới một nhánh khác thì mới đo được đi trước.
		tm(root, 'branch', 'theo-doi');
		tm(root, 'switch', 'main');
		// `config` là lệnh gốc của tm chứ không nằm dưới `vcs`.
		execFileSync(TM_BIN!, ['config', '-C', root, 'branch.main.merge', 'refs/heads/theo-doi'], { encoding: 'utf8' });
		execFileSync(TM_BIN!, ['config', '-C', root, 'branch.main.remote', 'gia'], { encoding: 'utf8' });
		write(root, 'a.txt', 'b\n');
		tm(root, 'add', '.');
		tm(root, 'commit', '-m', 'c2');

		const status = parseStatus(tm(root, 'status'));
		assert.equal(status.hasUpstream, true);
		assert.equal(status.ahead, 1);
		assert.equal(status.behind, 0);
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});

// ─── diff ───────────────────────────────────────────────────────

runTm('diff: dựng lại cả hai phía cho tệp đã sửa', () => {
	const root = init();
	try {
		write(root, 'a.txt', 'một\nhai\nba\n');
		tm(root, 'add', '.');
		tm(root, 'commit', '-m', 'c1');
		write(root, 'a.txt', 'một\nhai sửa\nba\nbốn\n');

		const patches = parsePatches(tm(root, 'diff', '-U1000000'));
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

runTm('diff: tệp thêm mới chỉ có phía mới', () => {
	const root = init();
	try {
		write(root, 'a.txt', 'x\n');
		tm(root, 'add', '.');
		tm(root, 'commit', '-m', 'c1');
		write(root, 'b.txt', 'mới\n');

		const patches = parsePatches(tm(root, 'diff', '-U1000000'));
		assert.equal(patches.length, 1);
		assert.equal(patches[0].status, 'A');
		assert.equal(patches[0].old, '');
		assert.equal(patches[0].new, 'mới\n');
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});

runTm('diff: tệp bị xoá chỉ có phía cũ', () => {
	const root = init();
	try {
		write(root, 'a.txt', 'x\n');
		tm(root, 'add', '.');
		tm(root, 'commit', '-m', 'c1');
		rmSync(path.join(root, 'a.txt'));

		const patches = parsePatches(tm(root, 'diff', '-U1000000'));
		assert.equal(patches.length, 1);
		assert.equal(patches[0].status, 'D');
		assert.equal(patches[0].old, 'x\n');
		assert.equal(patches[0].new, '');
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});

runTm('diff: lọc theo đường dẫn sau dấu hai gạch ngang', () => {
	const root = init();
	try {
		write(root, 'muon.txt', 'a\n');
		write(root, 'khong.txt', 'b\n');
		tm(root, 'add', '.');
		tm(root, 'commit', '-m', 'c1');
		write(root, 'muon.txt', 'sửa\n');
		write(root, 'khong.txt', 'sửa\n');

		const patches = parsePatches(tm(root, 'diff', '-U1000000', '--', 'muon.txt'));
		assert.equal(patches.length, 1);
		assert.equal(patches[0].path, 'muon.txt');
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});

runTm('diff: so với một commit cho trước', () => {
	const root = init();
	try {
		write(root, 'a.txt', 'một\n');
		tm(root, 'add', '.');
		tm(root, 'commit', '-m', 'c1');
		write(root, 'a.txt', 'một\nhai\n');
		tm(root, 'add', '.');
		tm(root, 'commit', '-m', 'c2');
		const hash = parseLogOneline(tm(root, 'log', '--oneline', '-n1'))[0].hash;

		const patches = parsePatches(tm(root, 'diff', hash, '-U1000000', '--', 'a.txt'));
		assert.equal(patches.length, 1);
		assert.equal(patches[0].old, 'một\n');
		assert.equal(patches[0].new, 'một\nhai\n');
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});

// ─── branch, tag, stash, log ────────────────────────────────────

runTm('branch: đánh dấu nhánh đang đứng và nhánh phụ', () => {
	const root = init();
	try {
		write(root, 'a.txt', 'a\n');
		tm(root, 'add', '.');
		tm(root, 'commit', '-m', 'c1');
		tm(root, 'branch', 'nhanh-phu');

		const branches = parseBranches(tm(root, 'branch', '-vv'));
		assert.equal(branches.length, 2);
		// Tạo nhánh bằng `tm vcs branch <tên>` cũng chuyển sang nhánh đó, và nhánh
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

runTm('tag: đọc tag nhẹ và tag có chú thích', () => {
	const root = init();
	try {
		write(root, 'a.txt', 'a\n');
		tm(root, 'add', '.');
		tm(root, 'commit', '-m', 'c1');
		tm(root, 'tag', 'nhe');
		tm(root, 'tag', '-a', 'co-chu-thich', '-m', 'ghi chú');

		const tags = parseTags(tm(root, 'tag', '-l'));
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

runTm('stash: đọc danh sách bản lưu tạm', () => {
	const root = init();
	try {
		write(root, 'a.txt', 'a\n');
		tm(root, 'add', '.');
		tm(root, 'commit', '-m', 'c1');
		write(root, 'a.txt', 'việc dở\n');
		tm(root, 'stash', '-m', 'việc dở');

		const stashes = parseStashes(tm(root, 'stash', '--list'));
		assert.equal(stashes.length, 1);
		assert.equal(stashes[0].index, 0);
		assert.equal(stashes[0].message, 'việc dở');
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});

runTm('log: tách mã băm, tiêu đề và tham chiếu', () => {
	const root = init();
	try {
		write(root, 'a.txt', 'a\n');
		tm(root, 'add', '.');
		tm(root, 'commit', '-m', 'tiêu đề có dấu');

		const entries = parseLogOneline(tm(root, 'log', '--oneline', '-n5'));
		assert.equal(entries.length, 1);
		assert.equal(entries[0].summary, 'tiêu đề có dấu');
		assert.deepEqual(entries[0].refs, ['HEAD -> main']);
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});

// ─── Kiểm thử bằng dữ liệu mẫu, không cần lệnh tm ───────────────

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