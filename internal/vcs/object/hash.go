// Package object định nghĩa các loại object cơ bản của kho mã nguồn:
// blob, tree, commit, tag. Mỗi object có một mã băm SHA1 duy nhất tính từ
// loại và nội dung của nó, dùng để định danh và kiểm tra tính toàn vẹn.
package object

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
)

// Hash là mã băm 20 byte của một object.
type Hash [20]byte

// ZeroHash là giá trị rỗng, dùng cho object không tồn tại.
var ZeroHash Hash

// String trả về dạng hex 40 ký tự.
func (h Hash) String() string { return hex.EncodeToString(h[:]) }

// IsZero báo xem hash có rỗng hay không.
func (h Hash) IsZero() bool { return h == ZeroHash }

// Short trả về dạng hex rút gọn (mặc định 7 ký tự) để hiển thị.
func (h Hash) Short(n int) string {
	s := h.String()
	if n <= 0 || n > len(s) {
		n = 7
	}
	return s[:n]
}

// ParseHash chuyển chuỗi hex (40 ký tự) thành Hash.
func ParseHash(s string) (Hash, error) {
	var h Hash
	if len(s) != 40 {
		return h, fmt.Errorf("hash không hợp lệ: %q", s)
	}
	b, err := hex.DecodeString(s)
	if err != nil {
		return h, fmt.Errorf("hash không hợp lệ: %q", s)
	}
	copy(h[:], b)
	return h, nil
}

// Type là loại object.
type Type string

// Các loại object được hỗ trợ.
const (
	TypeBlob   Type = "blob"
	TypeTree   Type = "tree"
	TypeCommit Type = "commit"
	TypeTag    Type = "tag"
)

// ErrObjectType là lỗi khi gặp loại object không hỗ trợ.
var ErrObjectType = errors.New("loại object không hỗ trợ")

// ComputeHash tính mã băm của object: SHA1 của chuỗi
// "<loại> <kích thước>" theo sau bởi một byte NUL rồi tới nội dung.
// Byte NUL ở giữa giúp phân biệt rõ ranh giới giữa phần mô tả và nội dung.
func ComputeHash(t Type, content []byte) Hash {
	h := sha1.New()
	fmt.Fprintf(h, "%s %d%c", t, len(content), byte(0))
	h.Write(content)
	var out Hash
	copy(out[:], h.Sum(nil))
	return out
}
