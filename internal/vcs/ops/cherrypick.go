package ops

import (
	"fmt"

	"github.com/tonguyenducmanh/devcli/internal/vcs/object"
	"github.com/tonguyenducmanh/devcli/internal/vcs/repo"
)

// CherryPickOptions điều khiển lệnh cherry-pick.
type CherryPickOptions struct {
	// Commits là danh sách commit cần áp dụng, theo thứ tự.
	Commits []object.Hash
	// NoCommit áp dụng thay đổi vào index nhưng không tạo commit.
	NoCommit bool
	// Abort huỷ thao tác đang dở dang.
	Abort bool
	// Continue tiếp tục sau khi giải quyết xung đột.
	Continue bool
	// X bỏ qua commit đang gây xung đột.
	Skip bool
}

// CherryPickResult là kết quả áp dụng một hoặc nhiều commit.
type CherryPickResult struct {
	// Commits là các hash commit mới được tạo.
	Commits []object.Hash
	// Conflicts là danh sách file xung đột.
	Conflicts []string
	// Skipped là số commit bị bỏ qua do xung đột.
	Skipped int
}

// CherryPick áp dụng thay đổi của một hoặc nhiều commit lên nhánh hiện tại.
func CherryPick(r *repo.Repo, opts CherryPickOptions) (*CherryPickResult, error) {
	if opts.Abort {
		return abortCherryPick(r)
	}
	if opts.Continue {
		return continueCherryPick(r)
	}
	if opts.Skip {
		return skipCherryPick(r)
	}
	if len(opts.Commits) == 0 {
		return nil, fmt.Errorf("thiếu đối số: cần chỉ định commit cần áp dụng")
	}

	res := &CherryPickResult{}
	for i, target := range opts.Commits {
		applied, conflicts, err := applyOneCommit(r, target, opts.NoCommit)
		if err != nil {
			return res, err
		}
		if len(conflicts) > 0 {
			// Ghi trạng thái để có thể --continue/--abort sau đó.
			_ = r.WriteState(CherryPickHead, target.String()+"\n")
			res.Conflicts = conflicts
			res.Skipped = len(opts.Commits) - i
			return res, nil
		}
		if applied != object.ZeroHash {
			res.Commits = append(res.Commits, applied)
		}
	}
	return res, nil
}

// applyOneCommit áp dụng một commit bằng cách hợp nhất nội dung với HEAD hiện tại.
func applyOneCommit(r *repo.Repo, target object.Hash, noCommit bool) (object.Hash, []string, error) {
	src, err := r.Objects.ReadCommit(target)
	if err != nil {
		return object.ZeroHash, nil, err
	}
	head, err := r.Head()
	if err != nil {
		return object.ZeroHash, nil, err
	}

	// Base là tree của phụ huynh đầu tiên (hoặc rỗng nếu là commit gốc).
	base := object.ZeroHash
	if len(src.Parents) > 0 {
		base, err = r.CommitTree(src.Parents[0])
		if err != nil {
			return object.ZeroHash, nil, err
		}
	}
	headTree := object.ZeroHash
	if !head.IsZero() {
		headTree, err = r.CommitTree(head)
		if err != nil {
			return object.ZeroHash, nil, err
		}
	}

	// Nếu nội dung đích đã giống hệt thì không có gì để làm.
	if headTree == src.Tree {
		if noCommit {
			return object.ZeroHash, nil, nil
		}
		return head, nil, nil
	}

	conflicts, err := mergeTreesInto(r, base, headTree, src.Tree, src.Summary())
	if err != nil {
		return object.ZeroHash, nil, err
	}
	if err := r.SaveIndex(); err != nil {
		return object.ZeroHash, nil, err
	}
	if len(conflicts) > 0 {
		return object.ZeroHash, conflicts, nil
	}
	if noCommit {
		return object.ZeroHash, nil, nil
	}

	tree, err := r.TreeFromIndex()
	if err != nil {
		return object.ZeroHash, nil, err
	}
	msg := fmt.Sprintf("cherry-pick: %s\n\n%s", src.Summary(), src.Message)
	h, err := writeCommit(r, tree, []object.Hash{head}, msg)
	if err != nil {
		return object.ZeroHash, nil, err
	}
	if err := r.UpdateHeadCommit(h, "cherry-pick: "+src.Summary()); err != nil {
		return object.ZeroHash, nil, err
	}
	// Nếu cherry-pick xong thì xóa trạng thái cũ.
	_ = r.ClearState(CherryPickHead)
	return h, nil, nil
}

// abortCherryPick huỷ cherry-pick đang dở dang.
func abortCherryPick(r *repo.Repo) (*CherryPickResult, error) {
	if !r.HasState(CherryPickHead) {
		return nil, fmt.Errorf("không có lần cherry-pick nào đang dở dang")
	}
	head, err := r.Head()
	if err != nil {
		return nil, err
	}
	if head.IsZero() {
		return nil, fmt.Errorf("không có commit nào để khôi phục")
	}
	tree, err := r.CommitTree(head)
	if err != nil {
		return nil, err
	}
	if err := r.ResetWorktreeTo(tree); err != nil {
		return nil, err
	}
	if err := r.ClearState(CherryPickHead); err != nil {
		return nil, err
	}
	return &CherryPickResult{}, nil
}

// continueCherryPick hoàn tất cherry-pick sau khi giải quyết xung đột.
func continueCherryPick(r *repo.Repo) (*CherryPickResult, error) {
	if !r.HasState(CherryPickHead) {
		return nil, fmt.Errorf("không có lần cherry-pick nào đang dở dang")
	}
	raw, err := r.ReadState(CherryPickHead)
	if err != nil {
		return nil, err
	}
	target, err := parseHashString(raw)
	if err != nil {
		return nil, err
	}
	if r.Index.HasConflicts() {
		return nil, fmt.Errorf("vẫn còn %d file xung đột", len(r.Index.Conflicts()))
	}
	head, err := r.Head()
	if err != nil {
		return nil, err
	}
	tree, err := r.TreeFromIndex()
	if err != nil {
		return nil, err
	}
	src, err := r.Objects.ReadCommit(target)
	if err != nil {
		return nil, err
	}
	msg := fmt.Sprintf("cherry-pick: %s\n\n%s", src.Summary(), src.Message)
	h, err := writeCommit(r, tree, []object.Hash{head}, msg)
	if err != nil {
		return nil, err
	}
	if err := r.UpdateHeadCommit(h, "cherry-pick (tiếp tục): "+src.Summary()); err != nil {
		return nil, err
	}
	if err := r.SaveIndex(); err != nil {
		return nil, err
	}
	_ = r.ClearState(CherryPickHead)
	return &CherryPickResult{Commits: []object.Hash{h}}, nil
}

// skipCherryPick bỏ qua commit đang gây xung đột.
func skipCherryPick(r *repo.Repo) (*CherryPickResult, error) {
	if !r.HasState(CherryPickHead) {
		return nil, fmt.Errorf("không có lần cherry-pick nào đang dở dang")
	}
	head, err := r.Head()
	if err != nil {
		return nil, err
	}
	tree, err := r.CommitTree(head)
	if err != nil {
		return nil, err
	}
	if err := r.ResetWorktreeTo(tree); err != nil {
		return nil, err
	}
	if err := r.ClearState(CherryPickHead); err != nil {
		return nil, err
	}
	return &CherryPickResult{Skipped: 1}, nil
}

// CherryPickInProgress báo xem đang cherry-pick dở dang hay không.
func CherryPickInProgress(r *repo.Repo) bool { return r.HasState(CherryPickHead) }
