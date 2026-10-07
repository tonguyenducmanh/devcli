package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// moduleRoot trả về thư mục gốc của module.
func moduleRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("không tìm thấy go.mod")
		}
		dir = parent
	}
}

// sourceFiles liệt kê mọi tệp Go trong module.
func sourceFiles(t *testing.T) []string {
	t.Helper()
	root := moduleRoot(t)
	var out []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// Bỏ qua thư mục cache và thư mục version control.
		if d.IsDir() {
			switch d.Name() {
			case ".git", ".tdx", "vendor":
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(path, ".go") {
			out = append(out, path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// TestKhongGoiLenhHeThong bao đảm mã nguồn không chạy bất kỳ chương trình
// ngoài nào. Mọi thao tác phải nằm gọn trong quá trình hiện tại để dữ liệu
// của td không bị công cụ khác can thiệp.
func TestKhongGoiLenhHeThong(t *testing.T) {
	for _, file := range sourceFiles(t) {
		if strings.HasSuffix(file, "_test.go") {
			continue // tệp kiểm thử được phép chạy chương trình khác
		}
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		for i, line := range strings.Split(string(data), "\n") {
			if strings.Contains(line, "os/exec") || strings.Contains(line, "exec.Command") {
				t.Errorf("%s:%d phát hiện gọi chương trình ngoài: %s",
					file, i+1, strings.TrimSpace(line))
			}
		}
	}
}

// TestBinaryKhongPhuThuocLenhNgoai chạy bản nhị phân với PATH rỗng để chắc
// chắn nó vẫn hoạt động, tức là không cần bất kỳ chương trình nào khác.
func TestBinaryKhongPhuThuocLenhNgoai(t *testing.T) {
	root := moduleRoot(t)
	bin := filepath.Join(t.TempDir(), "td")

	// Dùng chính `go` của phiên chạy kiểm thử để biên dịch rồi xoá khỏi PATH.
	goBin, err := exec.LookPath("go")
	if err != nil {
		t.Skipf("không tìm thấy go: %v", err)
	}
	build := exec.Command(goBin, "build", "-o", bin, ".")
	build.Dir = root
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("biên dịch thất bại: %v\n%s", err, out)
	}

	// Chạy lệnh version với PATH trỏ tới một thư mục rỗng.
	emptyDir := t.TempDir()
	cmd := exec.Command(bin, "version")
	cmd.Dir = emptyDir
	cmd.Env = append(os.Environ(), "PATH="+emptyDir)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("chạy với PATH rỗng thất bại: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "td phiên bản") {
		t.Fatalf("kết quả lạ: %s", out)
	}
}
