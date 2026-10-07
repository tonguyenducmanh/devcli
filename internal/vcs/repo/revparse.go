package repo

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/tonguyenducmanh/devcli/internal/vcs/object"
	"github.com/tonguyenducmanh/devcli/internal/vcs/storage"
)

// ResolveRev chuyển một biểu thức tham chiếu thành hash commit.
// Hỗ trợ: hash đầy đủ/ngắn, tên nhánh, tên tag, HEAD, HEAD~n, HEAD^, ref@{n}.
func (r *Repo) ResolveRev(name string) (object.Hash, error) {
	if name == "" {
		return r.Head()
	}
	// Tham chiếu dạng "<rev>~<n>" nghĩa là lùi n đời trên đồ thị phụ huynh.
	if i := strings.IndexByte(name, '~'); i > 0 {
		base := name[:i]
		n, err := strconv.Atoi(name[i+1:])
		if err != nil {
			return object.ZeroHash, fmt.Errorf("số lần lùi không hợp lệ trong %q", name)
		}
		h, err := r.ResolveRev(base)
		if err != nil {
			return object.ZeroHash, err
		}
		return r.nthParent(h, n)
	}
	// Dạng "<rev>^" nghĩa là phụ huynh thứ nhất.
	if strings.HasSuffix(name, "^") {
		h, err := r.ResolveRev(strings.TrimSuffix(name, "^"))
		if err != nil {
			return object.ZeroHash, err
		}
		return r.nthParent(h, 1)
	}
	// Dạng "<rev>^{commit}" ép phân giải về commit.
	if strings.HasSuffix(name, "^{commit}") {
		return r.ResolveRev(strings.TrimSuffix(name, "^{commit}"))
	}
	// Dạng "name@{n}" lấy commit từ reflog.
	if i := strings.Index(name, "@{"); i > 0 && strings.HasSuffix(name, "}") {
		base := name[:i]
		idxStr := name[i+2 : len(name)-1]
		n, err := strconv.Atoi(idxStr)
		if err != nil {
			return object.ZeroHash, fmt.Errorf("chỉ số reflog không hợp lệ trong %q", name)
		}
		return r.resolveReflog(base, n)
	}

	// Thử xem có phải hash hợp lệ không.
	if h, err := object.ParseHash(name); err == nil {
		if !r.Objects.Has(h) {
			return object.ZeroHash, fmt.Errorf("không tìm thấy object %s", h.Short(12))
		}
		return h, nil
	}
	if h, ok := r.lookupShortHash(name); ok {
		return h, nil
	}

	// Thử các ref phổ biến theo thứ tự độ ưu tiên.
	for _, cand := range []string{
		"refs/heads/" + name,
		"refs/tags/" + name,
		"refs/remotes/" + name,
		"refs/" + name,
	} {
		if h, err := r.Refs.Resolve(cand); err == nil {
			// Tag có chú thích trỏ tới một object tag, cần bóc ra commit.
			return r.peelToCommit(h)
		}
	}
	// Cuối cùng thử HEAD.
	if strings.EqualFold(name, "head") {
		h, err := r.Head()
		if err != nil {
			return object.ZeroHash, fmt.Errorf("không tìm thấy HEAD")
		}
		return h, nil
	}
	return object.ZeroHash, fmt.Errorf("không phân giải được tham chiếu %q", name)
}

// peelToCommit bóc một object tag có chú thích để tới commit mà nó trỏ tới.
// Với object thường hoặc tag nhẹ thì trả về nguyên hash.
func (r *Repo) peelToCommit(h object.Hash) (object.Hash, error) {
	typ, data, err := r.Objects.Read(h)
	if err != nil {
		return object.ZeroHash, err
	}
	if typ != object.TypeTag {
		return h, nil
	}
	// Header của object tag bắt đầu bằng dòng "object <hash>".
	for _, line := range strings.Split(string(data), "\n") {
		if rest, ok := strings.CutPrefix(line, "object "); ok {
			return object.ParseHash(strings.TrimSpace(rest))
		}
		if line == "" {
			break
		}
	}
	return object.ZeroHash, fmt.Errorf("object tag %s không hợp lệ", h.Short(12))
}

// nthParent lùi n bước trên đồ thị phụ huynh (n=0 trả về chính nó).
func (r *Repo) nthParent(h object.Hash, n int) (object.Hash, error) {
	cur := h
	for i := 0; i < n; i++ {
		c, err := r.Objects.ReadCommit(cur)
		if err != nil {
			return object.ZeroHash, fmt.Errorf("không có đủ %d đời trên %s", n, h.Short(8))
		}
		if len(c.Parents) == 0 {
			return object.ZeroHash, fmt.Errorf("commit %s là commit gốc", cur.Short(8))
		}
		cur = c.Parents[0]
	}
	return cur, nil
}

// resolveReflog lấy hash tại vị trí n trong reflog của một ref.
func (r *Repo) resolveReflog(name string, n int) (object.Hash, error) {
	ref := name
	if !strings.HasPrefix(name, "refs/") {
		if strings.EqualFold(name, "head") {
			ref = "HEAD"
		} else {
			ref = "refs/heads/" + name
		}
	}
	entries, err := r.ReadReflog(ref)
	if err != nil {
		return object.ZeroHash, fmt.Errorf("không đọc được reflog của %s", name)
	}
	// Reflog đọc theo thứ tự thời gian tăng dần, vị trí 0 là mới nhất.
	idx := len(entries) - 1 - n
	if idx < 0 || idx >= len(entries) {
		return object.ZeroHash, fmt.Errorf("reflog của %s không có mục %d", name, n)
	}
	return entries[idx].New, nil
}

// lookupShortHash tìm object bằng tiền tố hex của hash.
func (r *Repo) lookupShortHash(prefix string) (object.Hash, bool) {
	if len(prefix) < 4 || len(prefix) > 40 {
		return object.ZeroHash, false
	}
	for _, c := range prefix {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			return object.ZeroHash, false
		}
	}
	var found object.Hash
	count := 0
	err := r.Objects.Each(func(h object.Hash) error {
		if !strings.HasPrefix(h.String(), prefix) {
			return nil
		}
		found = h
		count++
		return nil
	})
	if err != nil || count != 1 {
		return object.ZeroHash, false
	}
	return found, true
}

// AncestorSet gom toàn bộ tổ tiên (bao gồm chính nó) của một commit.
// Đây là bản công khai của hàm nội bộ, dùng cho thuật toán rebase.
func (r *Repo) AncestorSet(h object.Hash) (map[object.Hash]bool, error) {
	return r.ancestorSet(h)
}

// CommitDesc tạo mô tả ngắn gọn cho một commit, ví dụ "abc1234 (HEAD -> main)".
func (r *Repo) CommitDesc(h object.Hash, extra string) string {
	desc := h.Short(7)
	if extra != "" {
		desc += " (" + extra + ")"
	}
	return desc
}

// HeadDecorate trả về nhãn trang trí cho HEAD: "HEAD", "HEAD -> main".
func (r *Repo) HeadDecorate() string {
	headRef, err := r.ReadHeadRef()
	if err != nil || headRef == "" {
		head, err := r.Head()
		if err != nil || head.IsZero() {
			return ""
		}
		return "HEAD detached"
	}
	return "HEAD -> " + storage.ShortName(headRef)
}
