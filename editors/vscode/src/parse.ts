/**
 * Cách hiểu output của lệnh `td`.
 *
 * td in thông tin cho người đọc bằng tiếng Việt, không phải cho kịch bản đọc.
 * Extension vì thế phải dịch từng dòng output thành dữ liệu có kiểu rõ ràng.
 * Toàn bộ phần dịch nằm gọn trong tệp này để khi cây lệnh của td đổi thì chỉ
 * chỗ này phải sửa.
 */

/** Ký hiệu trạng thái của một tệp, theo đúng byte mà td dùng. */
export type TdStatusCode = 'A' | 'M' | 'D' | 'U' | '?';

/** Một tệp trong output của `td vcs status`. */
export interface TdStatusEntry {
	path: string;
	status: TdStatusCode;
}

/** Toàn bộ trạng thái của một kho, đọc từ `td vcs status`. */
export interface TdStatus {
	/** Tên nhánh đang đứng, rỗng khi HEAD đang tách rời hoặc chưa có nhánh. */
	branch: string;
	/** true khi HEAD không gắn với nhánh nào. */
	detached: boolean;
	/** Mã băm ngắn của HEAD, rỗng khi kho chưa có commit nào. */
	head: string;
	/** Nhánh hiện tại có theo dõi nhánh nào không. */
	hasUpstream: boolean;
	ahead: number;
	behind: number;
	staged: TdStatusEntry[];
	unstaged: TdStatusEntry[];
	untracked: TdStatusEntry[];
	conflicts: TdStatusEntry[];
}

/** Nội dung hai phía của một tệp, dựng lại từ diff của td. */
export interface TdPatch {
	path: string;
	status: 'A' | 'M' | 'D';
	binary: boolean;
	/** Nội dung phía cũ, rỗng với tệp mới thêm vào. */
	old: string;
	/** Nội dung phía mới, rỗng với tệp đã xoá. */
	new: string;
	added: number;
	deleted: number;
}

/** Một dòng của `td vcs log --oneline`. */
export interface TdLogEntry {
	hash: string;
	summary: string;
	refs: string[];
}

/** Một nhánh trong output của `td vcs branch`. */
export interface TdBranch {
	name: string;
	current: boolean;
	hash: string;
	subject: string;
	upstream: string;
	ahead: number;
	behind: number;
}

/** Một tag trong output của `td vcs tag`. */
export interface TdTag {
	name: string;
	hash: string;
	annotated: boolean;
	message: string;
}

/** Một mục trong output của `td vcs stash --list`. */
export interface TdStash {
	/** Vị trí trong danh sách, 0 là mới nhất. */
	index: number;
	message: string;
}

/**
 * Dấu nhận biết đầu dòng tiêu đề của một tệp trong output diff.
 *
 * Nhãn được td in dài đều nhau rồi thêm dấu hai chấm, ví dụ `Sửa     : a.txt`.
 */
const FILE_HEADER = /^(Thêm|Sửa|Xoá)\s*:\s?(.*)$/;

/** Đầu dòng của một hunk, ví dụ `@@ -1,3 +1,4 @@`. */
const HUNK_HEADER = /^@@ -(\d+),(\d+) \+(\d+),(\d+) @@/;

/** Dòng báo tệp nhị phân, td không in nội dung cho loại tệp này. */
const BINARY_LINE = '(file nhị phân, không hiển thị nội dung)';

/** Các nhóm tệp trong output của `td vcs status`. */
const SECTION_STAGED = 'Thay đổi đã stage:';
const SECTION_UNSTAGED = 'Thay đổi chưa stage:';
const SECTION_UNTRACKED = 'File chưa được theo dõi:';
const SECTION_CONFLICTS = 'Xung đột cần giải quyết:';

/** Từ khoá tiếng Việt mà td dùng cho ký hiệu trạng thái. */
const STATUS_WORDS: Record<string, TdStatusCode> = {
	'thêm': 'A',
	'sửa': 'M',
	'xoá': 'D',
	'?': '?',
	// Trong nhóm "chưa stage" td cũng liệt kê tệp chưa theo dõi, ghi là "mới".
	// Nhóm chưa theo dõi có mục riêng nên mục trùng ở đây bị bỏ qua.
	'mới': '?',
	'U': 'U'
};

