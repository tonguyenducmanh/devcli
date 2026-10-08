// Package ops chứa nghiệp vụ cấp cao cho các lệnh version control của tm.
package ops

import (
	"fmt"
	"strings"

	"github.com/tonguyenducmanh/devcli/internal/vcs/object"
	"github.com/tonguyenducmanh/devcli/internal/vcs/repo"
)

// AddOptions điều khiển lệnh add.
type AddOptions struct {
	// All thêm tất cả thay đổi trong toàn repo.
	All bool
	// Update chỉ cập nhật file đã được theo dõi, không thêm file mới.
	Update bool
	// Paths là danh sách đường dẫn hoặc mẫu glob do người dùng chỉ định.
	Paths []string
}

// Add đưa thay đổi vào staging area.
func Add(r *repo.Repo, opts AddOptions) error {
	if opts.All {
		return addAll(r, false)
	}
	if opts.Update {
		return addUpdate(r)
	}
	if len(opts.Paths) == 0 {
		// Không có đường dẫn thì chỉ cập nhật các tệp đã được theo dõi,
		// không thêm tệp mới xuất hiện trên đĩa.
		return addUpdate(r)
	}

	st, err := r.Status()
	if err != nil {
		return err
	}
	matched := false
	for _, e := range st.Entries {
		if !matchesAnyPath(e.Path, opts.Paths) {
			continue
		}
		// Tệp đã bị xoá khỏi đĩa thì gỡ khỏi index, để commit ghi nhận việc
		// xoá.
		//
		// Không dùng UnstagePath ở đây: hàm đó khôi phục lại nội dung của tệp
		// từ HEAD, tức là giữ tệp trong index, nên việc xoá không bao giờ được
		// ghi nhận. Cách gỡ khỏi index giống hệt addAll và addUpdate.
		if e.WorkStatus == 'D' {
			r.Index.Remove(e.Path)
			matched = true
			continue
		}
		if err := r.StageFile(e.Path); err != nil {
			return err
		}
		matched = true
	}
	if !matched {
		return fmt.Errorf("không khớp với đường dẫn nào được chỉ định")
	}
	return r.SaveIndex()
}

// addAll stage mọi thay đổi, bao gồm cả file mới.
func addAll(r *repo.Repo, keepConflicts bool) error {
	st, err := r.Status()
	if err != nil {
		return err
	}
	for _, e := range st.Entries {
		switch {
		case e.WorkStatus == 'D':
			r.Index.Remove(e.Path)
		default:
			if err := r.StageFile(e.Path); err != nil {
				return err
			}
		}
	}
	return r.SaveIndex()
}

// addUpdate chỉ cập nhật các file đã được theo dõi.
func addUpdate(r *repo.Repo) error {
	st, err := r.Status()
	if err != nil {
		return err
	}
	for _, e := range st.Entries {
		if e.IsUntracked {
			continue // không thêm file mới
		}
		if e.WorkStatus == 'D' {
			r.Index.Remove(e.Path)
			continue
		}
		if err := r.StageFile(e.Path); err != nil {
			return err
		}
	}
	return r.SaveIndex()
}

// CommitOptions điều khiển lệnh commit.
type CommitOptions struct {
	Message    string
	Amend      bool
	AllowEmpty bool
	All        bool
	// NoVerify bỏ qua bước kiểm tra (hiện dành cho tương thích CLI).
	NoVerify bool
}

// Commit tạo commit mới từ nội dung staging area.
func Commit(r *repo.Repo, opts CommitOptions) (object.Hash, error) {
	// Cờ -a đưa mọi thay đổi trên đĩa vào stage trước khi kiểm tra.
	if opts.All {
		if err := addAll(r, false); err != nil {
			return object.ZeroHash, err
		}
	}

	// Trạng thái được đọc sau khi stage để phản ánh đúng nội dung sắp commit.
	st, err := r.Status()
	if err != nil {
		return object.ZeroHash, err
	}

	if len(st.Conflicts) > 0 {
		return object.ZeroHash, fmt.Errorf("còn %d file xung đột, hãy giải quyết trước khi commit", len(st.Conflicts))
	}

	if !opts.Amend && !opts.AllowEmpty {
		if len(st.Staged()) == 0 {
			if len(st.Unstaged()) > 0 {
				return object.ZeroHash, fmt.Errorf("không có gì để commit (hãy dùng `tm vcs add` để stage thay đổi)")
			}
			return object.ZeroHash, fmt.Errorf("không có gì để commit, cây làm việc sạch")
		}
	}

	head, err := r.Head()
	if err != nil {
		return object.ZeroHash, err
	}

	tree, err := r.TreeFromIndex()
	if err != nil {
		return object.ZeroHash, err
	}

	var parents []object.Hash
	msg := opts.Message

	if opts.Amend {
		if head.IsZero() {
			return object.ZeroHash, fmt.Errorf("không có commit nào để amend")
		}
		old, err := r.Objects.ReadCommit(head)
		if err != nil {
			return object.ZeroHash, err
		}
		parents = old.Parents
		if msg == "" {
			msg = old.Message // giữ nguyên message cũ
		}
		// Cập nhật reflog của nhánh/HEAD sau khi ghi.
		h, err := writeCommit(r, tree, parents, msg)
		if err != nil {
			return object.ZeroHash, err
		}
		if err := r.UpdateHeadCommit(h, "commit (amend)"); err != nil {
			return object.ZeroHash, err
		}
		return h, nil
	}

	if !head.IsZero() {
		parents = append(parents, head)
	}

	h, err := writeCommit(r, tree, parents, msg)
	if err != nil {
		return object.ZeroHash, err
	}
	if err := r.UpdateHeadCommit(h, "commit: "+firstLine(msg)); err != nil {
		return object.ZeroHash, err
	}

	// Dọn trạng thái merge/revert nếu có sau khi commit thành công.
	if r.HasState("MERGE_HEAD") {
		_ = r.ClearState("MERGE_HEAD")
	}
	if r.HasState("REVERT_HEAD") {
		_ = r.ClearState("REVERT_HEAD")
	}
	if r.HasState("CHERRY_PICK_HEAD") {
		_ = r.ClearState("CHERRY_PICK_HEAD")
	}
	return h, nil
}

// writeCommit ghi một object commit với thông tin tác giả hiện tại.
func writeCommit(r *repo.Repo, tree object.Hash, parents []object.Hash, msg string) (object.Hash, error) {
	if msg == "" {
		return object.ZeroHash, fmt.Errorf("thiếu nội dung commit (dùng -m để viết nhanh)")
	}
	if !strings.HasSuffix(msg, "\n") {
		msg += "\n"
	}
	id := r.Identity()
	c := &object.Commit{
		Tree:      tree,
		Parents:   parents,
		Author:    id,
		Committer: id,
		Message:   msg,
	}
	return r.Objects.WriteCommit(c)
}
