## td use

Đặt nhóm công cụ mặc định

### Synopsis

Đặt một nhóm công cụ làm mặc định để tiết kiệm thời gian gõ lệnh.

Ví dụ, thay vì lúc nào cũng phải gõ 'td vcs status' hay 'td vcs commit', bạn
chỉ cần đặt nhóm mặc định là 'vcs' bằng lệnh 'td use vcs'. Từ đó, mọi lệnh gọi
không thuộc nhóm nào khác (như 'td status', 'td commit') sẽ tự động được
chuyển hướng sang nhóm 'vcs'.

Để tắt tính năng này và quay về trạng thái bình thường, hãy chạy 'td use'
không kèm theo đối số nào.

```
td use [tên nhóm] [flags]
```

### Examples

```
  td use vcs       đặt nhóm 'vcs' làm mặc định
  td status        lệnh này giờ sẽ tương đương với 'td vcs status'
  td use           gỡ bỏ nhóm mặc định, mọi lệnh phải ghi rõ nhóm
```

### Options

```
  -h, --help   hiển thị phần trợ giúp của lệnh này
```

### SEE ALSO

* [td](td.md)	 - Bộ công cụ dòng lệnh cá nhân, tất cả gói dưới một lệnh duy nhất

