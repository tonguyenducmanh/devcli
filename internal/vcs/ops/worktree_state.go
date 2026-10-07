package ops

import (
	"github.com/tonguyenducmanh/devcli/internal/vcs/object"
	"github.com/tonguyenducmanh/devcli/internal/vcs/repo"
	"github.com/tonguyenducmanh/devcli/internal/vcs/worktree"
)

// StageAllWorkingTree đưa toàn bộ file trên đĩa (kể cả chưa theo dõi) vào index.
func StageAllWorkingTree(r *repo.Repo) error {
	nodes, err := r.ScanWorktree()
	if err != nil {
		return err
	}
	for _, n := range nodes {
		if err := r.StageFile(n.Path); err != nil {
			return err
		}
	}
	return r.SaveIndex()
}

// RestoreIndexFromHEAD đặt index về đúng nội dung của HEAD.
// Nếu chưa có commit thì index trở nên rỗng.
func RestoreIndexFromHEAD(r *repo.Repo) error {
	head, err := r.Head()
	if err != nil || head.IsZero() {
		r.Index.Clear()
		return r.SaveIndex()
	}
	tree, err := r.CommitTree(head)
	if err != nil {
		return err
	}
	return r.WriteIndexFromTree(tree)
}

// RemoveUntrackedFiles xóa mọi file chưa được theo dõi khỏi worktree.
func RemoveUntrackedFiles(r *repo.Repo) error {
	st, err := r.Status()
	if err != nil {
		return err
	}
	for _, e := range st.Untracked() {
		if err := worktree.RemoveFile(r.WorkPath(e.Path), r.Root); err != nil {
			return err
		}
	}
	return nil
}

// RestoreUntrackedFromStash khôi phục lại file chưa được theo dõi từ một entry stash.
// Nếu entry không chứa file chưa theo dõi thì không làm gì.
func RestoreUntrackedFromStash(r *repo.Repo, stash object.Hash) error {
	c, err := r.Objects.ReadCommit(stash)
	if err != nil {
		return err
	}
	// File chưa được theo dõi nằm trong phụ huynh có message đặc biệt ở vị trí 2.
	// Vị trí này chỉ tồn tại khi stash được tạo kèm cờ -u.
	for _, p := range c.Parents[min(2, len(c.Parents)):] {
		uc, err := r.Objects.ReadCommit(p)
		if err != nil || uc.Summary() != stashUntrackedMessage {
			continue
		}
		nodes, err := r.Flatten(uc.Tree)
		if err != nil {
			return err
		}
		for _, n := range nodes {
			data, err := r.Objects.ReadBlob(n.Hash)
			if err != nil {
				return err
			}
			if err := worktree.WriteFileSymlinkAware(r.WorkPath(n.Path), n.Mode, data); err != nil {
				return err
			}
		}
		return nil
	}
	return nil
}
