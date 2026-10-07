package sys

import (
	"fmt"
	"io"
	"os"
)

// Rm xoá file hoặc thư mục.
func Rm(w io.Writer, files []string, r, f bool) error {
	for _, file := range files {
		var err error
		if r {
			err = os.RemoveAll(file)
		} else {
			err = os.Remove(file)
		}
		if err != nil && !f {
			return fmt.Errorf("lỗi xoá %s: %w", file, err)
		}
	}
	return nil
}
