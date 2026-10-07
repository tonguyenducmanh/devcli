// Package worktree quản lý cây thư mục làm việc: đọc file, quét thay đổi,
// áp dụng cây object xuống đĩa và bỏ qua file theo quy tắc ignore.
package worktree

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/tonguyenducmanh/devcli/internal/vcs/object"
)

// FileModeFromInfo suy ra chế độ lưu của một tệp từ thông tin stat
// do hệ điều hành cung cấp.
func FileModeFromInfo(fi os.FileInfo) object.FileMode {
	switch {
	case fi.Mode()&os.ModeSymlink != 0:
		return object.ModeSymlink
	case fi.Mode().Perm()&0o111 != 0:
		return object.ModeExec
	default:
		return object.ModeBlob
	}
}

// StatTimes trả về thời điểm sửa đổi dạng giây.nanô giây để lưu vào index.
func StatTimes(fi os.FileInfo) uint64 {
	mt := fi.ModTime().UnixNano()
	if mt < 0 {
		mt = 0
	}
	return uint64(mt)
}

// ReadFileSymlinkAware đọc nội dung một entry trong worktree.
// Với liên kết tượng trưng, nội dung là đường dẫn đích.
func ReadFileSymlinkAware(abs string, mode object.FileMode) ([]byte, error) {
	if mode == object.ModeSymlink {
		// Nội dung của một symlink chính là đường dẫn đích, lưu dạng byte.
		target, err := os.Readlink(abs)
		if err != nil {
			return nil, err
		}
		return []byte(target), nil
	}
	return os.ReadFile(abs)
}

// WriteFileSymlinkAware ghi nội dung lên đĩa theo đúng chế độ.
func WriteFileSymlinkAware(abs string, mode object.FileMode, content []byte) error {
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return err
	}
	// Xóa trước để tránh symlink cũ gây ghi đè nhầm sang nơi khác.
	if mode == object.ModeSymlink {
		_ = os.Remove(abs)
		return os.Symlink(string(content), abs)
	}
	perm := os.FileMode(0o644)
	if mode.IsExec() {
		perm = 0o755
	}
	tmp := abs + ".tdtmp"
	if err := os.WriteFile(tmp, content, perm); err != nil {
		return err
	}
	if err := os.Chmod(tmp, perm); err != nil {
		os.Remove(tmp)
		return err
	}
	// Rename cần xóa đích trước nếu đích tồn tại.
	_ = os.Remove(abs)
	return os.Rename(tmp, abs)
}

// RemoveFile xóa file và dọn thư mục cha nếu rỗng.
func RemoveFile(abs string, repoRoot string) error {
	if err := os.Remove(abs); err != nil && !os.IsNotExist(err) {
		return err
	}
	pruneEmptyDirs(filepath.Dir(abs), repoRoot)
	return nil
}

// pruneEmptyDirs xóa các thư mục rỗng từ dir ngược lên tới repoRoot.
func pruneEmptyDirs(dir, repoRoot string) {
	for dir != repoRoot && strings.HasPrefix(dir, repoRoot) {
		entries, err := os.ReadDir(dir)
		if err != nil || len(entries) > 0 {
			return
		}
		if err := os.Remove(dir); err != nil {
			return
		}
		dir = filepath.Dir(dir)
	}
}

// CopyFile sao chép file từ nguồn sang đích, giữ quyền.
func CopyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	fi, err := in.Stat()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, fi.Mode().Perm())
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}

// SameContents so sánh nội dung file trên đĩa với blob trong store.
func SameContents(abs string, mode object.FileMode, want []byte) bool {
	got, err := ReadFileSymlinkAware(abs, mode)
	if err != nil {
		return false
	}
	if len(got) != len(want) {
		return false
	}
	return string(got) == string(want)
}

// ToSlash chuẩn hóa đường dẫn sang dạng dùng cho index và tree.
func ToSlash(p string) string {
	return filepath.ToSlash(filepath.Clean(p))
}

// ErrOutsideRepo báo lỗi khi đường dẫn nằm ngoài thư mục repo.
type ErrOutsideRepo struct{ Path string }

func (e ErrOutsideRepo) Error() string {
	return fmt.Sprintf("đường dẫn %s nằm ngoài thư mục repo", e.Path)
}
