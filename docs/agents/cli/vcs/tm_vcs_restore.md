## tm vcs restore

Khôi phục lại nội dung tệp

### Synopsis

Đưa lại nội dung tệp từ một nguồn khác về cây làm việc hoặc vùng chuẩn bị.

Mặc định lấy nội dung đang có trong vùng chuẩn bị, tức là huỷ các sửa đổi chưa
stage. Dùng --staged để chỉ gỡ khỏi vùng chuẩn bị mà giữ nguyên cây làm việc,
và --source để lấy từ một commit bất kỳ.

```
tm vcs restore <tệp>... [flags]
```

### Examples

```
  tm vcs restore main.go         lấy lại nội dung đang stage
  tm vcs restore --staged main.go  gỡ thay đổi đã stage
  tm vcs restore --source=HEAD~1 main.go  lấy từ một commit khác
```

### Options

```
  -h, --help            hiển thị phần trợ giúp của lệnh này
  -s, --source string   lấy nội dung từ một commit thay vì index
  -S, --staged          chỉ thay đổi vùng stage, giữ nguyên cây làm việc
  -v, --verbose         in ra từng tệp đã khôi phục
  -W, --worktree        chỉ thay đổi cây làm việc
```

### Options inherited from parent commands

```
  -C, --dir string   chạy lệnh tại thư mục khác
```

### SEE ALSO

* [tm vcs](tm_vcs.md)	 - Quản lý phiên bản mã nguồn cục bộ

