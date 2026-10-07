# Cấu trúc dự án và cách lưu dữ liệu

## Cấu trúc thư mục

```
main.go                      điểm khởi động
cmd/                         cây lệnh
  root.go                    lệnh gốc `td`, đăng ký các nhóm công cụ và các biến toàn cục
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
internal/env/                hàm quản lý môi trường và đường dẫn toàn cục (vd: ~/.td)
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
