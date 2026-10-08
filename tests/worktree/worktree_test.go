package worktree_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tonguyenducmanh/devcli/internal/vcs/object"
	"github.com/tonguyenducmanh/devcli/internal/vcs/worktree"
)

func TestFileModeFromStat(t *testing.T) {
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
	if worktree.FileModeFromInfo(fi) != object.ModeBlob {
		t.Fatalf("file 644 phải là ModeBlob: %s", worktree.FileModeFromInfo(fi))
	}

	// File thực thi.
	exec := filepath.Join(dir, "chuong-trinh.sh")
	if err := os.WriteFile(exec, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	fi, _ = os.Lstat(exec)
	if worktree.FileModeFromInfo(fi) != object.ModeExec {
		t.Fatalf("file 755 phải là ModeExec: %s", worktree.FileModeFromInfo(fi))
	}

	// Liên kết tượng trưng.
	link := filepath.Join(dir, "lien-ket")
	if err := os.Symlink(plain, link); err != nil {
		t.Fatal(err)
	}
	fi, _ = os.Lstat(link)
	if worktree.FileModeFromInfo(fi) != object.ModeSymlink {
		t.Fatalf("symlink phải là ModeSymlink: %s", worktree.FileModeFromInfo(fi))
	}
}

func TestWriteAndReadFileOnDisk(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "thu-muc", "tệp.txt")

	content := []byte("nội dung tiếng Việt\n")
	if err := worktree.WriteFileSymlinkAware(path, object.ModeBlob, content); err != nil {
		t.Fatal(err)
	}
	got, err := worktree.ReadFileSymlinkAware(path, object.ModeBlob)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(content) {
		t.Fatalf("nội dung không khớp: %q", got)
	}

	// Ghi đè lại phải cho đúng nội dung mới.
	updated := []byte("nội dung mới\n")
	if err := worktree.WriteFileSymlinkAware(path, object.ModeBlob, updated); err != nil {
		t.Fatal(err)
	}
	got, _ = worktree.ReadFileSymlinkAware(path, object.ModeBlob)
	if string(got) != string(updated) {
		t.Fatalf("ghi đè không thành công: %q", got)
	}
}

func TestWriteSymlink(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "dich")
	if err := os.WriteFile(target, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "lien-ket")
	if err := worktree.WriteFileSymlinkAware(link, object.ModeSymlink, []byte(target)); err != nil {
		t.Fatal(err)
	}
	fi, err := os.Lstat(link)
	if err != nil {
		t.Fatal(err)
	}
	if fi.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("phải tạo ra liên kết tượng trưng")
	}
	got, err := worktree.ReadFileSymlinkAware(link, object.ModeSymlink)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != target {
		t.Fatalf("nội dung symlink phải là đường dẫn đích, nhận %q", got)
	}
}

func TestRemoveFileAndEmptyDirs(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a", "b", "tệp.txt")
	if err := worktree.WriteFileSymlinkAware(path, object.ModeBlob, []byte("x")); err != nil {
		t.Fatal(err)
	}
	if err := worktree.RemoveFile(path, dir); err != nil {
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
	if !worktree.SameContents(path, object.ModeBlob, []byte("nội dung")) {
		t.Fatalf("nội dung giống phải trả về true")
	}
	if worktree.SameContents(path, object.ModeBlob, []byte("khác")) {
		t.Fatalf("nội dung khác phải trả về false")
	}
}

// viếtIgnoreFile ghi nội dung vào một tệp ignore rồi trả về đường dẫn.
func viếtIgnoreFile(t *testing.T, dir, content string) string {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, worktree.IgnoreFileName)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// TestIgnoreNestedFileOnlyAppliesInside bảo đảm .tmxignore trong thư mục con chỉ
// có tác dụng bên trong thư mục đó.
//
// Quy tắc đặt sai phạm vi thì rất khó phát hiện: tệp bị bỏ qua ngoài ý muốn ở
// một nơi, và tệp lẽ ra phải bị bỏ qua lại không bị.
func TestIgnoreNestedFileOnlyAppliesInside(t *testing.T) {
	root := t.TempDir()
	ig := worktree.NewIgnore(root)

	if err := ig.AddFile(viếtIgnoreFile(t, root, "*.log\n")); err != nil {
		t.Fatal(err)
	}
	sub := filepath.Join(root, "sub")
	if err := ig.AddFile(viếtIgnoreFile(t, sub, "only-here.txt\n")); err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		path string
		want bool
	}{
		{"only-here.txt", false},         // ngoài sub, quy tắc không tới
		{"sub/only-here.txt", true},      // trong sub
		{"sub/deep/only-here.txt", true}, // vẫn trong sub
		{"other.txt", false},
		{"sub/other.txt", false},
		{"sub/a.log", true}, // quy tắc ở gốc vẫn áp dụng mọi cấp
	}
	for _, tc := range cases {
		if got := ig.Matches(root, tc.path); got != tc.want {
			t.Errorf("%s: mong đợi bỏ qua=%v, nhận %v", tc.path, tc.want, got)
		}
	}
}

