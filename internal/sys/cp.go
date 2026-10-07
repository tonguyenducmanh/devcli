package sys

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func cpFile(src, dst string) error {
	s, err := os.Open(src)
	if err != nil {
		return err
	}
	defer s.Close()
	d, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer d.Close()
	_, err = io.Copy(d, s)
	return err
}

// Cp sao chép file hoặc thư mục.
func Cp(w io.Writer, sources []string, target string, r bool) error {
	info, err := os.Stat(target)
	isDir := err == nil && info.IsDir()
	if len(sources) > 1 && !isDir {
		return fmt.Errorf("đích đến %s phải là thư mục khi sao chép nhiều tệp", target)
	}
	for _, src := range sources {
		dst := target
		if isDir {
			dst = filepath.Join(target, filepath.Base(src))
		}
		srcInfo, err := os.Stat(src)
		if err != nil {
			return err
		}
		if srcInfo.IsDir() {
			if !r {
				return fmt.Errorf("không có cờ -r, bỏ qua thư mục %s", src)
			}
			err = filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				rel, _ := filepath.Rel(src, path)
				d := filepath.Join(dst, rel)
				if info.IsDir() {
					return os.MkdirAll(d, info.Mode())
				}
				return cpFile(path, d)
			})
			if err != nil {
				return err
			}
		} else {
			if err := cpFile(src, dst); err != nil {
				return err
			}
		}
	}
	return nil
}
