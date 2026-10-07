## td vcs stash

Lưu tạm và khôi phục các thay đổi chưa commit

### Synopsis

Cất các thay đổi chưa commit vào một kho tạm rồi đưa cây làm việc
về trạng thái sạch.

Mỗi lần lưu tạo thêm một mục trong danh sách. Dùng -u để cất kèm cả tệp chưa
theo dõi; những tệp này sẽ bị gỡ khỏi đĩa và trở lại khi áp dụng lại.

Số ở đối số chỉ vị trí trong danh sách, tính từ 0 cho mục mới nhất.

```
td vcs stash [flags]
```

### Examples

```
  td vcs stash                 lưu thay đổi hiện tại
  td vcs stash -u              kèm cả file chưa được theo dõi
  td vcs stash list            xem các bản đã lưu
  td vcs stash apply           áp dụng bản mới nhất, giữ lại trong danh sách
  td vcs stash pop             áp dụng rồi xóa bản đó
  td vcs stash drop            xóa một bản
  td vcs stash clear           xóa toàn bộ
```

### Options

```
      --apply               áp dụng mà không xóa khỏi danh sách
      --clear               xóa toàn bộ bản lưu tạm
      --drop                xóa một bản lưu tạm
  -h, --help                help for stash
  -u, --include-untracked   kèm cả file chưa được theo dõi
  -l, --list                liệt kê các bản đã lưu
  -m, --message string      mô tả ngắn cho bản lưu tạm
      --pop                 áp dụng rồi xóa khỏi danh sách
```

### Options inherited from parent commands

```
  -C, --dir string   chạy lệnh tại thư mục khác
      --no-color     tắt màu trong output
  -v, --verbose      in thêm thông tin chi tiết
```

### SEE ALSO

* [td vcs](td_vcs.md)	 - Quản lý phiên bản mã nguồn cục bộ

