package ops

import (
	"fmt"
	"sort"
	"strings"

	"github.com/tonguyenducmanh/devcli/internal/vcs/index"
	"github.com/tonguyenducmanh/devcli/internal/vcs/merge"
	"github.com/tonguyenducmanh/devcli/internal/vcs/object"
	"github.com/tonguyenducmanh/devcli/internal/vcs/repo"
	"github.com/tonguyenducmanh/devcli/internal/vcs/worktree"
)

// Tên file trạng thái cho các thao tác nhiều bước.
const (
	MergeHead      = "MERGE_HEAD"
	MergeMsg       = "MERGE_MSG"
	MergeBranch    = "MERGE_BRANCH"
	RebaseDir      = "rebase-apply"
	CherryPickHead = "CHERRY_PICK_HEAD"
	RevertHead     = "REVERT_HEAD"
	RevertCommit   = "REVERT_COMMIT"
)

// MergeOptions điều khiển lệnh merge.
type MergeOptions struct {
	// Branch là nhánh hoặc commit cần hợp nhất vào nhánh hiện tại.
	Branch string
	// NoFF buộc tạo commit merge ngay cả khi fast-forward được.
	NoFF bool
	// FFOnly chỉ cho phép fast-forward.
	FFOnly bool
	// Message ghi đè message của commit merge.
	Message string
	// Abort huỷ một lần merge đang dở dang.
	Abort bool
	// Continue hoàn tất merge sau khi đã giải quyết xung đột.
	Continue bool
	// Squash gộp thay đổi vào index mà không tạo commit merge.
	Squash bool
}

// MergeResult mô tả kết quả của lệnh merge.
type MergeResult struct {
	// FastForward true nghĩa là chỉ cần di chuyển con trỏ nhánh.
	FastForward bool
	// MergeCommit là hash commit mới tạo ra (nếu có).
	MergeCommit object.Hash
	// Conflicts là danh sách file xung đột.
	Conflicts []string
	// AlreadyUpToDate báo repo đã ở đúng trạng thái mong muốn.
	AlreadyUpToDate bool
}

// Merge hợp nhất một nhánh vào nhánh hiện tại.
func Merge(r *repo.Repo, opts MergeOptions) (*MergeResult, error) {
	if opts.Abort {
		return abortMerge(r)
	}
	if opts.Continue {
		return continueMerge(r)
	}
	if r.HasState(MergeHead) {
		return nil, fmt.Errorf("đang có một lần merge chưa hoàn tất, chạy `td vcs merge --continue` hoặc `--abort`")
	}

	target, err := resolveCommitish(r, opts.Branch)
	if err != nil {
		return nil, err
	}
	head, err := r.Head()
	if err != nil {
		return nil, err
	}
	if head.IsZero() {
		// Chưa có lịch sử: chỉ cần trỏ nhánh hiện tại tới đích.
		if err := r.UpdateHeadCommit(target, "merge: "+opts.Branch); err != nil {
			return nil, err
		}
		return &MergeResult{FastForward: true}, nil
	}

	// Fast-forward: đích là hậu duệ trực tiếp của HEAD.
	isFF, err := r.IsAncestor(head, target)
	if err != nil {
		return nil, err
	}
	if isFF && !opts.NoFF && !opts.Squash {
		tree, err := r.CommitTree(target)
		if err != nil {
			return nil, err
		}
		if err := ensureCleanForSwitch(r, opts.Branch); err != nil {
			return nil, err
		}
		if err := r.ResetWorktreeTo(tree); err != nil {
			return nil, err
		}
		if err := r.UpdateHeadCommit(target, "merge --ff: "+opts.Branch); err != nil {
			return nil, err
		}
		return &MergeResult{FastForward: true}, nil
	}
	if opts.FFOnly {
		return nil, fmt.Errorf("không thể fast-forward sang %s, đã hủy", opts.Branch)
	}

	// Nhánh đã chứa HEAD: không cần làm gì.
	if isTargetReachable(r, head, target) {
		return &MergeResult{AlreadyUpToDate: true}, nil
	}

	// Tính base chung để hợp nhất ba phía.
	baseCommit, err := r.MergeBase(head, target)
	if err != nil {
		return nil, err
	}
	// Hàm hợp nhất làm việc trên cây nội dung, cần chuyển commit thành tree.
	base, err := r.CommitTree(baseCommit)
	if err != nil {
		return nil, err
	}
	headTree, err := r.CommitTree(head)
	if err != nil {
		return nil, err
	}
	targetTree, err := r.CommitTree(target)
	if err != nil {
		return nil, err
	}
	conflicts, err := mergeTreesInto(r, base, headTree, targetTree, opts.Branch)
	if err != nil {
		return nil, err
	}

	if opts.Squash {
		// Squash: index đã chứa kết quả hợp nhất, người dùng tự commit.
		_ = r.SaveIndex()
		if len(conflicts) > 0 {
			_ = r.WriteState(MergeMsg, "squash merge")
			return &MergeResult{Conflicts: conflicts}, nil
		}
		return &MergeResult{}, nil
	}

	// Ghi trạng thái merge để có thể --abort hoặc --continue.
	if err := r.WriteState(MergeHead, target.String()+"\n"); err != nil {
		return nil, err
	}
	msg := opts.Message
	if msg == "" {
		msg = fmt.Sprintf("Hợp nhất nhánh '%s' vào %s", opts.Branch, currentName(r))
	}
	if err := r.WriteState(MergeMsg, msg+"\n"); err != nil {
		return nil, err
	}
	_ = r.WriteState(MergeBranch, opts.Branch+"\n")

	if len(conflicts) > 0 {
		_ = r.SaveIndex()
		return &MergeResult{Conflicts: conflicts}, nil
	}

	// Không có xung đột: tạo commit merge ngay.
	tree, err := r.TreeFromIndex()
	if err != nil {
		return nil, err
	}
	h, err := writeCommit(r, tree, []object.Hash{head, target}, msg)
	if err != nil {
		return nil, err
	}
	if err := r.UpdateHeadCommit(h, "merge: "+opts.Branch); err != nil {
		return nil, err
	}
	if err := r.SaveIndex(); err != nil {
		return nil, err
	}
	_ = r.ClearState(MergeHead)
	_ = r.ClearState(MergeMsg)
	_ = r.ClearState(MergeBranch)
	return &MergeResult{MergeCommit: h}, nil
}

