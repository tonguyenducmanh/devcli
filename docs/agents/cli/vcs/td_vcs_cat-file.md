## td vcs cat-file

In nội dung của một object (blob, tree, commit, tag)

### Synopsis

Đọc một object trong kho và kiểm tra loại của nó rồi in nội dung.

Loại hợp lệ: blob, tree, commit, tag. Nếu loại truyền vào không khớp với loại
thật của object thì lệnh báo lỗi. Kèm cờ -v để in chi tiết: với blob là nội
dung đầy đủ, với tree là từng entry kèm chế độ và mã băm, với commit và tag là
toàn bộ phần thô.

```
td vcs cat-file <loại> <mã-băm> [flags]
```

### Examples

```
  # Xác minh loại và in nội dung
  td vcs cat-file blob a1b2c3d4

  # In toàn bộ nội dung thô của commit HEAD
  td vcs cat-file commit HEAD -v
```

### Options

```
  -h, --help      hiển thị phần trợ giúp của lệnh này
  -v, --verbose   in ra dạng đã rút gọn cho dễ đọc
```

### Options inherited from parent commands

```
  -C, --dir string   chạy lệnh tại thư mục khác
```

### SEE ALSO

* [td vcs](td_vcs.md)	 - Quản lý phiên bản mã nguồn cục bộ

