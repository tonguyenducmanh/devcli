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

/**
 * Tệp ignore của tm phải được tô màu như tệp ignore, không phải văn bản thuần.
 *
 * Không khai báo thì VS Code mở `.tmxignore` ra với ngôn ngữ `plaintext`, người
 * dùng không thấy chỗ nào là chú thích, chỗ nào là phủ định lại quy tắc trước.
 * Ngôn ngữ `ignore` do tiện ích git của VS Code khai báo, nên chỉ cần gắn tên
 * tệp vào ngôn ngữ đó chứ không cần mang theo cú pháp.
 */
test('tệp ignore của tm dùng ngôn ngữ ignore', () => {
	const pkg = manifest();
	const ignore = (pkg.contributes.languages ?? []).find(language => language.id === 'ignore');
	if (!ignore) {
		throw new Error('chưa gắn tệp ignore của tm vào ngôn ngữ ignore');
	}
	if (!(ignore.extensions ?? []).includes('.tmxignore')) {
		throw new Error('tệp .tmxignore chưa được gắn vào ngôn ngữ ignore');
	}
	// Tệp của riêng máy có tên là `exclude`, không có dấu chấm nào để bám vào,
	// nên chỉ có thể khớp theo đường dẫn.
	const patterns = ignore.filenamePatterns ?? [];
	if (!patterns.includes('**/.tmx/info/exclude')) {
		throw new Error('tệp .tmx/info/exclude chưa được gắn vào ngôn ngữ ignore');
	}
});

/**
 * Mọi view trong khung Source Control phải luôn hiện, không giấu bằng `when`.
 *
 * Một view có `when` mà điều kiện chưa đúng lúc VS Code dựng khung thì không
 * biến mất: nó bị tách sang một khung Source Control riêng. Kết quả là activity
 * bar hiện hai biểu tượng giống hệt nhau, một cái cho thay đổi tệp và một cái
 * cho nhánh, commit, stash, tag.
 *
 * Khung trống thì không sao, vì `viewsWelcome` đã có lời nhắc cho từng view.
 */
test('mọi view trong khung Source Control luôn hiện để không bị tách khung', () => {
	const pkg = manifest();
	const hidden = (pkg.contributes.views.scm ?? [])
		.filter(view => view.when)
		.map(view => `${view.id} (${view.when})`);
	if (hidden.length > 0) {
		throw new Error(`view có "when" sẽ bị tách sang khung Source Control riêng: ${hidden.join(', ')}`);
	}
});

test('bấm tệp trong khung Source Control mở khung so sánh', () => {
	const pkg = manifest();
	const properties = pkg.contributes.configuration.flatMap(c => Object.entries(c.properties));
	const openDiff = properties.find(([key]) => key === 'tm.openDiffOnClick');
	if (!openDiff) {
		throw new Error('thiếu cấu hình tm.openDiffOnClick');
	}
	if (openDiff[1].default !== true) {
		throw new Error('tm.openDiffOnClick phải mặc định là true để bấm tệp ra khung so sánh');
	}
});

test('nút cạnh tệp chỉ hiện một lệnh mở', () => {
	const pkg = manifest();
	const nav = pkg.contributes.menus['scm/resourceState/context']
		.filter(item => item.group === 'navigation');
	// Mỗi nhóm chỉ được có một nút cạnh tệp, không phải hai nút cùng chức năng.
	const openChange = nav.filter(item => item.command === 'tm.openChange');
	if (openChange.length !== 1) {
		throw new Error(`nút mở khung so sánh phải hiện đúng một chỗ, nhận ${openChange.length}`);
	}
	// Tệp chưa theo dõi mở thẳng tệp, giống git.
	for (const item of nav) {
		if (item.command === 'tm.openChange' && (item.when ?? '').includes('untracked')) {
			throw new Error('tệp chưa theo dõi không được mở khung so sánh ở nút cạnh tệp');
		}
	}
});