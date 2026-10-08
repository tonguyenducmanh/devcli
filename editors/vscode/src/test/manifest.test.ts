import { test } from 'node:test';

import { manifest, nls } from './harness';

/**
 * Kiểm tra phần khai báo trong package.json.
 *
 * Những lỗi ở đây không biểu hiện lúc chạy mã nguồn nên kiểm thử kiểu kích
 * hoạt không bắt được. Đặc biệt là submenu rỗng: khai báo một submenu mà
 * không khai báo lệnh nào thuộc về nó thì nút ba chấm trên thanh Source Control
 * mở ra trống, người dùng thấy là không có thao tác nào.
 */
test('mọi submenu khai báo đều có lệnh', () => {
	const pkg = manifest();
	const empty = pkg.contributes.submenus
		.filter(submenu => !(pkg.contributes.menus[submenu.id] ?? []).length)
		.map(submenu => submenu.id);
	if (empty.length > 0) {
		throw new Error(`submenu không có lệnh nào, bấm vào sẽ mở ra trống: ${empty.join(', ')}`);
	}
});

test('mọi lệnh xuất hiện trong menu đều được khai báo', () => {
	const pkg = manifest();
	const declared = new Set(pkg.contributes.commands.map(c => c.command));
	const submenus = new Set(pkg.contributes.submenus.map(s => s.id));
	const unknown: string[] = [];
	for (const [menu, items] of Object.entries(pkg.contributes.menus)) {
		for (const item of items) {
			// Bảng lệnh dùng lệnh để ẩn đi nên "when": "false" là hợp lệ.
			const known = (id: string | undefined, pool: Set<string>) =>
				!id || pool.has(id) || (menu === 'commandPalette' && item.when === 'false');
			if (!known(item.command, declared)) {
				unknown.push(`${menu}: lệnh ${item.command}`);
			}
			if (!known(item.submenu, submenus)) {
				unknown.push(`${menu}: submenu ${item.submenu}`);
			}
			if (!known(item.alt, declared)) {
				unknown.push(`${menu}: lệnh thay thế ${item.alt}`);
			}
		}
	}
	if (unknown.length > 0) {
		throw new Error(`menu trỏ tới thứ chưa khai báo: ${unknown.join(', ')}`);
	}
});

test('mọi submenu dùng trong thanh Source Control đều được khai báo', () => {
	const pkg = manifest();
	const declared = new Set(pkg.contributes.submenus.map(s => s.id));
	const missing = pkg.contributes.menus['scm/title']
		.map(item => item.submenu)
		.filter((id): id is string => !!id && !declared.has(id));
	if (missing.length > 0) {
		throw new Error(`thanh Source Control dùng submenu chưa khai báo: ${missing.join(', ')}`);
	}
});

test('mọi lệnh khai báo đều có nhãn dịch', () => {
	const pkg = manifest();
	const labels = nls();
	const missing: string[] = [];
	for (const command of pkg.contributes.commands) {
		if (!command.title) {
			missing.push(`${command.command} (thiếu title)`);
			continue;
		}
		const key = command.title.replace(/^%/, '').replace(/%$/, '');
		if (!labels[key]) {
			missing.push(`${command.command} (${key})`);
		}
	}
	if (missing.length > 0) {
		throw new Error(`lệnh thiếu nhãn trong package.nls.json: ${missing.join(', ')}`);
	}
});

test('mỗi khung bên có lời nhắc khi còn trống', () => {
	const pkg = manifest();
	const views = pkg.contributes.views.scm ?? [];
	const welcome = new Set((pkg.contributes.viewsWelcome ?? []).map(w => w.view));
	const empty = views.filter(view => !welcome.has(view.id)).map(view => view.id);
	if (empty.length > 0) {
		throw new Error(`khung bên không có lời nhắc khi trống: ${empty.join(', ')}`);
	}
});

test('bấm tệp trong khung Source Control mở khung so sánh', () => {
	const pkg = manifest();
	const properties = pkg.contributes.configuration.flatMap(c => Object.entries(c.properties));
	const openDiff = properties.find(([key]) => key === 'td.openDiffOnClick');
	if (!openDiff) {
		throw new Error('thiếu cấu hình td.openDiffOnClick');
	}
	if (openDiff[1].default !== true) {
		throw new Error('td.openDiffOnClick phải mặc định là true để bấm tệp ra khung so sánh');
	}
});

test('nút cạnh tệp chỉ hiện một lệnh mở', () => {
	const pkg = manifest();
	const nav = pkg.contributes.menus['scm/resourceState/context']
		.filter(item => item.group === 'navigation');
	// Mỗi nhóm chỉ được có một nút cạnh tệp, không phải hai nút cùng chức năng.
	const openChange = nav.filter(item => item.command === 'td.openChange');
	if (openChange.length !== 1) {
		throw new Error(`nút mở khung so sánh phải hiện đúng một chỗ, nhận ${openChange.length}`);
	}
	// Tệp chưa theo dõi mở thẳng tệp, giống git.
	for (const item of nav) {
		if (item.command === 'td.openChange' && (item.when ?? '').includes('untracked')) {
			throw new Error('tệp chưa theo dõi không được mở khung so sánh ở nút cạnh tệp');
		}
	}
});