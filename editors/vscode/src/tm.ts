import { ChildProcess, execFile } from 'node:child_process';
import { statSync } from 'node:fs';
import { homedir } from 'node:os';
import * as path from 'node:path';

import {
	parseBranches,
	parseChangedPaths,
	parseLogOneline,
	parsePatches,
	parseStashes,
	parseStatus,
	parseTags,
	TmBranch,
	TmLogEntry,
	TmPatch,
	TmStash,
	TmStatus,
	TmTag
} from './parse';

/**
 * Số dòng ngữ cảnh dùng khi xin diff.
 *
 * Khung so sánh của VS Code cần *nội dung đầy đủ* của cả hai phía rồi tự tính
 * vạch hiệu, nên phải xin diff ôm trọn tệp chứ không phải ba dòng quanh thay
 * đổi. Giá trị này lớn hơn mọi tệp thực tế nên hunk luôn bao trọn tệp.
 */
const FULL_CONTEXT = 1000000;

/** Thời gian chờ tối đa cho một lần gọi tm, tính bằng mili giây. */
const DEFAULT_TIMEOUT = 30000;

/** Tên thư mục dữ liệu của tm, dùng để nhận ra một kho khi dò thư mục. */
const DATA_DIR = '.tmx';

/** Kết quả một lần gọi tm: mã thoát cùng toàn bộ output. */
export interface TmResult {
	code: number;
	stdout: string;
	stderr: string;
}

/** Lỗi của một lần gọi tm thất bại, giữ nguyên thông điệp tiếng Việt của tm. */
export class TmError extends Error {
	constructor(message: string, readonly result: TmResult) {
		super(message);
		this.name = 'TmError';
	}
}

/** Cách gọi diff: so sánh vùng chuẩn bị với HEAD, hay với cây làm việc. */
export interface DiffOptions {
	staged?: boolean;
	revision?: string;
	paths?: string[];
}

/** Nội dung một tệp ở một điểm lịch sử, cùng lý do khi không đọc được. */
export interface TmFileContent {
	/** Nội dung nguyên văn, rỗng khi tệp không tồn tại ở điểm đó. */
	content: string;
	/** false khi tệp không có ở điểm đó. */
	found: boolean;
	/** true khi lệnh tm trên máy chưa có `vcs show-file`. */
	unsupported: boolean;
}

/**
 * Dấu nhận biết thông báo của tm khi lệnh hoặc cờ chưa có trên máy.
 *
 * Cần để phân biệt "tm trên máy cũ hơn tiện ích" với "tệp không có trong
 * commit", vì cả hai đều làm lệnh thoát với mã khác không. Cờ lạ thì tm báo
 * "unknown flag" chứ không nói tới lệnh con, nên phải nhận ra cả hai kiểu.
 */
const UNKNOWN_COMMAND = /không có lệnh nào tên|unknown flag|unknown shorthand flag/;

/**
 * Bọc lệnh `tm vcs`.
 *
 * Mọi việc đọc trạng thái và thay đổi kho đều đi qua đây, không có đường nào
 * khác chạm vào tệp `.tmx`. Nhờ vậy phần còn lại của extension chỉ cần quan tâm
 * tới dữ liệu mà tm trả về.
 */
export class Tm {
	private readonly commandLog: string[] = [];

	constructor(private readonly executable: string, private readonly report?: (line: string) => void) {}

	/**
	 * Kiểm tra lệnh tm có chạy được không.
	 *
	 * Dùng `version` vì lệnh đó không cần kho nào, chỉ cần chính tệp thực thi.
	 */
	async available(): Promise<boolean> {
		try {
			const result = await this.spawn(['version']);
			return result.code === 0;
		} catch {
			return false;
		}
	}

	/** Lấy đường dẫn tệp thực thi đang dùng. */
	get command(): string {
		return this.executable;
	}

	/** Ghi một dòng ghi chú vào kênh log mà không phải chạy lệnh nào. */
	note(line: string): void {
		this.report?.(line);
	}

