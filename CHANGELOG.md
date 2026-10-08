# Changelog

Ghi lại những thay đổi đáng kể của td theo từng mốc, kèm tệp đã đụng tới, cách
kiểm chứng và những gì còn chưa làm. Mục tiêu là người đọc trên máy khác dựng
lại được bối cảnh mà không cần đọc lịch sử git.

Định dạng theo [Keep a Changelog](https://keepachangelog.com/vi/1.1.0/), phiên
bản theo [SemVer](https://semver.org/lang/vi/).

## Chưa phát hành · 2026-10-08

Ngày 2026-10-08, một phiên làm việc kéo dài từ 07:55 đến 08:47 theo giờ máy.
Các mốc dưới đây là các chặng trong phiên đó, không phải các ngày khác nhau.

Nội dung đợt này: đưa `td` vào khung Source Control của VS Code, sửa những lỗi
trong `td` chặn việc đó, và đóng gói tiện ích thành tệp `.vsix` cài được.

---

### Mốc 1 · 2026-10-08 07:55 — Sửa ba lỗi trong `td`

Bắt đầu từ việc đọc mã nguồn và dựng khung extension, phát hiện ba chỗ hành vi
không khớp với chính phần trợ giúp của lệnh. Cả ba đều chặn không thể làm một
tính năng nào đó của khung Source Control, nên phải sửa trước khi viết tiếp.

| Tệp | Sửa gì |
| --- | --- |
| `internal/vcs/ops/checkout.go` | `restore --staged` lấy nội dung từ HEAD thay vì từ chính vùng chuẩn bị |
| `internal/vcs/ops/checkout.go` | `restore` không kèm cờ thì chỉ đụng cây làm việc, không ghi vùng chuẩn bị |
| `cmd/flags.go` | Thêm hàm chung `splitPathsAtDash` đọc vị trí dấu `--` |
| `cmd/vcs_diff.go`, `cmd/vcs_basic.go`, `cmd/vcs_branch.go` | `diff`, `log`, `reset` dùng `splitPathsAtDash` |
| `cmd/vcs_init.go` | `init` nghe cờ `-C` như mọi lệnh khác |

Ba lỗi và hậu quả cụ thể:

1. **`restore --staged` không gỡ được gì.** Lệnh lấy nội dung từ chính vùng chuẩn
   bị rồi stage lại, nên tệp vẫn ở trạng thái đã stage. Khung Source Control không
   có nút *Unstage Changes*.
2. **`restore` không kèm cờ lại vô tình stage thay đổi.** Mặc định trước đó ghi cả
   hai vùng, nên thao tác "huỷ sửa đổi" xong lại biến thành "đã stage". Không có nút
   *Discard Changes* nào dùng được.
3. **Danh sách tệp sau dấu `--` không được dùng.** Các lệnh tự dò `--` trong danh
   sách đối số, nhưng pflag đã ăn mất dấu đó khi phân tích cờ.
   `td vcs diff -- a.txt` báo *không phân giải được tham chiếu "a.txt"*. Không lọc
   được diff theo tệp thì không mở được khung so sánh.

Kiểm chứng: `tests/ops/ops_test.go` có `TestRestoreStagedUnstage` và
`TestRestoreDefaultOnlyWorktree`; `tests/cli/cli_test.go` có
`TestDiffPathsAfterDash`, `TestDiffUnstagedPathsAfterDash`, `TestLogPathsAfterDash`,
`TestResetPathsAfterDash`. `./scripts/check.sh` qua cả bốn bước.

---

### Mốc 2 · 2026-10-08 08:05 — Dựng tiện ích VS Code

Tạo `editors/vscode/` bằng TypeScript với API `vscode.scm`. Cách làm giống tiện
ích git tích hợp sẵn: cùng tên nhóm, cùng nhãn, cùng vị trí menu, cùng bảng màu.

| Tệp | Vai trò |
| --- | --- |
| `package.json` | 49 lệnh, 7 điểm menu, 4 khung bên, 10 màu trang trí |
| `src/extension.ts` | Khởi động, cấp nội dung ảo cho khung so sánh |
| `src/model.ts` | Dò kho trong workspace, theo dõi tệp, thanh trạng thái |
| `src/repository.ts` | Một kho: SourceControl, nhóm thay đổi, quick diff |
| `src/td.ts` | Bọc lệnh `td vcs` |
| `src/parse.ts` | Dịch output tiếng Việt của `td` thành dữ liệu có kiểu |
| `src/commands.ts` | Toàn bộ lệnh |
| `src/views.ts` | Bốn khung nhánh, commit, stash, tag |
| `src/decorations.ts` | Chữ viết tắt trên cây thư mục |
| `src/uri.ts` | Lược đồ `td:` cho nội dung nằm trong kho |

Khối Source Control dựng được: ô nhập commit kèm nút dấu tích, bốn nhóm *Merge
Changes* / *Staged Changes* / *Changes* / *Untracked Changes*, số đếm trên biểu
tượng thanh hoạt động, nút stage/unstage/huỷ ở tiêu đề nhóm và cạnh tệp. Trong
trình soạn thảo có vạch khác biệt ở rìa và dải khác biệt trong tệp. Trên cây thư
mục có chữ viết tắt `M`, `A`, `D`, `U`.

Ranh giới quan trọng đã chốt ngay từ đầu: tiện ích **không** đọc tệp `.tdx` bằng
tay, mọi thao tác đi qua `td vcs`. Định dạng dữ liệu của kho vì thế không bị phụ
thuộc vào tiện ích.

Hai luật kỹ thuật phải giữ khi sửa `src/td.ts`, ghi ở đó và ở README:

- **`-C` luôn đặt trước mọi thứ khác.** Đặt cuối dòng lệnh thì khi lệnh có danh
  sách tệp sau `--`, `td` hiểu `-C` là một đường dẫn nữa và chạy nhầm ở thư mục
  hiện tại. Trong lúc làm đã dính lỗi này một lần.
- **Diff phải xin với số dòng ngữ cảnh rất lớn** (`-U1000000`). Khung so sánh của
  VS Code cần nội dung đầy đủ của cả hai phía rồi tự tính vạch hiệu, nên phải có
  diff ôm trọn tệp. Lệnh luôn gọi cho đúng một tệp nên lượng dữ liệu vẫn nhỏ.

---

### Mốc 3 · 2026-10-08 08:20 — Kiểm thử

Thêm kiểm thử ở cả hai vùng nghiệp vụ.

- `tests/cli/cli_test.go`: bộ kiểm thử tầng dòng lệnh, chạy cây lệnh trong bộ nhớ
  rồi kiểm tra output. Trước đó chỉ có kiểm thử nghiệp vụ trong `tests/ops/`.
- `editors/vscode/src/test/parse.test.ts`: kiểm thử từng hàm dịch output, chạy lệnh
  `td` thật trên kho tạm rồi so với dữ liệu mẫu.
- `editors/vscode/src/test/td.test.ts`: kiểm thử lớp bọc lệnh, gồm stage, unstage,
  huỷ, commit, nhánh, tag, stash.
- `editors/vscode/src/test/smoke.test.ts`: chạy phần kích hoạt của tiện ích với
  một bản sao đủ dùng của API VS Code, rồi kiểm tra từng lệnh khai báo trong
  `package.json` có được đăng ký thật không, ô nhập có nhắc nhánh không, và ba
  nhóm thay đổi có đúng số tệp không.

Kiểm thử này bắt được ba lỗi thật trong lúc viết: hàm dịch diff bỏ sót phần nội
dung vì đoán sai điểm bắt đầu hunk, dòng lệnh `td` bị đặt `-C` sai chỗ, và cách
đọc tên nhánh trong output của `td`.

Kiểm chứng: `cd editors/vscode && npm test` cho 27 kiểm thử, tất cả đều pass.

---

### Mốc 4 · 2026-10-08 08:35 — Bổ sung `td vcs clean` và cải thiện tiện ích

Thêm lệnh `td vcs clean` vào `td` để khung Source Control làm được nút *Discard
All Untracked Changes`. Đây là tính năng duy nhất còn thiếu mà trước đó phải báo
chưa làm được.

- `internal/vcs/ops/clean.go` (mới), `cmd/vcs_clean.go` (mới),
  `docs/agents/cli/vcs/td_vcs_clean.md` (sinh tự động).
- Nguyên tắc an toàn: chỉ xoá tệp chưa từng được đưa vào vùng chuẩn bị; tệp đã
  theo dõi thuộc về lịch sử nên lệnh không đụng tới. Không có `-f` thì chỉ liệt
  kê, không hỏi tương tác, đúng với kiểu CLI của `td`.
- Kiểm thử: `TestCleanRemovesOnlyUntracked`, `TestCleanWithPaths`,
  `TestCleanListsBeforeDeleting`, `TestCleanOnlyNamedPaths`.

Cùng đợt, bốn cải thiện trong tiện ích:

1. Tìm lệnh `td` theo thứ tự `td.path` → `PATH` → `~/go/bin` → `~/bin` →
   `/usr/local/bin` → `/opt/homebrew/bin` → `out/` của kho. Trước đó chỉ dùng
   `PATH`, nên cài lên máy mới phải sửa cấu hình.
2. Ghi từng lệnh `td` đã chạy vào kênh Output, kèm mã thoát khi hỏng.
3. Báo một lần khi máy chưa có `td`, thay vì im lặng.
4. Báo đang xử lý trên thanh trạng thái khi thao tác chạy lâu.

---

### Mốc 5 · 2026-10-08 08:40 — Đóng gói `.vsix` và nối vào build

- `scripts/build_extension.sh` (mới): biên dịch rồi đóng gói tiện ích thành
  `out/devcli-vscode-0.1.0.vsix`. Tự bỏ qua nhẹ nhàng khi máy không có Node.js,
  phần build Go không được phụ thuộc Node.
- `build_all.sh`: thêm bước thứ ba. Cờ mới `--no-extension` để bỏ qua.
- `editors/vscode/.vscodeignore`, `LICENSE`, `.vscode/launch.json`,
  `.vscode/tasks.json` (mới).

Một tệp `.vsix` dùng được cho cả ba nền tảng vì chỉ chứa JavaScript đã biên dịch
và một tệp biểu tượng, không có mã native theo hệ điều hành. Kích thước 44 KB.

```bash
code --install-extension out/devcli-vscode-0.1.0.vsix
```

Phần phải làm riêng theo máy là đặt lệnh `td`, vì binary Go không chạy được trên
máy khác. Bảng hướng dẫn cho từng nền tảng nằm trong `editors/vscode/README.md`.

---

### Tổng kết kiểm chứng · 2026-10-08

```bash
./scripts/check.sh                            # định dạng, phân tích, kiểm thử, đối chiếu tài liệu
cd editors/vscode && npm install && npm test  # 27 kiểm thử
```

Kiểm thử của tiện ích chạy lệnh `td` thật trên kho tạm nên máy phải build binary
trước, ví dụ `./scripts/build_binaries.sh`. Đặt biến môi trường `TD_BIN` để trỏ
tới một tệp thực thi cụ thể.

---

### Mốc 6 · 2026-10-08 — Sửa ba lỗi tiện ích thấy được khi dùng thật

Các mốc trên chỉ kiểm chứng tới mức kích hoạt với API giả, chưa mở VS Code thật
để soi bằng mắt. Đợt này bắt được ba lỗi mà người dùng thấy ngay lần đầu mở
khung Source Control.

| Tệp | Sửa gì |
| --- | --- |
| `editors/vscode/src/commands.ts` | `vscode.diff` truyền thừa một đối số, làm lệnh bị VS Code từ chối |
| `editors/vscode/src/commands.ts` | `vscode.changes` nhận sai hình dạng danh sách tệp |
| `editors/vscode/src/repository.ts`, `package.json` | Bấm tệp mở thẳng tệp, không mở khung so sánh |
| `editors/vscode/package.json` | Năm submenu khai báo mà không có lệnh nào, nên nút `...` mở ra trống |
| `internal/vcs/ops/plumbing.go`, `cmd/vcs_basic.go` | Thêm `td vcs show-file` |
| `editors/vscode/src/td.ts`, `repository.ts` | Nội dung ở HEAD đọc bằng `show-file` thay vì dựng từ khác biệt |

**1. Bấm tệp không ra khung so sánh.** `td.openDiffOnClick` mặc định là `false`
và mọi tệp trong khung Source Control đều gắn lệnh `td.openFile`, nên bấm vào là
mở tệp. Đổi mặc định thành `true` và chọn lệnh theo cấu hình, giống cách
extension git làm trong `ResourceCommandResolver`.

**2. `vscode.diff` bị từ chối.** Lệnh nhận `(trái, phải, tiêu đề)` và kiểm tra
thứ tự đối số theo tên; mã cũ truyền `(trái, phải, địa chỉ, tiêu đề)` nên VS Code
báo *Invalid argument 'title'*. Đã sửa lại đúng ba đối số.

**3. `vscode.changes` bị từ chối khi bấm vào một commit.** Đối số thứ hai phải là
danh sách bộ ba `[địa chỉ hiển thị, phía gốc, phía đã sửa]`; mã cũ truyền danh
sách địa chỉ thuần nên VS Code báo *Invalid argument 'resourceList'* và không mở
gì cả. Đúng lúc này người dùng không thấy tệp nào được commit đó đổi.

**4. Không có thao tác nào trong menu.** Năm submenu `td.commit`, `td.changes`,
`td.branch`, `td.stash`, `td.tags` chỉ khai báo nhãn, không có mục menu nào trỏ
tới, nên bấm nút ba chấm ở thanh Source Control ra một danh sách rỗng. Đã điền
đủ lệnh cho cả năm, thêm nút trên tiêu đề bốn khung bên và lời nhắc khi khung
trống.

**5. `td vcs show-file`.** Trước đợt này nội dung phía HEAD được dựng lại từ
`td vcs diff --staged`, chỉ đúng với tệp đang chờ commit: tệp sạch không nằm
trong khác biệt nào nên phía đó rỗng và khung so sánh hiện sai. Lệnh mới đọc thẳng
blob trong cây của commit nên đúng với mọi tệp, kể cả tệp vừa được thêm (trả về
rỗng vì tệp chưa có ở đó). Khi gặp `td` cũ hơn, tiện ích nhận ra qua thông báo
*không có lệnh nào tên* rồi rơi về cách cũ và ghi một dòng ra kênh log.

Ngoài ra, khung so sánh của một commit giờ chỉ xin tên tệp bằng
`td vcs diff --name-only`, còn nội dung từng phía đọc sau đúng lúc khung so sánh
cần tới. Trước đây phải nạp toàn bộ nội dung của mọi tệp trong commit ngay khi
mở.

Kiểm chứng:

```bash
./scripts/check.sh
cd editors/vscode && npm run compile && npm test   # 43 kiểm thử
```

Kiểm thử mới:

- `editors/vscode/src/test/manifest.test.ts` (mới): mọi submenu phải có lệnh,
  mọi lệnh trong menu phải được khai báo, mọi lệnh phải có nhãn dịch, mỗi khung
  bên phải có lời nhắc khi trống. Bắt được lỗi submenu rỗng ngay từ tệp khai báo.
- `editors/vscode/src/test/smoke.test.ts`: bấm tệp phải mở `vscode.diff` với ba
  đối số; bấm commit phải mở `vscode.changes` với danh sách bộ ba; nội dung ở
  HEAD phải là nội dung thật kể cả với tệp sạch; tệp bị xoá và tệp chưa theo dõi
  phải có đúng một phía rỗng; nút *Open File* trên khung so sánh phải ra tệp thật.
- `editors/vscode/src/test/harness.ts` (mới): bản giả API VS Code dùng chung, có
  ghi lại lệnh đã chạy kèm đối số để soi được tham số truyền vào.
- `editors/vscode/src/test/td.test.ts`: `showFile` trả về `unsupported` khi lệnh
  td trên máy chưa có `show-file`; `changedFiles` chỉ trả về tên tệp.
- `tests/ops/ops_test.go`: `TestShowFileReadsContentAtRevision`,
  `TestShowFileAfterAddAndRemove`.
- `tests/cli/cli_test.go`: `TestShowFileReadsFromRevision`.

---

## Chưa làm (tính đến 2026-10-08)

### Vì `td` chưa có, cần làm ở phần td trước

| Việc | Vì sao chưa làm |
| --- | --- |
| Pull, push, fetch, publish, sync | `td` quản lý phiên bản cục bộ, không có khái niệm remote. Thanh trạng thái vẫn hiện số commit đi trước/đi sau khi nhánh có cấu hình `branch.<tên>.merge`. |
| Stage / unstage từng đoạn đang chọn | Cần `td` ghi được entry từng phần vào vùng chuẩn bị. `td` hiện chỉ lưu trạng thái toàn tệp, giống `git add` không có `-p`. |
| Cờ ký commit | `td` chưa có. Khi có, cần thêm thao tác ký vào lệnh commit và thêm trường người ký vào object commit. |
| Kiểm tra nội dung commit trước khi tạo | `td` chỉ chặn nội dung rỗng, không chặn nội dung chỉ gồm khoảng trắng. |

### Vì API của VS Code chưa có trong bản ổn định

| Việc | Vì sao chưa làm |
| --- | --- |
| Nút phụ cạnh nút commit (Commit Staged, Amend) trong dải ô nhập | Nằm ở API đề xuất `scmActionButton`. Bản ổn định chỉ có một nút dấu tích, nên các biến thể commit đang nằm trong menu `...` của khung Source Control. |
| Kiểm tra nội dung ngay khi gõ trong ô nhập | `validateInput` của ô nhập thuộc API đề xuất. Tiện ích tự hở hộp thoại khi ô nhập rỗng, gần với trải nghiệm git nhất có thể với API ổn định. |
| Kho con trong Source Control, biểu tượng riêng cho từng kho | `createSourceControl` ở bản ổn định chỉ nhận ba tham số; phần còn lại nằm ở API đề xuất. |

### Còn thiếu trong tiện ích

| Việc | Vì sao chưa làm |
| --- | --- |
| Chưa chạy thử trong cửa sổ VS Code thật | Ba lỗi ở mốc 6 đều do chưa mở VS Code thật mà lọt qua. Mốc đó đã soi lại từng đối số truyền cho `vscode.diff` và `vscode.changes`, tức là đúng những chỗ VS Code thật kiểm tra, nhưng vẫn chưa soi bằng mắt trên cửa sổ thật. |
| Chưa thử trên macOS và Windows thật | Mới build được ba nền tảng. Phần đa nền tảng của tệp `.vsix` là chắc chắn vì không có mã native, nhưng đường dẫn cài `td` trên từng hệ chưa kiểm tra. |
| Chưa có kiểm thử giao diện tự động | Bộ kiểm thử dừng ở tầng mô hình và tầng lệnh, chưa tới thao tác chuột và bàn phím trong khung. |
| Chưa phát hành lên marketplace | `publisher` trong `package.json` đang là `td`, cần đổi thành tên nhà phát hành thật và khai báo `repository`. |
| Chưa có CI | Mọi kiểm tra hiện chạy tay bằng `./scripts/check.sh`. |
| Tô màu tệp bị bỏ qua trong cây thư mục | `git` hỏi tệp có bị bỏ qua không bằng một lệnh riêng. `td vcs ignore` chỉ in ra quy tắc đang có, chưa có lệnh hỏi cho một tệp cụ thể. Tự viết lại logic khớp mẫu trong TypeScript thì dễ lệch với `td`, nên chưa làm. Cần thêm lệnh ở phần `td` trước. |
| Chưa hiện kho lồng nhau sâu hơn một tầng | Khi mở thư mục không phải gốc kho, tiện ích dò kho trong thư mục con ở tầng ngay bên dưới. Kho nằm sâu hơn thì không thấy. |
| Chưa có nhóm *Merge Changes* dùng thật | Nhóm đã khai báo và sẽ hiện khi `td vcs status` báo xung đột, nhưng chưa chạy thử tình huống merge xung đột thật. |

### Hạn chế đã biết

- **Output của `td` bị phụ thuộc.** Tiện ích đọc output tiếng Việt của `td`, nên đổi
  câu chữ trong `cmd/` thì phải sửa `editors/vscode/src/parse.ts` theo. Giải pháp
  lâu dài là thêm một chế độ in dạng máy đọc được vào `td`, chưa làm vì nó động
  vào bề mặt lệnh.
- **Tệp không có dấu xuống dòng cuối.** `td vcs diff` không in dấu báo như git, nên
  khi dựng lại nội dung để mở khung so sánh, phía dựng sẽ có thêm dấu xuống dòng ở
  cuối. Khung so sánh có thể hiện dòng cuối là khác biệt dù nội dung thật như nhau.
  Chỉ xảy ra với tệp không có dấu xuống dòng cuối, ví dụ tệp do Windows ghi. Phía
  trên đĩa và phía ở HEAD đã đọc thẳng nên không bị; còn phía ở vùng chuẩn bị và
  phía ở một commit thì vẫn dựng từ khác biệt.
- **Tệp nhị phân không hiện khác biệt.** `td` không in nội dung tệp nhị phân nên
  tiện ích không dựng được khung so sánh cho loại tệp đó, giống hành vi của git.
  Cả hai phía đều hiện dòng *tệp nhị phân, không hiển thị nội dung* để khung so
  sánh không báo nhầm là có khác biệt.