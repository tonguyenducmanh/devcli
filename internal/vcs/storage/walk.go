package storage

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/tonguyenducmanh/devcli/internal/vcs/object"
)

// Each duyệt toàn bộ object đang có trong store (chỉ dạng loose).
// Dùng cho tìm kiếm theo tiền tố hash.
func (s *ObjectStore) Each(fn func(h object.Hash) error) error {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, e := range entries {
		if !e.IsDir() || len(e.Name()) != 2 {
			continue
		}
		files, err := os.ReadDir(filepath.Join(s.dir, e.Name()))
		if err != nil {
			continue
		}
		for _, f := range files {
			name := e.Name() + f.Name()
			if len(name) != 40 {
				continue
			}
			h, err := object.ParseHash(name)
			if err != nil {
				continue
			}
			if err := fn(h); err != nil {
				return err
			}
		}
	}
	_ = fs.ErrInvalid
	return nil
}
