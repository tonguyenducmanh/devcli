# devcli

> One command line app you need.

Mọi công cụ dòng lệnh bạn dùng hằng ngày, gói lại dưới **một lệnh duy nhất**.
Gõ `tm`, xem hết danh sách, không cần nhớ tên từng công cụ.

```bash
tm vcs init            # khởi tạo kho mã nguồn
tm use vcs             # đặt nhóm vcs làm mặc định (sau đó chỉ cần gõ tm commit, tm status)
tm commit -m "..."     # ghi lại thay đổi (tương đương tm vcs commit)
tm config --list       # xem cấu hình
```

`devcli` viết bằng Go, dựng cây lệnh bằng [cobra](https://github.com/spf13/cobra).
Mỗi công cụ là một nhóm lệnh con, và mỗi nhóm tự quản lý toàn bộ dữ liệu của
nó trong thư mục riêng.

**Vì sao tên là devcli mà chạy bằng `tm`?** `devcli` là tên dự án, còn `tm` là
tên lệnh gọi trên terminal, ngắn để gõ nhanh. Muốn đổi thì sửa một dòng
`CMD_NAME` trong [`scripts/build_binaries.sh`](scripts/build_binaries.sh).

Hiện tại có 2 nhóm chính:
- `tm vcs`: quản lý phiên bản mã nguồn cục bộ.
- `tm sys`: các tiện ích hệ thống (ls, mkdir, rm, cp, mv, grep, wc...).
Các nhóm khác sẽ được bổ sung theo cùng khuôn mẫu.

## Yêu cầu

- Go 1.23 trở lên, kiểm tra bằng `go version`.
- Không cần gì khác lúc chạy. devcli là một tệp thực thi duy nhất, không
  phụ thuộc thư viện hệ thống.

## Build

```bash
./build_all.sh                # tài liệu + binary mọi nền tảng + tiện ích VS Code
./build_all.sh --no-extension # bỏ qua bước đóng gói tiện ích
```

Kết quả nằm trong `out/`. Tệp `.vsix` cài vào VS Code trên mọi nền tảng:

```bash
code --install-extension out/devcli-tm-vscode-0.1.0.vsix
```

## Tài liệu chi tiết

Mọi hướng dẫn cụ thể đã được chia nhỏ. Vui lòng tìm và đọc các tài liệu tương ứng bên trong thư mục `docs/`.

Những thay đổi đáng kể của từng đợt nằm trong [`CHANGELOG.md`](CHANGELOG.md).

## Trình soạn thảo

| Thư mục | Vai trò |
| --- | --- |
| [`editors/vscode/`](editors/vscode/) | Tiện ích đưa `tm` vào khung Source Control của VS Code |