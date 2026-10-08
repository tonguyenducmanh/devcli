import * as path from 'node:path';
import {
	CancellationToken,
	Disposable,
	EventEmitter,
	QuickDiffProvider,
	scm,
	SourceControl,
	SourceControlResourceGroup,
	SourceControlResourceState,
	ThemeColor,
	Uri,
	window,
	workspace
} from 'vscode';

import { TdPatch, TdStatus, TdStatusCode } from './parse';
import { Td } from './td';
import { Ref, Side, tdUri } from './uri';

/** Nhóm tệp. Tên id trùng với extension git để cách hiển thị là như nhau. */
export const enum GroupId {
	Merge = 'merge',
	Index = 'index',
	WorkingTree = 'workingTree',
	Untracked = 'untracked'
}

/** Nhãn của từng nhóm, giống extension git. */
export const GROUP_LABELS: Record<GroupId, string> = {
	[GroupId.Merge]: 'Merge Changes',
	[GroupId.Index]: 'Staged Changes',
	[GroupId.WorkingTree]: 'Changes',
	[GroupId.Untracked]: 'Untracked Changes'
};

/** Chữ viết tắt hiện ở góc tệp trong cây thư mục, giống git. */
export const BADGES: Record<TdStatusCode, string> = {
	A: 'A',
	M: 'M',
	D: 'D',
	U: 'U',
	'?': 'U'
};

/** Một mục tệp của khung Source Control, dùng lại cho cả giao diện lẫn lệnh. */
export interface TdResourceState {
	path: string;
	group: GroupId;
	status: TdStatusCode;
}

