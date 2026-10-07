package sys

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// RmEmpty xoá các thư mục rỗng đệ quy bắt đầu từ dir.
func RmEmpty(w io.Writer, dir string) error {
	removedCount := 0

	var walk func(path string) error
	walk = func(path string) error {
		entries, err := os.ReadDir(path)
		if err != nil {
			return fmt.Errorf("không thể đọc thư mục %s: %w", path, err)
		}

		for _, e := range entries {
			if e.IsDir() {
				childPath := filepath.Join(path, e.Name())
				_ = walk(childPath)
			}
		}

		if path != "." {
			if err := os.Remove(path); err == nil {
				fmt.Fprintf(w, "Đã xoá thư mục rỗng: %s\n", path)
				removedCount++
			}
		}
		return nil
	}

	err := walk(dir)
	if removedCount == 0 && err == nil {
		fmt.Fprintln(w, "Không có thư mục rỗng nào.")
	}
	return err
}
