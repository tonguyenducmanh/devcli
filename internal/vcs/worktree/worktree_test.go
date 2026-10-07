package worktree

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tonguyenducmanh/devcli/internal/vcs/object"
)

func TestFileModeTuStat(t *testing.T) {
	dir := t.TempDir()

	// File thường.
	plain := filepath.Join(dir, "van-ban.txt")
	if err := os.WriteFile(plain, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	fi, err := os.Lstat(plain)
	if err != nil {
		t.Fatal(err)
	}
	if FileModeFromInfo(fi) != object.ModeBlob {
		t.Fatalf("file 644 phải là ModeBlob: %s", FileModeFromInfo(fi))
	}

	// File thực thi.
	exec := filepath.Join(dir, "chuong-trinh.sh")
	if err := os.WriteFile(exec, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	fi, _ = os.Lstat(exec)
	if FileModeFromInfo(fi) != object.ModeExec {
		t.Fatalf("file 755 phải là ModeExec: %s", FileModeFromInfo(fi))
	}

	// Liên kết tượng trưng.
	link := filepath.Join(dir, "lien-ket")
	if err := os.Symlink(plain, link); err != nil {
		t.Fatal(err)
	}
	fi, _ = os.Lstat(link)
	if FileModeFromInfo(fi) != object.ModeSymlink {
		t.Fatalf("symlink phải là ModeSymlink: %s", FileModeFromInfo(fi))
	}
}

func TestGhiVaDocFileTrenDia(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "thu-muc", "tệp.txt")

	content := []byte("nội dung tiếng Việt\n")
	if err := WriteFileSymlinkAware(path, object.ModeBlob, content); err != nil {
		t.Fatal(err)
	}
	got, err := ReadFileSymlinkAware(path, object.ModeBlob)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(content) {
		t.Fatalf("nội dung không khớp: %q", got)
	}

	// Ghi đè lại phải cho đúng nội dung mới.
	updated := []byte("nội dung mới\n")
	if err := WriteFileSymlinkAware(path, object.ModeBlob, updated); err != nil {
		t.Fatal(err)
	}
	got, _ = ReadFileSymlinkAware(path, object.ModeBlob)
	if string(got) != string(updated) {
		t.Fatalf("ghi đè không thành công: %q", got)
	}
}

func TestGhiSymlink(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "dich")
	if err := os.WriteFile(target, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "lien-ket")
	if err := WriteFileSymlinkAware(link, object.ModeSymlink, []byte(target)); err != nil {
		t.Fatal(err)
	}
	fi, err := os.Lstat(link)
	if err != nil {
		t.Fatal(err)
	}
	if fi.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("phải tạo ra liên kết tượng trưng")
	}
	got, err := ReadFileSymlinkAware(link, object.ModeSymlink)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != target {
		t.Fatalf("nội dung symlink phải là đường dẫn đích, nhận %q", got)
	}
}

func TestXoaFileVaDonThuMuc(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a", "b", "tệp.txt")
	if err := WriteFileSymlinkAware(path, object.ModeBlob, []byte("x")); err != nil {
		t.Fatal(err)
	}
	if err := RemoveFile(path, dir); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("file phải bị xóa")
	}
	// Các thư mục cha rỗng cũng phải được dọn.
	if _, err := os.Stat(filepath.Join(dir, "a")); !os.IsNotExist(err) {
		t.Fatalf("thư mục rỗng phải được dọn")
	}
	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("thư mục gốc không được xóa: %v", err)
	}
}

func TestSameContents(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "f.txt")
	if err := os.WriteFile(path, []byte("nội dung"), 0o644); err != nil {
		t.Fatal(err)
	}
	if !SameContents(path, object.ModeBlob, []byte("nội dung")) {
		t.Fatalf("nội dung giống phải trả về true")
	}
	if SameContents(path, object.ModeBlob, []byte("khác")) {
		t.Fatalf("nội dung khác phải trả về false")
	}
}

func TestIgnoreKhacNhau(t *testing.T) {
	root := t.TempDir()
	ig := NewIgnore(root)

	// Nạp quy tắc từ chuỗi.
	rules := `
# đây là chú thích
build/
*.log
!quan-trong.log
docs/*.tmp
**/cache/
`
	f, err := os.CreateTemp(t.TempDir(), "ignore")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(rules); err != nil {
		t.Fatal(err)
	}
	if _, err := f.Seek(0, 0); err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := ig.AddReader(f, root); err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		path string
		want bool
	}{
		{"build/app.bin", true},
		{"build", true},
		{"src/build/x.o", true},
		{"debug.log", true},
		{"src/debug.log", true},
		{"quan-trong.log", false}, // dấu ! phủ định lại quy tắc *.log
		{"docs/notes.tmp", true},
		{"docs/a/b.tmp", false},
		{"src/cache/temp.dat", true},
		{"src/main.go", false},
	}
	for _, tc := range cases {
		if got := ig.Matches(root, tc.path); got != tc.want {
			t.Fatalf("%s: mong đợi bỏ qua=%v, nhận %v", tc.path, tc.want, got)
		}
	}
	if ig.Len() == 0 {
		t.Fatalf("phải nạp được quy tắc")
	}
}

func TestGlobMatch(t *testing.T) {
	cases := []struct {
		pattern, path string
		want          bool
	}{
		{"*.go", "main.go", true},
		{"*.go", "main.js", false},
		{"src/*.go", "src/main.go", true},
		{"src/*.go", "src/deep/main.go", false},
		{"a?c.txt", "abc.txt", true},
		{"a?c.txt", "ac.txt", false},
	}
	for _, tc := range cases {
		if got := globMatch(tc.pattern, tc.path); got != tc.want {
			t.Fatalf("mẫu %q với đường dẫn %q: mong đợi %v, nhận %v", tc.pattern, tc.path, tc.want, got)
		}
	}
}

func TestCopyFile(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "nguon.txt")
	dst := filepath.Join(dir, "thu-muc", "dich.txt")
	if err := os.WriteFile(src, []byte("dữ liệu"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := CopyFile(src, dst); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "dữ liệu" {
		t.Fatalf("sai nội dung sau khi sao chép: %q", data)
	}
}

func TestErrOutsideRepo(t *testing.T) {
	err := ErrOutsideRepo{Path: "/etc/passwd"}
	if err.Error() == "" {
		t.Fatalf("thông báo lỗi không được rỗng")
	}
}