/** Dòng đầu tiên của status cho biết đang đứng ở đâu. */
const LINE_BRANCH = /^Trên nhánh (.+)$/;
const LINE_DETACHED = /^Trên HEAD tách rời tại (\S+)/;
const LINE_DETACHED_EMPTY = /^Trên HEAD tách rời/;
const LINE_NO_BRANCH = /^Chưa có nhánh nào/;
const LINE_UPSTREAM = /^Theo dõi: đi trước (\d+), đi sau (\d+)$/;
const LINE_SYNCED = /^Đã đồng bộ với nhánh theo dõi$/;

/**
 * Đọc output của `td vcs status`.
 *
 * Tệp chưa được theo dõi xuất hiện ở cả nhóm chưa stage lẫn nhóm chưa theo
 * dõi. Nhóm chưa stage bỏ qua chúng để mỗi tệp chỉ xuất hiện một lần, đúng
 * như khung Source Control của VS Code.
 */
export function parseStatus(stdout: string): TdStatus {
	const status: TdStatus = {
		branch: '',
		detached: false,
		head: '',
		hasUpstream: false,
		ahead: 0,
		behind: 0,
		staged: [],
		unstaged: [],
		untracked: [],
		conflicts: []
	};

	let section = '';
	for (const raw of stdout.split('\n')) {
		const line = raw.replace(/\r$/, '');
		const trimmed = line.trim();

		if (!trimmed) {
			section = '';
			continue;
		}

		if (trimmed === SECTION_STAGED) {
			section = 'staged';
			continue;
		}
		if (trimmed === SECTION_UNSTAGED) {
			section = 'unstaged';
			continue;
		}
		if (trimmed === SECTION_UNTRACKED) {
			section = 'untracked';
			continue;
		}
		if (trimmed === SECTION_CONFLICTS) {
			section = 'conflicts';
			continue;
		}

		// Dòng đầu tiên mô tả vị trí đứng, đứng ngoài mọi nhóm.
		const branch = LINE_BRANCH.exec(trimmed);
		if (branch) {
			status.branch = branch[1].trim();
			continue;
		}
		if (LINE_NO_BRANCH.test(trimmed)) {
			status.branch = '';
			continue;
		}
		const detached = LINE_DETACHED.exec(trimmed);
		if (detached) {
			status.detached = true;
			status.head = detached[1];
			continue;
		}
		if (LINE_DETACHED_EMPTY.test(trimmed)) {
			status.detached = true;
			continue;
		}
		const upstream = LINE_UPSTREAM.exec(trimmed);
		if (upstream) {
			status.hasUpstream = true;
			status.ahead = Number(upstream[1]);
			status.behind = Number(upstream[2]);
			continue;
		}
		// Nhánh có theo dõi mà đang ngang bằng thì td in dạng khác.
		if (LINE_SYNCED.test(trimmed)) {
			status.hasUpstream = true;
			continue;
		}

		if (!section) {
			continue;
		}
		const entry = parseStatusEntry(line);
		if (!entry) {
			continue;
		}
		switch (section) {
			case 'staged':
				status.staged.push(entry);
				break;
			case 'unstaged':
				// Tệp chưa theo dõi không thuộc nhóm này, nó có nhóm riêng.
				if (entry.status !== '?') {
					status.unstaged.push(entry);
				}
				break;
			case 'untracked':
				status.untracked.push(entry);
				break;
			case 'conflicts':
				status.conflicts.push(entry);
				break;
		}
	}

	return status;
}

/**
 * Đọc một dòng tệp trong status.
 *
 * Dòng gồm ký hiệu trạng thái rồi đến đường dẫn, ví dụ `  sửa  thư mục/a.txt`.
 * Đường dẫn có thể chứa cả khoảng trắng nên phần còn lại của dòng được giữ
 * nguyên thay vì cắt theo từ.
 */
