## tm vcs merge

Hợp nhất một nhánh vào nhánh hiện tại

### Synopsis

Kết hợp lịch sử của một nhánh khác vào nhánh đang đứng.

Khi nhánh đích nằm ngay sau nhánh hiện tại, con trỏ chỉ dịch thẳng sang đó mà
không tạo mốc mới. Trong trường hợp lệch nhánh, tm so từng tệp và hợp nhất nội
dung ba phía; tệp không thể tự động hợp nhất sẽ được đánh dấu xung đột.

Gặp xung đột thì lệnh dừng lại và ghi lại trạng thái, dùng --continue để hoàn
tất hoặc --abort để huỷ.

```
tm vcs merge <nhánh> [flags]
```

### Examples

```
  tm vcs merge main            hợp nhất nhánh main
  tm vcs merge --no-ff main    luôn tạo commit merge
  tm vcs merge --abort         huỷ lần merge đang dở dang
  tm vcs merge --continue      hoàn tất sau khi giải quyết xung đột
```

### Options

```
      --abort            huỷ lần merge đang dở dang
      --continue         hoàn tất merge sau khi giải quyết xung đột
      --ff-only          chỉ cho phép fast-forward
  -h, --help             hiển thị phần trợ giúp của lệnh này
  -m, --message string   nội dung cho commit merge
      --no-ff            luôn tạo commit merge dù có thể fast-forward
      --squash           gộp thay đổi vào vùng stage mà không tạo commit merge
```

### Options inherited from parent commands

```
  -C, --dir string   chạy lệnh tại thư mục khác
```

### SEE ALSO

* [tm vcs](tm_vcs.md)	 - Quản lý phiên bản mã nguồn cục bộ

