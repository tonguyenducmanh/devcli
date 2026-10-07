## td vcs rebase

Di chuyển các commit hiện tại lên trên một điểm khác

### Synopsis

Áp dụng lại các commit của một nhánh lên trên một điểm khác.

Mỗi commit được áp dụng lại bằng cách hợp nhất nội dung với trạng thái hiện
tại, nên lịch sử trở nên gọn và tuyến tính thay vì nhiều nhánh song song.

Có thể rebase một nhánh khác bằng cách truyền cả hai tham số: điểm đích trước,
tên nhánh sau. Dùng --onto khi muốn điểm đích khác điểm gốc.

```
td vcs rebase [đích] [flags]
```

### Examples

```
  td vcs rebase main            đưa commit hiện tại lên trên main
  td vcs rebase <đích> <nhánh>   rebase một nhánh khác lên trên đích
  td vcs rebase --onto <đích>   chỉ định điểm đích khác
  td vcs rebase --continue      tiếp tục sau khi giải quyết xung đột
  td vcs rebase --abort         quay lại trạng thái trước rebase
```

### Options

```
      --abort           huỷ rebase và trở về trạng thái trước đó
      --branch string   rebase cho nhánh này thay vì nhánh hiện tại
      --continue        tiếp tục rebase sau khi giải quyết xung đột
  -h, --help            hiển thị phần trợ giúp của lệnh này
      --onto string     commit đích thay cho điểm gốc
      --skip            bỏ qua commit hiện tại
```

### Options inherited from parent commands

```
  -C, --dir string   chạy lệnh tại thư mục khác
      --no-color     tắt màu trong output
  -v, --verbose      in thêm thông tin chi tiết
```

### SEE ALSO

* [td vcs](td_vcs.md)	 - Quản lý phiên bản mã nguồn cục bộ

