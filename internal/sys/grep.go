package sys

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
)

// Grep tìm chuỗi.
func Grep(w io.Writer, pattern string, files []string) error {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("biểu thức không hợp lệ: %w", err)
	}

	for _, file := range files {
		f, err := os.Open(file)
		if err != nil {
			return err
		}
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			if re.MatchString(line) {
				if len(files) > 1 {
					fmt.Fprintf(w, "%s:%s\n", file, line)
				} else {
					fmt.Fprintln(w, line)
				}
			}
		}
		f.Close()
	}
	return nil
}
