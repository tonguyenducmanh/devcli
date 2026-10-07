# AGENTS.md

Tài liệu cho trợ lý lập trình nằm trong thư mục [`docs/agents/`](docs/agents/).

**Đọc [`docs/agents/AGENTS.md`](docs/agents/AGENTS.md) trước khi sửa bất cứ điều gì.**
Tệp đó chứa kiến trúc, bất biến kiến trúc, quy trình và quy ước viết mã.

Tóm tắt nhanh:

```bash
./build_all.sh                # build Mac, Linux, Windows vào out/
go test ./...                 # chạy kiểm thử trong tests/
./scripts/check.sh            # định dạng + phân tích + kiểm thử + đối chiếu tài liệu
./scripts/build_agent_docs.sh # sinh lại docs/agents/cli khi cây lệnh thay đổi
```

Bản đồ mã nguồn chi tiết: [`docs/agents/README.md`](docs/agents/README.md).
Tham chiếu từng lệnh: [`docs/agents/cli/`](docs/agents/cli/).