// currentName trả về tên nhánh hiện tại để ghép vào message.
func currentName(r *repo.Repo) string {
	b, _ := r.CurrentBranch()
	if b == "" {
		return "HEAD detached"
	}
	return b
}

// isTargetReachable báo xem target đã nằm trong lịch sử của head chưa.
func isTargetReachable(r *repo.Repo, target, head object.Hash) bool {
	ok, err := r.IsAncestor(target, head)
	if err != nil {
		return false
	}
	return ok
}

// mergeTreesInto hợp nhất nội dung base/ours/theirs và cập nhật index + worktree.
// Trả về danh sách file xung đột.
func mergeTreesInto(r *repo.Repo, base, ours, theirs object.Hash, label string) ([]string, error) {
	baseNodes, err := r.Flatten(base)
	if err != nil {
		return nil, err
	}
	ourNodes, err := r.Flatten(ours)
	if err != nil {
		return nil, err
	}
	theirNodes, err := r.Flatten(theirs)
	if err != nil {
		return nil, err
	}

	allPaths := map[string]bool{}
	for p := range baseNodes {
		allPaths[p] = true
	}
	for p := range ourNodes {
		allPaths[p] = true
	}
	for p := range theirNodes {
		allPaths[p] = true
	}
	paths := make([]string, 0, len(allPaths))
	for p := range allPaths {
		paths = append(paths, p)
	}
	sort.Strings(paths)

	var conflicts []string
	// Chuyển index về trạng thái ours trước khi áp dụng kết quả hợp nhất.
	r.Index.Clear()
	if !ours.IsZero() {
		for _, p := range paths {
			if n, ok := ourNodes[p]; ok {
				r.Index.Add(indexEntry(n))
			}
		}
	}

	for _, p := range paths {
		b, hasB := baseNodes[p]
		o, hasO := ourNodes[p]
		t, hasT := theirNodes[p]

		// Trường hợp dễ: hai phía giống nhau hoặc một phía không đổi.
		switch {
		case hasO && hasT && o.Hash == t.Hash && o.Mode == t.Mode:
			// Nội dung đã giống nhau, giữ nguyên.
			continue
		case !hasO && !hasT:
			// Cả hai cùng xóa file.
			r.Index.Remove(p)
			_ = worktree.RemoveFile(r.WorkPath(p), r.Root)
			continue
		case hasO && !hasT:
			// Bên kia đã xóa file. Nếu phía chúng ta cũng không sửa gì
			// thì chấp nhận việc xóa, ngược lại mới là xung đột.
			if !hasB {
				continue
			}
			if b.Hash == o.Hash && b.Mode == o.Mode {
				r.Index.Remove(p)
				if err := worktree.RemoveFile(r.WorkPath(p), r.Root); err != nil {
					return nil, fmt.Errorf("xoá %s: %w", p, err)
				}
				continue
			}
			r.Index.Remove(p)
			conflicts = append(conflicts, p)
			continue
		case !hasO && hasT:
			// Bên kia thêm file mới, đưa vào index.
			r.Index.Add(indexEntry(t))
			if err := writeBlobToWorktree(r, t); err != nil {
				return nil, err
			}
			continue
		}

		// Cả hai cùng có file, so sánh với base.
		hashB := object.ZeroHash
		modeB := object.ModeBlob
		if hasB {
			hashB, modeB = b.Hash, b.Mode
		}
		if hashB == o.Hash && modeB == o.Mode {
			// Phía chúng ta không đổi, lấy bên kia.
			r.Index.Remove(p)
			r.Index.RemoveStages(p)
			r.Index.Add(indexEntry(t))
			if err := writeBlobToWorktree(r, t); err != nil {
				return nil, err
			}
			continue
		}
		if hashB == t.Hash && modeB == t.Mode {
			// Bên kia không đổi, giữ bản chúng ta.
			continue
		}

		// Cả hai cùng sửa: thử hợp nhất nội dung ba phía.
		merged, mode, conflict, err := mergeFileContent(r, hashB, o, t, label)
		if err != nil {
			return nil, err
		}
		h, err := r.Objects.WriteBlob(merged)
		if err != nil {
			return nil, err
		}
		if conflict {
			// Ghi lại các stage để người dùng giải quyết được.
			r.Index.Remove(p)
			if hasB {
				r.Index.Add(index.Entry{Mode: modeB, Hash: hashB, Name: p, Stage: index.StageBase})
			}
			r.Index.Add(index.Entry{Mode: o.Mode, Hash: o.Hash, Name: p, Stage: index.StageOurs})
			r.Index.Add(index.Entry{Mode: t.Mode, Hash: t.Hash, Name: p, Stage: index.StageTheirs})
			conflicts = append(conflicts, p)
		} else {
			r.Index.Remove(p)
			r.Index.RemoveStages(p)
			r.Index.Add(index.Entry{Mode: mode, Hash: h, Name: p})
		}
		// Ghi kết quả xuống worktree kể cả khi còn xung đột,
		// để người dùng thấy dấu báo xung đột.
		if err := worktree.WriteFileSymlinkAware(r.WorkPath(p), mode, merged); err != nil {
			return nil, fmt.Errorf("ghi %s: %w", p, err)
		}
	}
	return conflicts, nil
}

