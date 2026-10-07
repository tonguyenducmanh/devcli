## td vcs

Quản lý phiên bản mã nguồn cục bộ

### Synopsis

Nhóm lệnh quản lý phiên bản mã nguồn cục bộ.

Toàn bộ dữ liệu của nhóm này nằm trong thư mục .tdx cạnh dự án, gồm lịch
sử commit, các nhánh, các tag và vùng chuẩn bị.

Mỗi lệnh dưới đây chạy trên kho tìm thấy bằng cách đi lên từ thư mục hiện
tại. Dùng -C để chỉ định thư mục khác.

```
td vcs [flags]
```

### Examples

```
  # Khởi tạo kho rồi ghi lại thay đổi đầu tiên
  td vcs init
  td vcs add .
  td vcs commit -m "tin nhắn đầu tiên"

  # Xem nhánh hiện tại và lịch sử gọn
  td vcs status
  td vcs log --oneline -n 10

  # Tạo nhánh, làm việc rồi hợp nhất về nhánh chính
  td vcs switch -c tinh-nang
  td vcs commit -am "bổ sung tính năng"
  td vcs switch main
  td vcs merge tinh-nang
```

### Options

```
  -C, --dir string   chạy lệnh tại thư mục khác
  -h, --help         hiển thị phần trợ giúp của lệnh này
```

### Options inherited from parent commands

```
      --no-color   tắt màu trong output
  -v, --verbose    in thêm thông tin chi tiết
```

### SEE ALSO

* [td](td.md)	 - Bộ công cụ dòng lệnh cá nhân, tất cả gói dưới một lệnh duy nhất
* [td vcs add](td_vcs_add.md)	 - Đưa thay đổi vào vùng chuẩn bị commit
* [td vcs branch](td_vcs_branch.md)	 - Liệt kê, tạo hoặc xóa nhánh
* [td vcs cat-file](td_vcs_cat-file.md)	 - In nội dung của một object (blob, tree, commit, tag)
* [td vcs checkout](td_vcs_checkout.md)	 - Chuyển sang nhánh hoặc commit khác
* [td vcs cherry-pick](td_vcs_cherry-pick.md)	 - Áp dụng thay đổi của một commit cụ thể
* [td vcs commit](td_vcs_commit.md)	 - Ghi lại các thay đổi đã stage thành một commit
* [td vcs diff](td_vcs_diff.md)	 - Hiển thị khác biệt giữa các phiên bản
* [td vcs fsck](td_vcs_fsck.md)	 - Kiểm tra tính toàn vẹn của kho và các tham chiếu
* [td vcs hash-object](td_vcs_hash-object.md)	 - Tính và in mã băm của nội dung tệp
* [td vcs init](td_vcs_init.md)	 - Khởi tạo kho mã nguồn td trong thư mục cho trước
* [td vcs log](td_vcs_log.md)	 - Xem lịch sử commit
* [td vcs merge](td_vcs_merge.md)	 - Hợp nhất một nhánh vào nhánh hiện tại
* [td vcs mv](td_vcs_mv.md)	 - Đổi tên hoặc di chuyển một tệp đã được theo dõi
* [td vcs rebase](td_vcs_rebase.md)	 - Di chuyển các commit hiện tại lên trên một điểm khác
* [td vcs reflog](td_vcs_reflog.md)	 - Xem nhật ký di chuyển của HEAD hoặc một tham chiếu
* [td vcs reset](td_vcs_reset.md)	 - Di chuyển con trỏ HEAD về commit khác
* [td vcs restore](td_vcs_restore.md)	 - Khôi phục lại nội dung tệp
* [td vcs revert](td_vcs_revert.md)	 - Hoàn tác thay đổi của một commit
* [td vcs rm](td_vcs_rm.md)	 - Gỡ tệp khỏi theo dõi và khỏi cây làm việc
* [td vcs show](td_vcs_show.md)	 - Hiển thị chi tiết của một commit
* [td vcs stash](td_vcs_stash.md)	 - Lưu tạm và khôi phục các thay đổi chưa commit
* [td vcs status](td_vcs_status.md)	 - Hiển thị trạng thái thay đổi của cây làm việc
* [td vcs switch](td_vcs_switch.md)	 - Chuyển sang nhánh khác
* [td vcs tag](td_vcs_tag.md)	 - Liệt kê, tạo hoặc xóa tag

