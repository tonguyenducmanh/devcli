// Command docgen sinh tài liệu tham chiếu cho toàn bộ cây lệnh của td.
//
// Mặc định sinh Markdown: mỗi lệnh một tệp gồm phần mô tả, cú pháp, các cờ
// và các ví dụ. Đây là dạng đầu vào thuận tiện để trợ lý lập trình đọc và
// hiểu CLI mà không cần chạy thử từng lệnh.
//
// Trang man sinh theo yêu cầu khi phát hành, không commit vào kho mã.
//
// Cách dùng:
//
//	go run ./internal/tools/docgen -out agents/cli
//	go run ./internal/tools/docgen -out out/man -format man
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra/doc"

	"github.com/tonguyenducmanh/devcli/cmd"
)

func main() {
	out := flag.String("out", "./agents/cli", "thư mục đích để ghi tài liệu")
	format := flag.String("format", "markdown", "định dạng: markdown hoặc man")
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
			// Nơi phát hành đi vào trường Source của dòng .TH. Khuôn trang man
			// của cobra không có chỗ cho tên tác giả nên phần AUTHORS được
			// thêm vào sau, xem hàm themPhanTacGia.
			Source: cmd.RepoURL,
			Manual: cmd.AppName,
		}
		if err := doc.GenManTree(root, hdr, *out); err != nil {
			log.Fatalf("sinh tài liệu man thất bại: %v", err)
		}
		themPhanTacGia(*out, root.Name())

	default:
		log.Fatalf("định dạng không hợp lệ: %s (chọn markdown hoặc man)", *format)
	}

	fmt.Printf("Đã sinh tài liệu định dạng %s vào %s\n", *format, *out)
}

// themPhanTacGia chèn mục AUTHORS vào trang man của lệnh gốc.
//
// Khuôn trang man của cobra không có chỗ cho tên tác giả, nên phải tự thêm
// vào. Chỉ trang của lệnh gốc được thêm, vì đó là trang người đọc tìm về tác
// giả trước tiên.
func themPhanTacGia(thuMuc, tenLenh string) {
	if cmd.Author == "" {
		return
	}

	tenTep := filepath.Join(thuMuc, tenLenh+".1")
	noiDung, err := os.ReadFile(tenTep)
	if err != nil {
		log.Fatalf("không đọc được trang man %s: %v", tenTep, err)
	}

	phan := "# AUTHORS\n" + cmd.Author + "\n"
	if cmd.RepoURL != "" {
		phan += "\n# NƠI PHÁT HÀNH\n" + cmd.RepoURL + "\n"
	}

	noiDungMoi := strings.Replace(string(noiDung), "# SEE ALSO", phan+"\n# SEE ALSO", 1)
	if noiDungMoi == string(noiDung) {
		// Trang không có mục SEE ALSO thì ghi nối vào cuối.
		noiDungMoi = string(noiDung) + "\n" + phan
	}
	if err := os.WriteFile(tenTep, []byte(noiDungMoi), 0o644); err != nil {
		log.Fatalf("không ghi được trang man %s: %v", tenTep, err)
	}
}