// mergeFileContent hợp nhất nội dung một file từ ba phía.
func mergeFileContent(r *repo.Repo, baseHash object.Hash, ours, theirs repo.TreeNode, label string) ([]byte, object.FileMode, bool, error) {
	// File nhị phân hoặc liên kết tượng trưng không hợp nhất nội dung được.
	if ours.Mode.IsExec() != theirs.Mode.IsExec() {
		return nil, object.ModeBlob, true, nil
	}
	if ours.Mode == object.ModeSymlink || theirs.Mode == object.ModeSymlink {
		// Liên kết tượng trưng: chỉ nhận nếu hai bên giống nhau.
		if ours.Hash == theirs.Hash {
			return nil, ours.Mode, false, nil
		}
		return nil, object.ModeBlob, true, nil
	}

	readLines := func(h object.Hash) ([]string, bool, error) {
		if h.IsZero() {
			return nil, false, nil
		}
		data, err := r.Objects.ReadBlob(h)
		if err != nil {
			return nil, false, err
		}
		if isBinaryData(data) {
			return nil, true, nil
		}
		return splitData(string(data)), false, nil
	}

	baseLines, baseBin, err := readLines(baseHash)
	if err != nil {
		return nil, "", false, err
	}
	ourLines, ourBin, err := readLines(ours.Hash)
	if err != nil {
		return nil, "", false, err
	}
	theirLines, theirBin, err := readLines(theirs.Hash)
	if err != nil {
		return nil, "", false, err
	}
	if baseBin || ourBin || theirBin {
		// File nhị phân: chỉ hợp nhất được khi hai phía giống nhau.
		if ours.Hash == theirs.Hash {
			return nil, ours.Mode, false, nil
		}
		return nil, object.ModeBlob, true, nil
	}

	ourLabel := "HEAD"
	if b, _ := r.CurrentBranch(); b != "" {
		ourLabel = b
	}
	res := merge.Merge3(baseLines, ourLines, theirLines, ourLabel, label)
	out := joinData(res.Lines)
	return out, ours.Mode, res.Conflict, nil
}

