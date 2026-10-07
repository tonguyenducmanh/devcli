package ops

import (
	"fmt"
	"os"
	"strings"

	"github.com/tonguyenducmanh/devcli/internal/vcs/object"
	"github.com/tonguyenducmanh/devcli/internal/vcs/worktree"
)

// lstatPath lấy thông tin stat của một đường dẫn trên đĩa.
func lstatPath(p string) (os.FileInfo, error) { return os.Lstat(p) }

// readWorktreeData đọc nội dung file trên đĩa theo chế độ đã biết.
func readWorktreeData(abs string, mode object.FileMode) ([]byte, error) {
	return worktree.ReadFileSymlinkAware(abs, mode)
}

// fileModeFromInfo suy ra chế độ lưu của một tệp từ thông tin stat.
func fileModeFromInfo(fi os.FileInfo) object.FileMode {
	return worktree.FileModeFromInfo(fi)
}

// splitDiffData tách nội dung thành các dòng cho thuật toán diff.
func splitDiffData(s string) []string {
	if s == "" {
		return nil
	}
	s = strings.ReplaceAll(s, "\r\n", "\n")
	if strings.HasSuffix(s, "\n") {
		s = s[:len(s)-1]
	}
	return strings.Split(s, "\n")
}

// printOut in ra stdout, giúp chỗ gọi ngắn gọn trong các hàm nghiệp vụ.
func printOut(format string, args ...any) {
	fmt.Fprintf(os.Stdout, format, args...)
}
