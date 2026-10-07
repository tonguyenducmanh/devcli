## td vcs status

Hiển thị trạng thái thay đổi của cây làm việc

### Synopsis

In trạng thái hiện tại của cây làm việc theo ba nhóm:

  đã stage       nội dung sẽ được ghi vào commit kế tiếp
  chưa stage     đã sửa trên đĩa nhưng chưa đưa vào vùng chuẩn bị
  chưa theo dõi  tệp mới xuất hiện, td chưa quản lý

Ký hiệu đầu mỗi dòng cho biết thao tác: thêm, sửa, xoá hoặc mới.

Dòng đầu tiên cho biết đang ở nhánh nào, HEAD có đang tách rời không, và nhánh
đó đi trước hay đi sau nhánh theo dõi bao nhiêu commit.

Tệp nào không xuất hiện ở đây thì đã bị bỏ qua theo .tdxignore hoặc
.tdx/info/exclude, xem "td vcs --help" để biết cách viết mẫu.

```
td vcs status [flags]
```

### Examples

```
  td vcs status
  td vcs status -C thư-mục-khác
  td vcs st
```

### Options

```
  -h, --help      hiển thị phần trợ giúp của lệnh này
  -v, --verbose   liệt kê cả các tệp chưa được theo dõi
```

### Options inherited from parent commands

```
  -C, --dir string   chạy lệnh tại thư mục khác
```

### SEE ALSO

* [td vcs](td_vcs.md)	 - Quản lý phiên bản mã nguồn cục bộ

