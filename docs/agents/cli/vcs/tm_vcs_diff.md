## tm vcs diff

Hiển thị khác biệt giữa các phiên bản

### Synopsis

Hiển thị khác biệt giữa hai vùng bất kỳ.

Mặc định so sánh vùng đã stage với cây làm việc, tức là những sửa đổi chưa
được stage. Dùng --staged để so sánh HEAD với vùng đã stage.

Tham số phạm vi nhận một tên nhánh, một mã băm, hoặc hai mã băm nối bằng hai
dấu chấm để so sánh trực tiếp, ba dấu chấm để so sánh từ điểm chung gần nhất.

Sau dấu hai gạch ngang là danh sách tệp cần lọc.

```
tm vcs diff [phạm vi] [-- tệp...] [flags]
```

### Examples

```
  tm vcs diff                       đã stage so với cây làm việc
  tm vcs diff --staged              HEAD so với vùng đã stage
  tm vcs diff main                  commit hiện tại so với main
  tm vcs diff main..feature         so sánh hai nhánh
  tm vcs diff main...feature        so sánh từ điểm chung gần nhất
  tm vcs diff --stat HEAD~1         chỉ xem thống kê thay đổi
```

### Options

```
      --color string   màu output: auto, always, never (default "auto")
  -h, --help           hiển thị phần trợ giúp của lệnh này
      --name-only      chỉ hiển thị tên tệp thay đổi
  -c, --staged         so sánh HEAD với vùng đã stage
      --stat           chỉ hiển thị thống kê thay đổi
  -U, --unified int    số dòng ngữ cảnh quanh thay đổi (default 3)
```

### Options inherited from parent commands

```
  -C, --dir string   chạy lệnh tại thư mục khác
```

### SEE ALSO

* [tm vcs](tm_vcs.md)	 - Quản lý phiên bản mã nguồn cục bộ