// writeBlobToWorktree ghi nội dung blob xuống đĩa.
func writeBlobToWorktree(r *repo.Repo, n repo.TreeNode) error {
	data, err := r.Objects.ReadBlob(n.Hash)
	if err != nil {
		return err
	}
	return worktree.WriteFileSymlinkAware(r.WorkPath(n.Path), n.Mode, data)
}

// abortMerge huỷ một lần merge đang dở dang bằng cách đưa về HEAD.
func abortMerge(r *repo.Repo) (*MergeResult, error) {
	if !r.HasState(MergeHead) {
		return nil, fmt.Errorf("không có lần merge nào đang dở dang")
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
	// Đặt cả index và worktree về trạng thái trước khi bắt đầu merge.
	if err := r.ResetWorktreeTo(tree); err != nil {
		return nil, err
	}
	if err := clearMergeState(r); err != nil {
		return nil, err
	}
	return &MergeResult{}, nil
}

// continueMerge hoàn tất merge sau khi đã giải quyết xung đột.
func continueMerge(r *repo.Repo) (*MergeResult, error) {
	if !r.HasState(MergeHead) {
		return nil, fmt.Errorf("không có lần merge nào đang dở dang")
	}
	raw, err := r.ReadState(MergeHead)
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
	if r.Index.HasConflicts() {
		return nil, fmt.Errorf("vẫn còn %d file xung đột", len(r.Index.Conflicts()))
	}

	msg, _ := r.ReadState(MergeMsg)
	if msg == "" {
		msg = "Merge"
	}
	tree, err := r.TreeFromIndex()
	if err != nil {
		return nil, err
	}
	h, err := writeCommit(r, tree, []object.Hash{head, target}, msg)
	if err != nil {
		return nil, err
	}
	if err := r.UpdateHeadCommit(h, "merge (tiếp tục)"); err != nil {
		return nil, err
	}
	if err := r.SaveIndex(); err != nil {
		return nil, err
	}
	if err := clearMergeState(r); err != nil {
		return nil, err
	}
	return &MergeResult{MergeCommit: h}, nil
}

// clearMergeState xóa toàn bộ file trạng thái của merge.
func clearMergeState(r *repo.Repo) error {
	for _, name := range []string{MergeHead, MergeMsg, MergeBranch} {
		if err := r.ClearState(name); err != nil {
			return err
		}
	}
	return nil
}

// MergeInProgress báo xem đang có một lần merge chưa hoàn tất hay không.
func MergeInProgress(r *repo.Repo) bool { return r.HasState(MergeHead) }

// isBinaryData báo xem dữ liệu có phải file nhị phân hay không.
func isBinaryData(data []byte) bool {
	limit := len(data)
	if limit > 8000 {
		limit = 8000
	}
	for i := 0; i < limit; i++ {
		if data[i] == 0 {
			return true
		}
	}
	return false
}

// splitData tách nội dung thành các dòng.
func splitData(s string) []string {
	if s == "" {
		return nil
	}
	s = strings.ReplaceAll(s, "\r\n", "\n")
	if strings.HasSuffix(s, "\n") {
		s = s[:len(s)-1]
	}
	return strings.Split(s, "\n")
}

// joinData nối các dòng thành nội dung, thêm xuống dòng cuối.
func joinData(lines []string) []byte {
	if len(lines) == 0 {
		return nil
	}
	return []byte(strings.Join(lines, "\n") + "\n")
}
