package main

// Kiểm thử cho tên tệp mà build sinh ra. Tên tệp xuất hiện ở nhiều nơi: script
// build đặt tên, tài liệu hướng dẫn cài đặt, và bản giả kiểm thử của tiện ích VS
// Code dò tệp đã build. Chỗ nào lệch chỗ nào thì người đọc tài liệu cài được một
// tên còn máy thật lại sinh ra tên khác, và kiểm thử tiện ích thì âm thầm bị bỏ
// qua. Vì vậy ở đây kiểm tra cho cả ba khớp nhau.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// appNameFromScript đọc biến APP_NAME trong scripts/build_binaries.sh.
//
// Đây là nguồn duy nhất của tiền tố tên tệp: scripts/build_extension.sh đọc lại
// từ đây thay vì khai báo lần thứ hai.
func appNameFromScript(t *testing.T) string {
	t.Helper()
	root := moduleRoot(t)
	body, err := os.ReadFile(filepath.Join(root, "scripts", "build_binaries.sh"))
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Split(string(body), "\n") {
		if value, ok := strings.CutPrefix(strings.TrimSpace(line), "APP_NAME="); ok {
			return strings.TrimSpace(value)
		}
	}
	t.Fatal("không tìm thấy biến APP_NAME trong scripts/build_binaries.sh")
	return ""
}

// TestBuildExtensionUsesBinaryAppName chặn việc khai báo lại tiền tố tên tệp ở
// script đóng gói tiện ích.
//
// Hai script nằm cùng nền tảng build mà trước đây cùng ghi APP_NAME, nên đổi
// tên ở một bên quên bên kia thì out/ có hai kiểu tên cùng lúc. Nay script
// đóng gói đọc từ script build nên chỉ còn một chỗ phải sửa, và kiểm thử này giữ cho
// nó không quay lại ghi tay.
func TestBuildExtensionUsesBinaryAppName(t *testing.T) {
	root := moduleRoot(t)
	body, err := os.ReadFile(filepath.Join(root, "scripts", "build_extension.sh"))
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Split(string(body), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			continue
		}
		value, ok := strings.CutPrefix(trimmed, "APP_NAME=")
		if !ok {
			continue
		}
		// Gán bằng kết quả của lệnh khác thì hợp lệ. Gán thẳng một chuỗi tức là
		// khai báo lần thứ hai và sẽ lệch với script build.
		if !strings.HasPrefix(value, "$") {
			t.Errorf("scripts/build_extension.sh gán thẳng APP_NAME=%s, hãy đọc từ scripts/build_binaries.sh", value)
		}
	}
}

// TestDocsUseCurrentArtifactName bảo đảm tài liệu viết đúng tên tệp build sinh ra.
//
// Tài liệu cài đặt chỉ rõ tên tệp phải copy. Tên sai ở đó thì người đọc làm
// theo rồi hỏng, mà không có gì báo trước.
func TestDocsUseCurrentArtifactName(t *testing.T) {
	root := moduleRoot(t)
	appName := appNameFromScript(t)

	docs := []string{
		"README.md",
		"docs/development.md",
		"docs/install.md",
		"scripts/README.md",
		"build_all.sh",
		filepath.Join("editors", "vscode", "README.md"),
	}
	// Dạng tên tệp trong tài liệu: tiền tố rồi tới dấu gạch nối và nền tảng.
	marker := appName + "-"
	for _, rel := range docs {
		body, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil {
			t.Fatalf("đọc %s: %v", rel, err)
		}
		for number, line := range strings.Split(string(body), "\n") {
			at := strings.Index(line, "out/")
			if at < 0 {
				continue
			}
			rest := line[at+len("out/"):]
			end := strings.IndexAny(rest, "`\"' )")
			if end < 0 {
				end = len(rest)
			}
			name := rest[:end]
			// Chỉ quan tâm tới tên tệp build, không phải thư mục trong out/.
			if !strings.Contains(name, "-") || strings.Contains(name, "/") {
				continue
			}
			if !strings.HasPrefix(name, marker) {
				t.Errorf("%s dòng %d viết tên tệp %q, còn build sinh ra tên bắt đầu bằng %q",
					rel, number+1, name, marker)
			}
		}
	}
}

// TestTestHarnessAcceptsCurrentArtifactName bảo đảm bản giả kiểm thử của tiện ích
// VS Code nhận ra tên tệp mới.
//
// Bản giả dò tệp trong out/ bằng một danh sách tiền tố viết tay. Lệch khỏi tên
// thật thì findTm trả về undefined và toàn bộ kiểm thử cần lệnh tm bị bỏ qua,
// báo cáo vẫn là xanh trong khi chẳng kiểm thử nào chạy.
func TestTestHarnessAcceptsCurrentArtifactName(t *testing.T) {
	root := moduleRoot(t)
	appName := appNameFromScript(t)

	body, err := os.ReadFile(filepath.Join(root, "editors", "vscode", "src", "test", "harness.ts"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), appName) {
		t.Errorf("harness.ts không nêu %q trong danh sách tiền tố tên tệp build, kiểm thử sẽ tự bỏ qua", appName)
	}
}
