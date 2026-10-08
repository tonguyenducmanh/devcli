# TD VCS — tiện ích Source Control cho VS Code

Tiện ích đưa `td` vào khung **Source Control** của VS Code với đúng hình dạng
của tiện ích git tích hợp sẵn: ô nhập commit, các nhóm thay đổi, thanh trạng
thái, chữ viết tắt trên cây thư mục, khung so sánh, và các khung nhánh, lịch
sử commit, bản lưu tạm, tag.

## Cần có gì

- VS Code 1.90 trở lên.
- Lệnh `td` trên máy. Tiện ích tự tìm `td` trong `PATH`, rồi tới `~/go/bin`,
  `~/bin`, `/usr/local/bin` và `/opt/homebrew/bin`. Muốn chỉ định riêng thì đặt
  `td.path`.

## Cài từ tệp .vsix

Một tệp `.vsix` chạy được trên mọi nền tảng: nó chỉ chứa mã JavaScript đã biên
dịch và một tệp biểu tượng, không có mã gốc theo hệ điều hành. Vì vậy chỉ cần
**một** tệp cho cả macOS, Linux và Windows.

```bash
# Build từ gốc kho td
./build_all.sh
# hoặc chỉ đóng gói tiện ích
./scripts/build_extension.sh

# Cài vào VS Code
code --install-extension out/devcli-vscode-0.1.0.vsix
```

Trên cùng một máy cài nhiều lần thì cài đè bản mới:

```bash
code --install-extension out/devcli-vscode-0.1.0.vsix --force
```

Cài cho một người dùng cụ thể mà không cần quyền quản trị:

```bash
code --install-extension out/devcli-vscode-0.1.0.vsix --user
```

### Đặt lệnh `td` trên từng nền tảng

Tệp `.vsix` không chứa lệnh `td`, vì một binary Go không chạy được trên máy
khác. Đặt `td` trên máy đích rồi mở khung Source Control:

| Nền tảng | Cách đặt |
| --- | --- |
| macOS (Apple silicon) | copy `out/devcli-mac-arm-0.1.0` thành `~/go/bin/td`, rồi `chmod +x` |
| macOS (Intel) | build với `GOARCH=amd64` |
| Linux | copy `out/devcli-linux-0.1.0` thành `~/go/bin/td`, rồi `chmod +x` |
| Windows | copy `out/devcli-windows-0.1.0.exe` thành `%USERPROFILE%\go\bin\td.exe` |
| Mọi nền tảng, bỏ qua việc đặt trong PATH | đặt cấu hình `td.path` trỏ tới tệp thực thi |

Kiểm tra nhanh trong khung Output của tiện ích (lệnh **TD: Show Output**): dòng
đầu tiên ghi lệnh td đang dùng.

Muốn phát hành cho nhiều người thì chỉ cần một tệp `.vsix` cộng với hướng dẫn
đặt `td` như bảng trên.

### Nhấn F5 để thử trong lúc code

```bash
cd editors/vscode
npm install
npm run compile
```

Mở VS Code tại thư mục này rồi nhấn `F5`. Cửa sổ thử nghiệm mở kho `td` trong
chính kho này làm ví dụ.

## Những gì tiện ích làm được

| Việc | Cách làm trong khung Source Control |
| --- | --- |
| Xem thay đổi | Nhóm **Staged Changes**, **Changes**, **Merge Changes**, **Untracked Changes** |
| Commit | Gõ nội dung vào ô nhập rồi bấm dấu tích, hoặc `TD: Commit` |
| Stage, unstage, huỷ thay đổi | Nút cộng/trừ/thùng rác cạnh từng tệp, và trên tiêu đề nhóm |
| Xem khác biệt | Bấm tên tệp để mở khung so sánh, hoặc bấm biểu tượng ngay cạnh |
| Đường ngữ cảnh trong trình soạn thảo | Dấu nháy và dải khác biệt ở rìa tệp, giống git |
| Chữ viết tắt trên cây thư mục | `M` đã sửa, `A` thêm, `D` xoá, `U` mới hoặc xung đột |
| Nhánh, tag, bản lưu tạm | Các khung **Branches**, **Commits**, **Stashes**, **Tags** |
| Thanh trạng thái | Nhánh hiện tại, số commit đi trước/đi sau, số tệp chờ commit, lỗi |

## Cách tiện ích nói chuyện với td

Tiện ích không tự đọc tệp `.tdx`. Mọi thứ đều đi qua `td vcs`, ví dụ:

```
td vcs -C /du-an status
td vcs -C /du-an diff --staged -U1000000 -- main.go
td vcs -C /du-an commit -m "nội dung"
```

Hai điểm đáng lưu ý:

