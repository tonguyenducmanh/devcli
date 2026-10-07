package coreutils

import (
	"fmt"
	"io"
	"os"
)

// Cat đọc và in toàn bộ nội dung của các tệp tin được truyền vào.
func Cat(w io.Writer, filenames []string) error {
	if len(filenames) == 0 {
		_, err := io.Copy(w, os.Stdin)
		return err
	}

	for _, filename := range filenames {
		if filename == "-" {
			if _, err := io.Copy(w, os.Stdin); err != nil {
				return err
			}
			continue
		}

		f, err := os.Open(filename)
		if err != nil {
			return fmt.Errorf("không thể mở %s: %w", filename, err)
		}

		_, err = io.Copy(w, f)
		f.Close()

		if err != nil {
			return fmt.Errorf("lỗi đọc %s: %w", filename, err)
		}
	}
	return nil
}
