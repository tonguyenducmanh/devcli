package sys

import (
	"io"
	"os"
)

// Mkdir tạo thư mục.
func Mkdir(w io.Writer, dirs []string, p bool) error {
	for _, d := range dirs {
		if p {
			if err := os.MkdirAll(d, 0755); err != nil {
				return err
			}
		} else {
			if err := os.Mkdir(d, 0755); err != nil {
				return err
			}
		}
	}
	return nil
}
