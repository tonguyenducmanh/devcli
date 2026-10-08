import { execFileSync } from 'node:child_process';
import { existsSync, mkdirSync, mkdtempSync, readdirSync, readFileSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import * as path from 'node:path';
import { test } from 'node:test';
import Module = require('node:module');

/** Gốc kho của td, thư mục cha của thư mục extension. */
const REPO_ROOT = path.join(__dirname, '..', '..', '..', '..');

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
const runSmoke = TD_BIN ? test : test.skip;

class Disposable {
	constructor(private readonly callOnDispose: () => void) {}
	dispose(): void {
		this.callOnDispose();
	}
}

class EventEmitter<T> {
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
 * Một bản sao đủ dùng của API VS Code cho kiểm thử.
 *
 * Chỉ những phần tiện ích thật sự gọi tới mới có mặt: SourceControl, cây thư
 * mục, sự kiện cấu hình, ô nhập nội dung. Mục đích là bắt được lỗi chạy thật
 * trong lúc kích hoạt mà không cần mở VS Code.
 */
/** Những gì kiểm thử cần soi sau khi kích hoạt. */
interface FakeState {
	resourceGroups: { id: string; label: string; resourceStates: unknown[]; hideWhenEmpty?: boolean }[];
	sourceControls: Record<string, unknown>[];
	registeredCommands: Map<string, (...args: unknown[]) => unknown>;
	shown: string[];
	executed: string[];
	uri(fsPath: string): unknown;
}

function fakeVscode(tdPath: string): { vscode: Record<string, unknown>; state: FakeState } {
	const uri = (fsPath: string) => ({ fsPath, scheme: 'file', path: fsPath, toString: () => `file://${fsPath}` });
	const resourceGroups: { id: string; label: string; resourceStates: unknown[]; hideWhenEmpty?: boolean }[] = [];
	const sourceControls: Record<string, unknown>[] = [];
	const registeredCommands = new Map<string, (...args: unknown[]) => unknown>();
	const shown: string[] = [];
	const executed: string[] = [];

	const vscode = {
		Uri: {
			file: uri,
			parse: (value: string) => ({ fsPath: value, scheme: 'file', path: value }),
			from: (parts: { scheme: string; path: string; query?: string }) => ({ ...parts, fsPath: parts.path, toString: () => `${parts.scheme}:${parts.path}` })
		},
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
			// Chỉ td.path có giá trị thật, còn lại trả về mặc định.
			getConfiguration: () => ({
				get: <T>(key: string, fallback?: T): T => (key === 'path' ? (tdPath as unknown as T) : (fallback as T))
			}),
			onDidChangeConfiguration: () => new Disposable(() => { }),
			onDidChangeWorkspaceFolders: () => new Disposable(() => { }),
			createFileSystemWatcher: () => ({
				onDidCreate: () => new Disposable(() => { }),
				onDidChange: () => new Disposable(() => { }),
				onDidDelete: () => new Disposable(() => { }),
				dispose: () => { }
			}),
			registerTextDocumentContentProvider: () => new Disposable(() => { }),
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
			showWarningMessage: () => Promise.resolve(undefined),
			showInformationMessage: () => Promise.resolve(undefined),
			showInputBox: () => Promise.resolve(undefined),
			showQuickPick: () => Promise.resolve(undefined),
			showWorkspaceFolderPick: () => Promise.resolve(undefined),
			activeTextEditor: undefined,
			setStatusBarMessage: () => new Disposable(() => { })
		},
		commands: {
			registerCommand: (id: string, handler: (...args: unknown[]) => unknown) => {
				registeredCommands.set(id, handler);
				return new Disposable(() => registeredCommands.delete(id));
			},
			executeCommand: (id: string) => {
				executed.push(id);
				return Promise.resolve(undefined);
			}
		},
		FileDecoration: class { }
	};

	const state: FakeState = { resourceGroups, sourceControls, registeredCommands, shown, executed, uri };
	return { vscode, state };
}

/**
 * Chạy phần kích hoạt của tiện ích với API VS Code giả.
 *
 * Bản giả đủ để phần kích hoạt chạy hết: dò kho, đọc trạng thái, đăng ký lệnh,
 * đăng ký bốn khung. Kiểm thử này bắt được lỗi chỉ xuất hiện lúc chạy, ví dụ
 * gọi sai API hoặc quên đăng ký lệnh nào đó.
 */
runSmoke('kích hoạt: dò kho, đọc trạng thái và đăng ký lệnh', async () => {
	const { vscode, state } = fakeVscode(TD_BIN!);

	const root = mkdtempSync(path.join(tmpdir(), 'td-smoke-'));
	execFileSync(TD_BIN!, ['vcs', '-C', root, 'init']);
	mkdirSync(path.join(root, 'thu-muc'), { recursive: true });
	writeFileSync(path.join(root, 'a.txt'), 'một\nhai\n', 'utf8');
	writeFileSync(path.join(root, 'thu-muc/b.txt'), 'x\n', 'utf8');
	execFileSync(TD_BIN!, ['vcs', '-C', root, 'add', '.']);
	execFileSync(TD_BIN!, ['vcs', '-C', root, 'commit', '-m', 'c1']);
	// Rời kho ở trạng thái có thay đổi: một tệp đã stage, một tệp sửa trên đĩa,
	// một tệp mới chưa theo dõi.
	writeFileSync(path.join(root, 'a.txt'), 'một\nhai sửa\n', 'utf8');
	execFileSync(TD_BIN!, ['vcs', '-C', root, 'add', 'a.txt']);
	writeFileSync(path.join(root, 'thu-muc/b.txt'), 'x\ny\n', 'utf8');
	writeFileSync(path.join(root, 'moi.txt'), 'mới\n', 'utf8');

	const workspace = (vscode.workspace as { workspaceFolders: unknown });
	workspace.workspaceFolders = [{ uri: state.uri(root), name: 'kho', index: 0 }];

	// Chặn mọi lần require('vscode') trong mã nguồn.
	const load = (Module as unknown as { _load(request: string, parent: unknown, isMain: boolean): unknown })._load;
	(Module as unknown as { _load: unknown })._load = function (request: string, parent: unknown, isMain: boolean) {
		if (request === 'vscode') {
			return vscode;
		}
		return load.call(this, request, parent, isMain);
	};

	try {
		// Nạp lại từ đầu để mã nguồn nhận bản giả thay vì module thật của VS Code.
		for (const key of Object.keys(require.cache)) {
			if (key.includes(path.join('out', 'extension.js')) || key.includes(path.join('out', 'model.js'))
				|| key.includes(path.join('out', 'repository.js')) || key.includes(path.join('out', 'commands.js'))
				|| key.includes(path.join('out', 'views.js')) || key.includes(path.join('out', 'decorations.js'))) {
				delete require.cache[key];
			}
		}

		const extension = require('../extension') as { activate(context: unknown): void };
		const context = { subscriptions: [] as { dispose(): void }[] };
		extension.activate(context);

		// activate() gọi việc dò kho theo lời hứa, chờ nốt thì đó.
		await new Promise(resolve => setTimeout(resolve, 1500));

		const pkg = JSON.parse(
			readFileSync(path.join(__dirname, '..', '..', 'package.json'), 'utf8')
		) as { contributes: { commands: { command: string }[] } };

		const missing = pkg.contributes.commands
			.map(c => c.command)
			.filter(id => !state.registeredCommands.has(id));
		if (missing.length > 0) {
			throw new Error(`các lệnh khai báo mà không đăng ký: ${missing.join(', ')}`);
		}

		if (state.sourceControls.length !== 1) {
			throw new Error(`phải có đúng một SourceControl, nhận ${state.sourceControls.length}`);
		}
		const source = state.sourceControls[0] as {
			label: string;
			count: number;
			inputBox: { placeholder: string };
			acceptInputCommand: { command: string } | undefined;
		};
		if (source.acceptInputCommand?.command !== 'td.commit') {
			throw new Error('ô nhập commit phải có nút commit');
		}
		if (!source.inputBox.placeholder.includes("'main'")) {
			throw new Error(`ô nhập phải nhắc nhánh đang đứng, nhận "${source.inputBox.placeholder}"`);
		}

		// Bốn nhóm thay đổi phải đúng tên và đúng số tệp.
		const byId = new Map(state.resourceGroups.map(g => [g.id, g]));
		if (byId.get('index')?.resourceStates.length !== 1) {
			throw new Error('nhóm Staged Changes phải có đúng một tệp');
		}
		if (byId.get('workingTree')?.resourceStates.length !== 1) {
			throw new Error('nhóm Changes phải có đúng một tệp');
		}
		if (byId.get('untracked')?.resourceStates.length !== 1) {
			throw new Error('nhóm Untracked Changes phải có đúng một tệp');
		}
		if (source.count !== 3) {
			throw new Error(`số đếm phải là 3, nhận ${source.count}`);
		}

		const leaks = context.subscriptions.filter(d => typeof (d as { dispose?: unknown }).dispose !== 'function');
		if (leaks.length > 0) {
			throw new Error(`có ${leaks.length} tài nguyên không huỷ được: ${leaks.map(d => d?.constructor?.name).join(', ')}`);
		}
		for (const d of context.subscriptions) {
			try {
				d.dispose();
			} catch (error) {
				throw new Error(`không huỷ được ${(d as object)?.constructor?.name}: ${String(error)}`);
			}
		}
	} finally {
		(Module as unknown as { _load: unknown })._load = load;
		rmSync(root, { recursive: true, force: true });
	}
});