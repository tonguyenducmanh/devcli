package ops

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tonguyenducmanh/devcli/internal/vcs/object"
	"github.com/tonguyenducmanh/devcli/internal/vcs/repo"
)

// RebaseOptions điều khiển lệnh rebase.
type RebaseOptions struct {
	// Upstream là điểm gốc để rebase lên, ví dụ "main" hoặc "abc123".
	Upstream string
	// Onto ghi đè điểm đích, khác với điểm gốc Upstream.
	Onto string
	// Branch là nhánh cần rebase, mặc định là nhánh hiện tại.
	Branch string
	// Continue tiếp tục sau khi đã giải quyết xung đột.
	Continue bool
	// Abort huỷ rebase đang dở dang.
	Abort bool
	// Skip bỏ qua commit đang gây xung đột.
	Skip bool
}

// RebaseResult là kết quả của một lần rebase.
type RebaseResult struct {
	// Applied là số commit đã áp dụng thành công.
	Applied int
	// Commits là các hash commit mới tạo ra.
	Commits []object.Hash
	// Conflicts là danh sách file xung đột.
	Conflicts []string
	// Skipped là số commit bị bỏ qua.
	Skipped int
	// Done báo rebase đã hoàn tất.
	Done bool
}

// Rebase di chuyển các commit của nhánh hiện tại nằm phía trên một điểm khác.
// Mỗi commit được áp dụng lại bằng hợp nhất ba phía nên vẫn giữ được lịch sử sạch.
func Rebase(r *repo.Repo, opts RebaseOptions) (*RebaseResult, error) {
	switch {
	case opts.Abort:
		return rebaseAbort(r)
	case opts.Continue:
		return rebaseContinue(r)
	case opts.Skip:
		return rebaseSkip(r)
	}

	// Khi rebase một nhánh khác thì nhánh đó sẽ là nơi nhận commit mới.
	targetBranch := opts.Branch
	if targetBranch == "" {
		targetBranch, _ = r.CurrentBranch()
	}

	// Khi không chỉ định nhánh nào thì rebase nhánh đang đứng lên chính nó
	// theo một điểm gốc khác. Nếu chỉ định nhánh thì bắt buộc có điểm gốc,
	// tránh rebase một nhánh lên chính nó một cách vô nghĩa.
	upstreamName := opts.Upstream
	if upstreamName == "" && targetBranch == "" {
		return nil, fmt.Errorf("thiếu đối số: cần chỉ định nhánh hoặc commit đích để rebase lên")
	}
	if upstreamName == "" {
		return nil, fmt.Errorf("cần chỉ định điểm đích để rebase nhánh %s", targetBranch)
	}
	upstream, err := resolveCommitish(r, upstreamName)
	if err != nil {
		return nil, err
	}

	// Điểm đích thực sự: --onto được ưu tiên, nếu không thì chính là upstream.
	onto := upstream
	ontoName := upstreamName
	if opts.Onto != "" {
		onto, err = resolveCommitish(r, opts.Onto)
		if err != nil {
			return nil, err
		}
		ontoName = opts.Onto
	}

	// Đầu nguồn của rebase là commit đầu của nhánh được chỉ định,
	// hoặc HEAD khi không chỉ định nhánh nào.
	head := object.ZeroHash
	if opts.Branch != "" {
		if !r.BranchExists(opts.Branch) {
			return nil, fmt.Errorf("không tìm thấy nhánh %s", opts.Branch)
		}
		head, err = r.BranchHash(opts.Branch)
		if err != nil {
			return nil, err
		}
	} else {
		head, err = r.Head()
		if err != nil {
			return nil, err
		}
	}
	if head.IsZero() {
		return nil, fmt.Errorf("chưa có commit nào để rebase")
	}

	// Danh sách commit cần di chuyển: commit có trong nhánh nguồn nhưng
	// không có trong lịch sử của điểm gốc.
	toReplay, err := commitsToReplay(r, head, upstream)
	if err != nil {
		return nil, err
	}
	if len(toReplay) == 0 {
		// Không có gì cần di chuyển, chỉ cần trỏ nhánh sang điểm đích.
		if onto != head {
			if err := r.ResetWorktreeTo(onto); err != nil {
				return nil, err
			}
			if err := r.UpdateHeadCommit(onto, "rebase: không có gì để di chuyển"); err != nil {
				return nil, err
			}
		}
		return &RebaseResult{Done: true}, nil
	}

	if err := ensureCleanForSwitch(r, upstreamName); err != nil {
		return nil, err
	}

	// Lưu trạng thái để hỗ trợ --continue, --abort, --skip.
	// Nhánh đang đứng trước khi rebase cũng phải được ghi lại,
	// vì rebase có thể chuyển sang nhánh khác rồi phải quay về.
	if _, err := r.StateDir(RebaseDir); err != nil {
		return nil, err
	}
	origBranch, _ := r.CurrentBranch()
	origHead, _ := r.Head()
	state := rebaseState{
		Upstream:      upstream.String(),
		Onto:          onto.String(),
		OrigHead:      origHead.String(),
		OrigBranch:    origBranch,
		Commits:       hashesToStrings(toReplay),
		Branch:        targetBranch,
		BranchOrigTip: head.String(),
		OntoName:      ontoName,
	}
	if err := saveRebaseState(r, state); err != nil {
		return nil, err
	}
	return runRebase(r, targetBranch)
}

