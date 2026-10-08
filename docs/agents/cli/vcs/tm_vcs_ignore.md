## tm vcs ignore

In các quy tắc bỏ qua tệp đang có trong kho

### Synopsis

In nội dung mọi tệp chứa quy tắc bỏ qua tệp của kho.

Mỗi tệp in ra kèm đường dẫn để biết quy tắc đó có hiệu lực ở đâu. Tệp ở thư mục
nông in trước, quy tắc ở thư mục sâu hơn in sau và thắng, đúng như khi thực thi.

Nếu kho chưa có quy tắc nào thì lệnh in luôn phần trợ giúp này.

Tệp ignore nằm ở đâu:

  .tmxignore         ở gốc dự án hoặc trong bất kỳ thư mục con nào. Quy tắc chỉ
                     áp dụng bên trong thư mục chứa nó, và không lan sang thư mục
                     khác. Nên commit tệp này để cả nhóm cùng dùng.
  .tmx/info/exclude  riêng cho máy này, nằm trong .tmx nên không được commit.

Cách viết mẫu:

  *.log           bỏ qua mọi tệp kết thúc bằng .log, ở mọi cấp thư mục
  build           bỏ qua thư mục build, ở mọi cấp
  /build          chỉ bỏ qua thư mục build nằm ở gốc dự án
  docs/*.tmp      bỏ qua tệp .tmp nằm trong thư mục docs
  **/cache/       bỏ qua thư mục cache, ở mọi cấp
  !giữ-lại.log    phủ định lại quy tắc trước, tệp này lại được theo dõi
  # ghi chú       dòng bắt đầu bằng dấu # là chú thích

Dấu / ở cuối mẫu nói mẫu đó chỉ áp dụng cho thư mục. Dấu * không vượt qua dấu
/, dấu ** vượt được nhiều cấp. Quy tắc ở dưới thắng quy tắc ở trên.

Hai tệp này được tm đọc tự động, không cần khai báo ở đâu. Tệp bị bỏ qua sẽ
không xuất hiện trong status và không được add vào vùng chuẩn bị.

```
tm vcs ignore [flags]
```

### Examples

```
  tm vcs ignore          in các quy tắc bỏ qua đang có
  tm vcs ignore --help   xem cách viết mẫu bỏ qua tệp và thư mục
```

### Options

```
  -h, --help   hiển thị phần trợ giúp của lệnh này
```

### Options inherited from parent commands

```
  -C, --dir string   chạy lệnh tại thư mục khác
```

### SEE ALSO

* [tm vcs](tm_vcs.md)	 - Quản lý phiên bản mã nguồn cục bộ

