// Package index quản lý vùng chuẩn bị (staging area) của kho mã nguồn:
// tập hợp nội dung tệp sẽ được ghi vào commit kế tiếp.
//
// Định dạng file là index phiên bản 2, lưu tại .tmx/index.
package index

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"

	"github.com/tonguyenducmanh/devcli/internal/vcs/object"
)

// Các hằng số của định dạng file index.
const (
	signature      = "DIRC"
	indexVersion   = 2
	entryHeaderLen = 62 // độ dài phần cố định của một entry
	entryNameLen   = 0xFFF
)

// Stage là mức xung đột khi merge (0 = bình thường).
type Stage uint8

// Các mức xung đột.
const (
	StageNormal Stage = 0
	StageBase   Stage = 1
	StageOurs   Stage = 2
	StageTheirs Stage = 3
)

// Entry là một bản ghi trong staging area.
type Entry struct {
	CTime uint64 // thời điểm đổi inode (giây.nanô giây)
	MTime uint64
	Dev   uint32
	Ino   uint32
	Mode  object.FileMode
	UID   uint32
	GID   uint32
	Size  uint32
	Hash  object.Hash
	Stage Stage
	Name  string // đường dẫn tương đối so với thư mục gốc, dùng dấu "/"
}

// Index là toàn bộ staging area.
type Index struct {
	path    string
	entries []Entry
}

// Open đọc index từ file. Nếu file chưa tồn tại thì trả về index rỗng.
func Open(path string) (*Index, error) {
	idx := &Index{path: path}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return idx, nil
		}
		return nil, err
	}
	if len(data) < 12 || string(data[:4]) != signature {
		return nil, fmt.Errorf("file index không hợp lệ")
	}
	version := binary.BigEndian.Uint32(data[4:8])
	if version != 2 {
		return nil, fmt.Errorf("phiên bản index không được hỗ trợ: %d", version)
	}
	count := int(binary.BigEndian.Uint32(data[8:12]))
	off := 12
	for i := 0; i < count; i++ {
		if off+entryHeaderLen > len(data) {
			return nil, fmt.Errorf("index bị cắt cụt")
		}
		e := Entry{}
		e.CTime = binary.BigEndian.Uint64(data[off:])
		e.MTime = binary.BigEndian.Uint64(data[off+8:])
		e.Dev = binary.BigEndian.Uint32(data[off+16:])
		e.Ino = binary.BigEndian.Uint32(data[off+20:])
		e.Mode = formatMode(binary.BigEndian.Uint32(data[off+24:]))
		e.UID = binary.BigEndian.Uint32(data[off+28:])
		e.GID = binary.BigEndian.Uint32(data[off+32:])
		e.Size = binary.BigEndian.Uint32(data[off+36:])
		copy(e.Hash[:], data[off+40:off+60])
		flags := binary.BigEndian.Uint16(data[off+60:])
		e.Stage = Stage((flags >> 12) & 0x3)
		nameLen := int(flags & entryNameLen)

		// Tên đường dẫn, sau đó đệm NUL tới khi tổng chiều dài entry
		// chia hết cho 8, nhờ vậy phần tên của mọi entry luôn nằm
		// trên cùng một địa chỉ chia hết cho 8.
		nameStart := off + entryHeaderLen
		if nameStart+nameLen > len(data) {
			return nil, fmt.Errorf("index bị cắt cụt")
		}
		e.Name = string(data[nameStart : nameStart+nameLen])
		entryLen := (entryHeaderLen + nameLen + 8) &^ 7
		off = nameStart + entryLen - entryHeaderLen
		idx.entries = append(idx.entries, e)
	}
	// Bỏ qua phần mở rộng (extensions) và checksum ở cuối file.
	return idx, nil
}

// checksum tính SHA1 của toàn bộ phần dữ liệu phía trên, ghi ở cuối file.
func checksum(b []byte) []byte {
	s := sha1.Sum(b)
	return s[:]
}

// Các chế độ file ở dạng số nguyên tương ứng với hằng số ở trên.
const (
	modeBlob    = 0o100644 // file thường, không quyền thực thi
	modeExec    = 0o100755 // file thực thi
	modeSymlink = 0o120000 // liên kết tượng trưng
	modeGitlink = 0o160000 // submodule
)

// parseMode chuyển chế độ file dạng văn bản sang số nguyên cho file index.
func parseMode(m object.FileMode) uint32 {
	n, err := strconv.ParseUint(string(m), 8, 32)
	if err != nil {
		return modeBlob
	}
	return uint32(n)
}

// formatMode chuyển số nguyên trong file index về dạng văn bản.
func formatMode(v uint32) object.FileMode {
	// Chế độ submodule không dùng trong tm nên quy về file thường.
	if v == modeGitlink {
		return object.ModeBlob
	}
	return object.FileMode(strconv.FormatUint(uint64(v), 8))
}

