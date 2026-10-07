package ops

import (
	"fmt"

	"github.com/tonguyenducmanh/devcli/internal/vcs/object"
	"github.com/tonguyenducmanh/devcli/internal/vcs/repo"
)

// stashRef là ref chứa các entry stash, mỗi entry là một commit.
const stashRef = "refs/stash"

// Message cố định của các commit phụ trợ tạo nội bộ cho stash.
const (
	stashIndexMessage     = "td vcs stash index"
	stashUntrackedMessage = "td vcs stash untracked"
)

// StashEntry là một bản stash đã lưu.
type StashEntry struct {
	// Index là vị trí trong danh sách (0 là mới nhất).
	Index int
	// Hash là commit của entry này.
	Hash object.Hash
	// Message là mô tả ngắn.
	Message string
}

// Stash lưu các thay đổi chưa commit vào một entry stash.
// includeUntracked cho biết có lưu cả file chưa được theo dõi hay không.
func Stash(r *repo.Repo, message string, includeUntracked bool) (object.Hash, error) {
	st, err := r.Status()
	if err != nil {
		return object.ZeroHash, err
	}
	if st.IsClean() && !includeUntracked {
		return object.ZeroHash, fmt.Errorf("không có thay đổi nào để lưu")
	}

	head, err := r.Head()
	if err != nil {
		return object.ZeroHash, err
	}
	headTree := object.ZeroHash
	if !head.IsZero() {
		headTree, err = r.CommitTree(head)
		if err != nil {
			return object.ZeroHash, err
		}
	}

	msg := stashMessage(message, st)

	// Cây đầu tiên chụp lại phần đã stage, dùng làm phụ huynh thứ hai
	// để sau này khôi phục lại được đúng trạng thái index.
	indexTree, err := r.TreeFromIndex()
	if err != nil {
		return object.ZeroHash, err
	}

	// Cây thứ hai chụp toàn bộ worktree, bao gồm phần chưa stage.
	workTree := indexTree
	if len(st.Unstaged()) > 0 {
		if err := StageAllWorkingTree(r); err != nil {
			return object.ZeroHash, err
		}
		workTree, err = r.TreeFromIndex()
		if err != nil {
			return object.ZeroHash, err
		}
	}

	// Phụ huynh của commit stash: HEAD, rồi commit index, rồi commit file chưa theo dõi.
	var parents []object.Hash
	if !head.IsZero() {
		parents = append(parents, head)
	}

	// File chưa được theo dõi được lưu vào một commit riêng trước khi đụng index.
	if includeUntracked && len(st.Untracked()) > 0 {
		untrackedTree, err := untrackedTree(r)
		if err != nil {
			return object.ZeroHash, err
		}
		if untrackedTree != object.ZeroHash {
			uc, err := writeCommit(r, untrackedTree, []object.Hash{head}, stashUntrackedMessage+"\n")
			if err != nil {
				return object.ZeroHash, err
			}
			parents = append(parents, uc)
		}
		// Đưa index về đúng phần đã stage ban đầu.
		if err := r.WriteIndexFromTree(indexTree); err != nil {
			return object.ZeroHash, err
		}
	}

	// Commit index để có thể khôi phục lại trạng thái staged khi apply.
	indexCommit, err := writeCommit(r, indexTree, []object.Hash{head}, stashIndexMessage+"\n")
	if err != nil {
		return object.ZeroHash, err
	}
	// Chèn commit index vào vị trí thứ hai nếu chưa có phụ huynh nào khác.
	parents = insertParent(parents, indexCommit, 1)

	// Commit stash cũ trở thành phụ huynh tiếp theo của chuỗi stash.
	oldStash, _ := r.Refs.Resolve(stashRef)
	stashParents := parents
	if !oldStash.IsZero() {
		stashParents = append(stashParents, oldStash)
	}

	id := r.Identity()
	// Tree của commit stash là trạng thái worktree đầy đủ, để khi apply
	// thì khôi phục được cả phần đã stage lẫn chưa stage.
	c := &object.Commit{
		Tree:      workTree,
		Parents:   stashParents,
		Author:    id,
		Committer: id,
		Message:   msg,
	}
	h, err := r.Objects.WriteCommit(c)
	if err != nil {
		return object.ZeroHash, err
	}
	if err := r.Refs.Write(stashRef, h); err != nil {
		return object.ZeroHash, err
	}
	_ = r.Refs.AppendReflog(stashRef, oldStash, h, "stash: "+msg)

	// Đưa worktree về trạng thái sạch theo HEAD.
	if head.IsZero() {
		r.Index.Clear()
		if err := r.SaveIndex(); err != nil {
			return object.ZeroHash, err
		}
	} else {
		if err := r.ResetWorktreeTo(headTree); err != nil {
			return object.ZeroHash, err
		}
	}
	// Xóa file chưa theo dõi nếu đã lưu vào stash.
	if includeUntracked {
		if err := RemoveUntrackedFiles(r); err != nil {
			return object.ZeroHash, err
		}
	}
	// Ghi lại reflog HEAD vì worktree đã thay đổi.
	_ = r.AppendReflog("HEAD", head, head, "stash: "+msg)
	return h, nil
}

