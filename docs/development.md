# Hướng dẫn đóng góp và phát triển (Development)

## Build

```bash
./build_all.sh
```

Lệnh trên sinh lại tài liệu lệnh trong `docs/agents/cli/`, rồi build cho mọi nền
tảng khai báo trong `scripts/build_binaries.sh`. Kết quả nằm ở thư mục `out/`,
tên file chứa kèm số phiên bản:

```
out/td-devcli-mac-arm-0.1.0          macOS trên chip Apple Silicon
out/td-devcli-linux-0.1.0            Linux 64 bit
out/td-devcli-windows-0.1.0.exe      Windows 64 bit
```

Chỉ cần một nền tảng thì nêu tên:

```bash
./scripts/build_binaries.sh mac-arm windows   # build 2 nền tảng
./scripts/build_binaries.sh --list           # xem danh sách có thể build
```

Muốn thêm hoặc bỏ một nền tảng, sửa mục `TARGETS` ở đầu `scripts/build_binaries.sh`,
không cần đụng vào phần logic của script. Cùng chỗ đó còn cấu hình tên lệnh,
tên tệp thực thi, thư mục kết quả và các cờ biên dịch. Chi tiết ở
[`scripts/README.md`](../scripts/README.md).

### Phiên bản

Số phiên bản nằm ở biến `VERSION` trong phần cấu hình của
[`scripts/build_binaries.sh`](../scripts/build_binaries.sh), là nguồn duy nhất cho
toàn bộ dự án. Đổi phiên bản thì sửa dòng đó rồi build lại:

```sh
VERSION=1.2.3
```

Phiên bản được gắn vào tệp thực thi lúc biên dịch, nên `tm version` luôn cho
biết đúng bản đang chạy, khớp với tên file trong `out/`.

### Build bằng lệnh go thuần

```bash
go build -o tm .    # cho máy đang chạy
go install .        # cài vào $GOPATH/bin
```

## Kiểm thử

```bash
go test ./...
```

Bao gồm:

- Kiểm thử tính đúng đắn của thuật toán diff trên nhiều trường hợp biên.
- Kiểm thử theo dòng cho nghiệp vụ của các lệnh: commit, nhánh, hợp nhất,
  hợp nhất lại, hoàn tác, lưu tạm, đặt lại, khôi phục tệp.
- Kiểm thử vòng đọc-ghi cho mã băm, cây, vùng chuẩn bị và các tham chiếu.
- Kiểm thử bảo đảm công cụ tự trị: mã nguồn không gọi chương trình ngoài nào,
  và bản nhị phân vẫn chạy được khi `PATH` bị đặt rỗng.

Để kiểm tra toàn diện trước khi commit, chạy script:
```bash
./scripts/check.sh
```

## Thêm nhóm công cụ mới

1. Tạo gói nghiệp vụ trong `internal/<tên>/` nếu nhóm mới cần kiểu dữ liệu riêng.
2. Tạo `cmd/<tên>.go` trả về một `*cobra.Command` làm nhóm lệnh.
3. Gọi `registerToolGroup(...)` trong `init()` của `cmd/root.go`.

Nhóm `vcs` chỉ là một ví dụ. Các nhóm khác theo đúng khuôn mẫu đó mà không
cần đụng tới phần version control.

## Tài liệu cho trợ lý lập trình

| Tệp | Dành cho |
| --- | --- |
| `docs/agents/README.md` | Mô tả thư mục tài liệu cho trợ lý lập trình |
| `docs/agents/AGENTS.md` | Kiến trúc, bất biến, quy trình và quy ước viết mã |
| `docs/agents/cli/` | Tham chiếu từng lệnh, sinh tự động |
| `CONTRIBUTING.md` | Quy trình đóng góp |
| `tests/README.md` | Vì sao kiểm thử nằm ở thư mục riêng |

`docs/agents/cli/` được sinh từ cây lệnh bằng `internal/tools/docgen`, không sửa tay:

```bash
./scripts/build_agent_docs.sh
```

Mỗi tệp Markdown có cấu trúc ổn định: mô tả, cú pháp, các ví dụ, các cờ.
Nhờ vậy người đọc và trợ lý AI nắm được chính xác từng lệnh làm gì mà không
cần chạy thử. `./scripts/check.sh` sẽ báo nếu tài liệu lệch với câu lệnh.
