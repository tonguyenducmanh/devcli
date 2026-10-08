# AGENTS.md

Tài liệu dành cho trợ lý lập trình làm việc trên kho mã này. Đọc tệp này trước
khi sửa bất kỳ điều gì.

Xem [`README.md`](README.md) trong thư mục này để biết từng tệp tài liệu dành
cho ai và khi nào cần sinh lại.

## Dự án là gì

`tm` là một ứng dụng dòng lệnh viết bằng Go, dựng cây lệnh bằng
[spf13/cobra](https://github.com/spf13/cobra). Lệnh gốc là `tm`, mỗi nhóm
công cụ là một lệnh con.

Nhóm công cụ hiện có: `tm vcs` (quản lý phiên bản mã nguồn cục bộ),
`tm config`, `tm version`.

## Bất biến kiến trúc

Bốn quy tắc này được kiểm chứng bằng kiểm thử, vi phạm làm kiểm thử đỏ.

1. **Không gọi chương trình ngoài.** Mã nguồn không được import `os/exec` hay
   gọi tiến trình khác. Mọi thao tác phải nằm gọn trong tiến trình của tm.
   Có kiểm thử ở `tests/architecture/` quét toàn bộ mã nguồn để chặn điều này.
2. **Dữ liệu chỉ nằm trong `.tmx`.** Không đọc hay ghi thư mục dữ liệu của
   công cụ khác.
3. **Phụ thuộc đi trong một chiều.** `object` → `storage`/`index` → `repo`
   → `ops` → `cmd`. Một gói chỉ được biết tới gói ngay dưới nó.
4. **Cái gì thuần mã nguồn thì bằng tiếng Anh, cái gì người dùng đọc thì bằng
   tiếng Việt.** Tên hàm, biến, kiểu dữ liệu, hằng số, tệp và thư mục là tiếng Anh.
   Chú thích và mọi thông điệp tới người dùng là tiếng Việt, kể cả phần mô tả
   trong `Short`, `Long`, `Example` và các bộ kiểm tra số đối số. Chi tiết ở
   [Quy tắc viết mã](#quy-tắc-viết-mã).

## Bản đồ mã nguồn

| Đường dẫn | Vai trò |
| --- | --- |
| `main.go` | Điểm khởi động, chỉ gọi `cmd.Execute()` |
| `cmd/root.go` | Lệnh gốc, cờ toàn cục, biến toàn cục, đăng ký nhóm công cụ |
| `cmd/help.go` | Khuôn trợ giúp, lệnh `help`, danh sách lệnh |
| `cmd/output.go` | In ấn, màu ANSI, bộ kiểm tra số đối số |
| `cmd/context.go` | `openRepo`: mở kho cho các lệnh con |
| `cmd/vcs*.go` | Khai báo lệnh của nhóm `vcs`, mỗi tệp một nhóm nhỏ |
| `internal/vcs/object/` | Kiểu dữ liệu blob, tree, commit, tag; tính mã băm |
| `internal/vcs/storage/` | Ghi đọc object nén zlib; quản lý tham chiếu và nhật ký |
| `internal/vcs/index/` | Vùng chuẩn bị, đọc ghi nhị phân có checksum |
| `internal/vcs/worktree/` | Đọc ghi tệp trên đĩa, quy tắc bỏ qua tệp, tệp ignore mẫu |
| `internal/vcs/repo/` | Lớp trừu tắng kho: cây nội dung, trạng thái, phân giải tham chiếu |
| `internal/vcs/diff/` | Thuật toán Myers, căn dòng, dựng hunk |
| `internal/vcs/merge/` | Hợp nhất ba phía kiểu diff3 |
| `internal/vcs/ops/` | Nghiệp vụ cấp cao, mỗi tệp một nhóm lệnh |
| `internal/vcs/config/` | Đọc ghi cấu hình dạng mục và khoá |
| `internal/tools/docgen/` | Sinh tài liệu Markdown cho cây lệnh |
| `build_all.sh` | Điểm vào để build, nằm ở gốc kho |
| `scripts/` | Cấu hình và script build còn lại (xem `scripts/README.md`) |
| `tests/` | Toàn bộ mã kiểm thử, tách theo vùng nghiệp vụ |
| `editors/vscode/` | Tiện ích VS Code đưa `tm` vào khung Source Control |
| `docs/agents/` | Tài liệu cho trợ lý lập trình: `AGENTS.md`, `README.md`, `cli/` |
| `CHANGELOG.md` | Những thay đổi đáng kể của từng đợt, kèm cách kiểm chứng |

## Mô hình dữ liệu

Mỗi đơn vị dữ liệu gọi là *object*, lưu thành một tệp riêng nén zlib tại
`.tmx/objects/<2 ký tự đầu>/<38 ký tự sau>`.

- Mã băm là SHA1 của chuỗi `"<loại> <kích thước>"` + byte NUL + nội dung.
- **Blob**: nội dung tệp.
- **Tree**: danh sách entry `<chế độ> <tên>` + byte NUL + 20 byte mã băm.
  Tên thư mục so sánh kèm dấu `/` phía sau.
- **Commit**: các dòng `tree`, `parent`, `author`, `committer`, dòng trống,
  rồi thông điệp.
- Có ba vùng trạng thái: **HEAD** (con trỏ đang đứng), **vùng chuẩn bị**
  (`.tmx/index`) và **cây làm việc** (tệp trên đĩa).

Cây nội dung được xây từ vùng chuẩn bị bởi `repo.TreeFromIndex`, dựng đệ quy
theo `repo.WriteTree`.

## Tệp ignore

Tên tệp là `.tmxignore`, đặt ở bất kỳ thư mục nào và chỉ có tác dụng bên trong
thư mục đó. Tệp của riêng máy là `.tmx/info/exclude`.

`tm vcs init` tạo sẵn `.tmxignore` ở gốc kho bằng `worktree.WriteDefaultIgnore`.
Nội dung mẫu nằm ở `internal/vcs/worktree/ignore_default.txt` và được nhúng vào
tệp thực thi bằng `//go:embed`, nên sửa tệp mẫu thì phải chạy lại build.

Hàm này **không** đè tệp đã có: người dùng có thể đã tự viết quy tắc riêng và
khởi tạo không được xoá mất lựa chọn đó. Vì vậy khi thêm tệp ignore mới vào bộ
mẫu, kiểm thử phải có cả tình huống tệp đã tồn tại.

## Thêm một lệnh mới

1. Thêm hàm nghiệp vụ vào tệp phù hợp trong `internal/vcs/ops/`. Hàm nhận
   `*repo.Repo` và một struct tuỳ chọn, trả về kết quả có kiểu rõ ràng.
2. Khai báo lệnh trong tệp `cmd/vcs_*.go` phù hợp. Bắt buộc có:
   - `Short`: một dòng ngắn.
   - `Long`: giải thích *tác dụng và lý do*, không chỉ liệt kê tham số.
   - `Example`: các lệnh mẫu chạy được, mỗi dòng một tình huống.
   - `Args`: dùng bộ kiểm tra trong `cmd/output.go` để báo lỗi tiếng Việt.
3. Thêm lệnh vào `cmd.AddCommand(...)` trong `cmd/vcs.go`.
4. Sinh lại tài liệu: `./scripts/build_agent_docs.sh`.

## Thêm một nhóm công cụ mới

1. Tạo gói nghiệp vụ trong `internal/<tên>/` nếu cần kiểu dữ liệu riêng.
2. Tạo `cmd/<tên>.go` trả về `*cobra.Command`, đặt `GroupID` phù hợp.
3. Thêm nhóm vào danh sách trong `rootCmd.AddGroup(...)` nếu dùng ID mới.
4. Gọi `registerToolGroup(...)` trong `init()` của `cmd/root.go`.

## Quy tắc viết mã

Chia làm hai nhóm: **cái thuần mã nguồn thì tiếng Anh, cái người dùng đọc thì
tiếng Việt.**

| | Ngôn ngữ |
| --- | --- |
| Tên hàm, biến, kiểu dữ liệu, hằng số | **Tiếng Anh** |
| Tên tệp và tên thư mục | **Tiếng Anh** |
| Chú thích trong mã nguồn (`//`, `/* */`, `/** */`) | **Tiếng Việt** |
| `Short`, `Long`, `Example` của mỗi lệnh | **Tiếng Việt** |
| Mô tả cờ (`Usage`) | **Tiếng Việt** |
| Thông báo lỗi trả về cho người dùng | **Tiếng Việt** |
| Mọi thứ in ra màn hình: `printLine`, `printOut`, tiêu đề, thống kê | **Tiếng Việt** |
| Thông báo của tiện ích VS Code mà người dùng đọc | **Tiếng Việt** |
| Tài liệu cho người dùng | **Tiếng Việt** |
| Chuỗi trong kiểm thử mà đối chiếu với output của lệnh | **Tiếng Việt** |

Ranh giới là *ai đọc phần đó*. Tên do người lẫn công cụ đọc, nên phải là tiếng Anh
cho quen thuộc và dùng được với mọi ngôn ngữ. Còn lời và thông điệp thì chỉ người
dùng tiếng Việt đọc, nên viết bằng tiếng Việt thì tự nhiên và dễ hiểu hơn.

Các quy tắc khác:

- Chú thích giải thích *tại sao*, không lặp lại *cái gì* mà tên hàm đã nói.
- Không dùng gói ở tầng thấp hơn so với tầng đang làm.
- Thông báo lỗi nói rõ cái gì hỏng và vì sao, kèm ngữ cảnh:
  `fmt.Errorf("không tìm thấy nhánh %s", name)`.

### Tên hàm, biến và kiểu

Theo chuẩn Go: chữ thường, không dấu, các từ ghép liền với nhau.

```go
func showFileDiff(repo *Repo, rel string) error       // đúng
func hienThiKhacBiet(kho *Kho, duongDan string) error  // sai
```

### Tên tệp và tên thư mục

Tên tệp và tên thư mục luôn viết bằng tiếng Anh, không dấu, không ký tự ngoài
ASCII. Tên có dấu vỡ ở nhiều nơi: terminal trên Windows, script build chạy trên
ba nền tảng, và người khác gõ lại đường dẫn từ thông báo lỗi thì phải gõ đúng dấu.

```bash
cmd/vcs_branch.go                            # đúng
cmd/lenh-nhanh.go                            # sai
internal/vcs/repo/ten-co-dau.txt             # sai, có dấu
editors/vscode/src/test/diff-view.test.ts    # đúng
```

Quy tắc cụ thể:

- Tệp Go theo chuẩn Go: chữ thường, từ ghép bằng dấu gạch dưới, ví dụ
  `vcs_branch.go`.
- Tệp TypeScript và Markdown dùng dấu gạch nối `-`, ví dụ `diff-view.test.ts`.
- Chữ hoa chỉ dành cho tệp quen thuộc ở gốc kho: `README.md`, `AGENTS.md`,
  `CHANGELOG.md`, `CONTRIBUTING.md`, `LICENSE`.
- Tên phải nói được tệp đó làm gì, không cần mở ra mới hiểu. `diff-view.test.ts`
  tốt hơn `nhanh.test.ts`.

Có kiểm thử ở `tests/architecture/naming_test.go` chặn bốn điều: ký tự ngoài
ASCII trong tên tệp, chữ hoa ở gốc kho, từ tiếng Việt viết không dấu trong tên tệp,
và chữ có dấu trong tên hàm hay biến. `docs/agents/cli/` được sinh tự động nên
tên ở đó lấy từ tên lệnh, không kiểm tra.

## Hai nguyên tắc riêng của phần trợ giúp

Hai điều dưới đây đã có kiểm thử trong `tests/architecture/help_test.go`.

### Mỗi lệnh phải chạy được

Gõ lệnh mà không kèm tham số thì lệnh đó vẫn phải làm được việc gì đó hữu
ích, chứ không in cả trang trợ giúp dài:

- Lệnh gốc in phiên bản và danh sách lệnh, xem `runRoot`.
- Nhóm lệnh in danh sách lệnh con, xem `printCommandList`.
- Trợ giúp đầy đủ chỉ hiện khi có `-h` hoặc `--help`.

Cờ `-v` vì vậy phải in thông tin môi trường. Một cờ mà in ra trợ giúp thì
vô dụng.

Every English label cobra generates is replaced in `setupHelp`. The help
trợ giúp nằm trong `cmd/help.go`; đừng gọi `InitDefaultHelpFlag` hay
`InitDefaultVersionFlag` của cobra, vì hai hàm đó ép nối cờ toàn cục của lệnh
cha vào lệnh con và làm hỏng lệnh con khai báo cờ trùng tên.

### Lệnh con không được khai báo lại cờ toàn cục

Cờ toàn cục áp dụng cho mọi lệnh. Nếu một lệnh con khai báo cờ trùng *tên*
hoặc trùng *chữ viết tắt*, thì cờ của lệnh con được ưu tiên và cờ toàn cục bị
bỏ qua trong lệnh đó. Việc này xảy ra **âm thầm, không có thông báo nào**, và
cùng một chữ viết tắt sẽ mang hai nghĩa khác nhau tuỳ lệnh.

Nên đặt tên cờ theo đúng việc nó làm: `tm vcs branch --hash` chứ không phải
`--verbose`, vì `-v` đã là cờ toàn cục in thêm thông tin chi tiết.

## Lệnh thường dùng

```bash
./build_all.sh                  # build Mac, Linux, Windows vào out/
./scripts/build_binaries.sh     # chỉ phần build binary
./scripts/build_agent_docs.sh   # sinh lại docs/agents/cli
go test ./...                   # chạy kiểm thử
./scripts/check.sh              # kiểm tra trọn vẹn trước khi đóng góp
```

`./scripts/check.sh` chạy định dạng, phân tích tĩnh, kiểm thử, rồi sinh lại
tài liệu vào thư mục tạm để so với bản đã commit. Đây là lệnh nên chạy trước
mỗi lần đóng góp.

## Phiên bản

Số phiên bản nằm ở biến `VERSION` trong phần cấu hình của
`scripts/build_binaries.sh`, là nguồn duy nhất. Script build đọc biến đó rồi
gắn vào tệp thực thi bằng cờ ldflags, nên `tm version` luôn khớp với tên file
trong `out/`.

Phát hành bản mới thì sửa đúng một dòng đó rồi chạy:

```bash
./build_all.sh
```

Biến `AppName`, `Version`, `Author` và `RepoURL` trong `cmd/root.go` phải khai
báo bằng `var` thì ldflags mới ghi được, đừng đổi thành `const`.

Khi chạy thẳng bằng `go build` mà không qua script, lệnh dùng giá trị mặc định
khai trong `cmd/root.go`. Ba giá trị `AppName`, `Author` và `RepoURL` phải
khớp với `build_binaries.sh`, có kiểm thử chặn.

## Ghi changelog

Thay đổi đủ lớn để người khác cần biết thì ghi vào `CHANGELOG.md` ở mục
`Chưa phát hành`: nêu thêm gì, sửa gì, đụng tới tệp nào, kiểm chứng bằng cách nào
và còn thiếu gì. Mục tiêu là người đọc trên máy khác dựng lại được bối cảnh mà
không cần đọc lịch sử git.

## Về kiểm thử

Mã kiểm thử nằm trong `tests/`, mỗi thư mục con kiểm thử một vùng nghiệp vụ.
Các tệp dùng gói kiểm thử ngoài (`package diff_test`) nên chỉ gọi được thứ
mà gói đó xuất ra.

`tests/architecture/` kiểm tra các bất biến kiến trúc: không gọi chương trình
ngoài, phụ thuộc giữa các tầng đi một chiều, tên tệp và tên hàm bằng tiếng Anh,
mọi lệnh đủ `Short`/`Long`/`Example` bằng tiếng Việt, và binary chạy được với
`PATH` rỗng.

`tests/cli/` chạy cây lệnh trong bộ nhớ để kiểm tra tầng dòng lệnh, ví dụ cách
lệnh đọc danh sách tệp sau dấu `--`.

Khi sửa một lỗi, thêm kiểm thử tái hiện lỗi đó trước khi sửa mã.

## Tiện ích VS Code

`editors/vscode/` là tiện ích TypeScript đưa `tm` vào khung Source Control của
VS Code, viết bằng API `vscode.scm`. Nó không thuộc module Go nên `go build`
và `go test` không đụng tới, và kiểm thử kiến trúc không quét tới.

Hai điều cần nhớ khi sửa cho khớp với phần còn lại của kho:

- Tiện ích **không** đọc tệp `.tmx` bằng tay, mọi thứ đi qua `tm vcs`. Nhờ vậy
  định dạng dữ liệu của kho không bị phụ thuộc vào nó.
- Toàn bộ phần đọc output tiếng Việt của `tm` nằm trong `editors/vscode/src/parse.ts`:
  các chữ trạng thái mà nó khớp và các thông báo lỗi mà nó nhận ra. Đổi câu chữ trong
  output của `tm` thì tệp này phải đổi theo.
- Lệnh `tm` luôn được gọi với `-C` đặt trước mọi thứ khác. Đặt `-C` cuối dòng lệnh
  thì khi lệnh có danh sách tệp sau dấu `--`, tm sẽ hiểu `-C` là một đường dẫn.

Các lệnh của VS Code có kiểm tra tham số theo tên, sai hình dạng là bị từ chối
ngay với thông báo *Invalid argument*. Hai lệnh hay dùng:

- `vscode.diff(trái, phải, tiêu đề)`: đúng ba đối số, đối số thứ ba là chuỗi.
- `vscode.changes(tiêu đề, [[địa chỉ, phía gốc, phía đã sửa], ...])`: mỗi mục
  phải là bộ ba địa chỉ, không phải danh sách địa chỉ thuần.

Kiểm thử của tiện ích nằm ở `editors/vscode/src/test/`. Bản giả API VS Code
dùng chung nằm ở `harness.ts`; nó ghi lại cả lệnh lẫn đối số truyền vào nên soi
được đúng cái khung so sánh sẽ được mở ra. `manifest.test.ts` kiểm tra phần khai
báo trong `package.json`, nơi mà lỗi không biểu hiện lúc chạy mã nguồn: một
submenu khai báo mà không có lệnh nào thuộc về nó thì menu ra trống mà mọi kiểm
thử khác vẫn xanh.

```bash
cd editors/vscode && npm install && npm run compile && npm test
```

Đóng gói thành tệp cài được:

```bash
./scripts/build_extension.sh     # ra out/devcli-vscode-<phiên bản>.vsix
```

`build_all.sh` đã gọi sẵn. Bước này cần Node.js và tự bỏ qua nếu máy không có.

## Tài liệu tham chiếu lệnh

`docs/agents/cli/` chứa tài liệu Markdown cho từng lệnh, sinh tự động từ cây lệnh
bằng `internal/tools/docgen`. Khi cần biết chính xác một lệnh nhận những gì và
làm gì, đọc tệp tương ứng thay vì đọc mã nguồn.

Tệp `docs/agents/cli/vcs/td_vcs.md` là mục lục của nhóm `vcs`, bắt đầu từ đó.

Thư mục này không sửa tay được. Sau khi thay đổi cây lệnh, chạy:

```bash
./scripts/build_agent_docs.sh
```