	/** Những dòng đã chạy gần đây, phục vụ phần log của extension. */
	history(): string[] {
		return this.commandLog.slice(-200);
	}

	/**
	 * Chạy một lệnh vcs trong kho cho trước và trả về output thô.
	 *
	 * `-C` đặt ngay sau `vcs`, trước mọi đối số khác. Đặt cuối dòng lệnh thì
	 * không được: khi lệnh có danh sách tệp sau dấu `--`, `-C` sẽ bị tm hiểu
	 * là một đường dẫn nữa và lệnh chạy nhầm ở thư mục hiện tại.
	 */
	async run(root: string, args: string[]): Promise<TmResult> {
		const full = ['vcs', '-C', root, ...args];
		const line = [this.executable, ...full].join(' ');
		this.commandLog.push(line);

		const result = await this.spawn(full, root);
		this.report?.(result.code === 0 ? line : `${line}  →  lỗi ${result.code}`);
		if (result.code !== 0) {
			throw new TmError(firstLine(result.stderr) || firstLine(result.stdout) || 'tm chạy không thành công', result);
		}
		return result;
	}

	/** Chạy lệnh và trả về stdout, dùng cho lệnh chỉ in thông tin. */
	private async text(root: string, args: string[]): Promise<string> {
		const result = await this.run(root, args);
		return result.stdout;
	}

	private spawn(args: string[], cwd?: string): Promise<TmResult> {
		return new Promise<TmResult>((resolve, reject) => {
			const env = {
				...process.env,
				// tm tự tắt màu khi không có thiết bị đầu ra, nhưng đặt thẳng
				// thì chắc chắn không còn ký tự ANSI nào lẫn vào output.
				NO_COLOR: '1',
				TERM: 'dumb'
			};
			let child: ChildProcess;
			try {
				child = execFile(
					this.executable,
					args,
					{ cwd, env, maxBuffer: 128 * 1024 * 1024, timeout: DEFAULT_TIMEOUT },
					(error, stdout, stderr) => {
						if (error && typeof (error as NodeJS.ErrnoException).code === 'string' && (error as NodeJS.ErrnoException).code === 'ENOENT') {
							reject(new Error(`không chạy được "${this.executable}", hãy cài tm hoặc đặt tm.path`));
							return;
						}
						const code = error ? (typeof error.code === 'number' ? error.code : 1) : 0;
						resolve({ code, stdout: stdout ?? '', stderr: stderr ?? '' });
					}
				);
			} catch (error) {
				reject(error);
				return;
			}
			child.on('error', reject);
		});
	}

	/** `tm vcs status` */
	async status(root: string): Promise<TmStatus> {
		return parseStatus(await this.text(root, ['status']));
	}

	/** `tm vcs diff`, có lọc theo danh sách tệp nếu có. */
	async diff(root: string, options: DiffOptions = {}): Promise<TmPatch[]> {
		const args = ['diff'];
		if (options.staged) {
			args.push('--staged');
		}
		if (options.revision) {
			args.push(options.revision);
		}
		args.push(`-U${FULL_CONTEXT}`);
		if (options.paths && options.paths.length > 0) {
			args.push('--', ...options.paths);
		}
		return parsePatches(await this.text(root, args));
	}

	/**
	 * Chỉ xin danh sách tệp thay đổi, không lấy nội dung.
	 *
	 * Dùng để mở khung so sánh nhiều tệp: liệt kê tệp rẻ hơn nhiều so với xin
	 * khác biệt đầy đủ, còn nội dung từng phía thì đợi khung so sánh thật sự cần
	 * mới đọc. Nhờ vậy một commit sửa hàng trăm tệp vẫn mở tức thì.
	 */
	async changedFiles(root: string, options: DiffOptions = {}): Promise<string[]> {
		const args = ['diff', '--name-only'];
		if (options.staged) {
			args.push('--staged');
		}
		if (options.revision) {
			args.push(options.revision);
		}
		if (options.paths && options.paths.length > 0) {
			args.push('--', ...options.paths);
		}
		return parseChangedPaths(await this.text(root, args));
	}

