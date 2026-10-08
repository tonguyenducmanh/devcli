import { execFileSync } from 'node:child_process';
import { existsSync, mkdirSync, mkdtempSync, readdirSync, readFileSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import * as path from 'node:path';
import Module = require('node:module');

/** Gốc kho của tm, thư mục cha của thư mục extension. */
const REPO_ROOT = path.join(__dirname, '..', '..', '..', '..');

/** Một lần gọi lệnh của VS Code mà bản giả ghi lại. */
export interface RecordedCall {
	id: string;
	args: unknown[];
}

export interface FakeState {
	resourceGroups: { id: string; label: string; resourceStates: unknown[]; hideWhenEmpty?: boolean }[];
	sourceControls: Record<string, unknown>[];
	registeredCommands: Map<string, (...args: unknown[]) => unknown>;
	/** Các lệnh VS Code đã được yêu cầu chạy, kèm đúng đối số. */
	calls: RecordedCall[];
	/** Nội dung các tài liệu đã mở, theo thứ tự mở. */
	opened: unknown[];
	shown: string[];
	uri(fsPath: string): unknown;
	/** Lần gọi cuối cùng của một lệnh, hoặc undefined nếu chưa gọi. */
	lastCall(id: string): RecordedCall | undefined;
	/** Nội dung ảo mà tiện ích cấp cho các địa chỉ tm:, đọc để kiểm thử. */
	contentOf(uri: unknown): Promise<string>;
	/**
	 * Câu trả lời cho các hộp thoại, theo thứ tự được dùng.
	 *
	 * Lệnh như tạo nhánh hay lưu tạm đều hỏi người dùng qua hộp nhập hoặc hộp
	 * chọn nhanh, mà bản giả không có ai để trả lời. Kiểm thử xếp trước câu
	 * trả lời vào đây rồi gọi lệnh; hết hàng thì hộp thoại bị huỷ, đúng như
	 * người dùng bấm Esc.
	 */
	answers: DialogAnswers;
	/** Sự kiện tệp trong workspace, để kiểm thử được phần làm mới trạng thái. */
	fireFileEvent(kind: 'create' | 'change' | 'delete', fsPath: string): void;
	/** Chờ cho các lệnh đã hứa hoàn tất, đúng như người dùng thấy. */
	settle(ms?: number): Promise<void>;
}

/** Câu trả lời dồn theo thứ tự cho từng loại hộp thoại. */
export interface DialogAnswers {
	input: (string | undefined)[];
	pick: unknown[];
	warning: (string | undefined)[];
	information: (string | undefined)[];
}

/** Lấy câu trả lời kế tiếp, hết hàng thì huỷ hộp thoại. */
function nextAnswer<T>(queue: T[]): T | undefined {
	return queue.shift();
}

export function findTm(): string | undefined {
	const fromEnv = process.env.TM_BIN;
	if (fromEnv && existsSync(fromEnv)) {
		return fromEnv;
	}
	try {
		// Tên tệp do build_binaries.sh đặt theo biến APP_NAME, nên thử cả
		// những tiền tố của các đời đặt tên trước cho chắc kho còn tệp build cũ.
		const built = readdirSync(path.join(REPO_ROOT, 'out'))
			.filter(name => /^(devcli-tm|td-devcli|devcli|tm)(-|\.)/.test(name))
			// Tệp đúng tên ứng dụng đứng trước, vì tệp theo nền tảng có thể là
			// bản build cho máy khác nằm cùng thư mục.
			.sort((a, b) => rankBinary(a) - rankBinary(b))
			.map(name => path.join(REPO_ROOT, 'out', name))
			// Chạy thử mới biết tệp nào dùng được: `out` chứa cả ba nền tảng
			// mà tệp của nền tảng khác sẽ báo lỗi khó hiểu như ENOEXEC.
			.find(candidate => canRun(candidate));
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

/** Tệp thực thi có chạy được trên máy này không. */
function canRun(candidate: string): boolean {
	try {
		execFileSync(candidate, ['version'], { stdio: 'ignore' });
		return true;
	} catch {
		return false;
	}
}

/** Thứ tự ưu tiên khi có nhiều tệp cùng nằm trong out. */
function rankBinary(name: string): number {
	const base = path.basename(name);
	return base === 'tm' || base === 'tm.exe' ? 0 : 1;
}

export class Disposable {
	constructor(private readonly callOnDispose: () => void) {}
	dispose(): void {
		this.callOnDispose();
	}
}

export class EventEmitter<T> {
	private listeners: ((value: T) => void)[] = [];
	readonly event = (listener: (value: T) => void): Disposable => {
		this.listeners.push(listener);
		return new Disposable(() => {
			this.listeners = this.listeners.filter(l => l !== listener);
		});
	};
	fire(value: T): void {
		for (const listener of [...this.listeners]) {
			listener(value);
		}
	}
	dispose(): void {
		this.listeners = [];
	}
}

/**
 * Địa chỉ giả, phải là một lớp thật.
 *
 * Mã nguồn kiểm tra địa chỉ bằng `instanceof Uri`, nên bản giả phải là lớp thì
 * mới cho ra kết quả giống hệt như trong VS Code.
 */
export class FakeUri {
	constructor(
		readonly scheme: string,
		readonly path: string,
		readonly query = ''
	) {}

	get fsPath(): string {
		return this.path;
	}

	toString(): string {
		return `${this.scheme}://${this.path}${this.query ? `?${this.query}` : ''}`;
	}

	static file(fsPath: string): FakeUri {
		return new FakeUri('file', fsPath);
	}

	static parse(value: string): FakeUri {
		const at = value.indexOf('://');
		return at < 0
			? new FakeUri('file', value)
			: new FakeUri(value.slice(0, at), value.slice(at + 3));
	}

	static from(parts: { scheme: string; path: string; query?: string }): FakeUri {
		return new FakeUri(parts.scheme, parts.path, parts.query);
	}
}

/**
 * Một bản sao đủ dùng của API VS Code cho kiểm thử.
 *
 * Chỉ những phần tiện ích thật sự gọi tới mới có mặt: SourceControl, cây thư
 * mục, sự kiện cấu hình, ô nhập nội dung, và lệnh chạy được ghi lại để soi
 * tham số truyền vào. Mục đích là bắt được lỗi chạy thật trong lúc kích hoạt
 * mà không cần mở VS Code.
 */
export function fakeVscode(tmPath: string, settings: Record<string, unknown> = {}): { vscode: Record<string, unknown>; state: FakeState } {
	const uri = (fsPath: string): FakeUri => FakeUri.file(fsPath);
	const resourceGroups: { id: string; label: string; resourceStates: unknown[]; hideWhenEmpty?: boolean }[] = [];
	const sourceControls: Record<string, unknown>[] = [];
	const registeredCommands = new Map<string, (...args: unknown[]) => unknown>();
	const calls: RecordedCall[] = [];
	const opened: unknown[] = [];
	const shown: string[] = [];
	const answers: DialogAnswers = { input: [], pick: [], warning: [], information: [] };

	// Mỗi bộ theo dõi tệp nhớ listener để kiểm thử bắn sự kiện như lúc chạy thật.
	type WatcherListener = (uri: FakeUri) => void;
	const watcherListeners = new Map<string, WatcherListener[]>();
	const watcherOf = (kind: string) => {
		const on = (event: string) => (fn: WatcherListener) => {
			const list = watcherListeners.get(event) ?? [];
			list.push(fn);
			watcherListeners.set(event, list);
			return new Disposable(() => {
				watcherListeners.set(event, (watcherListeners.get(event) ?? []).filter(l => l !== fn));
			});
		};
		return {
			onDidCreate: on(`${kind}:create`),
			onDidChange: on(`${kind}:change`),
			onDidDelete: on(`${kind}:delete`),
			dispose: () => { }
		};
	};

	const vscode = {
		Uri: FakeUri,
		EventEmitter,
		Disposable,
		ThemeIcon: class { constructor(public id: string) { } },
		ThemeColor: class { constructor(public id: string) { } },
		TreeItem: class {
			label?: string;
			description?: string;
			tooltip?: string;
			iconPath?: unknown;
			contextValue?: string;
			id?: string;
			resourceUri?: unknown;
			command?: unknown;
			constructor(label?: string) {
				this.label = label;
			}
		},
		scm: {
			createSourceControl: (id: string, label: string, rootUri: unknown) => {
				const groups = resourceGroups;
				const source = {
					id,
					label,
					rootUri,
					count: 0,
					inputBox: { value: '', placeholder: '', visible: true, enabled: true },
					acceptInputCommand: undefined as unknown,
					quickDiffProvider: undefined as unknown,
					statusBarCommands: [] as unknown[],
					createResourceGroup: (groupId: string, groupLabel: string) => {
						// Nhóm trong API thật cũng có dispose.
						const group = { id: groupId, label: groupLabel, resourceStates: [] as unknown[], dispose: () => { } };
						groups.push(group);
						return group;
					},
					dispose: () => { }
				};
				sourceControls.push(source);
				return source;
			}
		},
		workspace: {
			workspaceFolders: undefined as unknown,
			// tm.path có giá trị thật, cấu hình khác lấy từ tham số settings,
			// không có thì về mặc định.
			getConfiguration: () => ({
				get: <T>(key: string, fallback?: T): T => {
					if (key === 'path') {
						return tmPath as unknown as T;
					}
					return key in settings ? (settings[key] as T) : (fallback as T);
				}
			}),
			onDidChangeConfiguration: () => new Disposable(() => { }),
			onDidChangeWorkspaceFolders: () => new Disposable(() => { }),
			createFileSystemWatcher: (pattern: string) => watcherOf(pattern),
			registerTextDocumentContentProvider: (scheme: string, provider: ContentProvider) => {
				contentProviders.set(scheme, provider);
				return new Disposable(() => contentProviders.delete(scheme));
			},
			fs: {
				readFile: async (target: { fsPath: string }) => readFileSync(target.fsPath)
			}
		},
		window: {
			createOutputChannel: () => ({ appendLine: () => { }, show: () => { }, dispose: () => { } }),
			registerTreeDataProvider: () => new Disposable(() => { }),
			registerFileDecorationProvider: () => new Disposable(() => { }),
			showErrorMessage: (message: string) => {
				shown.push(message);
				return Promise.resolve(undefined);
			},
			showWarningMessage: (message: string, ...rest: unknown[]) => {
				shown.push(message);
				// Hộp xác nhận có nút bấm nên trả về nút mà answers.warning chỉ
				// định. Hộp chỉ báo không có nút nào thì luôn trả undefined.
				const hasButtons = rest.some(item => typeof item === 'string');
				return Promise.resolve(hasButtons ? nextAnswer(answers.warning) : undefined);
			},
			showInformationMessage: (message: string) => {
				shown.push(message);
				nextAnswer(answers.information);
				return Promise.resolve(undefined);
			},
			showInputBox: () => Promise.resolve(nextAnswer(answers.input)),
			showQuickPick: () => Promise.resolve(nextAnswer(answers.pick)),
			showWorkspaceFolderPick: () => Promise.resolve(undefined),
			showTextDocument: (target: unknown) => {
				opened.push(target);
				return Promise.resolve({});
			},
			activeTextEditor: undefined,
			setStatusBarMessage: () => new Disposable(() => { })
		},
		commands: {
			registerCommand: (id: string, handler: (...args: unknown[]) => unknown) => {
				registeredCommands.set(id, handler);
				return new Disposable(() => registeredCommands.delete(id));
			},
			executeCommand: (id: string, ...args: unknown[]) => {
				calls.push({ id, args });
				return Promise.resolve(undefined);
			}
		},
		FileDecoration: class { }
	};

	const state: FakeState = {
		resourceGroups,
		sourceControls,
		registeredCommands,
		calls,
		opened,
		shown,
		uri,
		lastCall: id => [...calls].reverse().find(call => call.id === id),
		answers,
		fireFileEvent: (kind, fsPath) => {
			for (const listener of watcherListeners.get(`**/*:${kind}`) ?? []) {
				listener(FakeUri.file(fsPath));
			}
		},
		settle: async (ms = 1500) => {
			await new Promise(resolve => setTimeout(resolve, ms));
		},
		contentOf: async uri => {
			const provider = contentProviders.get((uri as { scheme: string }).scheme);
			if (!provider) {
				throw new Error('chưa đăng ký trình cấp nội dung cho lược đồ này');
			}
			return provider.provideTextDocumentContent(uri, { isCancellationRequested: false });
		}
	};
	return { vscode, state };
}

/** Các trình cấp nội dung ảo đã đăng ký, dùng để gọi lại trong kiểm thử. */
export const contentProviders = new Map<string, ContentProvider>();

/** Phần API cần có để cấp nội dung cho một địa chỉ ảo. */
export interface ContentProvider {
	provideTextDocumentContent(uri: unknown, token: { isCancellationRequested: boolean }): Promise<string>;
}

/**
 * Nạp phần kích hoạt của tiện ích với API VS Code giả.
 *
 * Bản giả đủ để phần kích hoạt chạy hết: dò kho, đọc trạng thái, đăng ký lệnh,
 * đăng ký bốn khung. Hàm trả về cách gỡ bản giả ra khỏi bộ nạp module.
 */
export function loadExtension(vscode: Record<string, unknown>): { activate(context: unknown): void } {
	// Chặn mọi lần require('vscode') trong mã nguồn.
	const load = (Module as unknown as { _load(request: string, parent: unknown, isMain: boolean): unknown })._load;
	(Module as unknown as { _load: unknown })._load = function (request: string, parent: unknown, isMain: boolean) {
		if (request === 'vscode') {
			return vscode;
		}
		return load.call(this, request, parent, isMain);
	};

	// Nạp lại từ đầu để mã nguồn nhận bản giả thay vì module thật của VS Code.
	for (const key of Object.keys(require.cache)) {
		if (key.startsWith(path.join(__dirname, '..')) && !key.includes(path.join('out', 'test'))) {
			delete require.cache[key];
		}
	}
	const extension = require('../extension') as { activate(context: unknown): void };
	return {
		activate: context => {
			try {
				extension.activate(context);
			} finally {
				(Module as unknown as { _load: unknown })._load = load;
			}
		}
	};
}

/** package.json của tiện ích, đọc theo đường dẫn tới tệp đã biên dịch. */
export function manifest(): {
	contributes: {
		commands: { command: string; title?: string; category?: string }[];
		submenus: { id: string; label: string }[];
		menus: Record<string, { command?: string; submenu?: string; when?: string; group?: string; alt?: string }[]>;
		views: Record<string, { id: string; name: string }[]>;
		viewsWelcome?: { view: string; contents: string }[];
		configuration: { properties: Record<string, { default?: unknown }> }[];
	};
} {
	return JSON.parse(readFileSync(path.join(__dirname, '..', '..', 'package.json'), 'utf8'));
}

/** Tệp bản dịch nhãn của package.json. */
export function nls(): Record<string, string> {
	return JSON.parse(readFileSync(path.join(__dirname, '..', '..', 'package.nls.json'), 'utf8'));
}

/** Dựng một kho tm trên thư mục tạm rồi trả về thư mục gốc của nó. */
export function makeRepo(tmPath: string, prefix = 'tm-test-'): string {
	const root = mkdtempSync(path.join(tmpdir(), prefix));
	execFileSync(tmPath, ['vcs', '-C', root, 'init'], { stdio: 'ignore' });
	return root;
}

/** Ghi một tệp trong kho, tạo thư mục cha nếu thiếu. */
export function writeRepoFile(root: string, relative: string, content: string): void {
	const full = path.join(root, relative);
	mkdirSync(path.dirname(full), { recursive: true });
	writeFileSync(full, content, 'utf8');
}

/** Chạy một lệnh tm trong kho và bỏ qua output. */
export function runTm(tmPath: string, root: string, args: string[]): void {
	execFileSync(tmPath, ['vcs', '-C', root, ...args], { stdio: 'ignore' });
}