# Đóng góp cho tm

Cảm ơn bạn quan tâm tới dự án. Tài liệu này nói về **quy trình đóng góp**.

Quy ước viết mã, kiến trúc và bản đồ mã nguồn nằm ở
[agents/AGENTS.md](agents/AGENTS.md) — đọc tệp đó trước khi viết dòng mã đầu
tiên.

## Yêu cầu

- Go 1.23 trở lên, kiểm tra bằng `go version`.

## Thiết lập

```bash
git clone <url>
cd devcli
./build_all.sh
```

## Trước khi mở pull request

Chạy một lệnh, script sẽ báo cáo từng bước:

```bash
./scripts/check.sh
```

Bước cuối sinh lại tài liệu vào thư mục tạm rồi so với bản đã commit, nên
tài liệu lệch với cây lệnh sẽ bị chặn trước khi gửi.

## Quy trình

1. Mở một issue hoặc mô tả rõ trong pull request điều bạn định sửa.
2. Tạo nhánh theo dạng `them-<nội dung>` hoặc `sua-<nội dung>`.
3. Viết mã kèm kiểm thử đặt trong `tests/`.
4. Chạy `./scripts/build_agent_docs.sh` nếu có thay đổi cây lệnh.
5. Chạy `./scripts/check.sh` để chắc chắn mọi thứ đều qua.
6. Mở pull request mô tả *vì sao* thay đổi, không chỉ *thay đổi gì*.

## Trước khi gửi

Ba điều dễ quên, đều đã có trong [`agents/AGENTS.md`](agents/AGENTS.md):

- Chú thích và thông điệp người dùng bằng **tiếng Việt**.
- Không gọi chương trình ngoài, dữ liệu chỉ nằm trong `.tmx`.
- `agents/cli/` do công cụ sinh ra, **không sửa tay**.
