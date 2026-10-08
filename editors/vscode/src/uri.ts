import { Uri } from 'vscode';

/**
 * Lược đồ dùng cho nội dung "ảo" mà extension tự dựng.
 *
 * Khung so sánh của VS Code cần mỗi phía là một tài liệu riêng. Phía nằm trên
 * đĩa thì mở thẳng tệp thật, còn phía nằm trong kho (vùng chuẩn bị, HEAD, một
 * commit bất kỳ) thì không có tệp tương ứng trên đĩa, nên extension cấp nội
 * dung qua lược đồ riêng.
 */
export const TM_SCHEME = 'tm';

/** Các điểm trong lịch sử mà một phía của khung so sánh có thể trỏ tới. */
export const enum Ref {
	/** Nội dung đang ở trong vùng chuẩn bị. */
	Index = 'index',
	/** Nội dung của commit HEAD. */
	Head = 'head',
	/** Tệp trên đĩa. */
	Worktree = 'worktree'
}

/** Phía nào của một cặp khác biệt. */
export const enum Side {
	Old = 'old',
	New = 'new'
}

/** Thông tin đủ để dựng lại địa chỉ của nội dung ảo. */
export interface TmRef {
	repo: string;
	path: string;
	ref: Ref | string;
	side: Side;
}

/**
 * Dựng địa chỉ nội dung ảo.
 *
 * Thư mục gốc của kho nằm trong truy vấn chứ không nằm trong phần đường dẫn,
 * vì phần đường dẫn phải giữ đúng tên tệp để VS Code tự nhận diện ngôn ngữ và
 * đặt tên tab.
 */
export function tmUri(info: TmRef): Uri {
	const query = [
		`repo=${encodeURIComponent(info.repo)}`,
		`ref=${encodeURIComponent(info.ref)}`,
		`side=${info.side}`
	].join('&');
	return Uri.from({ scheme: TM_SCHEME, path: '/' + info.path, query });
}

/** Đọc lại thông tin từ một địa chỉ nội dung ảo, trả về undefined nếu không khớp. */
export function parseTdUri(uri: Uri): TmRef | undefined {
	if (uri.scheme !== TM_SCHEME) {
		return undefined;
	}
	const params = new URLSearchParams(uri.query);
	const repo = params.get('repo');
	const ref = params.get('ref');
	if (!repo || !ref) {
		return undefined;
	}
	const side = params.get('side') === Side.New ? Side.New : Side.Old;
	return { repo, ref, side, path: uri.path.replace(/^\//, '') };
}