// rebaseState là thông tin lưu trong thư mục trạng thái rebase.
type rebaseState struct {
	Upstream      string   // commit gốc dùng để tính danh sách commit cần di chuyển
	Onto          string   // commit đích sau rebase
	OrigHead      string   // HEAD trước khi bắt đầu rebase
	OrigBranch    string   // nhánh đang đứng trước khi rebase
	Commits       []string // danh sách commit cần áp dụng lại, theo thứ tự cũ đến mới
	Current       int      // vị trí commit đang xử lý
	Branch        string   // nhánh nhận commit mới
	BranchOrigTip string   // vị trí nhánh nhận commit mới trước khi rebase
	OntoName      string   // tên hiển thị của điểm đích
}

// commitsToReplay lấy danh sách commit thuộc head nhưng không thuộc upstream,
// theo thứ tự từ cũ đến mới.
func commitsToReplay(r *repo.Repo, head, upstream object.Hash) ([]object.Hash, error) {
	// Gom tất cả commit trong lịch sử của head theo thứ tự cũ đến mới
	// bằng cách duyệt hậu tố (đệ quy phụ huynh rồi thêm bản thân).
	var out []object.Hash
	seen := map[object.Hash]bool{}
	var walk func(h object.Hash, depth int)
	walk = func(h object.Hash, depth int) {
		if h.IsZero() || seen[h] || depth > 50000 {
			return
		}
		seen[h] = true
		c, err := r.Objects.ReadCommit(h)
		if err != nil {
			return
		}
		for _, p := range c.Parents {
			walk(p, depth+1)
		}
		out = append(out, h)
	}
	walk(head, 0)

	// Lọc bỏ các commit đã nằm trong lịch sử của điểm gốc.
	baseSet, err := r.AncestorSet(upstream)
	if err != nil {
		return nil, err
	}
	filtered := make([]object.Hash, 0, len(out))
	for _, h := range out {
		if !baseSet[h] {
			filtered = append(filtered, h)
		}
	}
	return filtered, nil
}

