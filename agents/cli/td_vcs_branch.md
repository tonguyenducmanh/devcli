## td vcs branch

Liệt kê, tạo hoặc xóa nhánh

### Synopsis

Liệt kê, tạo, đổi tên và xoá các nhánh cục bộ.

Nhánh chỉ là một con trỏ trỏ tới một commit nên tạo nhánh không tốn chi phí sao
chép. Dấu * đánh dấu nhánh đang đứng.

Xoá nhánh chỉ thành công nếu nhánh đó đã được hợp nhất vào nhánh hiện tại; dùng
-D để bỏ qua kiểm tra này. Không thể xoá nhánh đang đứng.

Không có tham số thì in danh sách. Một tham số là tạo nhánh mới tại HEAD, hai
tham số là tạo nhánh mới từ một điểm xuất phát cho trước.

```
td vcs branch [flags]
```

### Examples

```
  td vcs branch                  liệt kê các nhánh
  td vcs branch -d ten           xóa nhánh đã hợp nhất
  td vcs branch -D ten           xóa nhánh bất kể trạng thái
  td vcs branch ten              tạo nhánh tại HEAD
  td vcs branch ten main         tạo nhánh từ nhánh main
  td vcs branch -m cũ mới        đổi tên nhánh
```

### Options

```
  -d, --delete    xóa nhánh
  -D, --force     xóa nhánh bất kể đã hợp nhất hay chưa
  -h, --help      help for branch
  -l, --list      liệt kê các nhánh
  -m, --move      đổi tên nhánh
  -s, --switch    chuyển sang nhánh mới sau khi đổi tên
  -t, --track     ghi nhớ nhánh theo dõi cho nhánh mới
  -v, --verbose   kèm hash của từng nhánh
```

### Options inherited from parent commands

```
  -C, --dir string   chạy lệnh tại thư mục khác
      --no-color     tắt màu trong output
```

### SEE ALSO

* [td vcs](td_vcs.md)	 - Quản lý phiên bản mã nguồn cục bộ

