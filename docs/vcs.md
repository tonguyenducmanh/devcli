# Quản lý phiên bản (td vcs)

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

Lệnh cấu hình và lệnh chung:

| Lệnh | Công dụng |
| --- | --- |
| `td config` | đọc và ghi cấu hình, `-g` cho toàn cục, `-l` liệt kê, `--unset` xoá |
| `td use` | đặt nhóm công cụ làm mặc định để gõ lệnh ngắn hơn (vd: `td use vcs`) |
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

td không tự đoán tệp nào là tạm, tệp nào là dữ liệu do trình biên dịch sinh ra.
Muốn bỏ qua thì tạo tệp **`.tdxignore`** ở gốc dự án, mỗi dòng một mẫu:

```
*.log           bỏ qua mọi tệp kết thúc bằng .log, ở mọi cấp thư mục
build           bỏ qua thư mục build, ở mọi cấp
/build          chỉ bỏ qua thư mục build nằm ở gốc dự án
docs/*.tmp      bỏ qua tệp .tmp nằm trong thư mục docs
**/cache/       bỏ qua thư mục cache, ở mọi cấp
!giữ-lại.log    phủ định lại quy tắc trước, tệp này lại được theo dõi
# ghi chú       dòng bắt đầu bằng dấu # là chú thích
```

Dấu `/` ở cuối mẫu nói mẫu đó chỉ áp dụng cho thư mục. Dấu `*` không vượt qua
dấu `/`, dấu `**` vượt được nhiều cấp. Quy tắc ở dưới thắng quy tắc ở trên.

### Đặt quy tắc trong thư mục con

`.tdxignore` có thể đặt ở bất kỳ thư mục nào, và quy tắc trong đó **chỉ áp dụng
bên trong thư mục chứa nó**, không lan sang nơi khác:

```
.tdxignore              *.log          bỏ qua mọi tệp .log, mọi cấp
sub/.tdxignore          tmp/           chỉ bỏ qua thư mục sub/tmp và bên dưới
sub/deep/.tdxignore     !keep.log      trong sub/deep thì giữ lại keep.log
```

Quy tắc sâu hơn nạp sau nên thắng quy tắc ở cấp trên.

### Hai tệp ignore

| Tệp | Phạm vi | Có commit không |
| --- | --- | --- |
| `.tdxignore` | Cả nhóm cùng dùng, đặt ở gốc hoặc thư mục con | Nên commit |
| `.tdx/info/exclude` | Riêng máy này, nằm trong `.tdx` | Không, tự sinh khi `init` |

### Xem quy tắc đang có

```bash
td vcs ignore          # in nội dung mọi tệp ignore kèm đường dẫn
td vcs ignore --help   # hướng dẫn viết mẫu, in ra luôn nếu kho chưa có quy tắc
```

Tệp bị bỏ qua sẽ không xuất hiện trong `td vcs status` và không được
`td vcs add` đưa vào vùng chuẩn bị.

Xem thêm trong phần trợ giúp:

```bash
td vcs --help
td vcs ignore
```
