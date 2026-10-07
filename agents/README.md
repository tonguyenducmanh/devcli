# agents

Thư mục này chứa **toàn bộ tài liệu dành cho trợ lý lập trình**, tách biệt khỏi
tài liệu dành cho người dùng ở `README.md`.

Lý do tách riêng: người dùng cần biết *cài đặt và dùng như thế nào*, còn trợ lý
lập trình cần biết *mã nguồn tổ chức ra sao và quy ước ra sao*. Trộn hai loại
tài liệu khiến cả hai đều kém rõ ràng.

## Có những gì

| Tệp | Dành cho | Nội dung | Có sinh tự động? |
| --- | --- | --- | --- |
| `AGENTS.md` | Trợ lý lập trình | Kiến trúc, bất biến, quy ước viết mã | Không, viết tay |
| `llms.txt` | Trợ lý AI | Tóm tắt dự án trong vài dòng | Không, viết tay |
| `cli/` | Trợ lý AI | Tham chiếu từng lệnh: cú pháp, ví dụ, cờ | **Có** |

### `AGENTS.md`

Tài liệu nền cho trợ lý lập trình làm việc trên kho mã này. Gồm bốn phần:

1. **Bất biến kiến trúc** — những quy tắc mà vi phạm sẽ làm kiểm thử đỏ.
2. **Bản đồ mã nguồn** — mỗi tệp và thư mục làm việc gì.
3. **Mô hình dữ liệu** — các khái niệm cốt lõi của kho mã nguồn.
4. **Quy trình và quy ước** — thêm lệnh mới, thêm nhóm công cụ mới, viết mã.

Sửa `AGENTS.md` mỗi khi kiến trúc thay đổi.

### `llms.txt`

Bản tóm tắt ngắn gọn, theo định dạng thường dùng để trợ lý AI đọc nhanh.
Chỉ mô tả: dự án làm gì, có nhóm lệnh nào, tài liệu ở đâu.

### `cli/`

Tham chiếu đầy đủ từng lệnh, **sinh tự động từ cây lệnh**. Mỗi lệnh một tệp
Markdown với cấu trúc cố định:

```
## td vcs merge          <- tiêu đề
### Synopsis              <- mô tả dài
### Examples              <- ví dụ dùng được
### Options               <- các cờ của lệnh
### Options inherited...  <- các cờ toàn cục
### SEE ALSO              <- liên kết tới lệnh cha
```

Nhờ cấu trúc ổn định như vậy, tài liệu dễ được đọc theo từng đoạn hoặc đưa
vào vector index.

Tệp `cli/td_vcs.md` là mục lục của nhóm `vcs`, bắt đầu từ đó.

## Không sửa tay `cli/`

Thư mục này do chương trình sinh ra, sửa tay sẽ bị mất khi sinh lại.

```bash
./scripts/build_agent_docs.sh
```

Chạy lại lệnh này khi:

- Thêm, bỏ hoặc đổi tên một lệnh.
- Sửa `Short`, `Long` hoặc `Example` của một lệnh.
- Đổi kiến trúc câu lệnh, ví dụ thêm nhóm lệnh hay thêm cờ toàn cục.

Quên chạy sẽ khiến tài liệu lệch với thực tế. `./scripts/check.sh` có bước
đối chiếu để phát hiện ra điều đó trước khi đóng góp.

## Trợ lý nào đọc tệp nào

| Trợ lý | Tệp được đọc tự động |
| --- | --- |
| opencode, cursor, claude code, và nhiều công cụ khác | `AGENTS.md` ở thư mục gốc, đi theo đường dẫn tới tệp thật |
| Trợ lý AI tìm kiếm tài liệu | `llms.txt` |
| Bất kỳ trợ lý nào cần chi tiết về lệnh | `cli/td_vcs_<tên lệnh>.md` |

Vì nhiều công cụ chỉ tự tìm `AGENTS.md` ở thư mục gốc, có một tệp trỏ ngắn
cùng tên ở gốc. Tệp đó chỉ dẫn tới tệp thật trong thư mục này.

## Quy trình làm việc với trợ lý lập trình

1. Cho trợ lý đọc `AGENTS.md` trước khi sửa bất cứ tệp nào.
2. Cần biết chính xác một lệnh nhận gì và làm gì thì đọc
   `cli/td_vcs_<tên lệnh>.md`, đừng đoán từ mã nguồn.
3. Sau khi sửa cây lệnh, chạy `./scripts/build_agent_docs.sh`.
4. Trước khi đóng góp, chạy `./scripts/check.sh`.