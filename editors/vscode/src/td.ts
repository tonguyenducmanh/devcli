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
	TdBranch,
	TdLogEntry,
	TdPatch,
	TdStash,
	TdStatus,
	TdTag
} from './parse';

/**
 * Số dòng ngữ cảnh dùng khi xin diff.
 *
 * Khung so sánh của VS Code cần *nội dung đầy đủ* của cả hai phía rồi tự tính
 * vạch hiệu, nên phải xin diff ôm trọn tệp chứ không phải ba dòng quanh thay
 * đổi. Giá trị này lớn hơn mọi tệp thực tế nên hunk luôn bao trọn tệp.
 */
const FULL_CONTEXT = 1000000;

/** Thời gian chờ tối đa cho một lần gọi td, tính bằng mili giây. */
const DEFAULT_TIMEOUT = 30000;

/** Tên thư mục dữ liệu của td, dùng để nhận ra một kho khi dò thư mục. */
const DATA_DIR = '.tdx';

/** Kết quả một lần gọi td: mã thoát cùng toàn bộ output. */
export interface TdResult {
	code: number;
	stdout: string;
	stderr: string;
}

/** Lỗi của một lần gọi td thất bại, giữ nguyên thông điệp tiếng Việt của td. */
export class TdError extends Error {
	constructor(message: string, readonly result: TdResult) {
		super(message);
		this.name = 'TdError';
	}
}

/** Cách gọi diff: so sánh vùng chuẩn bị với HEAD, hay với cây làm việc. */
export interface DiffOptions {
	staged?: boolean;
	revision?: string;
	paths?: string[];
}

/** Nội dung một tệp ở một điểm lịch sử, cùng lý do khi không đọc được. */
export interface TdFileContent {
	/** Nội dung nguyên văn, rỗng khi tệp không tồn tại ở điểm đó. */
	content: string;
	/** false khi tệp không có ở điểm đó. */
	found: boolean;
	/** true khi lệnh td trên máy chưa có `vcs show-file`. */
	unsupported: boolean;
}

/**
 * Dấu nhận biết thông báo của td khi không có lệnh con tương ứng.
 *
 * Cần để phân biệt "td trên máy cũ hơn tiện ích" với "tệp không có trong
 * commit", vì cả hai đều làm lệnh thoát với mã khác không.
 */
const UNKNOWN_COMMAND = /không có lệnh nào tên/;

/**
 * Bọc lệnh `td vcs`.
 *
 * Mọi việc đọc trạng thái và thay đổi kho đều đi qua đây, không có đường nào
 * khác chạm vào tệp `.tdx`. Nhờ vậy phần còn lại của extension chỉ cần quan tâm
 * tới dữ liệu mà td trả về.
 */
export class Td {
	private readonly commandLog: string[] = [];

	constructor(private readonly executable: string, private readonly report?: (line: string) => void) {}

	/**
	 * Kiểm tra lệnh td có chạy được không.
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
	 * không được: khi lệnh có danh sách tệp sau dấu `--`, `-C` sẽ bị td hiểu
	 * là một đường dẫn nữa và lệnh chạy nhầm ở thư mục hiện tại.
	 */
	async run(root: string, args: string[]): Promise<TdResult> {
		const full = ['vcs', '-C', root, ...args];
		const line = [this.executable, ...full].join(' ');
		this.commandLog.push(line);

		const result = await this.spawn(full, root);
		this.report?.(result.code === 0 ? line : `${line}  →  lỗi ${result.code}`);
		if (result.code !== 0) {
			throw new TdError(firstLine(result.stderr) || firstLine(result.stdout) || 'td chạy không thành công', result);
		}
		return result;
	}

	/** Chạy lệnh và trả về stdout, dùng cho lệnh chỉ in thông tin. */
	private async text(root: string, args: string[]): Promise<string> {
		const result = await this.run(root, args);
		return result.stdout;
	}

