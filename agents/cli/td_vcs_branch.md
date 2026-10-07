## td vcs branch

Liệt kê, tạo, sao chép, đổi tên hoặc xoá các nhánh cục bộ

### Synopsis

Liệt kê, tạo, sao chép, đổi tên và xoá các nhánh cục bộ.

Nhánh chỉ là một con trỏ trỏ tới một commit nên tạo nhánh không tốn chi phí sao
chép. Dấu * đánh dấu nhánh đang đứng.

Xoá nhánh chỉ thành công nếu nhánh đó đã được hợp nhất vào nhánh hiện tại; dùng
-D để bỏ qua kiểm tra này. Không thể xoá nhánh đang đứng.

Không có tham số thì in danh sách. Một tham số là tạo nhánh mới tại HEAD, hai
tham số là tạo nhánh mới từ một điểm xuất phát cho trước.

Cờ liệt kê và cờ lọc dùng chung được, ví dụ --merged cùng --sort. Tham số
đưa vào không phải thao tác thì được hiểu là mẫu lọc tên nhánh, giống git.

```
td vcs branch [flags]
```

### Examples

```
  td vcs branch                  liệt kê các nhánh
  td vcs branch -v               liệt kê kèm mã băm và tiêu đề commit
  td vcs branch -vv              như trên, thêm cả nhánh đang theo dõi
  td vcs branch --show-current   in tên nhánh đang đứng
  td vcs branch -l 'tinh-*'      liệt kê các nhánh khớp mẫu
  td vcs branch --merged main    chỉ liệt kê nhánh đã hợp nhất vào main
  td vcs branch -d ten           xoá nhánh đã hợp nhất
  td vcs branch -D ten           xoá nhánh bất kể trạng thái
  td vcs branch ten              tạo nhánh và chuyển sang đó
  td vcs branch ten main         tạo nhánh từ nhánh main
  td vcs branch -m cũ mới        đổi tên nhánh
  td vcs branch -c ten bản-sao   sao chép nhánh ten thành bản-sao
```

### Options

```
      --abbrev string      số ký tự của mã băm rút gọn, mặc định 8, đặt 0 để in đầy đủ
      --contains string    chỉ in nhánh có chứa commit cho trước
  -c, --copy               sao chép nhánh
  -f, --create-force       ép tạo, ghi đè nhánh đã có
  -d, --delete             xoá nhánh đã hợp nhất
  -D, --force              xoá nhánh bất kể đã hợp nhất hay chưa
      --format string      định dạng từng dòng: %s tên, %h mã băm ngắn, %H mã băm đầy đủ, %d tiêu đề
  -h, --help               hiển thị phần trợ giúp của lệnh này
  -l, --list               liệt kê tên nhánh, có thể kèm mẫu lọc
      --merged string      chỉ in nhánh đã hợp nhất vào commit cho trước
  -m, --move               đổi tên nhánh
      --no-merged string   chỉ in nhánh chưa hợp nhất vào commit cho trước
      --points-at string   chỉ in nhánh trỏ tới đúng object cho trước
  -q, --quiet              chỉ báo lỗi, không in thông báo thành công
      --show-current       chỉ in tên nhánh đang đứng
      --sort string        sắp xếp theo khoá: name, -name, committerdate, -committerdate
  -t, --track              ghi nhớ nhánh theo dõi cho nhánh mới
  -v, --verbose count      liệt kê kèm mã băm và tiêu đề commit, dùng hai lần thì in thêm nhánh đang theo dõi
```

### Options inherited from parent commands

```
  -C, --dir string   chạy lệnh tại thư mục khác
```

### SEE ALSO

* [td vcs](td_vcs.md)	 - Quản lý phiên bản mã nguồn cục bộ

