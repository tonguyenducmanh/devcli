package sys

import (
	"io"
	"os"
	"time"
)

// Touch tạo file trống hoặc cập nhật thời gian.
func Touch(w io.Writer, files []string) error {
	now := time.Now()
	for _, f := range files {
		file, err := os.OpenFile(f, os.O_RDWR|os.O_CREATE, 0666)
		if err != nil {
			return err
		}
		file.Close()
		os.Chtimes(f, now, now)
	}
	return nil
}
