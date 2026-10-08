# tests

Toàn bộ mã kiểm thử nằm trong thư mục này, tách khỏi mã sản phẩm.

## Vì sao tách ra

Mã sản phẩm trong `cmd/` và `internal/` giữ đúng một việc: xây dựng và chạy
ứng dụng. Nếu để tệp `_test.go` nằm xen kẽ, người đọc phải lọc mới thấy được
những gì thực sự chạy khi biên dịch, và cấu trúc thư mục lẫn lộn giữa mã
chạy thật với mã kiểm chứng.

Tách ra giúp cây thư mục phản ánh đúng bản chất từng phần.

## Cách tổ chức

Mỗi thư mục con kiểm thử đúng một vùng nghiệp vụ, tên trùng với gói tương ứng:

```
tests/
  object/        kiểm thử kiểu dữ liệu và cách tính mã băm
  diff/          kiểm thử thuật toán so sánh
  merge/         kiểm thử hợp nhất ba phía
  index/         kiểm thử vùng chuẩn bị
  storage/       kiểm thử lưu trữ object và tham chiếu
  worktree/      kiểm thử đọc ghi tệp và quy tắc bỏ qua
  config/        kiểm thử cấu hình
  repo/          kiểm thử lớp trừu tượng kho mã nguồn
  ops/           kiểm thử nghiệp vụ cấp cao theo dòng lệnh
  cli/           kiểm thử cách dòng lệnh đọc đối số của nó
  architecture/  kiểm thử các bất biến kiến trúc
```

`tests/cli/` là nơi duy nhất chạy cây lệnh trong bộ nhớ, nên nó dành cho những
điều chỉ thấy được ở tầng dòng lệnh: đọc đối số, tên nhóm trong output, cách lệnh
tự báo lỗi. Phần nghiệp vụ thì kiểm thử ở `tests/ops/`.

Tệp kiểm thử dùng *gói kiểm thử ngoài* (`package object_test`), chỉ được
dùng những thứ mà gói đó xuất ra. Hệ quả là mã kiểm thử chỉ kiểm chứng được
hành vi công khai, và mọi thứ cần tới nội bộ đều phải được suy nghĩ lại trước
khi xuất ra. Đây cũng là cách đóng gói chuẩn của Go.

## Chạy kiểm thử

```bash
go test ./...                    # toàn bộ
go test ./tests/diff/            # một vùng
go test ./tests/... -v           # in chi tiết
go test ./tests/... -cover       # kèm độ phủ
./scripts/check.sh               # chạy kèm kiểm tra định dạng và phân tích tĩnh
```

## Thêm một bộ kiểm thử

1. Tạo thư mục `tests/<vùng>/` đặt cạnh vùng tương ứng.
2. Tệp đặt tên `<vùng>_test.go`, khai báo `package <vùng>_test`.
3. Chỉ dùng những thứ mà gói đó xuất ra. Cần thêm chức năng mới thì thêm ở
   gói sản phẩm, đừng thêm cờ nội bộ cho riêng kiểm thử.
4. Viết tiện ích dùng chung ngay trong tệp của vùng đó, đặt tên đủ hiểu để
   không đụng tên hàm của kiểm thử vùng khác.