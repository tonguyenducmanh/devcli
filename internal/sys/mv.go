package sys

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Mv di chuyển hoặc đổi tên file.
func Mv(w io.Writer, sources []string, target string) error {
	info, err := os.Stat(target)
	isDir := err == nil && info.IsDir()

	if len(sources) > 1 && !isDir {
		return fmt.Errorf("đích đến %s phải là thư mục khi di chuyển nhiều tệp", target)
	}
	for _, src := range sources {
		dst := target
		if isDir {
			dst = filepath.Join(target, filepath.Base(src))
		}
		if err := os.Rename(src, dst); err != nil {
			return fmt.Errorf("lỗi di chuyển %s: %w", src, err)
		}
	}
	return nil
}
