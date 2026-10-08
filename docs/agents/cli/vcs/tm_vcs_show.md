## tm vcs show

Hiển thị chi tiết của một commit

### Synopsis

In thông tin đầy đủ của một commit: mã băm, các tham chiếu trỏ tới nó,
tác giả, thời điểm, nội dung thông điệp và danh sách tệp bị thay đổi.

Mặc định lấy HEAD. Kèm -p để in luôn nội dung khác biệt của từng tệp.

```
tm vcs show [commit] [flags]
```

### Examples

```
  tm vcs show
  tm vcs show HEAD~2
  tm vcs show abc1234
```

### Options

```
  -h, --help    hiển thị phần trợ giúp của lệnh này
  -p, --patch   kèm nội dung diff (default true)
```

### Options inherited from parent commands

```
  -C, --dir string   chạy lệnh tại thư mục khác
```

### SEE ALSO

* [tm vcs](tm_vcs.md)	 - Quản lý phiên bản mã nguồn cục bộ

