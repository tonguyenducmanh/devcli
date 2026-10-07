# AGENTS.md

Tài liệu dành cho trợ lý lập trình làm việc trên kho mã này. Đọc tệp này trước
khi sửa bất kỳ điều gì.

Xem [`README.md`](README.md) trong thư mục này để biết từng tệp tài liệu dành
cho ai và khi nào cần sinh lại.

## Dự án là gì

`td` là một ứng dụng dòng lệnh viết bằng Go, dựng cây lệnh bằng
[spf13/cobra](https://github.com/spf13/cobra). Lệnh gốc là `td`, mỗi nhóm
công cụ là một lệnh con.

Nhóm công cụ hiện có: `td vcs` (quản lý phiên bản mã nguồn cục bộ),
`td config`, `td version`.

## Bất biến kiến trúc

Bốn quy tắc này được kiểm chứng bằng kiểm thử, vi phạm làm kiểm thử đỏ.

1. **Không gọi chương trình ngoài.** Mã nguồn không được import `os/exec` hay
   gọi tiến trình khác. Mọi thao tác phải nằm gọn trong tiến trình của td.
   Có kiểm thử ở `tests/architecture/` quét toàn bộ mã nguồn để chặn điều này.
2. **Dữ liệu chỉ nằm trong `.tdx`.** Không đọc hay ghi thư mục dữ liệu của
   công cụ khác.
3. **Phụ thuộc đi trong một chiều.** `object` → `storage`/`index` → `repo`
   → `ops` → `cmd`. Một gói chỉ được biết tới gói ngay dưới nó.
4. **Mọi thông điệp người dùng bằng tiếng Việt**, kể cả phần mô tả trong
   `Short`, `Long`, `Example` và các bộ kiểm tra số đối số.

## Bản đồ mã nguồn

| Đường dẫn | Vai trò |
| --- | --- |
| `main.go` | Điểm khởi động, chỉ gọi `cmd.Execute()` |
| `cmd/root.go` | Lệnh gốc, cờ toàn cục, đăng ký nhóm công cụ |
| `cmd/output.go` | In ấn, màu ANSI, bộ kiểm tra số đối số tiếng Việt |
| `cmd/context.go` | `openRepo`: mở kho cho các lệnh con |
| `cmd/vcs*.go` | Khai báo lệnh của nhóm `vcs`, mỗi tệp một nhóm nhỏ |
| `internal/vcs/object/` | Kiểu dữ liệu blob, tree, commit, tag; tính mã băm |
| `internal/vcs/storage/` | Ghi đọc object nén zlib; quản lý tham chiếu và nhật ký |
| `internal/vcs/index/` | Vùng chuẩn bị, đọc ghi nhị phân có checksum |
| `internal/vcs/worktree/` | Đọc ghi tệp trên đĩa, quy tắc bỏ qua tệp |
| `internal/vcs/repo/` | Lớp trừu tắng kho: cây nội dung, trạng thái, phân giải tham chiếu |
| `internal/vcs/diff/` | Thuật toán Myers, căn dòng, dựng hunk |
| `internal/vcs/merge/` | Hợp nhất ba phía kiểu diff3 |
| `internal/vcs/ops/` | Nghiệp vụ cấp cao, mỗi tệp một nhóm lệnh |
| `internal/vcs/config/` | Đọc ghi cấu hình dạng mục và khoá |
| `internal/tools/docgen/` | Sinh tài liệu Markdown cho cây lệnh |
| `scripts/` | Mọi cấu hình và script build (xem `scripts/README.md`) |
| `tests/` | Toàn bộ mã kiểm thử, tách theo vùng nghiệp vụ |
| `agents/` | Tài liệu cho trợ lý lập trình: `AGENTS.md`, `llms.txt`, `cli/` |
| `scripts/VERSION` | Nguồn duy nhất của số phiên bản |

## Mô hình dữ liệu

Mỗi đơn vị dữ liệu gọi là *object*, lưu thành một tệp riêng nén zlib tại
`.tdx/objects/<2 ký tự đầu>/<38 ký tự sau>`.

- Mã băm là SHA1 của chuỗi `"<loại> <kích thước>"` + byte NUL + nội dung.
- **Blob**: nội dung tệp.
- **Tree**: danh sách entry `<chế độ> <tên>` + byte NUL + 20 byte mã băm.
  Tên thư mục so sánh kèm dấu `/` phía sau.
- **Commit**: các dòng `tree`, `parent`, `author`, `committer`, dòng trống,
  rồi thông điệp.
- Có ba vùng trạng thái: **HEAD** (con trỏ đang đứng), **vùng chuẩn bị**
  (`.tdx/index`) và **cây làm việc** (tệp trên đĩa).

Cây nội dung được xây từ vùng chuẩn bị bởi `repo.TreeFromIndex`, dựng đệ quy
theo `repo.WriteTree`.

## Thêm một lệnh mới

1. Thêm hàm nghiệp vụ vào tệp phù hợp trong `internal/vcs/ops/`. Hàm nhận
   `*repo.Repo` và một struct tuỳ chọn, trả về kết quả có kiểu rõ ràng.
2. Khai báo lệnh trong tệp `cmd/vcs_*.go` phù hợp. Bắt buộc có:
   - `Short`: một dòng ngắn.
   - `Long`: giải thích *tác dụng và lý do*, không chỉ liệt kê tham số.
   - `Example`: các lệnh mẫu chạy được, mỗi dòng một tình huống.
   - `Args`: dùng bộ kiểm tra trong `cmd/output.go` để thông báo lỗi tiếng Việt.
3. Thêm lệnh vào `cmd.AddCommand(...)` trong `cmd/vcs.go`.
4. Sinh lại tài liệu: `./scripts/build_agent_docs.sh`.

## Thêm một nhóm công cụ mới

1. Tạo gói nghiệp vụ trong `internal/<tên>/` nếu cần kiểu dữ liệu riêng.
2. Tạo `cmd/<tên>.go` trả về `*cobra.Command`, đặt `GroupID` phù hợp.
3. Thêm nhóm vào danh sách trong `rootCmd.AddGroup(...)` nếu dùng ID mới.
4. Gọi `registerToolGroup(...)` trong `init()` của `cmd/root.go`.

## Quy ước viết mã

- Chú thích và thông điệp người dùng bằng tiếng Việt.
- Chú thích giải thích *tại sao*, không lặp lại *cái gì* mà tên hàm đã nói.
- Tên biến, hàm, kiểu dữ liệu bằng tiếng Anh theo chuẩn Go.
- Không dùng gói ở tầng thấp hơn so với tầng đang làm.
- Trả về lỗi kèm ngữ cảnh bằng tiếng Việt, ví dụ
  `fmt.Errorf("không tìm thấy nhánh %s", name)`.

## Lệnh thường dùng

```bash
./scripts/build_all.sh                  # build Mac, Linux, Windows vào out/
./scripts/build_binaries.sh     # chỉ phần build binary
./scripts/build_agent_docs.sh   # sinh lại agents/cli
go test ./...                   # chạy kiểm thử
./scripts/check.sh              # kiểm tra trọn vẹn trước khi đóng góp
```

`./scripts/check.sh` chạy định dạng, phân tích tĩnh, kiểm thử, rồi sinh lại
tài liệu vào thư mục tạm để so với bản đã commit. Đây là lệnh nên chạy trước
mỗi lần đóng góp.

## Phiên bản

Số phiên bản nằm trong file `scripts/VERSION`, là nguồn duy nhất.
`scripts/build_all.sh` đọc file đó rồi gắn vào binary bằng cờ ldflags, nên `td version`
luôn khớp với tên file trong `out/`.

Ghi đè tạm mà không cần sửa file:

```bash
VERSION=1.2.3 ./scripts/build_all.sh
```

Biến `Version` trong `cmd/root.go` phải khai báo bằng `var` thì ldflags mới ghi
được, đừng đổi thành `const`.

## Về kiểm thử

Mã kiểm thử nằm trong `tests/`, mỗi thư mục con kiểm thử một vùng nghiệp vụ.
Các tệp dùng gói kiểm thử ngoài (`package diff_test`) nên chỉ gọi được thứ
mà gói đó xuất ra.

`tests/architecture/` kiểm tra các bất biến kiến trúc: không gọi chương trình
ngoài, phụ thuộc giữa các tầng đi một chiều, mọi lệnh đủ `Short`/`Long`/
`Example` bằng tiếng Việt, và binary chạy được với `PATH` rỗng.

Khi sửa một lỗi, thêm kiểm thử tái hiện lỗi đó trước khi sửa mã.

## Tài liệu tham chiếu lệnh

`agents/cli/` chứa tài liệu Markdown cho từng lệnh, sinh tự động từ cây lệnh
bằng `internal/tools/docgen`. Khi cần biết chính xác một lệnh nhận những gì và
làm gì, đọc tệp tương ứng thay vì đọc mã nguồn.

Tệp `agents/cli/td_vcs.md` là mục lục của nhóm `vcs`, bắt đầu từ đó.

Thư mục này không sửa tay được. Sau khi thay đổi cây lệnh, chạy:

```bash
./scripts/build_agent_docs.sh
```