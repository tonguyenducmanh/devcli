// Command docgen sinh tài liệu tham chiếu cho toàn bộ cây lệnh của td.
//
// Tài liệu sinh ra có cấu trúc ổn định: mỗi lệnh một tệp Markdown, gồm phần
// mô tả, cú pháp, các cờ và các ví dụ. Đây cũng là dạng đầu vào thuận tiện
// để trợ lý lập trình đọc và hiểu CLI mà không cần chạy thử từng lệnh.
//
// Cách dùng:
//
//	go run ./internal/tools/docgen -out docs/cli
//	go run ./internal/tools/docgen -out docs/cli -format man
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"

	"github.com/tonguyenducmanh/devcli/cmd"
)

func main() {
	out := flag.String("out", "./docs/cli", "thư mục đích để ghi tài liệu")
	format := flag.String("format", "markdown", "định dạng: markdown, man hoặc rest")
	flag.Parse()

	if err := os.MkdirAll(*out, 0o755); err != nil {
		log.Fatalf("không tạo được thư mục đích: %v", err)
	}

	root := cmd.Root()
	// Tắt dòng thời gian sinh tự động để tài liệu sinh ra giống nhau
	// mỗi lần chạy, thuận tiện cho việc theo dõi thay đổi trong kho mã.
	root.DisableAutoGenTag = true

	switch *format {
	case "markdown":
		if err := doc.GenMarkdownTree(root, *out); err != nil {
			log.Fatalf("sinh tài liệu markdown thất bại: %v", err)
		}
	case "man":
		hdr := &doc.GenManHeader{
			Title:   strings.ToUpper(root.Name()),
			Section: "1",
		}
		if err := doc.GenManTree(root, hdr, *out); err != nil {
			log.Fatalf("sinh tài liệu man thất bại: %v", err)
		}
	case "rest":
		if err := doc.GenReSTTree(root, *out); err != nil {
			log.Fatalf("sinh tài liệu rest thất bại: %v", err)
		}
	default:
		log.Fatalf("định dạng không hợp lệ: %s (chọn markdown, man hoặc rest)", *format)
	}

	fmt.Printf("Đã sinh tài liệu định dạng %s vào %s\n", *format, filepath.Clean(*out))
}

// Bảo đảm gói cobra được dùng trong mã, giữ phụ thuộc rõ ràng.
var _ = cobra.Command{}
