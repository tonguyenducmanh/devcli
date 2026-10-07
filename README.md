# td

Một ứng dụng dòng lệnh viết bằng Go, dùng
[cobra](https://github.com/spf13/cobra) để dựng cây lệnh. Mọi công cụ được
gói dưới một lệnh gốc `td`, mỗi công cụ là một nhóm lệnh con.

Hiện tại có nhóm `td vcs` — quản lý phiên bản mã nguồn cục bộ. Các nhóm khác
sẽ bổ sung theo cùng khuôn mẫu.

## Yêu cầu

- Go 1.23 trở lên (kiểm tra bằng `go version`).
- Không cần thư viện nào khác lúc chạy, td là một binary tĩnh duy nhất.

## Build

```bash
./scripts/build_all.sh               # build Mac, Linux, Windows vào out/
./scripts/check.sh              # định dạng + phân tích tĩnh + kiểm thử + đối chiếu tài liệu
go test ./...                # chạy kiểm thử
```

Kết quả nằm trong `out/`, tên file chứa kèm phiên bản:

```
out/td-mac-arm-0.1.0
out/td-mac-intel-0.1.0
out/td-linux-0.1.0
out/td-windows-0.1.0.exe
```

### Phiên bản

Phiên bản lấy từ file `scripts/VERSION`, là nguồn duy nhất cho toàn bộ dự án. Ghi đè tạm khi build mà không cần sửa file:

```bash
VERSION=1.2.3 ./scripts/build_all.sh
```

Phiên bản được gắn vào binary lúc biên dịch nên `td version` luôn cho biết đúng
bản đang chạy. Muốn phát hành bản mới thì sửa `scripts/VERSION`, chạy `./scripts/build_all.sh`,
rồi đẩy lên trang phát hành.

### Build bằng lệnh go thuần

```bash
go build -o td .
go install .            # cài vào $GOPATH/bin
```

## Cài đặt

td là một binary độc lập, nên "cài đặt" chỉ là chép file executable vào một
thư mục nằm trong `PATH`.

```bash
# cài cho bản dùng thử
install -m 755 out/td-mac-arm-0.1.0 ~/bin/td

# hoặc chỉ cần dùng trong dự án mà không cài
./out/td-mac-arm-0.1.0 vcs status
```

### macOS

```bash
install -m 755 out/td-mac-arm-0.1.0 ~/bin/td
echo 'export PATH="$PATH:$HOME/bin"' >> ~/.zshrc
source ~/.zshrc
td version
```

Nếu tải binary bằng trình duyệt, macOS có thể chặn vì không có chữ ký:

```bash
xattr -d com.apple.quarantine ~/bin/td
```

### Linux

```bash
# không cần sudo
install -m 755 out/td-linux-0.1.0 ~/.local/bin/td
echo 'export PATH="$PATH:$HOME/.local/bin"' >> ~/.bashrc

# hoặc cài cho toàn hệ thống
sudo install -m 755 out/td-linux-0.1.0 /usr/local/bin/td
```

### Windows

```powershell
New-Item -ItemType Directory -Force "$env:LOCALAPPDATA\td"
Copy-Item out\td-windows-0.1.0.exe "$env:LOCALAPPDATA\td\td.exe"

# thêm vào Path của người dùng, vĩnh viễn
[Environment]::SetEnvironmentVariable(
  "Path",
  [Environment]::GetEnvironmentVariable("Path", "User") + ";$env:LOCALAPPDATA\td",
  "User"
)
```

Mở cửa sổ PowerShell mới rồi kiểm tra bằng `td version`.

### Cài bằng trình quản lý gói

Khi phát hành trên trang Releases, có thể đưa td vào các trình quản lý gói phổ
biến, người dùng cập nhật và gỡ cài đặt bằng một lệnh:

| Hệ điều hành | Trình quản lý | Lệnh cài | Lệnh cập nhật | Lệnh gỡ |
| --- | --- | --- | --- | --- |
| macOS | Homebrew | `brew install td` | `brew upgrade td` | `brew uninstall td` |
| Windows | Scoop | `scoop install td` | `scoop update td` | `scoop uninstall td` |
| Windows | winget | `winget install td` | `winget upgrade td` | `winget uninstall td` |
| Linux | Go | `go install ...@latest` | chạy lại lệnh | xoá trong `$GOPATH/bin` |

## Cập nhật

Vì cấu hình nằm ở `~/.config/td/config` và dữ liệu mỗi dự án nằm trong thư mục
`.tdx`, **cài đè binary là đủ để cập nhật**, không mất dữ liệu:

```bash
# tải bản mới, chép đè lên binary cũ
install -m 755 out/td-linux-1.2.0 /usr/local/bin/td
td version        # xác nhận đã lên phiên bản mới
```

Trên Windows:

```powershell
Copy-Item -Force out\td-windows-1.2.0.exe "$env:LOCALAPPDATA\td\td.exe"
```

Dùng trình quản lý gói thì không cần làm gì, chỉ cần lệnh `upgrade`.

## Gỡ cài đặt

```bash
# xoá binary
rm ~/bin/td            # hoặc: sudo rm /usr/local/bin/td

# xoá thêm cấu hình toàn cục
rm -rf ~/.config/td

# xoá thêm kho td trong các dự án (nếu muốn)
find . -type d -name .tdx -prune -exec rm -rf {} +
```

Chạy `td` sau khi gỡ không còn ý nghĩa, nhưng thư mục `.tdx` trong dự án vẫn là
dữ liệu thô. Xoá hay giữ là tuỳ bạn, nên lệnh xoá được tách riêng khỏi bước gỡ
binary.

## Bắt đầu nhanh

```bash
td config --global user.name "Tên của bạn"
td config --global user.email "ten@example.com"

td vcs init              # khởi tạo kho tại thư mục hiện tại
td vcs add .
td vcs commit -m "tin nhắn đầu tiên"
td vcs log --oneline
```

## Cấu trúc dự án

```
main.go                      điểm khởi động
cmd/                         cây lệnh
  root.go                    lệnh gốc `td`, đăng ký các nhóm công cụ
  context.go                 tiện ích mở kho cho các lệnh con
  vcs.go                     nhóm lệnh `td vcs`
  vcs_*.go                   mỗi tệp một nhóm lệnh nhỏ của `td vcs`
internal/vcs/
  object/                    các loại dữ liệu blob, tree, commit, tag và cách băm
  storage/                   lưu object nén zlib, tham chiếu và nhật ký
  index/                     vùng chuẩn bị (staging area)
  worktree/                  đọc ghi cây làm việc, quy tắc bỏ qua tệp
  repo/                      lớp trừu tượng kho mã nguồn, cây nội dung, trạng thái
  diff/                      thuật toán Myers và cách dựng hunk
  merge/                     hợp nhất ba phía theo kiểu diff3
  ops/                       nghiệp vụ cấp cao cho từng lệnh
  config/                    đọc ghi file cấu hình
```

Thứ tự phụ thuộc một chiều: `object` đứng độc lập, `storage` và `index` dùng
`object`, `repo` dùng các lớp phía trên, `ops` dùng `repo`, `cmd` dùng `ops`.
Mỗi tầng chỉ biết tới tầng dưới nó nên có thể thay đổi nội bộ mà không ảnh
hưởng tầng trên.

## Cách lưu dữ liệu

Dữ liệu của một kho nằm trong thư mục `.tdx` cạnh dự án. Mỗi đơn vị dữ liệu
(object) được ghi thành một tệp riêng, nén zlib, đặt theo mã băm của nó:

- **Mã băm**: SHA1 của chuỗi `"<loại> <kích thước>"` theo sau bởi một byte NUL
  rồi tới nội dung. Byte NUL phân tách rõ phần mô tả với nội dung.
- **Blob**: nội dung tệp, giữ nguyên như trên đĩa.
- **Tree**: nối các entry `<chế độ> <tên>` với byte NUL rồi 20 byte mã băm.
  Tên thư mục được so sánh kèm dấu `/` phía sau khi sắp xếp.
- **Commit**: các dòng `tree`, `parent`, `author`, `committer`, một dòng trống
  rồi tới nội dung thông điệp.
- **Vùng chuẩn bị**: danh sách entry sắp xếp theo tên, kèm checksum SHA1 ở
  cuối tệp.

Nội dung giống nhau luôn cho cùng một mã băm, nên kho không lưu trùng dữ liệu
và mọi thay đổi đều truy vết được về đúng nguồn.

## Danh sách lệnh của `td vcs`

| Lệnh | Công dụng |
| --- | --- |
| `init` | tạo kho, tuỳ chọn `--initial-branch` |
| `status` (`st`) | xem ba vùng thay đổi và các xung đột |
| `add` | đưa thay đổi vào vùng chuẩn bị, có `-A` và `-u` |
| `commit` | `-m` nhiều lần, `--amend`, `-a`, `--allow-empty` |
| `log` (`lg`) | `-n`, `--oneline`, `-p`, `--stat`, `--all`, lọc theo đường dẫn |
| `diff` (`df`) | `--staged`, `A..B`, `A...B`, `-U`, `--stat`, `--name-only` |
| `show` | chi tiết một commit |
| `branch` (`br`) | `-l`, `-d`, `-D`, `-m` đổi tên, `-v` |
| `checkout` (`co`) / `switch` (`sw`) | `-b`, `-c`, `--detach` |
| `restore` (`rst`) | `--staged`, `--worktree`, `--source` |
| `merge` (`mg`) | `--no-ff`, `--ff-only`, `--squash`, `--continue`, `--abort` |
| `rebase` (`rb`) | `--onto`, `--branch`, `--continue`, `--abort`, `--skip` |
| `cherry-pick` (`cp`) | nhiều commit, `-n`, `--continue`, `--abort`, `--skip` |
| `revert` (`rv`) | hoàn tác commit, các cờ xử lý xung đột |
| `stash` (`st`) | `-u`, `list`, `apply`, `pop`, `drop`, `clear` |
| `reset` (`rs`) | `--soft`, `--mixed`, `--hard`, hoặc theo đường dẫn |
| `tag` | `-a`, `-m`, `-d`, `-l` |
| `rm` / `mv` | `--cached` cho `rm` |
| `reflog` | lịch sử di chuyển của một tham chiếu |
| `hash-object` / `cat-file` | xem mã băm và nội dung object |
| `fsck` | kiểm tra toàn vẹn kho và các tham chiếu |

Lệnh cấu hình:

| Lệnh | Công dụng |
| --- | --- |
| `td config` | đọc và ghi cấu hình, `-g` cho toàn cục, `-l` liệt kê, `--unset` xoá |
| `td version` | hiện phiên bản |

Mọi lệnh trong `td vcs` nhận `-C <thư mục>` để chạy ở thư mục khác.

## Giải quyết xung đột

Khi `merge`, `rebase`, `cherry-pick` hoặc `revert` gặp xung đột, trạng thái
được ghi lại để có thể tiếp tục hoặc huỷ:

```bash
td vcs merge <nhánh>       # báo các tệp xung đột
# sửa tay tệp, chọn nội dung đúng
td vcs add <tệp>
td vcs commit              # hoặc: td vcs merge --continue
td vcs merge --abort       # quay lại trạng thái trước khi hợp nhất
```

Ba bản của một tệp xung đột được giữ ở ba stage khác nhau: stage 1 là bản chung,
stage 2 là phía của mình, stage 3 là phía đối tác.

## Bỏ qua tệp

Đặt mẫu vào `.tdxignore` ở gốc dự án hoặc `.tdx/info/exclude`:

```
build/
*.log
!giữ-lại.log
docs/*.tmp
**/cache/
```

Dấu `#` cho chú thích, dấu `!` để phủ định lại quy tắc trước, dấu `/` cuối mẫu
để chỉ áp dụng cho thư mục, `*` không vượt qua `/`, `**` vượt nhiều cấp.

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

## Tài liệu cho trợ lý lập trình

| Tệp | Dành cho |
| --- | --- |
| `agents/README.md` | Mô tả thư mục tài liệu cho trợ lý lập trình |
| `agents/AGENTS.md` | Kiến trúc, bất biến, quy trình và quy ước viết mã |
| `agents/llms.txt` | Tóm tắt ngắn cho trợ lý AI |
| `agents/cli/` | Tham chiếu từng lệnh, sinh tự động |
| `CONTRIBUTING.md` | Quy trình đóng góp |
| `tests/README.md` | Vì sao kiểm thử nằm ở thư mục riêng |

`agents/cli/` được sinh từ cây lệnh bằng `internal/tools/docgen`, không sửa tay:

```bash
./scripts/build_agent_docs.sh
```

Mỗi tệp Markdown có cấu trúc ổn định: mô tả, cú pháp, các ví dụ, các cờ.
Nhờ vậy người đọc và trợ lý AI nắm được chính xác từng lệnh làm gì mà không
cần chạy thử. `./scripts/check.sh` sẽ báo nếu tài liệu lệch với câu lệnh.

## Thêm nhóm công cụ mới

1. Tạo gói nghiệp vụ trong `internal/<tên>/` nếu nhóm mới cần kiểu dữ liệu riêng.
2. Tạo `cmd/<tên>.go` trả về một `*cobra.Command` làm nhóm lệnh.
3. Gọi `registerToolGroup(...)` trong `init()` của `cmd/root.go`.

Nhóm `vcs` chỉ là một ví dụ. Các nhóm khác theo đúng khuôn mẫu đó mà không
cần đụng tới phần version control.