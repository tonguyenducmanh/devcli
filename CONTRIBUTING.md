# Đóng góp cho td

Cảm ơn bạn quan tâm tới dự án. Tài liệu này mô tả cách thiết lập môi trường
và những quy ước cần giữ khi gửi thay đổi.

## Yêu cầu

- Go 1.23 trở lên.

## Thiết lập

```bash
git clone <url>
cd td
./scripts/build_all.sh
```

## Trước khi mở pull request

Chạy một lệnh, script sẽ báo cáo từng bước:

```bash
./scripts/check.sh
```

Bước cuối sinh lại tài liệu vào thư mục tạm rồi so với bản đã commit, nên tài
liệu lệch với cây lệnh sẽ bị chặn trước khi gửi.

## Quy trình

1. Mở một issue hoặc mô tả rõ trong pull request điều bạn định sửa.
2. Tạo nhánh theo dạng `them-<nội dung>` hoặc `sua-<nội dung>`.
3. Viết mã kèm kiểm thử đặt trong `tests/`.
4. Chạy `./scripts/build_agent_docs.sh` nếu có thay đổi cây lệnh.
5. Chạy `./scripts/check.sh` để chắc chắn mọi thứ đều qua.
5. Mở pull request mô tả *vì sao* thay đổi, không chỉ *thay đổi gì*.

## Quy ước mã nguồn

- Chú thích và mọi thông điệp người dùng bằng **tiếng Việt**.
- Tên hàm, biến, kiểu dữ liệu bằng tiếng Anh theo chuẩn Go.
- Chú thích giải thích *tại sao*, không lặp lại cái mà tên hàm đã nói.
- Mỗi lệnh mới cần đủ `Short`, `Long` và `Example`. Có kiểm thử bắt buộc.
- Giữ nguyên phụ thuộc một chiều giữa các tầng. Có kiểm thử bắt buộc.
- Không thêm phụ thuộc mới nếu chưa cần thiết; chuẩn bẩy gói đã đủ cho các
  tính năng hiện có.

Chi tiết về kiến trúc và bản đồ mã nguồn nằm trong
[agents/AGENTS.md](agents/AGENTS.md).

## Kiểm thử

- Kiểm thử nghiệp vụ đặt cạnh mã, đặt tên theo hành vi cần kiểm chứng.
- Một kiểm thử nên kiểm tra *điều gì đúng* chứ không chỉ *điều gì sai*.
- Kiểm thử phải chạy được mà không cần mạng và không phụ thuộc thứ tự chạy.
- Khi sửa một lỗi, thêm kiểm thử tái hiện lỗi đó trước khi sửa mã.

## Tài liệu

- `README.md` mô tả tổng quan cho người đọc.
- `AGENTS.md` mô tả kiến trúc cho người sửa mã.
- `agents/cli/` do công cụ sinh ra, **không sửa tay**. Chạy
  `./scripts/build_agent_docs.sh`.