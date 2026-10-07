# devcli

> One command line app you need.

Mọi công cụ dòng lệnh bạn dùng hằng ngày, gói lại dưới **một lệnh duy nhất**.
Gõ `td`, xem hết danh sách, không cần nhớ tên từng công cụ.

```bash
td vcs init            # khởi tạo kho mã nguồn
td use vcs             # đặt nhóm vcs làm mặc định (sau đó chỉ cần gõ td commit, td status)
td commit -m "..."     # ghi lại thay đổi (tương đương td vcs commit)
td config --list       # xem cấu hình
```

`devcli` viết bằng Go, dựng cây lệnh bằng [cobra](https://github.com/spf13/cobra).
Mỗi công cụ là một nhóm lệnh con, và mỗi nhóm tự quản lý toàn bộ dữ liệu của
nó trong thư mục riêng.

**Vì sao tên là devcli mà chạy bằng `td`?** `devcli` là tên dự án và tên tệp
thực thi. `td` là tên lệnh gọi trên terminal, ngắn để gõ nhanh. Muốn đổi thì
sửa một dòng trong [`scripts/build_binaries.sh`](scripts/build_binaries.sh).

Hiện tại có 2 nhóm chính:
- `td vcs`: quản lý phiên bản mã nguồn cục bộ.
- `td sys`: các tiện ích hệ thống (ls, mkdir, rm, cp, mv, grep, wc...).
Các nhóm khác sẽ được bổ sung theo cùng khuôn mẫu.

## Yêu cầu

- Go 1.23 trở lên, kiểm tra bằng `go version`.
- Không cần gì khác lúc chạy. devcli là một tệp thực thi duy nhất, không
  phụ thuộc thư viện hệ thống.

## Tài liệu chi tiết

Mọi hướng dẫn cụ thể đã được chia nhỏ. Vui lòng tìm và đọc các tài liệu tương ứng bên trong thư mục `docs/`.