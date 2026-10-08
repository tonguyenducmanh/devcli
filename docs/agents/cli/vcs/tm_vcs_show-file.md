## tm vcs show-file

In nội dung một tệp ở một điểm trong lịch sử

### Synopsis

In nguyên văn nội dung một tệp như nó nằm trong cây của một commit.

Điểm xuất phát là một commit, tên nhánh hoặc một tham chiệu khác; mặc định lấy
HEAD. Sau dấu hai gạch ngang là danh sách tệp cần in.

Khác với "tm vcs diff", lệnh này đọc thẳng nội dung đã lưu trong kho nên tệp
không bị thay đổi ở commit đó vẫn in ra được.

Kèm --index thì điểm xuất phát là vùng chuẩn bị thay vì lịch sử, in ra nội dung
sẽ được ghi vào commit kế tiếp. Cách này đúng cả khi tệp đã khớp vùng chuẩn bị:
lúc đó không còn khác biệt nào để dựng lại nội dung, mà vùng chuẩn bị vẫn còn đầy đủ.

```
tm vcs show-file <điểm> [-- tệp...] [flags]
```

### Examples

```
  tm vcs show-file HEAD -- cmd/root.go
  tm vcs show-file v1.0.0 -- README.md
  tm vcs show-file abc1234 -- a.txt b.txt
  tm vcs show-file --index -- a.txt   nội dung đã stage của a.txt
```

### Options

```
  -h, --help    hiển thị phần trợ giúp của lệnh này
      --index   đọc nội dung trong vùng chuẩn bị thay vì trong lịch sử
```

### Options inherited from parent commands

```
  -C, --dir string   chạy lệnh tại thư mục khác
```

### SEE ALSO

* [tm vcs](tm_vcs.md)	 - Quản lý phiên bản mã nguồn cục bộ

