## tm vcs reset

Di chuyển con trỏ HEAD về commit khác

### Synopsis

Di chuyển con trỏ HEAD về một commit khác, kèm mức độ áp dụng cho vùng chuẩn bị
và cây làm việc.

  --soft   chỉ di chuyển HEAD, giữ nguyên vùng chuẩn bị và cây làm việc
  --mixed  di chuyển HEAD và nạp lại vùng chuẩn bị (mặc định)
  --hard   di chuyển HEAD, vùng chuẩn bị và cây làm việc

Chế độ --hard ghi đè mọi thay đổi chưa lưu, dùng cẩn thận.

Nếu có danh sách tệp sau dấu hai gạch ngang, lệnh chỉ gỡ những tệp đó khỏi vùng
chuẩn bị, không động tới con trỏ HEAD.

```
tm vcs reset [commit] [tệp...] [flags]
```

### Examples

```
  tm vcs reset --soft HEAD~1
  tm vcs reset HEAD~1 -- main.go   chỉ gỡ main.go khỏi vùng stage
```

### Options

```
      --hard    di chuyển HEAD, index và cây làm việc
  -h, --help    hiển thị phần trợ giúp của lệnh này
      --mixed   di chuyển HEAD và nạp lại index
      --soft    chỉ di chuyển HEAD
```

### Options inherited from parent commands

```
  -C, --dir string   chạy lệnh tại thư mục khác
```

### SEE ALSO

* [tm vcs](tm_vcs.md)	 - Quản lý phiên bản mã nguồn cục bộ