function parseStatusEntry(line: string): TdStatusEntry | undefined {
	if (!/^\s+\S/.test(line)) {
		return undefined;
	}
	const withoutMargin = line.replace(/^\s+/, '');
	const space = withoutMargin.search(/\s/);
	if (space < 0) {
		return undefined;
	}
	const word = withoutMargin.slice(0, space);
	const path = withoutMargin.slice(space).trim();
	const status = STATUS_WORDS[word];
	if (!status || !path) {
		return undefined;
	}
	return { path, status };
}

/**
 * Đọc output của `td vcs diff` và dựng lại nội dung hai phía cho từng tệp.
 *
 * Extension cần nội dung đầy đủ của cả hai phía chứ không chỉ phần thay đổi,
 * vì VS Code tự tính vạch hiệu khi mở khung so sánh. Vì vậy lệnh diff luôn
 * được gọi với số dòng ngữ cảnh rất lớn, khiến hunk ôm trọn tệp.
 */
export function parsePatches(stdout: string): TdPatch[] {
	const patches: TdPatch[] = [];
	let current: TdPatch | undefined;
	let inHunk = false;

	const flush = () => {
		if (current) {
			patches.push(current);
			current = undefined;
		}
		inHunk = false;
	};

	for (const raw of stdout.split('\n')) {
		const line = raw.replace(/\r$/, '');
		if (!line) {
			// Dòng trống kết thúc tệp hiện tại. Mọi dòng của diff đều có tiền tố
			// nên dòng trống không bao giờ thuộc hunk.
			flush();
			continue;
		}

		if (!current) {
			const header = FILE_HEADER.exec(line);
			if (header) {
				current = {
					path: header[2].trim(),
					status: header[1] === 'Thêm' ? 'A' : header[1] === 'Xoá' ? 'D' : 'M',
					binary: false,
					old: '',
					new: '',
					added: 0,
					deleted: 0
				};
			}
			continue;
		}

		if (HUNK_HEADER.test(line)) {
			inHunk = true;
			continue;
		}
		if (!inHunk) {
			// Ngoài hunk thì chỉ còn dòng báo tệp nhị phân.
			if (line.includes(BINARY_LINE)) {
				current.binary = true;
			}
			continue;
		}

		const prefix = line[0];
		const body = line.slice(1);
		switch (prefix) {
			case '+':
				current.new += body + '\n';
				current.added += 1;
				break;
			case '-':
				current.old += body + '\n';
				current.deleted += 1;
				break;
			case ' ':
				current.old += body + '\n';
				current.new += body + '\n';
				break;
			default:
				break;
		}
	}
	flush();

	return patches;
}

/**
 * Đọc output của `td vcs diff --name-only`.
 *
 * Lệnh in mỗi tệp thay đổi trên một dòng, không kèm tiền tố. Dòng nào không
 * phải đường dẫn thì bỏ qua, vì lệnh có thể in thêm dòng trống ở cuối.
 */
export function parseChangedPaths(stdout: string): string[] {
	const out: string[] = [];
	for (const raw of stdout.split('\n')) {
		const line = raw.replace(/\r$/, '').trim();
		if (line) {
			out.push(line);
		}
	}
	return out;
}

/**
 * Đọc output của `td vcs log --oneline`.
 *
 * Mỗi dòng có dạng `<mã băm> <tiêu đề> (<các tham chiếu>)`, phần tham chiếu
 * là tuỳ chọn nên phải nhận ra bằng cách nhìn dấu ngoặc ở cuối dòng.
 */
export function parseLogOneline(stdout: string): TdLogEntry[] {
	const entries: TdLogEntry[] = [];
	for (const raw of stdout.split('\n')) {
		const line = raw.replace(/\r$/, '').trim();
		if (!line || line === 'Chưa có commit nào.') {
			continue;
		}
		const space = line.indexOf(' ');
		if (space <= 0) {
			continue;
		}
		const hash = line.slice(0, space);
		if (!/^[0-9a-f]{4,40}$/.test(hash)) {
			continue;
		}
		let rest = line.slice(space + 1).trim();
		let refs: string[] = [];
		if (rest.endsWith(')')) {
			const open = rest.lastIndexOf('(');
			if (open > 0) {
				refs = rest.slice(open + 1, -1).split(',').map(s => s.trim()).filter(s => !!s);
				rest = rest.slice(0, open).trim();
			}
		}
		entries.push({ hash, summary: rest, refs });
	}
	return entries;
}

