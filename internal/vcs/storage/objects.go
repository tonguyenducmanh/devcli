// Package storage quản lý việc đọc/ghi object và tham chiếu (refs) trên đĩa.
package storage

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/tonguyenducmanh/devcli/internal/vcs/object"
)

// ObjectStore lưu từng object thành một file riêng, nén zlib,
// nằm theo đường dẫn .tmx/objects/ab/cdef... trong đó ab là hai ký tự
// đầu của mã băm để phân tán đều các file trên nhiều thư mục.
type ObjectStore struct {
	dir string
}

// NewObjectStore tạo store trỏ tới thư mục objects.
func NewObjectStore(dir string) *ObjectStore { return &ObjectStore{dir: dir} }

// Dir trả về thư mục gốc của store.
func (s *ObjectStore) Dir() string { return s.dir }

// pathFor tính đường dẫn file của một hash.
func (s *ObjectStore) pathFor(h object.Hash) string {
	hex := h.String()
	return filepath.Join(s.dir, hex[:2], hex[2:])
}

// Write lưu object với loại t và nội dung content, trả về hash.
// Nếu object đã tồn tại thì không ghi lại (idempotent).
func (s *ObjectStore) Write(t object.Type, content []byte) (object.Hash, error) {
	h := object.ComputeHash(t, content)
	p := s.pathFor(h)
	if _, err := os.Stat(p); err == nil {
		return h, nil // đã tồn tại
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return h, err
	}
	var buf bytes.Buffer
	zw := zlib.NewWriter(&buf)
	// Ghi phần mô tả "<loại> <kích thước>" cùng byte NUL phân tách
	// rồi mới tới nội dung, đúng thứ tự mà ComputeHash dùng để tính mã băm.
	if _, err := zw.Write([]byte(fmt.Sprintf("%s %d%c", t, len(content), byte(0)))); err != nil {
		return h, err
	}
	if _, err := zw.Write(content); err != nil {
		return h, err
	}
	if err := zw.Close(); err != nil {
		return h, err
	}
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, buf.Bytes(), 0o444); err != nil {
		return h, err
	}
	return h, os.Rename(tmp, p)
}

// Has kiểm tra object đã tồn tại chưa.
func (s *ObjectStore) Has(h object.Hash) bool {
	_, err := os.Stat(s.pathFor(h))
	return err == nil
}

// Read đọc object từ store, tự động giải nén.
func (s *ObjectStore) Read(h object.Hash) (object.Type, []byte, error) {
	data, err := os.ReadFile(s.pathFor(h))
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil, fmt.Errorf("không tìm thấy object %s", h.Short(12))
		}
		return "", nil, err
	}
	zr, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return "", nil, fmt.Errorf("object %s hỏng: %w", h.Short(12), err)
	}
	defer zr.Close()
	raw, err := io.ReadAll(zr)
	if err != nil {
		return "", nil, fmt.Errorf("object %s hỏng: %w", h.Short(12), err)
	}
	// Tách header "<type> <size>\0".
	nul := bytes.IndexByte(raw, 0)
	if nul < 0 {
		return "", nil, fmt.Errorf("object %s hỏng: thiếu header", h.Short(12))
	}
	header := string(raw[:nul])
	var t object.Type
	if _, err := fmt.Sscanf(header, "%s", &t); err != nil {
		return "", nil, fmt.Errorf("object %s hỏng: %w", h.Short(12), err)
	}
	return t, raw[nul+1:], nil
}

// ReadTree đọc và phân tích một tree object.
func (s *ObjectStore) ReadTree(h object.Hash) (*object.Tree, error) {
	t, data, err := s.Read(h)
	if err != nil {
		return nil, err
	}
	if t != object.TypeTree {
		return nil, fmt.Errorf("%s không phải tree", h.Short(12))
	}
	return object.DecodeTree(data)
}

// ReadCommit đọc và phân tích một commit object.
func (s *ObjectStore) ReadCommit(h object.Hash) (*object.Commit, error) {
	t, data, err := s.Read(h)
	if err != nil {
		return nil, err
	}
	if t != object.TypeCommit {
		return nil, fmt.Errorf("%s không phải commit", h.Short(12))
	}
	return object.DecodeCommit(data)
}

// ReadBlob đọc nội dung của một blob object.
func (s *ObjectStore) ReadBlob(h object.Hash) ([]byte, error) {
	t, data, err := s.Read(h)
	if err != nil {
		return nil, err
	}
	if t != object.TypeBlob {
		return nil, fmt.Errorf("%s không phải blob", h.Short(12))
	}
	return data, nil
}

// WriteBlob tiện ích ghi blob.
func (s *ObjectStore) WriteBlob(content []byte) (object.Hash, error) {
	return s.Write(object.TypeBlob, content)
}

// WriteTree tiện ích ghi tree (tự sắp xếp trước khi ghi).
func (s *ObjectStore) WriteTree(t *object.Tree) (object.Hash, error) {
	t.Sort()
	return s.Write(object.TypeTree, t.Encode())
}

// WriteCommit tiện ích ghi commit.
func (s *ObjectStore) WriteCommit(c *object.Commit) (object.Hash, error) {
	return s.Write(object.TypeCommit, c.Encode())
}
