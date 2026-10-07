package sys

import (
	"fmt"
	"io"
	"os"
)

// Pwd in đường dẫn thư mục hiện tại.
func Pwd(w io.Writer) error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	fmt.Fprintln(w, dir)
	return nil
}