	private spawn(args: string[], cwd?: string): Promise<TdResult> {
		return new Promise<TdResult>((resolve, reject) => {
			const env = {
				...process.env,
				// td tự tắt màu khi không có thiết bị đầu ra, nhưng đặt thẳng
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
							reject(new Error(`không chạy được "${this.executable}", hãy cài td hoặc đặt td.path`));
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

	/** `td vcs status` */
	async status(root: string): Promise<TdStatus> {
		return parseStatus(await this.text(root, ['status']));
	}

	/** `td vcs diff`, có lọc theo danh sách tệp nếu có. */
	async diff(root: string, options: DiffOptions = {}): Promise<TdPatch[]> {
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
	 * `td vcs show-file <điểm> -- <tệp>`: nội dung tệp ở một điểm lịch sử.
	 *
	 * Phía đối diện của một khung so sánh cần nội dung rỗng đúng lúc tệp không
	 * tồn tại ở điểm đó, nên "không có" phải là một kết quả chứ không phải lỗi.
	 * Trường `unsupported` báo lệnh td trên máy còn cũ hơn tiện ích, lúc đó
	 * người gọi nên rơi về cách dựng nội dung từ khác biệt.
	 */
	async showFile(root: string, revision: string, relativePath: string): Promise<TdFileContent> {
		try {
			const result = await this.run(root, ['show-file', revision, '--', relativePath]);
			return { content: result.stdout, found: true, unsupported: false };
		} catch (error) {
			if (error instanceof TdError && UNKNOWN_COMMAND.test(error.result.stderr)) {
				return { content: '', found: false, unsupported: true };
			}
			// Tệp không có ở điểm đó là chuyện thường nên không báo, còn lỗi thật
			// thì ghi ra kênh log để không biến thành phía rỗng trong khung so sánh.
			if (error instanceof TdError && !/không có trong/.test(error.message)) {
				this.report?.(`đọc ${relativePath} ở ${revision} không được: ${error.message}`);
			}
			return { content: '', found: false, unsupported: false };
		}
	}

	/** `td vcs log --oneline`, có giới hạn số dòng theo cấu hình. */
	async log(root: string, maxCount: number, revision?: string, paths?: string[]): Promise<TdLogEntry[]> {
		const args = ['log', '--oneline', `-n${maxCount}`];
		if (revision) {
			args.push(revision);
		}
		if (paths && paths.length > 0) {
			args.push('--', ...paths);
		}
		return parseLogOneline(await this.text(root, args));
	}

	/** `td vcs branch -vv` */
	async branches(root: string): Promise<TdBranch[]> {
		return parseBranches(await this.text(root, ['branch', '-vv']));
	}

	/** `td vcs tag` */
	async tags(root: string): Promise<TdTag[]> {
		return parseTags(await this.text(root, ['tag', '-l']));
	}

	/** `td vcs stash --list` */
	async stashes(root: string): Promise<TdStash[]> {
		return parseStashes(await this.text(root, ['stash', '--list']));
	}

	/** `td vcs add <tệp...>` */
	async stage(root: string, paths: string[]): Promise<void> {
		await this.run(root, ['add', ...paths]);
	}

	/** `td vcs restore --staged <tệp...>`: gỡ thay đổi đã stage về HEAD. */
	async unstage(root: string, paths: string[]): Promise<void> {
		await this.run(root, ['restore', '--staged', ...paths]);
	}

	/** `td vcs restore <tệp...>`: huỷ sửa đổi trên đĩa, lấy lại nội dung đã stage. */
	async discard(root: string, paths: string[]): Promise<void> {
		await this.run(root, ['restore', ...paths]);
	}

	/** `td vcs commit` */
	async commit(root: string, message: string, amend = false): Promise<void> {
		const args = ['commit', '-m', message];
		if (amend) {
			args.push('--amend');
		}
		await this.run(root, args);
	}

	/** `td vcs commit --allow-empty` cho trường hợp muốn tạo mốc rỗng. */
	async commitEmpty(root: string, message: string): Promise<void> {
		await this.run(root, ['commit', '-m', message, '--allow-empty']);
	}

	/** `td vcs add --update .` rồi commit, dùng cho thao tác commit tất cả. */
	async stageAll(root: string): Promise<void> {
		await this.run(root, ['add', '-A', '.']);
	}

	/** `td vcs add --update .`: chỉ cập nhật tệp đã được theo dõi. */
	async stageAllTracked(root: string): Promise<void> {
		await this.run(root, ['add', '--update', '.']);
	}

	/** `td vcs switch` */
	async switchTo(root: string, ref: string): Promise<void> {
		await this.run(root, ['switch', ref]);
	}

	/** `td vcs switch --detach` cho một mã băm. */
	async detach(root: string, hash: string): Promise<void> {
		await this.run(root, ['checkout', '--detach', hash]);
	}

	/** `td vcs branch <tên>`: tạo nhánh tại HEAD rồi chuyển sang đó. */
	async createBranch(root: string, name: string, from?: string): Promise<void> {
		if (from) {
			await this.run(root, ['branch', name, from]);
			return;
		}
		await this.run(root, ['switch', '-c', name]);
	}

	/** `td vcs branch -d` hoặc -D */
	async deleteBranch(root: string, name: string, force = false): Promise<void> {
		await this.run(root, ['branch', force ? '-D' : '-d', name]);
	}

	/** `td vcs branch -m <cũ> <mới>` */
	async renameBranch(root: string, from: string, to: string): Promise<void> {
		await this.run(root, ['branch', '-m', from, to]);
	}

	/** `td vcs merge` */
	async merge(root: string, ref: string): Promise<void> {
		await this.run(root, ['merge', ref]);
	}

	/** `td vcs rebase` */
	async rebase(root: string, ref: string): Promise<void> {
		await this.run(root, ['rebase', ref]);
	}

	/** `td vcs cherry-pick` */
	async cherryPick(root: string, hash: string): Promise<void> {
		await this.run(root, ['cherry-pick', hash]);
	}

	/** `td vcs revert` */
	async revert(root: string, hash: string): Promise<void> {
		await this.run(root, ['revert', hash]);
	}

	/** `td vcs stash` */
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

	/** `td vcs stash --apply` hoặc --pop, không truyền số thì áp dụng mục mới nhất. */
	async applyStash(root: string, index: number | undefined, pop: boolean): Promise<void> {
		const args = ['stash', pop ? '--pop' : '--apply'];
		if (index !== undefined) {
			args.push(String(index));
		}
		await this.run(root, args);
	}

	/** `td vcs stash --drop` */
	async dropStash(root: string, index: number | undefined): Promise<void> {
		const args = ['stash', '--drop'];
		if (index !== undefined) {
			args.push(String(index));
		}
		await this.run(root, args);
	}

	/** `td vcs stash --clear` */
	async clearStashes(root: string): Promise<void> {
		await this.run(root, ['stash', '--clear']);
	}

	/** `td vcs tag` */
	async createTag(root: string, name: string, message?: string): Promise<void> {
		const args = ['tag'];
		if (message) {
			args.push('-a', '-m', message);
		}
		args.push(name);
		await this.run(root, args);
	}

	/** `td vcs tag -d` */
	async deleteTag(root: string, name: string): Promise<void> {
		await this.run(root, ['tag', '-d', name]);
	}

	/**
	 * `td vcs clean -f <tệp...>`: xoá tệp chưa được td theo dõi.
	 *
	 * Không có đối số nghĩa là xoá hết, nên lệnh gọi từ giao diện luôn truyền
	 * đúng danh sách tệp đang chọn.
	 */
	async clean(root: string, paths: string[]): Promise<void> {
		await this.run(root, ['clean', '-f', ...paths]);
	}

	/** `td vcs fsck`: kiểm tra tính toàn vẹn của kho. */
	async fsck(root: string): Promise<string> {
		return this.text(root, ['fsck']);
	}

	/** `td vcs init` */
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
 * Tìm thư mục gốc của kho td bằng cách đi lên từ một đường dẫn cho tới khi
 * gặp thư mục chứa `.tdx`.
 *
 * Cách này nhanh hơn gọi `td vcs status` mỗi lần thử, và cho phép bật trạng
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
 * Tìm tệp thực thi td trên máy người dùng.
 *
 * Thứ tự thử: đường dẫn người dùng đặt trong cấu hình, lệnh `td` trong PATH,
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
		'td',
		...extraDirs.map(dir => path.join(dir, 'td')),
		...extraDirs.map(dir => path.join(dir, 'td.exe')),
		path.join(home, 'go', 'bin', 'td'),
		path.join(home, 'go', 'bin', 'td.exe'),
		path.join(home, 'bin', 'td'),
		path.join(home, 'bin', 'td.exe'),
		'/usr/local/bin/td',
		'/opt/homebrew/bin/td'
	];
	for (const candidate of candidates) {
		if (candidate === 'td' || isRunnable(candidate)) {
			return candidate;
		}
	}
	return 'td';
}

/** Tệp có chạy được không: tồn tại và không phải thư mục. */
function isRunnable(candidate: string): boolean {
	try {
		return statSync(candidate).isFile();
	} catch {
		return false;
	}
}

/** Thư mục có chứa `.tdx` hay không. */
export function isRepositoryRoot(dir: string): boolean {
	try {
		return statSync(path.join(dir, DATA_DIR)).isDirectory();
	} catch {
		return false;
	}
}