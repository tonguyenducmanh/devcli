import {
	Event,
	EventEmitter,
	ProviderResult,
	ThemeIcon,
	TreeDataProvider,
	TreeItem,
	TreeItemCollapsibleState,
	workspace
} from 'vscode';

import { Model } from './model';
import { Repository } from './repository';
import { Td } from './td';
import { TdBranch, TdLogEntry, TdStash, TdTag } from './parse';

/**
 * Phần chung cho bốn view: nhánh, lịch sử commit, bản lưu tạm và tag.
 *
 * Bốn view đều đọc từ kho đang hoạt động và tự tải lại mỗi khi trạng thái kho
 * đổi, giống cách các view của extension git lấy dữ liệu.
 */
abstract class TdTreeProvider<T> implements TreeDataProvider<T> {
	protected readonly emitter = new EventEmitter<T | undefined>();
	readonly onDidChangeTreeData: Event<T | undefined> = this.emitter.event;

	private cached: T[] = [];
	private loading = false;

	constructor(protected readonly model: Model) {}

	/** Kho dùng làm nguồn dữ liệu. */
	protected get repository(): Repository | undefined {
		return this.model.active;
	}

	/** Nạp dữ liệu cho kho cho trước, trả về undefined nếu không có kho. */
	protected abstract load(td: Td, root: string): Promise<T[]>;

	/** Dựng một mục của view từ một dòng dữ liệu. */
	protected abstract toTreeItem(entry: T, repository: Repository): TreeItem;

	/** Đọc lại dữ liệu, dùng khi kho đổi hoặc người dùng bấm làm mới. */
	async refresh(): Promise<void> {
		if (this.loading) {
			return;
		}
		const repository = this.repository;
		if (!repository) {
			this.cached = [];
			this.emitter.fire(undefined);
			return;
		}
		this.loading = true;
		try {
			this.cached = await this.load(this.model.cli, repository.root);
		} catch {
			// Lỗi đã được Model đưa ra báo lỗi rồi, view chỉ cần giữ dữ liệu cũ.
			this.cached = [];
		} finally {
			this.loading = false;
			this.emitter.fire(undefined);
		}
	}

	getTreeItem(element: T): TreeItem {
		return this.toTreeItem(element, this.repository!);
	}

	getChildren(element?: T): ProviderResult<T[]> {
		if (element) {
			// Các view này là danh sách phẳng, không có phần tử con.
			return [];
		}
		return this.cached;
	}

	/** Bỏ dữ liệu đang giữ, dùng khi không còn kho nào. */
	clear(): void {
		this.cached = [];
		this.emitter.fire(undefined);
	}
}

/** View Branches: các nhánh cục bộ, nhánh đang đứng luôn ở đầu. */
export class TdBranchesProvider extends TdTreeProvider<TdBranch> {
	constructor(model: Model) {
		super(model);
		// Chỉ nạp lại khi chính kho đang xem đổi, tránh gọi td bốn lần mỗi
		// lần có tệp được lưu.
		model.onDidChangeRepository(repository => {
			if (repository === this.repository) {
				void this.refresh();
			}
		});
	}

	protected async load(td: Td, root: string): Promise<TdBranch[]> {
		return td.branches(root);
	}

	protected toTreeItem(entry: TdBranch, repository: Repository): TreeItem {
		const item = new TreeItem(entry.name, TreeItemCollapsibleState.None);
		item.description = entry.subject || entry.hash;
		item.iconPath = new ThemeIcon(entry.current ? 'git-branch' : 'git-branch');
		item.contextValue = entry.current ? 'branchCurrent' : 'branch';
		item.tooltip = branchTooltip(entry);
		item.resourceUri = repository.toAbsolutePath('.tdx');
		item.command = {
			command: 'td.checkout',
			title: 'Checkout',
			arguments: [repository, entry.name]
		};
		return item;
	}
}

