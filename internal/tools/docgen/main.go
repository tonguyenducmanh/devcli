// Command docgen sinh tài liệu tham chiếu cho toàn bộ cây lệnh của td.
//
// Tài liệu sinh ra có cấu trúc ổn định: mỗi lệnh một tệp Markdown, gồm phần
// mô tả, cú pháp, các cờ và các ví dụ. Đây cũng là dạng đầu vào thuận tiện
// để trợ lý lập trình đọc và hiểu CLI mà không cần chạy thử từng lệnh.
//
// Chỉ sinh một định dạng là Markdown, vì đó là thứ dự án thật sự dùng.
// Thêm định dạng khác khi thật sự cần, kèm kiểm thử và cập nhật check.sh.
//
// Cách dùng:
//
//	go run ./internal/tools/docgen -out agents/cli
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra/doc"

	"github.com/tonguyenducmanh/devcli/cmd"
)

func main() {
	out := flag.String("out", "./agents/cli", "thư mục đích để ghi tài liệu")
	flag.Parse()

	if err := os.MkdirAll(*out, 0o755); err != nil {
		log.Fatalf("không tạo được thư mục đích: %v", err)
	}

	root := cmd.Root()
	// Tắt dòng thời gian sinh tự động để tài liệu sinh ra giống nhau
	// mỗi lần chạy, thuận tiện cho việc theo dõi thay đổi trong kho mã.
	root.DisableAutoGenTag = true

	if err := doc.GenMarkdownTree(root, *out); err != nil {
		log.Fatalf("sinh tài liệu thất bại: %v", err)
	}

	fmt.Printf("Đã sinh tài liệu Markdown vào %s\n", *out)
}