// runRebase áp dụng tuần tự các commit còn lại của rebase.
// Nhánh đích được trỏ tới commit mới trước khi áp dụng commit đầu tiên.
func runRebase(r *repo.Repo, targetBranch string) (*RebaseResult, error) {
	res := &RebaseResult{}
	state, err := loadRebaseState(r)
	if err != nil {
		return nil, err
	}

	// Nhánh đích: nếu không phải nhánh hiện tại thì chuyển sang nó trước.
	if targetBranch != "" {
		current, _ := r.CurrentBranch()
		if current != targetBranch && r.BranchExists(targetBranch) {
			if err := switchToBranch(r, targetBranch); err != nil {
				return nil, err
			}
		}
	}

	for {
		state, err = loadRebaseState(r)
		if err != nil {
			return res, err
		}
		if state.Current >= len(state.Commits) {
			done, err := finishRebase(r, state)
			if err != nil {
				return res, err
			}
			done.Applied = res.Applied
			done.Commits = res.Commits
			return done, nil
		}
		target, err := object.ParseHash(state.Commits[state.Current])
		if err != nil {
			return res, err
		}

		head, err := r.Head()
		if err != nil {
			return res, err
		}
		headTree := object.ZeroHash
		if !head.IsZero() {
			headTree, err = r.CommitTree(head)
			if err != nil {
				return res, err
			}
		}
		src, err := r.Objects.ReadCommit(target)
		if err != nil {
			return res, err
		}
		base := object.ZeroHash
		if len(src.Parents) > 0 {
			base, err = r.CommitTree(src.Parents[0])
			if err != nil {
				return res, err
			}
		}

		// Trước khi áp dụng commit đầu tiên, nhánh đích phải được
		// trỏ về đúng điểm đích thì mọi commit sau đều nối tiếp từ đó.
		if state.Current == 0 {
			ontoHash := state2hash(state.Onto)
			if head != ontoHash {
				ontoTree, err := r.CommitTree(ontoHash)
				if err != nil {
					return res, err
				}
				if err := r.ResetWorktreeTo(ontoTree); err != nil {
					return res, err
				}
				if err := r.UpdateHeadCommit(ontoHash, "rebase: bắt đầu"); err != nil {
					return res, err
				}
				head = ontoHash
				headTree = ontoTree
			}
		}

		conflicts, err := mergeTreesInto(r, base, headTree, src.Tree, src.Summary())
		if err != nil {
			return res, err
		}
		if err := r.SaveIndex(); err != nil {
			return res, err
		}
		if len(conflicts) > 0 {
			res.Conflicts = conflicts
			res.Skipped = len(state.Commits) - state.Current
			return res, nil
		}

		// Ghi lại commit với nội dung đã hợp nhất.
		tree, err := r.TreeFromIndex()
		if err != nil {
			return res, err
		}
		id := r.Identity()
		c := &object.Commit{
			Tree:      tree,
			Parents:   []object.Hash{head},
			Author:    id,
			Committer: id,
			Message:   src.Message,
		}
		h, err := r.Objects.WriteCommit(c)
		if err != nil {
			return res, err
		}
		if err := r.UpdateHeadCommit(h, "rebase: "+src.Summary()); err != nil {
			return res, err
		}
		res.Commits = append(res.Commits, h)
		res.Applied++

		state.Current++
		if err := saveRebaseState(r, state); err != nil {
			return res, err
		}
	}
}

// state2hash chuyển chuỗi hash trong trạng thái thành đối tượng hash.
func state2hash(s string) object.Hash {
	h, err := object.ParseHash(s)
	if err != nil {
		return object.ZeroHash
	}
	return h
}

// mustTree trả về tree của một commit trong trạng thái rebase.
func mustTree(r *repo.Repo, commitHash string) object.Hash {
	h, err := object.ParseHash(commitHash)
	if err != nil {
		return object.ZeroHash
	}
	t, err := r.CommitTree(h)
	if err != nil {
		return object.ZeroHash
	}
	return t
}

