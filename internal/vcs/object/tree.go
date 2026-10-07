package object

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
)

// FileMode là quyền/thể loại của một entry trong tree.
type FileMode string

// Các chế độ file thông dụng mà td ghi nhận trong cây.
const (
	ModeBlob    FileMode = "100644" // file thường
	ModeExec    FileMode = "100755" // file thực thi
	ModeSymlink FileMode = "120000" // liên kết tượng trưng
	ModeTree    FileMode = "40000"  // thư mục
)

// IsTree báo xem entry có phải thư mục hay không.
func (m FileMode) IsTree() bool { return m == ModeTree }

// IsExec báo xem entry có phải file thực thi hay không.
func (m FileMode) IsExec() bool { return m == ModeExec }

// TreeEntry là một mục trong tree.
type TreeEntry struct {
	Mode FileMode
	Name string
	Hash Hash
}

// Tree là danh sách entry của một thư mục.
type Tree struct {
	Entries []TreeEntry
}

// Encode chuyển tree thành nội dung thô: mỗi entry là "<mode> <name>\0<hash>".
func (t *Tree) Encode() []byte {
	var b bytes.Buffer
	for _, e := range t.Entries {
		fmt.Fprintf(&b, "%s %s%c", e.Mode, e.Name, byte(0))
		b.Write(e.Hash[:])
	}
	return b.Bytes()
}

// Sort sắp xếp entry theo tên. Tên của một thư mục được so sánh kèm
// dấu "/" phía sau, nhờ vậy "src.txt" đứng trước "src" vì "src.txt"
// nhỏ hơn "src/" khi so sánh chuỗi.
func (t *Tree) Sort() {
	sort.Slice(t.Entries, func(i, j int) bool {
		return lessEntryName(t.Entries[i], t.Entries[j])
	})
}

func lessEntryName(a, b TreeEntry) bool {
	an, bn := a.Name, b.Name
	if a.Mode.IsTree() {
		an += "/"
	}
	if b.Mode.IsTree() {
		bn += "/"
	}
	return an < bn
}

// DecodeTree phân tích nội dung thô thành Tree.
func DecodeTree(data []byte) (*Tree, error) {
	t := &Tree{}
	for len(data) > 0 {
		space := bytes.IndexByte(data, ' ')
		if space < 0 {
			return nil, fmt.Errorf("tree không hợp lệ: thiếu khoảng trắng")
		}
		mode := FileMode(data[:space])
		rest := data[space+1:]
		nul := bytes.IndexByte(rest, 0)
		if nul < 0 || len(rest) < nul+1+20 {
			return nil, fmt.Errorf("tree không hợp lệ: entry thiếu dữ liệu")
		}
		var h Hash
		copy(h[:], rest[nul+1:nul+21])
		name := string(rest[:nul])
		if strings.ContainsAny(name, "/\x00") {
			return nil, fmt.Errorf("tên entry không hợp lệ: %q", name)
		}
		t.Entries = append(t.Entries, TreeEntry{Mode: mode, Name: name, Hash: h})
		data = rest[nul+21:]
	}
	return t, nil
}

// Get trả về entry theo tên, hoặc nil nếu không có.
func (t *Tree) Get(name string) *TreeEntry {
	for i := range t.Entries {
		if t.Entries[i].Name == name {
			return &t.Entries[i]
		}
	}
	return nil
}

// Upsert thêm hoặc cập nhật một entry rồi sắp xếp lại.
func (t *Tree) Upsert(e TreeEntry) {
	for i := range t.Entries {
		if t.Entries[i].Name == e.Name {
			t.Entries[i] = e
			return
		}
	}
	t.Entries = append(t.Entries, e)
	t.Sort()
}

// Remove xóa entry theo tên.
func (t *Tree) Remove(name string) {
	for i := range t.Entries {
		if t.Entries[i].Name == name {
			t.Entries = append(t.Entries[:i], t.Entries[i+1:]...)
			return
		}
	}
}