// Encode chuyển index thành nội dung file đầy đủ (kèm checksum).
func (idx *Index) Encode() []byte {
	idx.sort()
	var b bytes.Buffer
	b.WriteString(signature)
	var num [4]byte
	binary.BigEndian.PutUint32(num[:], indexVersion)
	b.Write(num[:])
	binary.BigEndian.PutUint32(num[:], uint32(len(idx.entries)))
	b.Write(num[:])

	for _, e := range idx.entries {
		var hdr [entryHeaderLen]byte
		binary.BigEndian.PutUint64(hdr[0:], e.CTime)
		binary.BigEndian.PutUint64(hdr[8:], e.MTime)
		binary.BigEndian.PutUint32(hdr[16:], e.Dev)
		binary.BigEndian.PutUint32(hdr[20:], e.Ino)
		binary.BigEndian.PutUint32(hdr[24:], parseMode(e.Mode))
		binary.BigEndian.PutUint32(hdr[28:], e.UID)
		binary.BigEndian.PutUint32(hdr[32:], e.GID)
		binary.BigEndian.PutUint32(hdr[36:], e.Size)
		copy(hdr[40:60], e.Hash[:])
		// Cờ: độ dài tên (bị chặn ở 0xFFF) và mức stage ở 3 bit cao.
		nameLen := len(e.Name)
		if nameLen > entryNameLen {
			nameLen = entryNameLen
		}
		flags := uint16(nameLen) | uint16(e.Stage)<<12
		binary.BigEndian.PutUint16(hdr[60:], flags)
		b.Write(hdr[:])

		// Tên được kết thúc bằng byte NUL rồi đệm thêm sao cho
		// tổng chiều dài entry luôn chia hết cho 8.
		b.WriteString(e.Name)
		base := entryHeaderLen + nameLen
		pad := (base + 8) &^ 7
		if extra := pad - base; extra > 0 {
			b.Write(make([]byte, extra))
		}
	}
	b.Write(checksum(b.Bytes()))
	return b.Bytes()
}

// Save ghi index xuống file.
func (idx *Index) Save() error {
	if err := os.MkdirAll(filepath.Dir(idx.path), 0o755); err != nil {
		return err
	}
	tmp := idx.path + ".lock"
	if err := os.WriteFile(tmp, idx.Encode(), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, idx.path)
}

// sort sắp xếp entry theo tên, các entry cùng tên thì theo stage tăng dần.
func (idx *Index) sort() {
	sort.SliceStable(idx.entries, func(i, j int) bool {
		if idx.entries[i].Name != idx.entries[j].Name {
			return idx.entries[i].Name < idx.entries[j].Name
		}
		return idx.entries[i].Stage < idx.entries[j].Stage
	})
}

// Entries trả về toàn bộ entry (bản sao).
func (idx *Index) Entries() []Entry {
	idx.sort()
	out := make([]Entry, len(idx.entries))
	copy(out, idx.entries)
	return out
}

// Get trả về entry ở stage bình thường (0) theo đường dẫn.
func (idx *Index) Get(path string) *Entry {
	for i := range idx.entries {
		if idx.entries[i].Name == path && idx.entries[i].Stage == StageNormal {
			return &idx.entries[i]
		}
	}
	return nil
}

// Stages trả về mọi entry của một đường dẫn, theo stage tăng dần.
func (idx *Index) Stages(path string) []Entry {
	var out []Entry
	for _, e := range idx.entries {
		if e.Name == path {
			out = append(out, e)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Stage < out[j].Stage })
	return out
}

// Set thêm hoặc thay thế entry của một đường dẫn ở stage cho trước.
func (idx *Index) Set(e Entry) {
	for i := range idx.entries {
		if idx.entries[i].Name == e.Name && idx.entries[i].Stage == e.Stage {
			idx.entries[i] = e
			return
		}
	}
	idx.entries = append(idx.entries, e)
}

// Add cập nhật entry ở stage 0.
func (idx *Index) Add(e Entry) {
	e.Stage = StageNormal
	idx.Set(e)
}

// Remove xóa mọi entry của một đường dẫn.
func (idx *Index) Remove(path string) {
	out := idx.entries[:0]
	for _, e := range idx.entries {
		if e.Name == path {
			continue
		}
		out = append(out, e)
	}
	idx.entries = out
}

// RemoveStages xóa các entry có stage khác 0 của một đường dẫn,
// dùng sau khi giải quyết xung đột.
func (idx *Index) RemoveStages(path string) {
	out := idx.entries[:0]
	for _, e := range idx.entries {
		if e.Name == path && e.Stage != StageNormal {
			continue
		}
		out = append(out, e)
	}
	idx.entries = out
}

// Clear làm rỗng index.
func (idx *Index) Clear() { idx.entries = nil }

// Len trả về số entry.
func (idx *Index) Len() int { return len(idx.entries) }

// HasConflicts báo xem index còn entry xung đột (stage > 0) hay không.
func (idx *Index) HasConflicts() bool {
	for _, e := range idx.entries {
		if e.Stage != StageNormal {
			return true
		}
	}
	return false
}

// Conflicts trả về danh sách đường dẫn đang xung đột.
func (idx *Index) Conflicts() []string {
	seen := map[string]bool{}
	var out []string
	for _, e := range idx.entries {
		if e.Stage != StageNormal && !seen[e.Name] {
			seen[e.Name] = true
			out = append(out, e.Name)
		}
	}
	sort.Strings(out)
	return out
}

// Paths trả về danh sách đường dẫn đang được theo dõi.
func (idx *Index) Paths() []string {
	seen := map[string]bool{}
	var out []string
	for _, e := range idx.entries {
		if !seen[e.Name] {
			seen[e.Name] = true
			out = append(out, e.Name)
		}
	}
	sort.Strings(out)
	return out
}

// Walk là duyệt index theo thứ tự, dùng cho các phép so sánh.
func (idx *Index) Walk(fn func(e Entry) bool) {
	for _, e := range idx.Entries() {
		if !fn(e) {
			return
		}
	}
}
