# TD VCS — tiện ích Source Control cho VS Code

Tiện ích đưa `tm` vào khung **Source Control** của VS Code với đúng hình dạng
của tiện ích git tích hợp sẵn: ô nhập commit, các nhóm thay đổi, thanh trạng
thái, chữ viết tắt trên cây thư mục, khung so sánh, và các khung nhánh, lịch
sử commit, bản lưu tạm, tag.

## Cần có gì

- VS Code 1.90 trở lên.
- Lệnh `tm` trên máy. Tiện ích tự tìm `tm` trong `PATH`, rồi tới `~/go/bin`,
  `~/bin`, `/usr/local/bin` và `/opt/homebrew/bin`. Muốn chỉ định riêng thì đặt
  `tm.path`.

## Cài từ tệp .vsix

Một tệp `.vsix` chạy được trên mọi nền tảng: nó chỉ chứa mã JavaScript đã biên
dịch và một tệp biểu tượng, không có mã gốc theo hệ điều hành. Vì vậy chỉ cần
**một** tệp cho cả macOS, Linux và Windows.

```bash
# Build từ gốc kho tm
./build_all.sh
# hoặc chỉ đóng gói tiện ích
./scripts/build_extension.sh

# Cài vào VS Code
code --install-extension out/devcli-tm-vscode-0.1.0.vsix
```

Trên cùng một máy cài nhiều lần thì cài đè bản mới:

```bash
code --install-extension out/devcli-tm-vscode-0.1.0.vsix --force
```

Cài cho một người dùng cụ thể mà không cần quyền quản trị:

```bash
code --install-extension out/devcli-tm-vscode-0.1.0.vsix --user
```

### Đặt lệnh `tm` trên từng nền tảng

Tệp `.vsix` không chứa lệnh `tm`, vì một binary Go không chạy được trên máy
khác. Đặt `tm` trên máy đích rồi mở khung Source Control:

| Nền tảng | Cách đặt |
| --- | --- |
| macOS (Apple silicon) | copy `out/devcli-tm-mac-arm-0.1.0` thành `~/go/bin/tm`, rồi `chmod +x` |
| macOS (Intel) | build với `GOARCH=amd64` |
| Linux | copy `out/devcli-tm-linux-0.1.0` thành `~/go/bin/tm`, rồi `chmod +x` |
| Windows | copy `out/devcli-tm-windows-0.1.0.exe` thành `%USERPROFILE%\go\bin\tm.exe` |
| Mọi nền tảng, bỏ qua việc đặt trong PATH | đặt cấu hình `tm.path` trỏ tới tệp thực thi |

Kiểm tra nhanh trong khung Output của tiện ích (lệnh **TD: Show Output**): dòng
đầu tiên ghi lệnh tm đang dùng.

Muốn phát hành cho nhiều người thì chỉ cần một tệp `.vsix` cộng với hướng dẫn
đặt `tm` như bảng trên.

### Nhấn F5 để thử trong lúc code

```bash
cd editors/vscode
npm install
npm run compile
```

Mở VS Code tại thư mục này rồi nhấn `F5`. Cửa sổ thử nghiệm mở kho `tm` trong
chính kho này làm ví dụ.

## Những gì tiện ích làm được

| Việc | Cách làm trong khung Source Control |
| --- | --- |
| Xem thay đổi | Nhóm **Staged Changes**, **Changes**, **Merge Changes**, **Untracked Changes** |
| Commit | Gõ nội dung vào ô nhập rồi bấm dấu tích, hoặc `TD: Commit` |
| Stage, unstage, huỷ thay đổi | Nút cộng/trừ/thùng rác cạnh từng tệp, và trên tiêu đề nhóm |
| Xem khác biệt | Bấm tên tệp để mở khung so sánh, chọn nhiều tệp để mở khung so sánh nhiều tệp |
| Xem lịch sử của một tệp | Chuột phải lên tệp trong cây tệp, trình soạn thảo hoặc khung Source Control, rồi **View File History...**; chọn một commit để xem thay đổi của tệp tại commit đó |
| Xem thay đổi của một commit | Bấm một dòng trong khung **Commits** |
| Đường ngữ cảnh trong trình soạn thảo | Dấu nháy và dải khác biệt ở rìa tệp, giống git |
| Chữ viết tắt trên cây thư mục | `M` đã sửa, `A` thêm, `D` xoá, `U` mới hoặc xung đột |
| Nhánh, tag, bản lưu tạm | Các khung **Branches**, **Commits**, **Stashes**, **Tags** |
| Thanh trạng thái | Nhánh hiện tại, số commit đi trước/đi sau, số tệp chờ commit, lỗi |

## Cách tiện ích nói chuyện với tm

Tiện ích không tự đọc tệp `.tmx`. Mọi thứ đều đi qua `tm vcs`, ví dụ:

```
tm vcs -C /du-an status
tm vcs -C /du-an diff --staged -U1000000 -- main.go
tm vcs -C /du-an log -n200 -- cmd/root.go
tm vcs -C /du-an commit -m "nội dung"
```

Hai điểm đáng lưu ý:

