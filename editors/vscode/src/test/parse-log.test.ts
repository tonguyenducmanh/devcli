import { test } from 'node:test';

import assert from 'node:assert/strict';

import { parseLogDetailed } from '../parse';

/**
 * Kiểm thử riêng cho bộ đọc log đầy đủ.
 *
 * Đây là nguồn của danh sách lịch sử tệp: thiếu một trường là người dùng mất
 * thông tin để chọn đúng commit, còn đọc lệch dòng thì ra tiêu đề sai.
 */
test('đọc đủ mã băm, tiêu đề, tham chiếu, tác giả và thời điểm', () => {
	const stdout = [
		'commit 3845daeee82b (HEAD -> main)',
		'Tác giả: Tô Nguyễn <ducmanh@example.com>',
		'Ngày:   2026-11-06 16:18:48',
		'',
		'    sửa tài liệu',
		'',
		'    giải thích dài hơn một dòng',
		'',
		'commit 359c076f9549 (main)',
		'Tác giả: Khác <khac@example.com>',
		'Ngày:   2026-11-05 09:00:00',
		'',
		'    thêm tệp'
	].join('\n');

	assert.deepEqual(parseLogDetailed(stdout), [
		{
			hash: '3845daeee82b',
			summary: 'sửa tài liệu',
			refs: ['HEAD -> main'],
			author: 'Tô Nguyễn',
			email: 'ducmanh@example.com',
			when: '2026-11-06 16:18:48'
		},
		{
			hash: '359c076f9549',
			summary: 'thêm tệp',
			refs: ['main'],
			author: 'Khác',
			email: 'khac@example.com',
			when: '2026-11-05 09:00:00'
		}
	]);
});

/**
 * Commit không có tham chiếu nào thì dòng đầu chỉ có mã băm.
 *
 * Commit cũ thường không còn nhánh nào trỏ tới, nên đây là trường hợp hay gặp.
 */
test('commit không có tham chiếu vẫn đọc được', () => {
	const entries = parseLogDetailed([
		'commit 1111111111111',
		'Tác giả: Ai đó <a@example.com>',
		'Ngày:   2026-01-02 03:04:05',
		'',
		'    nội dung'
	].join('\n'));

	assert.deepEqual(entries, [{
		hash: '1111111111111',
		summary: 'nội dung',
		refs: [],
		author: 'Ai đó',
		email: 'a@example.com',
		when: '2026-01-02 03:04:05'
	}]);
});

/**
 * Dòng thống kê tệp ở cuối không được lẫn vào tiêu đề.
 *
 * Dòng đó không thụt lề bốn khoảng trắng như phần thông điệp và nằm sau nó.
 */
test('dòng thống kê tệp không lẫn vào tiêu đề', () => {
	const entries = parseLogDetailed([
		'commit abc123456789',
		'Tác giả: Ai đó <a@example.com>',
		'Ngày:   2026-01-02 03:04:05',
		'',
		'    tiêu đề thật',
		'',
		' thêm    : tệp-mới.txt',
		' sửa     : tệp-cũ.txt'
	].join('\n'));

	assert.equal(entries.length, 1);
	assert.equal(entries[0].summary, 'tiêu đề thật');
});

/**
 * Tác giả không có địa chỉ thư thì lấy trọn phần sau nhãn.
 *
 * Cấu hình git trên máy không bắt buộc có địa chỉ thư, và lệnh log vẫn in ra.
 */
test('tác giả không có địa chỉ thư vẫn đọc được', () => {
	const entries = parseLogDetailed([
		'commit abc123456789',
		'Tác giả: Không Rõ',
		'Ngày:   2026-01-02 03:04:05',
		'',
		'    nội dung'
	].join('\n'));

	assert.equal(entries[0].author, 'Không Rõ');
	assert.equal(entries[0].email, undefined);
});

test('kho chưa có commit nào thì ra danh sách rỗng', () => {
	assert.deepEqual(parseLogDetailed('Chưa có commit nào.\n'), []);
});

test('output rỗng thì ra danh sách rỗng', () => {
	assert.deepEqual(parseLogDetailed(''), []);
});