// finishRebase dọn dẹp trạng thái và đưa người dùng về nhánh ban đầu
// sau khi rebase hoàn tất.
func finishRebase(r *repo.Repo, state rebaseState) (*RebaseResult, error) {
	// Nếu rebase nhánh khác thì trỏ ref của nhánh đó tới commit cuối cùng
	// vừa tạo, rồi quay lại nhánh đang đứng trước khi rebase.
	if state.Branch != "" {
		head, err := r.Head()
		if err != nil {
			return nil, err
		}
		if current, _ := r.CurrentBranch(); current == state.Branch {
			if err := r.UpdateBranch(state.Branch, head, "rebase: cập nhật nhánh "+state.Branch); err != nil {
				return nil, err
			}
		}
	}
	if state.OrigBranch != "" && state.OrigBranch != state.Branch && r.BranchExists(state.OrigBranch) {
		origHash := state2hash(state.OrigHead)
		origTree, err := r.CommitTree(origHash)
		if err == nil {
			if err := r.ResetWorktreeTo(origTree); err != nil {
				return nil, err
			}
		}
		if err := r.UpdateHeadRef("refs/heads/"+state.OrigBranch, "rebase: quay lại "+state.OrigBranch); err != nil {
			return nil, err
		}
	}
	if err := r.ClearState(RebaseDir); err != nil {
		return nil, err
	}
	return &RebaseResult{Done: true}, nil
}

// rebaseContinue tiếp tục rebase sau khi đã giải quyết xung đột.
func rebaseContinue(r *repo.Repo) (*RebaseResult, error) {
	state, err := loadRebaseState(r)
	if err != nil {
		return nil, err
	}
	if r.Index.HasConflicts() {
		return nil, fmt.Errorf("vẫn còn %d file xung đột", len(r.Index.Conflicts()))
	}
	if state.Current >= len(state.Commits) {
		return finishRebase(r, state)
	}

	head, err := r.Head()
	if err != nil {
		return nil, err
	}
	tree, err := r.TreeFromIndex()
	if err != nil {
		return nil, err
	}
	target, err := object.ParseHash(state.Commits[state.Current])
	if err != nil {
		return nil, err
	}
	src, err := r.Objects.ReadCommit(target)
	if err != nil {
		return nil, err
	}
	id := r.Identity()
	c := &object.Commit{
		Tree:      tree,
		Parents:   []object.Hash{head},
		Author:    id,
		Committer: id,
		Message:   src.Message,
	}
	h, err := r.Objects.WriteCommit(c)
	if err != nil {
		return nil, err
	}
	if err := r.UpdateHeadCommit(h, "rebase (tiếp tục): "+src.Summary()); err != nil {
		return nil, err
	}
	state.Current++
	if err := saveRebaseState(r, state); err != nil {
		return nil, err
	}
	// Tiếp tục áp dụng các commit còn lại.
	rest, err := runRebase(r, state.Branch)
	if err != nil {
		return nil, err
	}
	// Commit vừa giải quyết xung đột được tính vào kết quả.
	rest.Applied++
	rest.Commits = append([]object.Hash{h}, rest.Commits...)
	return rest, nil
}

// rebaseSkip bỏ qua commit hiện tại rồi tiếp tục với các commit còn lại.
func rebaseSkip(r *repo.Repo) (*RebaseResult, error) {
	state, err := loadRebaseState(r)
	if err != nil {
		return nil, err
	}
	// Khôi phục worktree về HEAD trước khi bỏ qua.
	head, err := r.Head()
	if err != nil {
		return nil, err
	}
	if headTree, err := r.CommitTree(head); err == nil {
		if err := r.ResetWorktreeTo(headTree); err != nil {
			return nil, err
		}
	}
	state.Current++
	if err := saveRebaseState(r, state); err != nil {
		return nil, err
	}
	res, err := runRebase(r, state.Branch)
	if err != nil {
		return nil, err
	}
	res.Skipped++
	return res, nil
}

