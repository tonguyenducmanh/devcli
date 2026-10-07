package ops

import (
	"fmt"
	"sort"
	"strings"

	"github.com/tonguyenducmanh/devcli/internal/vcs/object"
	"github.com/tonguyenducmanh/devcli/internal/vcs/repo"
)

// Branch là thông tin một nhánh để hiển thị.
type Branch struct {
	Name     string
	Hash     object.Hash
	Current  bool
	Upstream string
	Ahead    int
	Behind   int
}

// ListBranches liệt kê các nhánh cục bộ kèm thông tin nhánh hiện tại.
func ListBranches(r *repo.Repo) ([]Branch, error) {
	names, err := r.Branches()
	if err != nil {
		return nil, err
	}
	current, _ := r.CurrentBranch()

	out := make([]Branch, 0, len(names))
	for _, n := range names {
		h, err := r.BranchHash(n)
		if err != nil {
			continue
		}
		b := Branch{Name: n, Hash: h, Current: n == current}
		if merge := r.Config.GetString("branch."+n+".merge", ""); merge != "" {
			b.Upstream = shortRefName(merge)
			if uh, err := r.Refs.Resolve(merge); err == nil {
				b.Ahead, b.Behind = r.AheadBehind(h, uh)
			}
		}
		out = append(out, b)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

// shortRefName rút gọn tên ref đầy đủ.
func shortRefName(full string) string {
	for _, p := range []string{"refs/heads/", "refs/tags/", "refs/remotes/"} {
		if strings.HasPrefix(full, p) {
			return strings.TrimPrefix(full, p)
		}
	}
	return full
}

// CreateBranchOptions điều khiển lệnh tạo nhánh.
type CreateBranchOptions struct {
	Name        string
	StartPoint  string
	Switch      bool
	Track       bool
	ForceCreate bool
}

// CreateBranch tạo nhánh mới.
func CreateBranch(r *repo.Repo, opts CreateBranchOptions) error {
	if r.BranchExists(opts.Name) {
		if !opts.ForceCreate {
			return fmt.Errorf("nhánh %s đã tồn tại", opts.Name)
		}
		if err := r.DeleteBranch(opts.Name, true); err != nil {
			return err
		}
	}

	start := opts.StartPoint
	if start == "" {
		start = "HEAD"
	}
	h, err := resolveCommitish(r, start)
	if err != nil {
		return err
	}
	if err := r.CreateBranch(opts.Name, start, h); err != nil {
		return err
	}
	// Tự động ghi cấu hình theo dõi khi nhánh mới được tạo từ một nhánh khác.
	if opts.Track || opts.Switch {
		_ = r.SetUpstream(opts.Name, "refs/heads/"+opts.Name, "local")
	}
	if opts.Switch {
		return switchToBranch(r, opts.Name)
	}
	return nil
}

// DeleteBranchEntry là kết quả xóa nhánh.
type DeleteBranchEntry struct {
	Name    string
	Deleted bool
	Reason  string
}

// DeleteBranches xóa một hoặc nhiều nhánh, trả về báo cáo từng nhánh.
func DeleteBranches(r *repo.Repo, names []string, force bool) ([]DeleteBranchEntry, error) {
	current, _ := r.CurrentBranch()
	var out []DeleteBranchEntry
	for _, n := range names {
		if n == current {
			out = append(out, DeleteBranchEntry{Name: n, Reason: "không thể xóa nhánh đang ở"})
			continue
		}
		if !r.BranchExists(n) {
			out = append(out, DeleteBranchEntry{Name: n, Reason: "không tìm thấy nhánh"})
			continue
		}
		if !force {
			// Kiểm tra nhánh đã được hợp nhất vào nhánh hiện tại chưa.
			h, err := r.BranchHash(n)
			if err == nil && !currentHasCommit(r, current, h) {
				out = append(out, DeleteBranchEntry{Name: n, Reason: "chưa hợp nhất vào nhánh hiện tại (dùng -D)"})
				continue
			}
		}
		if err := r.DeleteBranch(n, force); err != nil {
			return nil, err
		}
		_ = r.UnsetUpstream(n)
		out = append(out, DeleteBranchEntry{Name: n, Deleted: true})
	}
	return out, nil
}

// currentHasCommit báo xem commit đã nằm trong lịch sử nhánh đích chưa.
func currentHasCommit(r *repo.Repo, branch string, h object.Hash) bool {
	target, err := r.BranchHash(branch)
	if err != nil {
		return false
	}
	ok, err := r.IsAncestor(h, target)
	if err != nil {
		return false
	}
	return ok
}

// switchToBranch chuyển HEAD sang một nhánh, đồng bộ index và worktree.
func switchToBranch(r *repo.Repo, name string) error {
	if !r.BranchExists(name) {
		return fmt.Errorf("không tìm thấy nhánh %s", name)
	}
	target, err := r.BranchHash(name)
	if err != nil {
		return err
	}
	tree := object.ZeroHash
	if !target.IsZero() {
		tree, err = r.CommitTree(target)
		if err != nil {
			return err
		}
	}
	// Không cho chuyển khi thay đổi chưa lưu sẽ bị mất.
	if err := ensureCleanForSwitch(r, name); err != nil {
		return err
	}
	if err := r.ResetWorktreeTo(tree); err != nil {
		return err
	}
	return r.UpdateHeadRef("refs/heads/"+name, "checkout: chuyển sang "+name)
}

// ensureCleanForSwitch kiểm tra trạng thái trước khi chuyển nhánh.
// Cho phép chuyển nếu file đã stage nhưng chưa commit, nhưng sẽ mang theo
// các thay đổi đó sang nhánh mới.
func ensureCleanForSwitch(r *repo.Repo, targetBranch string) error {
	st, err := r.Status()
	if err != nil {
		return err
	}
	if st.IsClean() {
		return nil
	}
	// Nếu thay đổi đã stage thì vẫn được mang sang, chỉ cần cảnh báo.
	if len(st.Unstaged()) == 0 && len(st.Untracked()) == 0 {
		return nil
	}
	// Nếu file sửa chưa stage sẽ bị ghi đè, báo lỗi và gợi ý commit trước.
	return fmt.Errorf("có thay đổi chưa lưu, hãy commit hoặc stash trước khi chuyển sang %s", targetBranch)
}

// RenameBranch đổi tên nhánh, có thể kèm chuyển HEAD sang nhánh mới.
func RenameBranch(r *repo.Repo, old, newName string, switchTo bool) error {
	if err := r.RenameBranch(old, newName); err != nil {
		return err
	}
	if switchTo {
		return r.UpdateHeadRef("refs/heads/"+newName, "branch: đổi tên từ "+old)
	}
	return nil
}