	/**
	 * `tm vcs show-file <điểm> -- <tệp>`: nội dung tệp ở một điểm lịch sử.
	 *
	 * Phía đối diện của một khung so sánh cần nội dung rỗng đúng lúc tệp không
	 * tồn tại ở điểm đó, nên "không có" phải là một kết quả chứ không phải lỗi.
	 * Trường `unsupported` báo lệnh tm trên máy còn cũ hơn tiện ích, lúc đó
	 * người gọi nên rơi về cách dựng nội dung từ khác biệt.
	 */
	async showFile(root: string, revision: string, relativePath: string): Promise<TmFileContent> {
		return this.readFileAt(root, ['show-file', revision, '--', relativePath], relativePath);
	}

	/**
	 * `tm vcs show-file --index -- <tệp>`: nội dung tệp trong vùng chuẩn bị.
	 *
	 * Phía gốc của khung so sánh thay đổi trên đĩa là nội dung đã stage, và nó
	 * phải đọc thẳng từ vùng chuẩn bị chứ không dựng lại từ khác biệt: người
	 * dùng hoàn tác hết thay đổi thì tệp khớp vùng chuẩn bị, không còn khác
	 * biệt nào để dựng, và phía gốc sẽ rơi về chuỗi rỗng trong khi phía phải
	 * vẫn có nội dung.
	 */
	async indexFile(root: string, relativePath: string): Promise<TmFileContent> {
		return this.readFileAt(root, ['show-file', '--index', '--', relativePath], relativePath);
	}

	/**
	 * Chạy một lệnh in nội dung tệp rồi chuẩn hoá kết quả về cùng một kiểu.
	 *
	 * Cần phân biệt "tệp không có ở điểm đó" với "lệnh tm trên máy cũ hơn tiện
	 * ích": cả hai đều làm lệnh thoát với mã khác không, nhưng chỉ trường hợp
	 * sau mới cần báo cho người dùng nâng cấp tm.
	 */
	private async readFileAt(root: string, args: string[], relativePath: string): Promise<TmFileContent> {
		try {
			const result = await this.run(root, args);
			return { content: result.stdout, found: true, unsupported: false };
		} catch (error) {
			if (error instanceof TmError && UNKNOWN_COMMAND.test(error.result.stderr)) {
				return { content: '', found: false, unsupported: true };
			}
			// Tệp không có ở điểm đó là chuyện thường nên không báo, còn lỗi thật
			// thì ghi ra kênh log để không biến thành phía rỗng trong khung so sánh.
			if (error instanceof TmError && !/không có trong/.test(error.message)) {
				this.report?.(`đọc ${relativePath} không được: ${error.message}`);
			}
			return { content: '', found: false, unsupported: false };
		}
	}

	/** `tm vcs log --oneline`, có giới hạn số dòng theo cấu hình. */
	async log(root: string, maxCount: number, revision?: string, paths?: string[]): Promise<TmLogEntry[]> {
		const args = ['log', '--oneline', `-n${maxCount}`];
		if (revision) {
			args.push(revision);
		}
		if (paths && paths.length > 0) {
			args.push('--', ...paths);
		}
		return parseLogOneline(await this.text(root, args));
	}

	/** `tm vcs branch -vv` */
	async branches(root: string): Promise<TmBranch[]> {
		return parseBranches(await this.text(root, ['branch', '-vv']));
	}

	/** `tm vcs tag` */
	async tags(root: string): Promise<TmTag[]> {
		return parseTags(await this.text(root, ['tag', '-l']));
	}

	/** `tm vcs stash --list` */
	async stashes(root: string): Promise<TmStash[]> {
		return parseStashes(await this.text(root, ['stash', '--list']));
	}

	/** `tm vcs add <tệp...>` */
	async stage(root: string, paths: string[]): Promise<void> {
		await this.run(root, ['add', ...paths]);
	}

