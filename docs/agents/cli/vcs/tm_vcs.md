## tm vcs

Quản lý phiên bản mã nguồn cục bộ

### Synopsis

Nhóm lệnh quản lý phiên bản mã nguồn cục bộ.

Toàn bộ dữ liệu của nhóm này nằm trong thư mục .tmx cạnh dự án, gồm lịch
sử commit, các nhánh, các tag và vùng chuẩn bị.

Mỗi lệnh dưới đây chạy trên kho tìm thấy bằng cách đi lên từ thư mục hiện
tại. Dùng -C để chỉ định thư mục khác.

BỎ QUA TỆP KHÔNG MUỐN THEO DÕI

tm không tự đoán tệp nào là tạm, tệp nào là dữ liệu do trình biên dịch sinh
ra. Muốn bỏ qua thì ghi mẫu vào tệp .tmxignore ở gốc dự án, mỗi dòng một mẫu:

  *.log           bỏ qua mọi tệp kết thúc bằng .log, ở mọi cấp thư mục
  build           bỏ qua thư mục build, ở mọi cấp
  /build          chỉ bỏ qua thư mục build nằm ở gốc dự án
  docs/*.tmp      bỏ qua tệp .tmp nằm trong thư mục docs
  **/cache/       bỏ qua thư mục cache, ở mọi cấp
  !giữ-lại.log    phủ định lại quy tắc trước, tệp này lại được theo dõi
  # ghi chú       dòng bắt đầu bằng dấu # là chú thích

Dấu / ở cuối mẫu nói mẫu đó chỉ áp dụng cho thư mục. Dấu * không vượt qua dấu
/, dấu ** vượt được nhiều cấp. Quy tắc ở dưới thắng quy tắc ở trên.

Muốn quy tắc chỉ áp dụng cho riêng máy này thì ghi vào .tmx/info/exclude, tệp
đó nằm trong .tmx nên không được commit. Còn .tmxignore nằm ở gốc dự án nên
có thể commit để cả nhóm cùng dùng.

Quy tắc cũng đặt được trong thư mục con, khi đó nó chỉ áp dụng bên trong thư mục
đó chứ không lan sang nơi khác.

Hai tệp này được tm đọc tự động, không cần khai báo ở đâu. Tệp bị bỏ qua sẽ
không xuất hiện trong status và không được add vào vùng chuẩn bị.

Xem các quy tắc đang có trong kho: tm vcs ignore

```
tm vcs [flags]
```

### Examples

```
  # Khởi tạo kho rồi ghi lại thay đổi đầu tiên
  tm vcs init
  tm vcs add .
  tm vcs commit -m "tin nhắn đầu tiên"

  # Xem nhánh hiện tại và lịch sử gọn
  tm vcs status
  tm vcs log --oneline -n 10

  # Tạo nhánh, làm việc rồi hợp nhất về nhánh chính
  tm vcs switch -c tinh-nang
  tm vcs commit -am "bổ sung tính năng"
  tm vcs switch main
  tm vcs merge tinh-nang
```

### Options

```
  -C, --dir string   chạy lệnh tại thư mục khác
  -h, --help         hiển thị phần trợ giúp của lệnh này
  -v, --verbose      in thêm tình trạng kho mã nguồn hiện tại
```

### SEE ALSO

* [tm](../tm.md)	 - Bộ công cụ dòng lệnh cá nhân, tất cả gói dưới một lệnh duy nhất
* [tm vcs add](tm_vcs_add.md)	 - Đưa thay đổi vào vùng chuẩn bị commit
* [tm vcs branch](tm_vcs_branch.md)	 - Liệt kê, tạo, sao chép, đổi tên hoặc xoá các nhánh cục bộ
* [tm vcs cat-file](tm_vcs_cat-file.md)	 - In nội dung của một object (blob, tree, commit, tag)
* [tm vcs checkout](tm_vcs_checkout.md)	 - Chuyển sang nhánh hoặc commit khác
* [tm vcs cherry-pick](tm_vcs_cherry-pick.md)	 - Áp dụng thay đổi của một commit cụ thể
* [tm vcs clean](tm_vcs_clean.md)	 - Xoá tệp chưa được theo dõi
* [tm vcs commit](tm_vcs_commit.md)	 - Ghi lại các thay đổi đã stage thành một commit
* [tm vcs diff](tm_vcs_diff.md)	 - Hiển thị khác biệt giữa các phiên bản
* [tm vcs fsck](tm_vcs_fsck.md)	 - Kiểm tra tính toàn vẹn của kho và các tham chiếu
* [tm vcs hash-object](tm_vcs_hash-object.md)	 - Tính và in mã băm của nội dung tệp
* [tm vcs ignore](tm_vcs_ignore.md)	 - In các quy tắc bỏ qua tệp đang có trong kho
* [tm vcs init](tm_vcs_init.md)	 - Khởi tạo kho mã nguồn tm trong thư mục cho trước
* [tm vcs log](tm_vcs_log.md)	 - Xem lịch sử commit
* [tm vcs merge](tm_vcs_merge.md)	 - Hợp nhất một nhánh vào nhánh hiện tại
* [tm vcs mv](tm_vcs_mv.md)	 - Đổi tên hoặc di chuyển một tệp đã được theo dõi
* [tm vcs rebase](tm_vcs_rebase.md)	 - Di chuyển các commit hiện tại lên trên một điểm khác
* [tm vcs reflog](tm_vcs_reflog.md)	 - Xem nhật ký di chuyển của HEAD hoặc một tham chiếu
* [tm vcs reset](tm_vcs_reset.md)	 - Di chuyển con trỏ HEAD về commit khác
* [tm vcs restore](tm_vcs_restore.md)	 - Khôi phục lại nội dung tệp
* [tm vcs revert](tm_vcs_revert.md)	 - Hoàn tác thay đổi của một commit
* [tm vcs rm](tm_vcs_rm.md)	 - Gỡ tệp khỏi theo dõi và khỏi cây làm việc
* [tm vcs show](tm_vcs_show.md)	 - Hiển thị chi tiết của một commit
* [tm vcs show-file](tm_vcs_show-file.md)	 - In nội dung một tệp ở một điểm trong lịch sử
* [tm vcs stash](tm_vcs_stash.md)	 - Lưu tạm và khôi phục các thay đổi chưa commit
* [tm vcs status](tm_vcs_status.md)	 - Hiển thị trạng thái thay đổi của cây làm việc
* [tm vcs switch](tm_vcs_switch.md)	 - Chuyển sang nhánh khác
* [tm vcs tag](tm_vcs_tag.md)	 - Liệt kê, tạo hoặc xóa tag

