import {
	CancellationToken,
	Disposable,
	Event,
	EventEmitter,
	ExtensionContext,
	TextDocumentContentProvider,
	Uri,
	window,
	workspace
} from 'vscode';

import { registerCommands } from './commands';
import { Model } from './model';
import { parseTdUri } from './uri';

/**
 * Điểm khởi động của extension.
 *
 * Kích hoạt theo ba bước: dò kho tm trong workspace, đăng ký lệnh, rồi cấp nội
 * dung ảo cho các khung so sánh. Bản thân VS Code không cần biết gì về tm, nó chỉ
 * thấy một nguồn quản lý phiên bản như bất kỳ extension nguồn quản lý nào khác.
 */
export function activate(context: ExtensionContext): void {
	const log = window.createOutputChannel('TM VCS', { log: true });
	context.subscriptions.push(log);

	const model = new Model(log);
	context.subscriptions.push(model);
	log.appendLine(`Sử dụng lệnh tm: ${model.command}`);

	context.subscriptions.push(...registerCommands(model, log));
	context.subscriptions.push(new TmContentProvider(model));

	// Dò kho trước khi báo sẵn sàng để khung Source Control có nội dung ngay.
	void model.discover().then(async () => {
		for (const repository of model.all) {
			log.appendLine(`Đã mở kho ${repository.root} (${repository.headLabel})`);
		}
		await model.checkMissing();
		if (model.all.length === 0) {
			log.appendLine('Không tìm thấy kho tm nào trong workspace.');
		}
	});

	// Cập nhật lại danh sách kho khi người dùng đổi cấu hình tm.path.
	context.subscriptions.push(workspace.onDidChangeConfiguration(event => {
		if (event.affectsConfiguration('tm.path')) {
			log.appendLine('Cấu hình tm.path đổi, dò kho lại.');
			void model.discover(true);
		}
	}));
}

/** Dọn dẹp khi extension bị gỡ hay tắt. */
export function deactivate(): void {
	// Mọi tài nguyên đã nằm trong context.subscriptions.
}

/**
 * Cấp nội dung cho các địa chỉ `tm:` mà khung so sánh của VS Code mở ra.
 *
 * Đây là phần tương đương của việc git phục vụ nội dung qua lược đồ `git:`:
 * mỗi địa chỉ mang theo kho, đường dẫn, điểm trong lịch sử và phía cần xem.
 */
class TmContentProvider implements TextDocumentContentProvider {
	private readonly onDidChangeEmitter = new EventEmitter<Uri>();
	readonly onDidChange: Event<Uri> = this.onDidChangeEmitter.event;

	private readonly disposables: Disposable[] = [];

	/** Các địa chỉ đã cấp nội dung, dùng để bảo VS Code đọc lại đúng những cái đó. */
	private readonly served = new Set<string>();

	constructor(private readonly model: Model) {
		this.disposables.push(workspace.registerTextDocumentContentProvider('tm', this));
		this.disposables.push(model.onDidChangeRepository(() => this.fire()));
	}

	async provideTextDocumentContent(uri: Uri, token: CancellationToken): Promise<string> {
		const info = parseTdUri(uri);
		if (!info) {
			return '';
		}
		const repository = this.model.byRoot(info.repo);
		if (!repository) {
			return '';
		}
		this.served.add(uri.toString());
		try {
			return await repository.contentOf(info.path, info.ref, info.side, token);
		} catch (error) {
			// Không dựng được nội dung thì hiện lỗi ngay trong khung so sánh,
			// còn im lặng thì người dùng chỉ thấy một khung trống.
			return `Không dựng được nội dung của ${info.path}:\n${error instanceof Error ? error.message : String(error)}`;
		}
	}

	/** Bảo VS Code đọc lại nội dung của những địa chỉ ảo đã cấp. */
	private fire(): void {
		for (const key of this.served) {
			this.onDidChangeEmitter.fire(Uri.parse(key));
		}
		this.served.clear();
	}

	dispose(): void {
		for (const d of this.disposables.splice(0)) {
			d.dispose();
		}
		this.onDidChangeEmitter.dispose();
	}
}