- **`-C` luôn đặt trước mọi thứ khác.** Đặt cuối dòng lệnh thì khi lệnh có danh
  sách tệp sau dấu `--`, tm sẽ hiểu `-C` là một đường dẫn nữa và chạy nhầm ở
  thư mục hiện tại.
- **Diff được xin với số dòng ngữ cảnh rất lớn.** Khung so sánh của VS Code cần
  nội dung đầy đủ của cả hai phía rồi tự tính vạch hiệu, nên phải có diff ôm
  trọn tệp. Lệnh luôn được gọi cho đúng một tệp nên lượng dữ liệu vẫn nhỏ, kể
  cả trong kho lớn.

Toàn bộ phần dịch output của tm nằm trong `src/parse.ts`. Nếu cây lệnh của tm
đổi câu chữ, đó là chỗ duy nhất phải sửa.

## Lệnh của `tm` mà tiện ích dùng

| Việc | Lệnh |
| --- | --- |
| Đọc trạng thái | `tm vcs status` |
| Khác biệt hai phía | `tm vcs diff [--staged] -U1000000 -- <tệp>` |
| Danh sách tệp của một commit | `tm vcs diff --name-only <mã băm>` |
| Nội dung tệp ở HEAD | `tm vcs show-file HEAD -- <tệp>` |
| Lịch sử | `tm vcs log --oneline -n <số>` |
| Nhánh, tag, bản lưu tạm | `tm vcs branch -vv`, `tm vcs tag -l`, `tm vcs stash --list` |
| Stage, unstage, huỷ | `tm vcs add`, `tm vcs restore --staged`, `tm vcs restore` |
| Xoá tệp chưa theo dõi | `tm vcs clean -f <tệp>` |

`show-file` cần `tm` 0.1.0 trở lên. Gặp `tm` cũ hơn thì tiện ích ghi một dòng
ra kênh log và dựng phía HEAD từ khác biệt đã stage như trước, tức là tệp sạch
vẫn hiện rỗng.

## Những chỗ chưa làm được

tm là công cụ quản lý phiên bản **cục bộ**, không có remote, nên tiện ích cố
tình không có những phần mà git dựa vào remote:

- Không có Pull, Push, Fetch, Publish, Sync, Remote.
- Không có Stage Selected Ranges: git đưa từng đoạn đang chọn vào vùng chuẩn bị,
  việc đó cần tm ghi được entry từng phần vào index nên `tm` chưa làm.
- Không có Commit & Push, Commit & Sync, vì không có nơi để đẩy lên.

Thanh trạng thái vẫn hiện số commit đi trước/đi sau khi nhánh có cấu hình
`branch.<tên>.merge`, vì tm có đọc cấu hình đó để tính.

## Cấu trúc mã nguồn

| Tệp | Vai trò |
| --- | --- |
| `src/extension.ts` | Khởi động, cấp nội dung ảo cho khung so sánh |
| `src/model.ts` | Dò kho trong workspace, theo dõi tệp, đặt thanh trạng thái |
| `src/repository.ts` | Một kho: SourceControl, các nhóm thay đổi, quick diff |
| `src/tm.ts` | Bọc lệnh `tm vcs`, không đọc `.tmx` bằng tay |
| `src/parse.ts` | Dịch output tiếng Việt của tm thành dữ liệu có kiểu |
| `src/commands.ts` | Toàn bộ lệnh |
| `src/views.ts` | Bốn khung nhánh, commit, stash, tag |
| `src/decorations.ts` | Chữ viết tắt trên cây thư mục |
| `src/uri.ts` | Lược đồ `tm:` cho nội dung nằm trong kho |
| `src/test/harness.ts` | Bản giả API VS Code dùng chung cho kiểm thử |

## Cấu hình

| Khoá | Mặc định | Ý nghĩa |
| --- | --- | --- |
| `tm.path` | `tm` | Đường dẫn lệnh tm |
| `tm.enabled` | `true` | Bật dò kho trong workspace |
| `tm.autoRefresh` | `true` | Tự làm mới khi tệp trên đĩa đổi |
| `tm.autorefreshDelay` | `1000` | Chờ bao nhiêu mili giây trước khi làm mới |
| `tm.decorations.enabled` | `true` | Hiện chữ viết tắt trên cây thư mục |
| `tm.showCommitInput` | `true` | Hiện ô nhập nội dung commit |
| `tm.alwaysShowStagedChangesResourceGroup` | `true` | Luôn hiện nhóm Staged Changes |
| `tm.openDiffOnClick` | `true` | Bấm tệp thì mở khung so sánh thay vì mở tệp |
| `tm.confirmEmptyCommits` | `true` | Hỏi lại trước khi tạo commit rỗng |
| `tm.logMaxCount` | `500` | Số commit tối đa ở khung Commits |

## Kiểm thử

```bash
npm run compile
npm test
```

Phần kiểm thử chạy lệnh `tm` thật trên các kho tạm trong thư mục tạm. Đặt biến
môi trường `TD_BIN` để trỏ tới một tệp thực thi cụ thể. Không tìm thấy `tm` thì
các kiểm thử cần lệnh thật được bỏ qua, phần kiểm thử dữ liệu mẫu vẫn chạy.