package sys

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// Wc đếm số dòng, từ, ký tự.
func Wc(w io.Writer, files []string, l, wo, c bool) error {
	if !l && !wo && !c {
		l, wo, c = true, true, true
	}
	for _, file := range files {
		f, err := os.Open(file)
		if err != nil {
			return err
		}

		scanner := bufio.NewScanner(f)
		lines, words, chars := 0, 0, 0
		for scanner.Scan() {
			lines++
			text := scanner.Text()
			chars += len(text) + 1 // newline
			words += len(strings.Fields(text))
		}
		f.Close()

		out := ""
		if l {
			out += fmt.Sprintf("%8d ", lines)
		}
		if wo {
			out += fmt.Sprintf("%8d ", words)
		}
		if c {
			out += fmt.Sprintf("%8d ", chars)
		}
		fmt.Fprintf(w, "%s %s\n", out, file)
	}
	return nil
}