/** View Commits: lịch sử từ HEAD đi ngược về. */
export class TdCommitsProvider extends TdTreeProvider<TdLogEntry> {
	constructor(model: Model) {
		super(model);
		// Chỉ nạp lại khi chính kho đang xem đổi, tránh gọi td bốn lần mỗi
		// lần có tệp được lưu.
		model.onDidChangeRepository(repository => {
			if (repository === this.repository) {
				void this.refresh();
			}
		});
	}

	protected async load(td: Td, root: string): Promise<TdLogEntry[]> {
		const max = workspace.getConfiguration('td').get<number>('logMaxCount', 500);
		return td.log(root, max);
	}

	protected toTreeItem(entry: TdLogEntry, repository: Repository): TreeItem {
		const item = new TreeItem(entry.summary || '(không có tiêu đề)', TreeItemCollapsibleState.None);
		item.description = entry.refs.length > 0 ? entry.refs.join(', ') : entry.hash;
		item.id = `${repository.root}:${entry.hash}`;
		item.iconPath = new ThemeIcon('git-commit');
		item.contextValue = 'commit';
		item.tooltip = `${entry.hash}\n\n${entry.summary}`;
		item.command = {
			command: 'td.openCommitChanges',
			title: 'Open Changes',
			arguments: [repository, entry.hash]
		};
		return item;
	}
}

/** View Stashes: các bản lưu tạm, mục mới nhất ở trên cùng. */
export class TdStashesProvider extends TdTreeProvider<TdStash> {
	constructor(model: Model) {
		super(model);
		// Chỉ nạp lại khi chính kho đang xem đổi, tránh gọi td bốn lần mỗi
		// lần có tệp được lưu.
		model.onDidChangeRepository(repository => {
			if (repository === this.repository) {
				void this.refresh();
			}
		});
	}

	protected async load(td: Td, root: string): Promise<TdStash[]> {
		return td.stashes(root);
	}

	protected toTreeItem(entry: TdStash, repository: Repository): TreeItem {
		const item = new TreeItem(entry.message || `stash@{${entry.index}}`, TreeItemCollapsibleState.None);
		item.description = `stash@{${entry.index}}`;
		item.id = `${repository.root}:${entry.index}`;
		item.iconPath = new ThemeIcon('git-stash');
		item.contextValue = 'stash';
		item.tooltip = `stash@{${entry.index}}: ${entry.message}`;
		return item;
	}
}

/** View Tags: các điểm đánh dấu trong lịch sử. */
export class TdTagsProvider extends TdTreeProvider<TdTag> {
	constructor(model: Model) {
		super(model);
		// Chỉ nạp lại khi chính kho đang xem đổi, tránh gọi td bốn lần mỗi
		// lần có tệp được lưu.
		model.onDidChangeRepository(repository => {
			if (repository === this.repository) {
				void this.refresh();
			}
		});
	}

	protected async load(td: Td, root: string): Promise<TdTag[]> {
		return td.tags(root);
	}

	protected toTreeItem(entry: TdTag, repository: Repository): TreeItem {
		const item = new TreeItem(entry.name, TreeItemCollapsibleState.None);
		item.description = entry.message || entry.hash;
		item.id = `${repository.root}:${entry.name}`;
		item.iconPath = new ThemeIcon('tag');
		item.contextValue = 'tag';
		item.tooltip = `${entry.name} (${entry.hash})`;
		item.command = {
			command: 'td.checkout',
			title: 'Checkout',
			arguments: [repository, entry.name]
		};
		return item;
	}
}

/** Lời nhắc của một nhánh trong view Branches. */
function branchTooltip(entry: TdBranch): string {
	const lines = [entry.name, entry.subject].filter(Boolean);
	if (entry.upstream) {
		lines.push(`theo dõi ${entry.upstream}: đi trước ${entry.ahead}, đi sau ${entry.behind}`);
	}
	return lines.join('\n');
}
