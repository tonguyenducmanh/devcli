package sys

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

// Tail đọc và in ra N dòng cuối cùng của tệp tin.
func Tail(w io.Writer, filename string, lines int) error {
	f, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("không thể mở %s: %w", filename, err)
	}
	defer f.Close()

	var buf []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		buf = append(buf, scanner.Text())
		if len(buf) > lines {
			buf = buf[1:]
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	for _, line := range buf {
		fmt.Fprintln(w, line)
	}
	return nil
}
