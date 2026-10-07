package coreutils

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

// Head đọc và in ra N dòng đầu tiên của tệp tin.
func Head(w io.Writer, filename string, lines int) error {
	f, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("không thể mở %s: %w", filename, err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	count := 0
	for scanner.Scan() && count < lines {
		fmt.Fprintln(w, scanner.Text())
		count++
	}
	return scanner.Err()
}