/** Trạng thái rỗng, dùng khi kho chưa đọc được gì. */
function emptyStatus(): TdStatus {
	return {
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
}

/**
 * Một kho td: giữ SourceControl cho VS Code và hỏi td về kho đó.
 *
 * Mọi thao tác thay đổi đều chạy qua td rồi đọc lại trạng thái, nên không có
 * đường nào sửa dữ liệu kho trực tiếp từ phía extension.
 */
export class Repository implements Disposable {
	private readonly onDidChangeStateEmitter = new EventEmitter<void>();
	/** Bắn khi trạng thái kho đổi, các view nghe để tự tải lại. */
	readonly onDidChangeState = this.onDidChangeStateEmitter.event;

	private readonly sourceControl: SourceControl;
	private readonly groups = new Map<GroupId, SourceControlResourceGroup>();
	private readonly disposables: Disposable[] = [];

	private status: TdStatus = emptyStatus();

	/** Khác biệt đã dựng, gom theo đường dẫn, mỗi vùng một bản riêng. */
	private readonly unstagedPatches = new Map<string, TdPatch>();
	private readonly stagedPatches = new Map<string, TdPatch>();

	private busy = false;
	private lastError: string | undefined;

	/** Danh sách tệp đang hiện, phục vụ trang trí cây thư mục và lệnh khác. */
	private states: TdResourceState[] = [];

	constructor(readonly root: string, private readonly td: Td) {
		this.sourceControl = scm.createSourceControl('td', 'TD', Uri.file(root));
		this.disposables.push(this.sourceControl);

		for (const id of [GroupId.Merge, GroupId.Index, GroupId.WorkingTree, GroupId.Untracked]) {
			const group = this.sourceControl.createResourceGroup(id, GROUP_LABELS[id]);
			this.groups.set(id, group);
			this.disposables.push(group);
		}
		// Nhóm Changes luôn hiện vì người dùng cần nhìn thấy nơi để thêm tệp mới.
		this.groups.get(GroupId.WorkingTree)!.hideWhenEmpty = false;
		this.groups.get(GroupId.Merge)!.hideWhenEmpty = true;
		this.groups.get(GroupId.Untracked)!.hideWhenEmpty = true;

		this.sourceControl.acceptInputCommand = {
			command: 'td.commit',
			title: 'Commit',
			arguments: [this.sourceControl]
		};
		this.sourceControl.quickDiffProvider = new TdQuickDiffProvider(this);
		this.sourceControl.inputBox.placeholder = inputPlaceholder('');
	}

	/** Các nút ở thanh trạng thái, đặt bởi Model. */
	set statusBarCommands(value: { command: string; title: string; tooltip: string }[]) {
		this.sourceControl.statusBarCommands = value;
	}

	/** SourceControl của kho này, dùng cho lệnh do VS Code gọi lại. */
	get scm(): SourceControl {
		return this.sourceControl;
	}

	/** Tên nhánh đang đứng, hoặc mô tả khi HEAD đang tách rời. */
	get headLabel(): string {
		if (this.status.detached) {
			return this.status.head ? `HEAD detached at ${this.status.head}` : 'HEAD detached';
		}
		return this.status.branch || 'No Branch';
	}

	/** Trạng thái mới nhất đọc từ td. */
	get current(): TdStatus {
		return this.status;
	}

	/** Có thao tác nào đang chạy không, dùng để khoá nút commit. */
	get isBusy(): boolean {
		return this.busy;
	}

	/** Lỗi của lần làm mới gần nhất, nếu có. */
	get error(): string | undefined {
		return this.lastError;
	}

	/** Danh sách tệp đang hiện trong khung Source Control. */
	get resourceStates(): TdResourceState[] {
		return this.states;
	}

	/** Số tệp đang hiện, dùng cho con số cạnh biểu tượng trên thanh hoạt động. */
	get changeCount(): number {
		return this.states.length;
	}

	/** Đường dẫn tương đối của một tệp trong kho, dùng khi ghép lệnh td. */
	toRelativePath(uri: Uri): string {
		return path.relative(this.root, uri.fsPath).split(path.sep).join('/');
	}

	/** Địa chỉ tuyệt đối của một tệp trong kho. */
	toAbsolutePath(relative: string): Uri {
		return Uri.file(path.join(this.root, relative));
	}

	/**
	 * Đọc lại trạng thái từ td rồi cập nhật khung Source Control.
	 *
	 * Khác biệt đã dựng bị bỏ vì trạng thái có thể đã đổi; chúng sẽ được dựng
	 * lại khi có ai đó thật sự mở khung so sánh.
	 */
	async refresh(): Promise<void> {
		try {
			this.status = await this.td.status(this.root);
			this.lastError = undefined;
		} catch (error) {
			this.lastError = messageOf(error);
			this.updateModel();
			this.onDidChangeStateEmitter.fire();
			return;
		}
		this.unstagedPatches.clear();
		this.stagedPatches.clear();
		this.updateModel();
		this.onDidChangeStateEmitter.fire();
	}

	/** Cập nhật những thứ VS Code hiển thị: nhóm tệp, ô nhập, số đếm. */
	private updateModel(): void {
		this.states = [
			...toStates(GroupId.Merge, this.status.conflicts),
			...toStates(GroupId.Index, this.status.staged),
			...toStates(GroupId.WorkingTree, this.status.unstaged),
			...toStates(GroupId.Untracked, this.status.untracked)
		];

		// Gom xong rồi mới gán, tránh khoảnh khắc nhóm nào đó trống rồi mới đầy
		// khiến khung Source Control nhấp nháy.
		for (const group of this.groups.values()) {
			group.resourceStates = this.states
				.filter(state => state.group === group.id)
				.map(state => this.toResourceState(state));
		}

		const config = workspace.getConfiguration('td', Uri.file(this.root));
		this.groups.get(GroupId.Index)!.hideWhenEmpty = !config.get('alwaysShowStagedChangesResourceGroup', true);
		this.sourceControl.inputBox.visible = config.get('showCommitInput', true);
		this.sourceControl.inputBox.placeholder = inputPlaceholder(this.headLabel);
		this.sourceControl.inputBox.enabled = !this.busy;
		this.sourceControl.count = this.changeCount;
	}

	/** Đổi một mục của nhóm thành đối tượng VS Code hiển thị được. */
	private toResourceState(state: TdResourceState): SourceControlResourceState {
		const uri = this.toAbsolutePath(state.path);
		const resource: SourceControlResourceState = {
			resourceUri: uri,
			contextValue: contextValueOf(state),
			command: {
				command: 'td.openFile',
				title: 'Open File',
				// Truyền cả kho lẫn tệp: khung Source Control gọi lệnh mà không
				// kèm gì, còn các lệnh khác gọi với đúng hai đối số này.
				arguments: [this.sourceControl, uri]
			},
			decorations: {
				// Gạch ngang tên tệp đã xoá trên đĩa, giống git.
				strikeThrough: state.status === 'D' && state.group === GroupId.WorkingTree,
				tooltip: tooltipOf(state),
				light: { iconPath: uri },
				dark: { iconPath: uri }
			}
		};
		return resource;
	}

	/**
	 * Lấy khác biệt của một tệp, dựng một lần rồi giữ tới lần làm mới sau.
	 *
	 * `td vcs diff` được gọi cho đúng một tệp nên phần ngữ cảnh ôm trọn tệp vẫn
	 * nhỏ, kể cả trong kho lớn.
	 */
	async patchFor(relativePath: string, staged: boolean): Promise<TdPatch | undefined> {
		const cache = staged ? this.stagedPatches : this.unstagedPatches;
		const known = cache.get(relativePath);
		if (known) {
			return known;
		}
		const patches = await this.td.diff(this.root, { staged, paths: [relativePath] });
		for (const patch of patches) {
			cache.set(patch.path, patch);
		}
		return cache.get(relativePath);
	}

	/**
	 * Nội dung của một phía trong khung so sánh.
	 *
	 * Phía trên đĩa thì đọc tệp thật để giữ đúng dấu xuống dòng cuối cùng; phía
	 * còn lại dựng lại từ khác biệt của td.
	 */
	async contentOf(relativePath: string, ref: Ref | string, side: Side, token: CancellationToken): Promise<string> {
		let patch: TdPatch | undefined;
		if (ref === Ref.Worktree) {
			return Buffer.from(await workspace.fs.readFile(this.toAbsolutePath(relativePath))).toString('utf8');
		}
		if (ref === Ref.Head) {
			patch = await this.patchFor(relativePath, true);
		} else if (ref === Ref.Index) {
			patch = await this.patchFor(relativePath, false);
		} else {
			patch = await this.patchForRevision(relativePath, String(ref));
		}
		if (token.isCancellationRequested || !patch) {
			return '';
		}
		return side === Side.New ? patch.new : patch.old;
	}

	/** Khác biệt của một tệp giữa một commit và phụ huynh của nó. */
	private async patchForRevision(relativePath: string, revision: string): Promise<TdPatch | undefined> {
		const patches = await this.td.diff(this.root, { revision, paths: [relativePath] });
		return patches.find(p => p.path === relativePath);
	}

	/** Địa chỉ nội dung ảo của một phía. */
	uriFor(relativePath: string, ref: Ref | string, side: Side): Uri {
		return tdUri({ repo: this.root, path: relativePath, ref, side });
	}

	/** Chạy một lệnh thay đổi rồi đọc lại trạng thái, báo lỗi khi hỏng. */
	async run(description: string, operation: () => Promise<void>): Promise<void> {
		this.busy = true;
		this.updateModel();
		try {
			await operation();
			await this.refresh();
		} catch (error) {
			this.busy = false;
			await this.refresh();
			void window.showErrorMessage(`${description}: ${messageOf(error)}`);
			throw error;
		} finally {
			this.busy = false;
			this.updateModel();
		}
	}

	/** Nội dung đang gõ trong ô nhập commit. */
	get inputValue(): string {
		return this.sourceControl.inputBox.value;
	}

	/** Xoá ô nhập sau khi commit thành công. */
	clearInput(): void {
		this.sourceControl.inputBox.value = '';
	}

	set inputValue(value: string) {
		this.sourceControl.inputBox.value = value;
	}

	/** Kho có tệp nào đang chờ commit không. */
	hasChanges(): boolean {
		return this.status.staged.length > 0
			|| this.status.unstaged.length > 0
			|| this.status.untracked.length > 0;
	}

	dispose(): void {
		this.onDidChangeStateEmitter.dispose();
		for (const d of this.disposables.splice(0)) {
			d.dispose();
		}
	}
}

/** Đổ danh sách trạng thái của td thành mục của khung Source Control. */
function toStates(group: GroupId, entries: { path: string; status: TdStatusCode }[]): TdResourceState[] {
	return entries.map(entry => ({ path: entry.path, group, status: entry.status }));
}

/** Màu của chữ viết tắt, theo nhóm và theo loại thay đổi. */
export function badgeColor(state: TdResourceState): ThemeColor {
	switch (state.status) {
		case 'A':
			return new ThemeColor('tdDecoration.addedResourceForeground');
		case 'U':
			return new ThemeColor('tdDecoration.conflictingResourceForeground');
		case 'D':
			return new ThemeColor(state.group === GroupId.Index
				? 'tdDecoration.stageDeletedResourceForeground'
				: 'tdDecoration.deletedResourceForeground');
		case '?':
			return new ThemeColor('tdDecoration.untrackedResourceForeground');
		default:
			return new ThemeColor(state.group === GroupId.Index
				? 'tdDecoration.stageModifiedResourceForeground'
				: 'tdDecoration.modifiedResourceForeground');
	}
}

/** Lời nhắc hiện khi rê chuột vào tệp trong khung Source Control. */
function tooltipOf(state: TdResourceState): string {
	const place: Record<GroupId, string> = {
		[GroupId.Merge]: 'Đang giải quyết xung đột',
		[GroupId.Index]: 'Đã stage',
		[GroupId.WorkingTree]: 'Đã sửa trên đĩa',
		[GroupId.Untracked]: 'Chưa được td theo dõi'
	};
	const change: Record<TdStatusCode, string> = {
		A: 'Thêm',
		M: 'Sửa',
		D: 'Xoá',
		U: 'Xung đột',
		'?': 'Mới'
	};
	return `${place[state.group]} (${change[state.status]})`;
}

/** contextValue để menu trong package.json khớp đúng nhóm và loại tệp. */
function contextValueOf(state: TdResourceState): string {
	if (state.group === GroupId.Merge) {
		return 'conflict';
	}
	if (state.group === GroupId.Untracked) {
		return 'untracked';
	}
	if (state.status === 'D') {
		return 'deleted';
	}
	if (state.status === 'A') {
		return 'added';
	}
	return 'modified';
}

/** Ô nhập commit, có tên nhánh đang đứng. */
function inputPlaceholder(branch: string): string {
	return branch
		? `Message (Ctrl+Enter to commit on '${branch}')`
		: 'Message (Ctrl+Enter to commit)';
}

/** Lấy thông điệp của một lỗi bất kỳ. */
export function messageOf(error: unknown): string {
	if (error instanceof Error) {
		return error.message;
	}
	return String(error);
}

/**
 * Cung cấp tệp gốc cho phần hiển thị khác biệt nhanh trong trình soạn thảo.
 *
 * VS Code dùng địa chỉ trả về để tô vạch ở rìa và dải khác biệt ngay trong
 * tệp đang mở, giống hệt git.
 */
class TdQuickDiffProvider implements QuickDiffProvider {
	constructor(private readonly repository: Repository) {}

	async provideOriginalResource(uri: Uri, token: CancellationToken): Promise<Uri | undefined> {
		if (uri.scheme !== 'file' || token.isCancellationRequested) {
			return undefined;
		}
		const relative = this.repository.toRelativePath(uri);
		if (!relative || relative.startsWith('..')) {
			return undefined;
		}
		const inWorkingTree = this.repository.current.unstaged.some(e => e.path === relative)
			|| this.repository.current.untracked.some(e => e.path === relative);
		if (inWorkingTree) {
			// So với vùng chuẩn bị: đây là phía gốc của thay đổi trên đĩa.
			return this.repository.uriFor(relative, Ref.Index, Side.Old);
		}
		const inIndex = this.repository.current.staged.some(e => e.path === relative);
		if (inIndex) {
			// Đã stage hết, so với HEAD.
			return this.repository.uriFor(relative, Ref.Head, Side.Old);
		}
		return undefined;
	}
}