	/** `tm vcs restore --staged <tệp...>`: gỡ thay đổi đã stage về HEAD. */
	async unstage(root: string, paths: string[]): Promise<void> {
		await this.run(root, ['restore', '--staged', ...paths]);
	}

	/** `tm vcs restore <tệp...>`: huỷ sửa đổi trên đĩa, lấy lại nội dung đã stage. */
	async discard(root: string, paths: string[]): Promise<void> {
		await this.run(root, ['restore', ...paths]);
	}

	/** `tm vcs commit` */
	async commit(root: string, message: string, amend = false): Promise<void> {
		const args = ['commit', '-m', message];
		if (amend) {
			args.push('--amend');
		}
		await this.run(root, args);
	}

	/** `tm vcs commit --allow-empty` cho trường hợp muốn tạo mốc rỗng. */
	async commitEmpty(root: string, message: string): Promise<void> {
		await this.run(root, ['commit', '-m', message, '--allow-empty']);
	}

	/** `tm vcs add --update .` rồi commit, dùng cho thao tác commit tất cả. */
	async stageAll(root: string): Promise<void> {
		await this.run(root, ['add', '-A', '.']);
	}

	/** `tm vcs add --update .`: chỉ cập nhật tệp đã được theo dõi. */
	async stageAllTracked(root: string): Promise<void> {
		await this.run(root, ['add', '--update', '.']);
	}

	/** `tm vcs switch` */
	async switchTo(root: string, ref: string): Promise<void> {
		await this.run(root, ['switch', ref]);
	}

	/** `tm vcs switch --detach` cho một mã băm. */
	async detach(root: string, hash: string): Promise<void> {
		await this.run(root, ['checkout', '--detach', hash]);
	}

	/** `tm vcs branch <tên>`: tạo nhánh tại HEAD rồi chuyển sang đó. */
	async createBranch(root: string, name: string, from?: string): Promise<void> {
		if (from) {
			await this.run(root, ['branch', name, from]);
			return;
		}
		await this.run(root, ['switch', '-c', name]);
	}

	/** `tm vcs branch -d` hoặc -D */
	async deleteBranch(root: string, name: string, force = false): Promise<void> {
		await this.run(root, ['branch', force ? '-D' : '-d', name]);
	}

	/** `tm vcs branch -m <cũ> <mới>` */
	async renameBranch(root: string, from: string, to: string): Promise<void> {
		await this.run(root, ['branch', '-m', from, to]);
	}

	/** `tm vcs merge` */
	async merge(root: string, ref: string): Promise<void> {
		await this.run(root, ['merge', ref]);
	}

	/** `tm vcs rebase` */
	async rebase(root: string, ref: string): Promise<void> {
		await this.run(root, ['rebase', ref]);
	}

	/** `tm vcs cherry-pick` */
	async cherryPick(root: string, hash: string): Promise<void> {
		await this.run(root, ['cherry-pick', hash]);
	}

	/** `tm vcs revert` */
	async revert(root: string, hash: string): Promise<void> {
		await this.run(root, ['revert', hash]);
	}

	/** `tm vcs stash` */
	async stash(root: string, message?: string, includeUntracked = false): Promise<void> {
		const args = ['stash'];
		if (includeUntracked) {
			args.push('-u');
		}
		if (message) {
			args.push('-m', message);
		}
		await this.run(root, args);
	}

	/** `tm vcs stash --apply` hoặc --pop, không truyền số thì áp dụng mục mới nhất. */
	async applyStash(root: string, index: number | undefined, pop: boolean): Promise<void> {
		const args = ['stash', pop ? '--pop' : '--apply'];
		if (index !== undefined) {
			args.push(String(index));
		}
		await this.run(root, args);
	}

	/** `tm vcs stash --drop` */
	async dropStash(root: string, index: number | undefined): Promise<void> {
		const args = ['stash', '--drop'];
		if (index !== undefined) {
			args.push(String(index));
		}
		await this.run(root, args);
	}

	/** `tm vcs stash --clear` */
	async clearStashes(root: string): Promise<void> {
		await this.run(root, ['stash', '--clear']);
	}