// TestIgnoreNestedFileCanOverrideParent bảo đảm quy tắc sâu hơn được nạp sau
// nên thắng quy tắc ở cấp trên.
func TestIgnoreNestedFileCanOverrideParent(t *testing.T) {
	root := t.TempDir()
	ig := worktree.NewIgnore(root)

	if err := ig.AddFile(viếtIgnoreFile(t, root, "*.log\n")); err != nil {
		t.Fatal(err)
	}
	if err := ig.AddFile(viếtIgnoreFile(t, filepath.Join(root, "sub"), "!keep.log\n")); err != nil {
		t.Fatal(err)
	}

	if !ig.Matches(root, "keep.log") {
		t.Error("keep.log ở gốc phải bị bỏ qua theo quy tắc ở gốc")
	}
	if ig.Matches(root, "sub/keep.log") {
		t.Error("sub/keep.log phải được giữ lại nhờ dấu ! trong sub/.tmxignore")
	}
}

// TestIgnoreReloadSameFileDoesNotDuplicate bảo đảm nạp lại cùng một tệp không
// sinh quy tắc trùng.
//
// Cây làm việc được quét lại nhiều lần, mỗi lần lại nạp các tệp ignore bên
// trong. Nếu không thay quy tắc cũ, số quy tắc sẽ tăng vô hạn theo số lần quét.
func TestIgnoreReloadSameFileDoesNotDuplicate(t *testing.T) {
	root := t.TempDir()
	ig := worktree.NewIgnore(root)
	path := viếtIgnoreFile(t, root, "*.log\nbuild/\n")

	for i := 0; i < 5; i++ {
		if err := ig.AddFile(path); err != nil {
			t.Fatal(err)
		}
	}
	if got := ig.Len(); got != 2 {
		t.Errorf("nạp 5 lần cùng một tệp phải giữ 2 quy tắc, nhận %d", got)
	}
}

func TestIgnorePatternsDiffer(t *testing.T) {
	root := t.TempDir()
	ig := worktree.NewIgnore(root)

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

// TestPathPatternMatch kiểm tra việc khớp mẫu thông qua giao diện công khai
// của bộ quy tắc bỏ qua, vì đó mới là hành vi người dùng thấy.
func TestPathPatternMatch(t *testing.T) {
	root := t.TempDir()
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
		ig := worktree.NewIgnore(root)
		if err := ig.AddReader(strings.NewReader(tc.pattern+"\n"), root); err != nil {
			t.Fatal(err)
		}
		if got := ig.Matches(root, tc.path); got != tc.want {
			t.Fatalf("mẫu %q với đường dẫn %q: mong đợi %v, nhận %v",
				tc.pattern, tc.path, tc.want, got)
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
	if err := worktree.CopyFile(src, dst); err != nil {
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
	err := worktree.ErrOutsideRepo{Path: "/etc/passwd"}
	if err.Error() == "" {
		t.Fatalf("thông báo lỗi không được rỗng")
	}
}
