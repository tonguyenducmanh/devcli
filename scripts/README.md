# scripts

Toàn bộ cấu hình build và script của dự án nằm trong thư mục này, giữ cho
thư mục gốc chỉ còn mã nguồn và tài liệu.

## Danh sách

| Tệp | Vai trò |
| --- | --- |
| `build_binaries.sh` | Toàn bộ cấu hình build nằm ở đầu tệp này, phần còn lại là logic |
| `build_agent_docs.sh` | Sinh lại tài liệu lệnh trong `agents/cli/` |
| `build_extension.sh` | Đóng gói tiện ích VS Code thành tệp `.vsix` trong `out/` |
| `check.sh` | Kiểm tra trọn vẹn trước khi đóng góp |
| `remove_old_tags.sh` | Xoá các tag cũ trong kho git, giữ lại danh sách tag chỉ định |

Điểm vào để build không nằm ở đây mà ở thư mục gốc, tên `build_all.sh`. Nó gọi
`build_agent_docs.sh`, `build_binaries.sh` rồi `build_extension.sh`.

`build_extension.sh` là bước duy nhất cần Node.js. Máy không có Node thì nó in
một dòng bỏ qua rồi thoát với mã 0, để phần build Go không bị ảnh hưởng.

Hai script cuối cùng là công cụ bảo trì, không liên quan tới build:
`check.sh` kiểm tra mã nguồn, `remove_old_tags.sh` dọn tag. Cấu hình build nằm
ngay trong `build_binaries.sh`, không tách tệp riêng. Thêm nền tảng, đổi phiên
bản, đổi tên lệnh hay đổi cờ biên dịch đều chỉ sửa một tệp duy nhất, đọc cũng
chỉ một chỗ.

## Cách chạy

Từ thư mục gốc của kho:

```bash
./build_all.sh                        # sinh tài liệu rồi build mọi nền tảng
./build_all.sh mac-arm linux          # chỉ build một vài nền tảng
./scripts/build_binaries.sh --list    # xem danh sách nền tảng
./scripts/check.sh                    # kiểm tra trước khi đóng góp
```

Mọi script đều tự tìm thư mục gốc qua vị trí của chính nó, nên chạy được từ bất
kỳ thư mục nào.

`build_all.sh` gọi lần lượt `build_agent_docs.sh` rồi `build_binaries.sh`. Muốn
build tệp thực thi mà không sinh lại tài liệu thì gọi `build_binaries.sh` trực
tiếp.

## Cấu hình

Phần cấu hình của `build_binaries.sh` nằm giữa hai đường kẻ `═══`, gồm bảy
biến khai báo thuần:

| Biến | Ý nghĩa |
| --- | --- |
| `CMD_NAME` | Tên gọi lệnh trên terminal, ví dụ `tm vcs status` |
| `REPO_URL` | Nơi phát hành, in ở cuối phần trợ giúp |
| `AUTHOR` | Tên tác giả, in ở lệnh version và cuối phần trợ giúp |
| `VERSION` | Số phiên bản của ứng dụng |
| `APP_NAME` | Tiền tố cho tên tệp trong `out/` |
| `OUT_DIR` | Thư mục chứa kết quả build |
| `BUILD_FLAGS` | Cờ biên dịch, mặc định `-s -w` |
| `CGO_ENABLED` | Đặt `0` để tệp thực thi tĩnh thật sự |

Cùng phần còn có `TARGETS`, danh sách nền tảng cần build.

`CMD_NAME`, `REPO_URL` và `AUTHOR` còn được khai lại làm giá trị mặc định trong
`cmd/root.go`, để chạy `go build` thẳng cũng ra kết quả đúng. Hai nơi phải khớp
nhau, có kiểm thử chặn. `AUTHOR` có dấu cách nên `build_binaries.sh` bọc nó
trong nháy đơn khi tạo ldflags.

### Đổi phiên bản

Sửa dòng `VERSION`, đây là nguồn duy nhất:

```sh
VERSION=1.2.3
./build_all.sh
```

Phiên bản được gắn vào tệp thực thi lúc biên dịch, nên `tm version` luôn khớp
với tên file trong `out/`.

### Đổi tên lệnh

Sửa `CMD_NAME`:

```sh
CMD_NAME=devtool
```

Đổi thành `devtool` thì chạy bằng `devtool vcs status`, phần trợ giúp cũng tự
đổi theo.

### Thêm hoặc bỏ một nền tảng

Sửa mục `TARGETS`. Mỗi dòng có dạng:

```
tên | goos | goarch | đuôi tệp
```

Danh sách mặc định gồm macOS Apple Silicon, Linux và Windows:

```
mac-arm|darwin|arm64|
linux|linux|amd64|
windows|windows|amd64|.exe
```

Ví dụ thêm bản cho Linux trên chip ARM:

```
mac-arm|darwin|arm64|
linux|linux|amd64|
linux-arm|linux|arm64|
windows|windows|amd64|.exe
```

Cột đuôi tệp bỏ trống nếu không cần. Xem kết quả sau khi sửa:

```bash
./scripts/build_binaries.sh --list
```

### Đổi cờ biên dịch

Sửa `BUILD_FLAGS` và `CGO_ENABLED`. Cờ gắn phiên bản và các biến toàn cục
(`-X ...cmd.Version`) do `make_ldflags` tự thêm, không khai báo ở đây.

Lưu ý: các biến `AppName`, `Version` và `RepoURL` trong `cmd/root.go` phải
khai báo bằng `var` thì `ldflags` mới ghi được. Đừng đổi thành `const`.

## Kiểm tra mã nguồn

`check.sh` chạy bốn bước, tất cả phải qua:

1. `gofmt -s -l .` — định dạng.
2. `go vet ./...` — phân tích tĩnh.
3. `go test ./...` — kiểm thử trong `tests/`.
4. Đối chiếu `agents/cli/` với cây lệnh.

## Đóng gói tiện ích VS Code

```bash
./scripts/build_extension.sh             # build rồi đóng gói vào out/
./scripts/build_extension.sh --package   # chỉ đóng gói, dùng lại kết quả build
```

Kết quả là `out/td-devcli-vscode-<phiên bản>.vsix`. Một tệp này cài được trên mọi
nền tảng vì nó chỉ chứa JavaScript đã biên dịch. Máy đích cần có lệnh `tm` riêng,
đặt trong `PATH`, `~/go/bin` hoặc qua cấu hình `tm.path`. Chi tiết ở
[`editors/vscode/README.md`](../editors/vscode/README.md).

Phiên bản của tiện ích nằm trong `editors/vscode/package.json`, còn tên tệp đầu
ra đặt theo `APP_NAME` trong `build_binaries.sh`.

## Sinh trang man

Chỉ cần khi phát hành, không commit trang man vào kho mã:

```bash
go run ./internal/tools/docgen -out out/man -format man
```

Tên tác giả và nơi phát hành lấy từ `AUTHOR` và `REPO_URL` qua biến toàn cục,
nên không phải nhập lại.

Muốn phân tích sâu hơn thì dùng `golangci-lint`, nhưng tự cài vì dự án không
kèm cấu hình:

```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
golangci-lint run
```