	/** `tm vcs tag` */
	async createTag(root: string, name: string, message?: string): Promise<void> {
		const args = ['tag'];
		if (message) {
			args.push('-a', '-m', message);
		}
		args.push(name);
		await this.run(root, args);
	}

	/** `tm vcs tag -d` */
	async deleteTag(root: string, name: string): Promise<void> {
		await this.run(root, ['tag', '-d', name]);
	}

	/**
	 * `tm vcs clean -f <tệp...>`: xoá tệp chưa được tm theo dõi.
	 *
	 * Không có đối số nghĩa là xoá hết, nên lệnh gọi từ giao diện luôn truyền
	 * đúng danh sách tệp đang chọn.
	 */
	async clean(root: string, paths: string[]): Promise<void> {
		await this.run(root, ['clean', '-f', ...paths]);
	}

	/** `tm vcs fsck`: kiểm tra tính toàn vẹn của kho. */
	async fsck(root: string): Promise<string> {
		return this.text(root, ['fsck']);
	}

	/** `tm vcs init` */
	async init(root: string, branch?: string): Promise<void> {
		const args = ['init'];
		if (branch) {
			args.push('--initial-branch', branch);
		}
		// init chưa có kho nên không cần chạy trong kho, chỉ cần trỏ đúng thư mục.
		await this.spawn(['vcs', '-C', root, ...args]);
	}
}

/** Lấy dòng đầu tiên của một khối output, bỏ khoảng trắng thừa. */
function firstLine(text: string): string {
	for (const line of text.split('\n')) {
		const trimmed = line.trim();
		if (trimmed) {
			return trimmed;
		}
	}
	return '';
}

/**
 * Tìm thư mục gốc của kho tm bằng cách đi lên từ một đường dẫn cho tới khi
 * gặp thư mục chứa `.tmx`.
 *
 * Cách này nhanh hơn gọi `tm vcs status` mỗi lần thử, và cho phép bật trạng
 * thái Source Control ngay khi mở cửa sổ thay vì phải chờ một lệnh chạy xong.
 */
export function findRepositoryRoot(start: string): string | undefined {
	let dir = start;
	for (;;) {
		if (isRepositoryRoot(dir)) {
			return dir;
		}
		const parent = path.dirname(dir);
		if (parent === dir) {
			return undefined;
		}
		dir = parent;
	}
}

/**
 * Tìm tệp thực thi tm trên máy người dùng.
 *
 * Thứ tự thử: đường dẫn người dùng đặt trong cấu hình, lệnh `tm` trong PATH,
 * thư mục bin mặc định của Go trên cả ba nền tảng, rồi tới thư mục `out/` của
 * chính kho này. Nhờ vậy cài extension lên máy mới chỉ cần có Go là chạy được,
 * không phải cài tay.
 */
export function resolveExecutable(configured: string, extraDirs: string[] = []): string {
	if (configured) {
		return configured;
	}
	const home = homedir();
	const candidates = [
		'tm',
		...extraDirs.map(dir => path.join(dir, 'tm')),
		...extraDirs.map(dir => path.join(dir, 'tm.exe')),
		path.join(home, 'go', 'bin', 'tm'),
		path.join(home, 'go', 'bin', 'tm.exe'),
		path.join(home, 'bin', 'tm'),
		path.join(home, 'bin', 'tm.exe'),
		'/usr/local/bin/tm',
		'/opt/homebrew/bin/tm'
	];
	for (const candidate of candidates) {
		if (candidate === 'tm' || isRunnable(candidate)) {
			return candidate;
		}
	}
	return 'tm';
}

/** Tệp có chạy được không: tồn tại và không phải thư mục. */
function isRunnable(candidate: string): boolean {
	try {
		return statSync(candidate).isFile();
	} catch {
		return false;
	}
}

/** Thư mục có chứa `.tmx` hay không. */
export function isRepositoryRoot(dir: string): boolean {
	try {
		return statSync(path.join(dir, DATA_DIR)).isDirectory();
	} catch {
		return false;
	}
}