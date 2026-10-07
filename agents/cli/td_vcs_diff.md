## td vcs diff

Hiển thị khác biệt giữa các phiên bản

### Synopsis

Hiển thị khác biệt giữa hai vùng bất kỳ.

Mặc định so sánh vùng đã stage với cây làm việc, tức là những sửa đổi chưa
được stage. Dùng --staged để so sánh HEAD với vùng đã stage.

Tham số phạm vi nhận một tên nhánh, một mã băm, hoặc hai mã băm nối bằng hai
dấu chấm để so sánh trực tiếp, ba dấu chấm để so sánh từ điểm chung gần nhất.

Sau dấu hai gạch ngang là danh sách tệp cần lọc.

```
td vcs diff [phạm vi] [-- tệp...] [flags]
```

### Examples

```
  td vcs diff                       đã stage so với cây làm việc
  td vcs diff --staged              HEAD so với vùng đã stage
  td vcs diff main                  commit hiện tại so với main
  td vcs diff main..feature         so sánh hai nhánh
  td vcs diff main...feature        so sánh từ điểm chung gần nhất
  td vcs diff --stat HEAD~1         chỉ xem thống kê thay đổi
```

### Options

```
      --color string   màu output: auto, always, never (default "auto")
  -h, --help           help for diff
      --name-only      chỉ hiển thị tên tệp thay đổi
  -c, --staged         so sánh HEAD với vùng đã stage
      --stat           chỉ hiển thị thống kê thay đổi
  -U, --unified int    số dòng ngữ cảnh quanh thay đổi (default 3)
```

### Options inherited from parent commands

```
  -C, --dir string   chạy lệnh tại thư mục khác
      --no-color     tắt màu trong output
  -v, --verbose      in thêm thông tin chi tiết
```

### SEE ALSO

* [td vcs](td_vcs.md)	 - Quản lý phiên bản mã nguồn cục bộ

