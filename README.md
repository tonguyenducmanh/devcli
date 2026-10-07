# td

Một ứng dụng dòng lệnh viết bằng Go, dùng
[cobra](https://github.com/spf13/cobra) để dựng cây lệnh. Mọi công cụ được
gói dưới một lệnh gốc `td`, mỗi công cụ là một nhóm lệnh con.

Hiện tại có nhóm `td vcs` — quản lý phiên bản mã nguồn cục bộ. Các nhóm khác
sẽ bổ sung theo cùng khuôn mẫu.

## Yêu cầu

- Go 1.23 trở lên, kiểm tra bằng `go version`.
- Không cần gì khác lúc chạy. td là một tệp thực thi duy nhất.

## Build

```bash
./scripts/build_all.sh
```

Lệnh trên sinh lại tài liệu lệnh trong `agents/cli/`, rồi build cho mọi nền
tảng khai báo trong `scripts/build_binaries.sh`. Kết quả nằm ở thư mục `out/`,
tên file chứa kèm số phiên bản:

```
out/td-mac-arm-0.1.0          macOS trên chip Apple Silicon
out/td-linux-0.1.0            Linux 64 bit
out/td-windows-0.1.0.exe      Windows 64 bit
```

Chỉ cần một nền tảng thì nêu tên:

```bash
./scripts/build_binaries.sh mac-arm windows   # build 2 nền tảng
./scripts/build_binaries.sh --list           # xem danh sách có thể build
```

Muốn thêm hoặc bỏ một nền tảng, sửa mục `TARGETS` ở đầu `scripts/build_binaries.sh`,
không cần đụng vào phần logic của script. Cùng chỗ đó còn cấu hình tên lệnh,
tên tệp thực thi, thư mục kết quả và các cờ biên dịch. Chi tiết ở
[`scripts/README.md`](scripts/README.md).

### Phiên bản

Số phiên bản nằm ở biến `VERSION` trong phần cấu hình của
[`scripts/build_binaries.sh`](scripts/build_binaries.sh), là nguồn duy nhất cho
toàn bộ dự án. Đổi phiên bản thì sửa dòng đó rồi build lại:

```sh
VERSION=1.2.3
```

Phiên bản được gắn vào tệp thực thi lúc biên dịch, nên `td version` luôn cho
biết đúng bản đang chạy, khớp với tên file trong `out/`.

### Build bằng lệnh go thuần

```bash
go build -o td .    # cho máy đang chạy
go install .        # cài vào $GOPATH/bin
```

## Dùng tệp thực thi trên từng hệ điều hành

Chọn đúng tệp trong `out/` theo hệ điều hành và kiến trúc máy, rồi chép vào một
thư mục nằm trong `PATH`.

### macOS

Bản dựng cho Apple Silicon, kiểm tra bằng `uname -m` cho ra `arm64`.

```bash
install -m 755 out/td-mac-arm-0.1.0 ~/bin/td

echo 'export PATH="$PATH:$HOME/bin"' >> ~/.zshrc
source ~/.zshrc
td version
```

Muốn cài cho toàn hệ thống thay vì chỉ tài khoản của bạn:

```bash
sudo install -m 755 out/td-mac-arm-0.1.0 /usr/local/bin/td
```

Nếu sao chép tệp bằng trình duyệt, macOS có thể chặn vì không có chữ ký:

```bash
xattr -d com.apple.quarantine ~/bin/td
```

### Linux

Bản dựng cho máy 64 bit thông thường, kiểm tra bằng `uname -m` cho ra
`x86_64`.

```bash
install -m 755 out/td-linux-0.1.0 ~/.local/bin/td

echo 'export PATH="$PATH:$HOME/.local/bin"' >> ~/.bashrc
source ~/.bashrc
```

Hoặc cài cho toàn hệ thống:

```bash
sudo install -m 755 out/td-linux-0.1.0 /usr/local/bin/td
```

Kiểm tra tệp đã quyền chạy chưa:

```bash
ls -l ~/bin/td          # phần cuối phải là rwxr-xr-x
chmod +x ~/bin/td       # nếu chưa có
```

### Windows

Mở PowerShell ở thư mục chứa thư mục `out/`:

```powershell
New-Item -ItemType Directory -Force "$env:LOCALAPPDATA\td"
Copy-Item out\td-windows-0.1.0.exe "$env:LOCALAPPDATA\td\td.exe"

# thêm vào Path của người dùng, có hiệu lực vĩnh viễn
[Environment]::SetEnvironmentVariable(
  "Path",
  [Environment]::GetEnvironmentVariable("Path", "User") + ";$env:LOCALAPPDATA\td",
  "User"
)
```

Mở một cửa sổ PowerShell mới rồi kiểm tra:

```powershell
td version
```

Windows không cần cấp quyền thực thi như trên macOS và Linux. Nếu PowerShell chặn
việc chạy tệp, xem [Execution Policy](https://learn.microsoft.com/powershell/module/microsoft.powershell.security/set-executionpolicy).

### Dùng tạm mà không cài

Không cần chép vào `PATH` cũng chạy được, gọi thẳng tệp thôi:

```bash
./out/td-mac-arm-0.1.0 vcs status
```

### Thay bằng bản mới

td không lưu gì cạnh tệp thực thi, nên chép tệp mới đè lên tệp cũ là xong.
Cấu hình nằm ở `~/.config/td/config` và dữ liệu mỗi dự án nằm trong thư mục
`.tdx`, cả hai đều không bị ảnh hưởng.

```bash
install -m 755 out/td-linux-1.2.0 ~/.local/bin/td
```

```powershell
Copy-Item -Force out\td-windows-1.2.0.exe "$env:LOCALAPPDATA\td\td.exe"
```

### Gỡ bỏ

```bash
rm ~/bin/td                    # hoặc ~/.local/bin/td, /usr/local/bin/td
```

Tệp thực thi là toàn bộ phần mềm, xoá nó là xong. Nếu muốn xoá nốt cấu hình:

```bash
rm -rf ~/.config/td
```

Dữ liệu kho mã nguồn trong thư mục `.tdx` của từng dự án không bị đụng tới,
vì đó là lịch sử công việc của bạn. Xoá thủ công khi không cần nữa:

```bash
find . -type d -name .tdx -prune -exec rm -rf {} +
```

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