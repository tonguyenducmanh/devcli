import {
	Disposable,
	Event,
	EventEmitter,
	FileDecoration,
	FileDecorationProvider,
	Uri,
	window
} from 'vscode';

import { badgeColor, BADGES, GroupId, Repository } from './repository';

/**
 * Vẽ chữ viết tắt M/A/D/U cạnh tên tệp trong cây thư mục.
 *
 * Khung Source Control đã có chữ của riêng nó, nên chỗ này lo phần ở cây thư mục
 * và ở thanh bên trình soạn thảo. Cách làm giống hệt extension git: một nhà cung
 * cấp cho mọi kho, nghe trạng thái của tất cả kho rồi tô lại các tệp vừa đổi.
 */
export class TdDecorations implements FileDecorationProvider, Disposable {
	private readonly onDidChangeFileDecorationsEmitter = new EventEmitter<Uri[] | undefined>();
	readonly onDidChangeFileDecorations: Event<Uri[] | undefined> = this.onDidChangeFileDecorationsEmitter.event;

	private readonly decorations = new Map<string, FileDecoration>();
	private readonly disposables: Disposable[] = [];
	private enabled = true;

	constructor(private readonly repositories: () => readonly Repository[]) {
		this.disposables.push(window.registerFileDecorationProvider(this));
	}

	/** Bật hay tắt việc vẽ chữ viết tắt, theo cấu hình `td.decorations.enabled`. */
	setEnabled(enabled: boolean): void {
		this.enabled = enabled;
	}

	/** Gọi lại khi trạng thái một kho đổi để vẽ lại chữ viết tắt. */
	refresh(): void {
		const next = new Map<string, FileDecoration>();
		for (const repository of this.repositories()) {
			for (const state of repository.resourceStates) {
				const uri = repository.toAbsolutePath(state.path).toString();
				// Tệp đã xoá không còn trên đĩa nên không tô trong cây thư mục.
				if (state.status === 'D' && state.group === GroupId.WorkingTree) {
					continue;
				}
				next.set(uri, {
					badge: BADGES[state.status],
					color: badgeColor(state),
					tooltip: badgeText(state.group)
				});
			}
		}
		// Chỉ báo cho VS Code những tệp vừa đổi trạng thái trang trí.
		const touched = new Set<string>([...this.decorations.keys(), ...next.keys()]);
		this.decorations.clear();
		for (const [key, value] of next) {
			this.decorations.set(key, value);
		}
		this.onDidChangeFileDecorationsEmitter.fire([...touched].map(key => Uri.parse(key)));
	}

	provideFileDecoration(uri: Uri): FileDecoration | undefined {
		if (!this.enabled) {
			return undefined;
		}
		return this.decorations.get(uri.toString());
	}

	dispose(): void {
		for (const d of this.disposables.splice(0)) {
			d.dispose();
		}
		this.onDidChangeFileDecorationsEmitter.dispose();
	}
}

/** Nhãn nhóm, dùng cho lời nhắc khi rê chuột vào tệp trong cây thư mục. */
function badgeText(group: GroupId): string {
	switch (group) {
		case GroupId.Merge:
			return 'Xung đột';
		case GroupId.Index:
			return 'Đã stage';
		case GroupId.WorkingTree:
			return 'Đã sửa';
		default:
			return 'Chưa được theo dõi';
	}
}