// rebaseAbort huỷ rebase và trả mọi nhánh về trạng thái trước khi bắt đầu:
// nhánh đang rebase về vị trí cũ, người dùng quay lại nhánh ban đầu.
func rebaseAbort(r *repo.Repo) (*RebaseResult, error) {
	state, err := loadRebaseState(r)
	if err != nil {
		return nil, err
	}
	// Đưa nhánh đang rebase về đúng vị trí cũ.
	origTip := state2hash(state.BranchOrigTip)
	if state.Branch != "" && !origTip.IsZero() && r.BranchExists(state.Branch) {
		if err := r.UpdateBranch(state.Branch, origTip, "rebase: hủy bỏ nhánh "+state.Branch); err != nil {
			return nil, err
		}
	}
	// Quay về nhánh ban đầu và đặt cây làm việc về trạng thái cũ.
	origHead := state2hash(state.OrigHead)
	if state.OrigBranch != "" && r.BranchExists(state.OrigBranch) {
		tree, err := r.CommitTree(origHead)
		if err == nil {
			if err := r.ResetWorktreeTo(tree); err != nil {
				return nil, err
			}
		}
		if err := r.UpdateHeadRef("refs/heads/"+state.OrigBranch, "rebase: hủy bỏ"); err != nil {
			return nil, err
		}
	} else if !origHead.IsZero() {
		tree, err := r.CommitTree(origHead)
		if err != nil {
			return nil, err
		}
		if err := r.ResetWorktreeTo(tree); err != nil {
			return nil, err
		}
		if err := r.UpdateHeadCommit(origHead, "rebase: hủy bỏ"); err != nil {
			return nil, err
		}
	}
	if err := r.ClearState(RebaseDir); err != nil {
		return nil, err
	}
	return &RebaseResult{Done: true}, nil
}

// RebaseInProgress báo xem rebase đang dở dang hay không.
func RebaseInProgress(r *repo.Repo) bool { return r.HasState(RebaseDir) }

// rebaseStatePath là đường dẫn file lưu trạng thái rebase.
func rebaseStatePath(r *repo.Repo) string { return r.StatePath(filepath.Join(RebaseDir, "state")) }

// loadRebaseState đọc trạng thái rebase từ file.
func loadRebaseState(r *repo.Repo) (rebaseState, error) {
	var st rebaseState
	data, err := os.ReadFile(rebaseStatePath(r))
	if err != nil {
		return st, fmt.Errorf("không có lần rebase nào đang dở dang")
	}
	// Định dạng văn bản đơn giản: mỗi trường một dòng "khoá=giá trị".
	for _, line := range strings.Split(string(data), "\n") {
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		switch key {
		case "upstream":
			st.Upstream = value
		case "onto":
			st.Onto = value
		case "orig_head":
			st.OrigHead = value
		case "orig_branch":
			st.OrigBranch = value
		case "branch":
			st.Branch = value
		case "branch_orig_tip":
			st.BranchOrigTip = value
		case "onto_name":
			st.OntoName = value
		case "current":
			fmt.Sscanf(value, "%d", &st.Current)
		case "commits":
			st.Commits = strings.Fields(value)
		}
	}
	return st, nil
}

// saveRebaseState ghi trạng thái rebase xuống file.
func saveRebaseState(r *repo.Repo, st rebaseState) error {
	if _, err := r.StateDir(RebaseDir); err != nil {
		return err
	}
	var b strings.Builder
	b.WriteString("upstream=" + st.Upstream + "\n")
	b.WriteString("onto=" + st.Onto + "\n")
	b.WriteString("orig_head=" + st.OrigHead + "\n")
	b.WriteString("orig_branch=" + st.OrigBranch + "\n")
	b.WriteString("branch=" + st.Branch + "\n")
	b.WriteString("branch_orig_tip=" + st.BranchOrigTip + "\n")
	b.WriteString("onto_name=" + st.OntoName + "\n")
	b.WriteString(fmt.Sprintf("current=%d\n", st.Current))
	b.WriteString("commits=" + strings.Join(st.Commits, " ") + "\n")
	return os.WriteFile(rebaseStatePath(r), []byte(b.String()), 0o644)
}

// hashesToStrings chuyển danh sách hash sang chuỗi.
func hashesToStrings(hs []object.Hash) []string {
	out := make([]string, 0, len(hs))
	for _, h := range hs {
		out = append(out, h.String())
	}
	return out
}