- **`-C` luôn đặt trước mọi thứ khác.** Đặt cuối dòng lệnh thì khi lệnh có danh
  sách tệp sau dấu `--`, td sẽ hiểu `-C` là một đường dẫn nữa và chạy nhầm ở
  thư mục hiện tại.
- **Diff được xin với số dòng ngữ cảnh rất lớn.** Khung so sánh của VS Code cần
  nội dung đầy đủ của cả hai phía rồi tự tính vạch hiệu, nên phải có diff ôm
  trọn tệp. Lệnh luôn được gọi cho đúng một tệp nên lượng dữ liệu vẫn nhỏ, kể
  cả trong kho lớn.

Toàn bộ phần dịch output của td nằm trong `src/parse.ts`. Nếu cây lệnh của td
đổi câu chữ, đó là chỗ duy nhất phải sửa.

## Lệnh của `td` mà tiện ích dùng

| Việc | Lệnh |
| --- | --- |
| Đọc trạng thái | `td vcs status` |
| Khác biệt hai phía | `td vcs diff [--staged] -U1000000 -- <tệp>` |
| Khác biệt của một commit | `td vcs diff <mã băm> -U1000000 -- <tệp>` |
| Lịch sử | `td vcs log --oneline -n <số>` |
| Nhánh, tag, bản lưu tạm | `td vcs branch -vv`, `td vcs tag -l`, `td vcs stash --list` |
| Stage, unstage, huỷ | `td vcs add`, `td vcs restore --staged`, `td vcs restore` |
| Xoá tệp chưa theo dõi | `td vcs clean -f <tệp>` |

## Những chỗ chưa làm được

td là công cụ quản lý phiên bản **cục bộ**, không có remote, nên tiện ích cố
tình không có những phần mà git dựa vào remote:

- Không có Pull, Push, Fetch, Publish, Sync, Remote.
- Không có Stage Selected Ranges: git đưa từng đoạn đang chọn vào vùng chuẩn bị,
  việc đó cần td ghi được entry từng phần vào index nên `td` chưa làm.
- Không có Commit & Push, Commit & Sync, vì không có nơi để đẩy lên.

Thanh trạng thái vẫn hiện số commit đi trước/đi sau khi nhánh có cấu hình
`branch.<tên>.merge`, vì td có đọc cấu hình đó để tính.

## Cấu trúc mã nguồn

| Tệp | Vai trò |
| --- | --- |
| `src/extension.ts` | Khởi động, cấp nội dung ảo cho khung so sánh |
| `src/model.ts` | Dò kho trong workspace, theo dõi tệp, đặt thanh trạng thái |
| `src/repository.ts` | Một kho: SourceControl, các nhóm thay đổi, quick diff |
| `src/td.ts` | Bọc lệnh `td vcs`, không đọc `.tdx` bằng tay |
| `src/parse.ts` | Dịch output tiếng Việt của td thành dữ liệu có kiểu |
| `src/commands.ts` | Toàn bộ lệnh |
| `src/views.ts` | Bốn khung nhánh, commit, stash, tag |
| `src/decorations.ts` | Chữ viết tắt trên cây thư mục |
| `src/uri.ts` | Lược đồ `td:` cho nội dung nằm trong kho |

## Cấu hình

| Khoá | Mặc định | Ý nghĩa |
| --- | --- | --- |
| `td.path` | `td` | Đường dẫn lệnh td |
| `td.enabled` | `true` | Bật dò kho trong workspace |
| `td.autoRefresh` | `true` | Tự làm mới khi tệp trên đĩa đổi |
| `td.autorefreshDelay` | `1000` | Chờ bao nhiêu mili giây trước khi làm mới |
| `td.decorations.enabled` | `true` | Hiện chữ viết tắt trên cây thư mục |
| `td.showCommitInput` | `true` | Hiện ô nhập nội dung commit |
| `td.alwaysShowStagedChangesResourceGroup` | `true` | Luôn hiện nhóm Staged Changes |
| `td.openDiffOnClick` | `false` | Bấm tệp thì mở khung so sánh thay vì mở tệp |
| `td.confirmEmptyCommits` | `true` | Hỏi lại trước khi tạo commit rỗng |
| `td.logMaxCount` | `500` | Số commit tối đa ở khung Commits |

## Kiểm thử

```bash
npm run compile
npm test
```

Phần kiểm thử chạy lệnh `td` thật trên các kho tạm trong thư mục tạm. Đặt biến
môi trường `TD_BIN` để trỏ tới một tệp thực thi cụ thể. Không tìm thấy `td` thì
các kiểm thử cần lệnh thật được bỏ qua, phần kiểm thử dữ liệu mẫu vẫn chạy.