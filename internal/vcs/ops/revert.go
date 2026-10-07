package ops

import (
	"fmt"

	"github.com/tonguyenducmanh/devcli/internal/vcs/object"
	"github.com/tonguyenducmanh/devcli/internal/vcs/repo"
)

// RevertOptions điều khiển lệnh revert.
type RevertOptions struct {
	// Commits là các commit cần hoàn tác, áp dụng từ cuối về đầu.
	Commits []object.Hash
	// NoCommit áp dụng thay đổi vào index mà không tạo commit.
	NoCommit bool
	// Abort huỷ thao tác đang dở dang.
	Abort bool
	// Continue hoàn tất sau khi giải quyết xung đột.
	Continue bool
	// Skip bỏ qua commit đang gây xung đột.
	Skip bool
}

// RevertResult là kết quả hoàn tác một hoặc nhiều commit.
type RevertResult struct {
	// Commits là các hash commit mới tạo ra.
	Commits []object.Hash
	// Conflicts là danh sách file xung đột.
	Conflicts []string
	// Skipped là số commit bị bỏ qua.
	Skipped int
}

// Revert hoàn tác thay đổi của một hoặc nhiều commit.
func Revert(r *repo.Repo, opts RevertOptions) (*RevertResult, error) {
	if opts.Abort {
		return abortRevert(r)
	}
	if opts.Continue {
		return continueRevert(r)
	}
	if opts.Skip {
		return skipRevert(r)
	}
	if len(opts.Commits) == 0 {
		return nil, fmt.Errorf("thiếu đối số: cần chỉ định commit cần hoàn tác")
	}

	res := &RevertResult{}
	// Hoàn tác từ commit mới nhất về cũ để tránh xung đột chồng lấn.
	for i := len(opts.Commits) - 1; i >= 0; i-- {
		target := opts.Commits[i]
		created, conflicts, err := revertOneCommit(r, target, opts.NoCommit)
		if err != nil {
			return res, err
		}
		if len(conflicts) > 0 {
			_ = r.WriteState(RevertHead, target.String()+"\n")
			_ = r.WriteState(RevertCommit, target.String()+"\n")
			res.Conflicts = conflicts
			res.Skipped = i
			return res, nil
		}
		if created != object.ZeroHash {
			res.Commits = append(res.Commits, created)
		}
	}
	return res, nil
}

// revertOneCommit hoàn tác một commit bằng cách hợp nhất ngược lại.
func revertOneCommit(r *repo.Repo, target object.Hash, noCommit bool) (object.Hash, []string, error) {
	src, err := r.Objects.ReadCommit(target)
	if err != nil {
		return object.ZeroHash, nil, err
	}
	head, err := r.Head()
	if err != nil {
		return object.ZeroHash, nil, err
	}

	// Tree trước khi commit: dùng làm nội dung đích khi hoàn tác.
	before := object.ZeroHash
	if len(src.Parents) > 0 {
		before, err = r.CommitTree(src.Parents[0])
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

	if headTree == before {
		// Nội dung đã trở về đúng trạng thái trước commit, không cần làm gì.
		if noCommit {
			return object.ZeroHash, nil, nil
		}
		return head, nil, nil
	}

	// Hoàn tác nghĩa là hợp nhất "trước khi commit" vào trạng thái hiện tại.
	// Base là tree sau khi commit, ours là HEAD hiện tại, theirs là tree trước commit.
	conflicts, err := mergeTreesInto(r, src.Tree, headTree, before, src.Summary())
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
	id := r.Identity()
	msg := fmt.Sprintf("Revert \"%s\"\n\nHoàn tác commit %s.", src.Summary(), target.Short(12))
	c := &object.Commit{
		Tree:      tree,
		Parents:   []object.Hash{head},
		Author:    id,
		Committer: id,
		Message:   msg,
	}
	h, err := r.Objects.WriteCommit(c)
	if err != nil {
		return object.ZeroHash, nil, err
	}
	if err := r.UpdateHeadCommit(h, "revert: "+src.Summary()); err != nil {
		return object.ZeroHash, nil, err
	}
	_ = r.ClearState(RevertHead)
	_ = r.ClearState(RevertCommit)
	return h, nil, nil
}

// abortRevert huỷ thao tác revert đang dở dang.
func abortRevert(r *repo.Repo) (*RevertResult, error) {
	if !r.HasState(RevertHead) {
		return nil, fmt.Errorf("không có lần revert nào đang dở dang")
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
	if err := r.ClearState(RevertHead); err != nil {
		return nil, err
	}
	_ = r.ClearState(RevertCommit)
	return &RevertResult{}, nil
}

// continueRevert hoàn tất revert sau khi giải quyết xung đột.
func continueRevert(r *repo.Repo) (*RevertResult, error) {
	if !r.HasState(RevertHead) {
		return nil, fmt.Errorf("không có lần revert nào đang dở dang")
	}
	if r.Index.HasConflicts() {
		return nil, fmt.Errorf("vẫn còn %d file xung đột", len(r.Index.Conflicts()))
	}
	raw, err := r.ReadState(RevertHead)
	if err != nil {
		return nil, err
	}
	target, err := parseHashString(raw)
	if err != nil {
		return nil, err
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
	id := r.Identity()
	msg := fmt.Sprintf("Revert \"%s\"\n\nHoàn tác commit %s.", src.Summary(), target.Short(12))
	c := &object.Commit{
		Tree:      tree,
		Parents:   []object.Hash{head},
		Author:    id,
		Committer: id,
		Message:   msg,
	}
	h, err := r.Objects.WriteCommit(c)
	if err != nil {
		return nil, err
	}
	if err := r.UpdateHeadCommit(h, "revert (tiếp tục): "+src.Summary()); err != nil {
		return nil, err
	}
	if err := r.SaveIndex(); err != nil {
		return nil, err
	}
	_ = r.ClearState(RevertHead)
	_ = r.ClearState(RevertCommit)
	return &RevertResult{Commits: []object.Hash{h}}, nil
}

// skipRevert bỏ qua commit đang gây xung đột.
func skipRevert(r *repo.Repo) (*RevertResult, error) {
	if !r.HasState(RevertHead) {
		return nil, fmt.Errorf("không có lần revert nào đang dở dang")
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
	_ = r.ClearState(RevertHead)
	_ = r.ClearState(RevertCommit)
	return &RevertResult{Skipped: 1}, nil
}

// RevertInProgress báo xem đang revert dở dang hay không.
func RevertInProgress(r *repo.Repo) bool { return r.HasState(RevertHead) }
