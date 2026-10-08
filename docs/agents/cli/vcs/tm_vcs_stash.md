## tm vcs stash

Lưu tạm và khôi phục các thay đổi chưa commit

### Synopsis

Cất các thay đổi chưa commit vào một kho tạm rồi đưa cây làm việc
về trạng thái sạch.

Mỗi lần lưu tạo thêm một mục trong danh sách. Dùng -u để cất kèm cả tệp chưa
theo dõi; những tệp này sẽ bị gỡ khỏi đĩa và trở lại khi áp dụng lại.

Số ở đối số chỉ vị trí trong danh sách, tính từ 0 cho mục mới nhất.

```
tm vcs stash [flags]
```

### Examples

```
  tm vcs stash                 lưu thay đổi hiện tại
  tm vcs stash -u              kèm cả file chưa được theo dõi
  tm vcs stash list            xem các bản đã lưu
  tm vcs stash apply           áp dụng bản mới nhất, giữ lại trong danh sách
  tm vcs stash pop             áp dụng rồi xóa bản đó
  tm vcs stash drop            xóa một bản
  tm vcs stash clear           xóa toàn bộ
```

### Options

```
      --apply               áp dụng mà không xóa khỏi danh sách
      --clear               xóa toàn bộ bản lưu tạm
      --drop                xóa một bản lưu tạm
  -h, --help                hiển thị phần trợ giúp của lệnh này
  -u, --include-untracked   kèm cả file chưa được theo dõi
  -l, --list                liệt kê các bản đã lưu
  -m, --message string      mô tả ngắn cho bản lưu tạm
      --pop                 áp dụng rồi xóa khỏi danh sách
```

### Options inherited from parent commands

```
  -C, --dir string   chạy lệnh tại thư mục khác
```

### SEE ALSO

* [tm vcs](tm_vcs.md)	 - Quản lý phiên bản mã nguồn cục bộ