/**
 * Đọc output của `td vcs branch -vv`.
 *
 * Dòng có dạng `* tên  mã-băm  tiêu đề  [nhánh theo dõi: đi trước x, đi sau y]`,
 * trong đó dấu `*` chỉ nhánh đang đứng và mọi thứ sau tên đều tuỳ chọn.
 */
export function parseBranches(stdout: string): TdBranch[] {
	const branches: TdBranch[] = [];
	// [tên-nhánh: đi trước N, đi sau M] ở cuối dòng khi nhánh có nhánh theo dõi.
	const tracking = /\[\s*([^:\]]+):\s*đi trước (\d+),\s*đi sau (\d+)\s*\]\s*$/;

	for (const raw of stdout.split('\n')) {
		const line = raw.replace(/\r$/, '');
		if (!line.trim()) {
			continue;
		}
		const current = line.startsWith('* ');
		if (!current && !/^\s\s\S/.test(line)) {
			// Cảnh báo của td khi không có nhánh nào khớp.
			continue;
		}

		let body = line.slice(2).trim();
		let upstream = '';
		let ahead = 0;
		let behind = 0;
		const found = tracking.exec(body);
		if (found) {
			upstream = found[1].trim();
			ahead = Number(found[2]);
			behind = Number(found[3]);
			body = body.slice(0, found.index).trim();
		}

		// Sau tên nhánh là mã băm ngắn rồi tới tiêu đề commit.
		const parts = /^(\S+)(?:\s+([0-9a-f]{4,40}))?(?:\s+(.*))?$/.exec(body);
		if (!parts || !parts[1]) {
			continue;
		}
		branches.push({
			name: parts[1],
			current,
			hash: parts[2] ?? '',
			subject: (parts[3] ?? '').trim(),
			upstream,
			ahead,
			behind
		});
	}
	return branches;
}

/**
 * Đọc output của `td vcs tag`.
 *
 * Dòng có dạng `ten  mã-băm`, kèm `(có chú thích) và nội dung` nếu là tag có
 * chú thích.
 */
export function parseTags(stdout: string): TdTag[] {
	const tags: TdTag[] = [];
	for (const raw of stdout.split('\n')) {
		const line = raw.replace(/\r$/, '').trim();
		if (!line) {
			continue;
		}
		const parts = line.split(/\s+/);
		const hash = parts[1] ?? '';
		if (!parts[0] || !/^[0-9a-f]{4,40}$/.test(hash)) {
			continue;
		}
		let message = '';
		const annotatedAt = line.indexOf('(có chú thích)');
		if (annotatedAt >= 0) {
			message = line.slice(annotatedAt + '(có chú thích)'.length).trim();
		}
		tags.push({ name: parts[0], hash, annotated: annotatedAt >= 0, message });
	}
	return tags;
}

/**
 * Đọc output của `td vcs stash --list`.
 *
 * Dòng có dạng `stash@{0}: On stash: mô tả`.
 */
export function parseStashes(stdout: string): TdStash[] {
	const stashes: TdStash[] = [];
	const pattern = /^stash@\{(\d+)\}:\s*(.*)$/;
	for (const raw of stdout.split('\n')) {
		const line = raw.replace(/\r$/, '').trim();
		if (!line) {
			continue;
		}
		const match = pattern.exec(line);
		if (!match) {
			continue;
		}
		let message = match[2];
		// td thêm tiền tố "On stash: " vào mọi mục nên bỏ đi cho gọn.
		message = message.replace(/^On stash:\s*/, '');
		stashes.push({ index: Number(match[1]), message });
	}
	return stashes;
}