// untrackedTree dựng tree chỉ chứa các file chưa được theo dõi.
func untrackedTree(r *repo.Repo) (object.Hash, error) {
	st, err := r.Status()
	if err != nil {
		return object.ZeroHash, err
	}
	var nodes []repo.TreeNode
	// Stage tạm để tạo blob cho từng file chưa theo dõi.
	for _, e := range st.Untracked() {
		if err := r.StageFile(e.Path); err != nil {
			return object.ZeroHash, err
		}
		if idx := r.Index.Get(e.Path); idx != nil {
			nodes = append(nodes, repo.IndexTreeNode(*idx))
		}
	}
	if len(nodes) == 0 {
		return object.ZeroHash, nil
	}
	// Người gọi chịu trách nhiệm nạp lại index sau khi dựng xong cây này.
	return r.WriteTree(nodes, "")
}

// insertParent chèn một phụ huynh vào danh sách tại vị trí đã cho,
// bổ sung phần tử nếu danh sách chưa đủ dài.
func insertParent(parents []object.Hash, h object.Hash, pos int) []object.Hash {
	if h.IsZero() {
		return parents
	}
	for len(parents) < pos {
		parents = append(parents, object.ZeroHash)
	}
	parents = append(parents, object.ZeroHash)
	copy(parents[pos+1:], parents[pos:])
	parents[pos] = h
	return parents
}

// stashMessage dựng message cho entry stash.
func stashMessage(message string, st *repo.Status) string {
	if message != "" {
		return "On stash: " + message
	}
	if st.Branch != "" {
		return fmt.Sprintf("WIP trên %s", st.Branch)
	}
	return "WIP trên HEAD tách rời"
}

// StashList liệt kê các entry stash đã lưu, mới nhất đứng đầu.
// Chuỗi stash được nối qua phụ huynh: mỗi entry stash giữ một entry cũ hơn,
// nên phải bỏ qua các commit phụ trợ (commit index và commit file chưa theo dõi).
func StashList(r *repo.Repo) ([]StashEntry, error) {
	h, err := r.Refs.Resolve(stashRef)
	if err != nil {
		return nil, nil // chưa có stash nào
	}
	var out []StashEntry
	seen := map[object.Hash]bool{}
	for !h.IsZero() && !seen[h] {
		seen[h] = true
		c, err := r.Objects.ReadCommit(h)
		if err != nil {
			break
		}
		out = append(out, StashEntry{Index: len(out), Hash: h, Message: c.Summary()})
		// Bỏ qua commit phụ trợ để tới entry stash liền trước.
		next := object.ZeroHash
		for _, p := range c.Parents[1:] {
			pc, err := r.Objects.ReadCommit(p)
			if err != nil {
				continue
			}
			if isStashHelperCommit(pc.Summary()) {
				continue
			}
			next = p
			break
		}
		h = next
	}
	return out, nil
}

// isStashHelperCommit nhận diện các commit phụ trợ do td tạo ra
// khi lưu stash, chúng không phải entry stash mà người dùng thấy.
func isStashHelperCommit(summary string) bool {
	switch summary {
	case stashIndexMessage, stashUntrackedMessage:
		return true
	default:
		return false
	}
}

// StashApply áp dụng lại một entry stash mà không xóa nó khỏi danh sách.
func StashApply(r *repo.Repo, which int) error {
	entry, err := resolveStashEntry(r, which)
	if err != nil {
		return err
	}
	head, err := r.Head()
	if err != nil {
		return err
	}
	stashTree, err := r.CommitTree(entry.Hash)
	if err != nil {
		return err
	}
	headTree := object.ZeroHash
	if !head.IsZero() {
		headTree, err = r.CommitTree(head)
		if err != nil {
			return err
		}
	}
	// Hợp nhất nội dung stash vào trạng thái hiện tại.
	if _, err := mergeTreesInto(r, headTree, headTree, stashTree, "stash"); err != nil {
		return err
	}
	if err := r.SaveIndex(); err != nil {
		return err
	}
	// Khôi phục lại file chưa theo dõi nếu entry có lưu chúng.
	if err := RestoreUntrackedFromStash(r, entry.Hash); err != nil {
		return err
	}
	return nil
}

// StashPop áp dụng một entry stash rồi xóa nó khỏi danh sách.
func StashPop(r *repo.Repo, which int) error {
	if _, err := resolveStashEntry(r, which); err != nil {
		return err
	}
	if err := StashApply(r, which); err != nil {
		return err
	}
	return StashDrop(r, which)
}

// StashDrop xóa một entry stash.
func StashDrop(r *repo.Repo, which int) error {
	entries, err := StashList(r)
	if err != nil {
		return err
	}
	if which < 0 || which >= len(entries) {
		return fmt.Errorf("không có stash ở vị trí %d", which)
	}
	remaining := make([]StashEntry, 0, len(entries)-1)
	for i, e := range entries {
		if i == which {
			continue
		}
		remaining = append(remaining, e)
	}
	if len(remaining) == 0 {
		return r.Refs.Remove(stashRef)
	}
	return r.Refs.Write(stashRef, remaining[0].Hash)
}

// StashClear xóa toàn bộ stash.
func StashClear(r *repo.Repo) error {
	return r.Refs.Remove(stashRef)
}

// resolveStashEntry tìm entry stash theo vị trí.
func resolveStashEntry(r *repo.Repo, which int) (StashEntry, error) {
	entries, err := StashList(r)
	if err != nil {
		return StashEntry{}, err
	}
	if len(entries) == 0 {
		return StashEntry{}, fmt.Errorf("chưa có stash nào")
	}
	// Vị trí âm được hiểu là tính từ cuối danh sách.
	if which < 0 {
		which = len(entries) + which
	}
	if which < 0 || which >= len(entries) {
		return StashEntry{}, fmt.Errorf("không có stash ở vị trí %d (hiện có %d)", which, len(entries))
	}
	return entries[which], nil
}
