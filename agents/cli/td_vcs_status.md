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
  -h, --help   help for status
```

### Options inherited from parent commands

```
  -C, --dir string   chạy lệnh tại thư mục khác
      --no-color     tắt màu trong output
  -v, --verbose      in thêm thông tin chi tiết
```

### SEE ALSO

* [td vcs](td_vcs.md)	 - Quản lý phiên bản mã nguồn cục bộ

