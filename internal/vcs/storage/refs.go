package storage

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/tonguyenducmanh/devcli/internal/vcs/object"
)

// RefStore quản lý các tham chiếu (refs) dưới dạng file văn bản,
// mỗi file chứa một mã băm hex hoặc một ref gián tiếp.
type RefStore struct {
	dir string // thư mục chứa refs/, ví dụ .tmx
}

// NewRefStore tạo RefStore từ thư mục gốc repo.
func NewRefStore(gitDir string) *RefStore { return &RefStore{dir: gitDir} }

func (r *RefStore) pathFor(name string) string {
	return filepath.Join(r.dir, filepath.FromSlash(name))
}

// Resolve trả về hash mà ref trỏ tới. Hỗ trợ ref gián tiếp (symbolic ref) qua
// tối đa 10 bước để tránh vòng lặp vô hạn.
func (r *RefStore) Resolve(name string) (object.Hash, error) {
	cur := name
	for i := 0; i < 10; i++ {
		data, err := os.ReadFile(r.pathFor(cur))
		if err != nil {
			if os.IsNotExist(err) {
				return object.ZeroHash, ErrRefNotFound
			}
			return object.ZeroHash, err
		}
		s := strings.TrimSpace(string(data))
		if strings.HasPrefix(s, "ref:") {
			cur = strings.TrimSpace(strings.TrimPrefix(s, "ref:"))
			continue
		}
		return object.ParseHash(s)
	}
	return object.ZeroHash, fmt.Errorf("ref %s quá nhiều mức gián tiếp", name)
}

// ErrRefNotFound trả về khi ref không tồn tại.
var ErrRefNotFound = fmt.Errorf("ref không tồn tại")

// Write ghi hash cho một ref.
func (r *RefStore) Write(name string, h object.Hash) error {
	p := r.pathFor(name)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	return os.WriteFile(p, []byte(h.String()+"\n"), 0o644)
}

// WriteSymbolic ghi một ref gián tiếp, ví dụ HEAD -> refs/heads/main.
func (r *RefStore) WriteSymbolic(name, target string) error {
	p := r.pathFor(name)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	return os.WriteFile(p, []byte("ref: "+target+"\n"), 0o644)
}

// Remove xóa ref và dọn thư mục rỗng.
func (r *RefStore) Remove(name string) error {
	p := r.pathFor(name)
	if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
		return err
	}
	// Dọn thư mục cha nếu rỗng (ví dụ refs/heads sau khi xóa nhánh cuối).
	dir := filepath.Dir(p)
	for dir != r.dir && strings.HasPrefix(dir, r.dir) {
		entries, err := os.ReadDir(dir)
		if err != nil || len(entries) > 0 {
			return nil
		}
		if err := os.Remove(dir); err != nil {
			return nil
		}
		dir = filepath.Dir(dir)
	}
	return nil
}

// Exists báo xem ref có tồn tại không.
func (r *RefStore) Exists(name string) bool {
	_, err := os.Stat(r.pathFor(name))
	return err == nil
}

// List trả về danh sách ref dưới prefix cho trước, đã sắp xếp theo tên.
// Ví dụ List("refs/heads/") trả về tên đầy đủ của các nhánh.
func (r *RefStore) List(prefix string) ([]string, error) {
	base := r.pathFor(prefix)
	if prefix == "" || prefix == "." {
		base = r.dir
	}
	var out []string
	err := filepath.WalkDir(base, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(r.dir, path)
		if err != nil {
			return err
		}
		name := filepath.ToSlash(rel)
		// Bỏ qua các file không phải ref (ví dụ index, HEAD đặc biệt).
		if name == "index" || name == "HEAD" || strings.Contains(name, "packed-refs") {
			return nil
		}
		out = append(out, name)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(out)
	return out, nil
}

// ShortName chuyển tên ref đầy đủ thành tên rút gọn dùng trên CLI.
func ShortName(full string) string {
	for _, prefix := range []string{"refs/heads/", "refs/tags/", "refs/remotes/"} {
		if strings.HasPrefix(full, prefix) {
			return strings.TrimPrefix(full, prefix)
		}
	}
	return full
}

// AppendReflog ghi thêm một dòng vào reflog của ref.
func (r *RefStore) AppendReflog(name string, old, new object.Hash, msg string) error {
	logPath := filepath.Join(r.dir, "logs", filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	fmt.Fprintf(w, "%s %s\t%s\n", old, new, msg)
	return w.Flush()
}
