## tm use

Đặt nhóm công cụ mặc định

### Synopsis

Đặt một nhóm công cụ làm mặc định để tiết kiệm thời gian gõ lệnh.

Ví dụ, thay vì lúc nào cũng phải gõ 'tm vcs status' hay 'tm vcs commit', bạn
chỉ cần đặt nhóm mặc định là 'vcs' bằng lệnh 'tm use vcs'. Từ đó, mọi lệnh gọi
không thuộc nhóm nào khác (như 'tm status', 'tm commit') sẽ tự động được
chuyển hướng sang nhóm 'vcs'.

Để tắt tính năng này và quay về trạng thái bình thường, hãy chạy 'tm use'
không kèm theo đối số nào.

```
tm use [tên nhóm] [flags]
```

### Examples

```
  tm use vcs       đặt nhóm 'vcs' làm mặc định
  tm status        lệnh này giờ sẽ tương đương với 'tm vcs status'
  tm use           gỡ bỏ nhóm mặc định, mọi lệnh phải ghi rõ nhóm
```

### Options

```
  -h, --help   hiển thị phần trợ giúp của lệnh này
```

### SEE ALSO

* [tm](../tm.md)	 - Bộ công cụ dòng lệnh cá nhân, tất cả gói dưới một lệnh duy nhất

