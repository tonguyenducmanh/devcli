## td vcs fsck

Kiểm tra tính toàn vẹn của kho và các tham chiếu

### Synopsis

Duyệt toàn bộ kho kiểm tra hai điều:

  - Mọi object trong kho có đọc được không (nén zlib không hỏng, header hợp lệ).
  - Mọi tham chiếu và HEAD có trỏ tới một object tồn tại không.

Lệnh trả về mã thoát khác 0 nếu phát hiện vấn đề.

```
td vcs fsck [flags]
```

### Examples

```
  td vcs fsck
  td vcs -C du-an-khac fsck
```

### Options

```
  -h, --help   hiển thị phần trợ giúp của lệnh này
```

### Options inherited from parent commands

```
  -C, --dir string   chạy lệnh tại thư mục khác
      --no-color     tắt màu trong output
  -v, --verbose      in thêm thông tin chi tiết
```

### SEE ALSO

* [td vcs](td_vcs.md)	 - Quản lý phiên bản mã nguồn cục bộ

