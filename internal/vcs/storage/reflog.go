package storage

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tonguyenducmanh/devcli/internal/vcs/object"
)

// ReflogEntry là một dòng trong reflog của một ref.
type ReflogEntry struct {
	Old     object.Hash
	New     object.Hash
	Author  string
	Message string
}

// ReadReflog đọc toàn bộ reflog của một ref theo thứ tự thời gian tăng dần.
// Trả về lỗi nếu reflog chưa tồn tại.
func (r *RefStore) ReadReflog(name string) ([]ReflogEntry, error) {
	path := filepath.Join(r.dir, "logs", filepath.FromSlash(name))
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrReflogNotFound
		}
		return nil, err
	}
	defer f.Close()

	var out []ReflogEntry
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		line := sc.Text()
		if line == "" {
			continue
		}
		// Định dạng: "<old> <new> <tác giả>\t<message>"
		tabIdx := strings.IndexByte(line, '\t')
		msg := ""
		head := line
		if tabIdx >= 0 {
			msg = line[tabIdx+1:]
			head = line[:tabIdx]
		}
		fields := strings.Fields(head)
		if len(fields) < 2 {
			continue
		}
		oldH, err1 := object.ParseHash(fields[0])
		newH, err2 := object.ParseHash(fields[1])
		if err1 != nil || err2 != nil {
			continue
		}
		author := ""
		if len(fields) > 2 {
			author = strings.Join(fields[2:], " ")
		}
		out = append(out, ReflogEntry{Old: oldH, New: newH, Author: author, Message: msg})
	}
	return out, sc.Err()
}

// ErrReflogNotFound trả về khi reflog chưa tồn tại.
var ErrReflogNotFound = fmt.Errorf("reflog chưa tồn tại")
