# AGENTS.md

Tài liệu cho trợ lý lập trình nằm trong thư mục [`agents/`](agents/).

**Đọc [`agents/AGENTS.md`](agents/AGENTS.md) trước khi sửa bất cứ điều gì.**
Tệp đó chứa kiến trúc, bất biến kiến trúc, quy trình và quy ước viết mã.

Tóm tắt nhanh:

```bash
./scripts/build_all.sh          # build Mac, Linux, Windows vào out/
go test ./...                   # chạy kiểm thử trong tests/
./scripts/check.sh              # định dạng + phân tích + kiểm thử + đối chiếu tài liệu
./scripts/build_agent_docs.sh   # sinh lại agents/cli khi cây lệnh thay đổi
```

Bản đồ mã nguồn chi tiết: [`agents/README.md`](agents/README.md).
Tham chiếu từng lệnh: [`agents/cli/`](agents/cli/).