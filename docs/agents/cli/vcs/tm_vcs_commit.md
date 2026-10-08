## tm vcs commit

Ghi lại các thay đổi đã stage thành một commit

### Synopsis

Ghi lại nội dung vùng chuẩn bị thành một mốc có tên trong lịch sử.

Mỗi commit lưu cây nội dung đầy đủ, tác giả và thời điểm, đồng thời trỏ tới
commit cha nên tạo thành một chuỗi lịch sử. Dùng --amend để viết lại commit
vừa tạo thay vì tạo mốc mới.

Thông điệp phải truyền bằng cờ -m, lặp lại -m để tách tiêu đề và phần mô tả
chi tiết thành hai đoạn.

```
tm vcs commit [flags]
```

### Examples

```
  tm vcs commit -m "tin nhắn ngắn"
  tm vcs commit -m "tiêu đề" -m "mô tả chi tiết"
  tm vcs commit --amend -m "sửa lại commit vừa tạo"
  tm vcs commit -a          stage mọi thay đổi rồi commit luôn
```

### Options

```
  -a, --all                   đưa mọi thay đổi vào stage trước khi commit
      --allow-empty           cho phép tạo commit dù không có thay đổi
      --amend                 sửa lại commit vừa tạo
  -h, --help                  hiển thị phần trợ giúp của lệnh này
  -m, --message stringArray   nội dung commit, lặp lại để thêm đoạn mô tả
```

### Options inherited from parent commands

```
  -C, --dir string   chạy lệnh tại thư mục khác
```

### SEE ALSO

* [tm vcs](tm_vcs.md)	 - Quản lý phiên bản mã nguồn cục bộ

