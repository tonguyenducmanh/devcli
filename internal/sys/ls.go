package sys

import (
	"fmt"
	"io"
	"io/fs"
	"os"
)

// Ls liệt kê các tệp tin trong thư mục được chỉ định.
func Ls(w io.Writer, dir string, all bool, long bool) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("không thể đọc thư mục %s: %w", dir, err)
	}
	var filtered []fs.DirEntry
	for _, e := range entries {
		// Bỏ qua các tệp ẩn nếu không có cờ all
		if !all && len(e.Name()) > 0 && e.Name()[0] == '.' {
			continue
		}
		filtered = append(filtered, e)
	}

	if long {
		for _, e := range filtered {
			info, err := e.Info()
			if err != nil {
				continue
			}
			fmt.Fprintf(w, "%v %10d %s %s\n", info.Mode(), info.Size(), info.ModTime().Format("Jan 02 15:04"), e.Name())
		}
	} else {
		for _, e := range filtered {
			fmt.Fprintln(w, e.Name())
		}
	}
	return nil
}
