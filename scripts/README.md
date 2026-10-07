# scripts

Toàn bộ cấu hình build và script của dự án nằm trong thư mục này, giữ cho
thư mục gốc chỉ còn mã nguồn và tài liệu.

## Danh sách

| Tệp | Vai trò |
| --- | --- |
| `VERSION` | Số phiên bản của ứng dụng, nguồn duy nhất |
| `build_all.sh` | Điểm vào chính để build |
| `build_binaries.sh` | Build binary cho Mac, Linux, Windows vào `out/` |
| `build_agent_docs.sh` | Sinh lại tài liệu lệnh trong `agents/cli/` |
| `check.sh` | Kiểm tra trọn vẹn trước khi đóng góp |
| `td_version.sh` | Các hàm đọc phiên bản và tạo ldflags, được `source` bởi script khác |
| `golangci.yml` | Cấu hình phân tích mã |

## Cách chạy

Tất cả script đều tự tìm thư mục gốc qua vị trí của chính nó, nên chạy được
từ bất kỳ thư mục nào:

```bash
./scripts/build_all.sh
./scripts/check.sh
```

## Đổi phiên bản

Sửa nội dung file `VERSION`:

```bash
echo "1.2.3" > scripts/VERSION
```

Rồi build lại:

```bash
./scripts/build_all.sh
```

Tên file trong `out/` và kết quả của `td version` đều lấy từ file này.

Muốn thử một phiên bản khác mà không sửa file:

```bash
VERSION=9.9.9 ./scripts/build_all.sh
```

## Về `td_version.sh`

Đây là tệp duy nhất được `source` chứ không phải chạy như script:

```sh
. "$(dirname "$0")/td_version.sh"
VERSION=$(td_get_version)
LDFLAGS=$(td_get_ldflags "$VERSION")
```

Các hàm cung cấp:

- `td_find_scripts_dir` — vị trí thư mục `scripts`.
- `td_get_version` — đọc phiên bản, ưu tiên biến môi trường `VERSION`.
- `td_get_ldflags` — tạo cờ ldflags gắn phiên bản vào binary.

Lưu ý: biến `Version` trong `cmd/root.go` phải khai báo bằng `var` thì ldflags
mới ghi được. Đừng đổi thành `const`.

## Về `golangci.yml`

Vì không nằm ở thư mục gốc nên khi chạy phải chỉ định đường dẫn:

```bash
golangci-lint run -c scripts/golangci.yml
```

Cài golangci-lint:

```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```