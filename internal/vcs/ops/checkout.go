package ops

import (
	"fmt"
	"os"

	"github.com/tonguyenducmanh/devcli/internal/vcs/object"
	"github.com/tonguyenducmanh/devcli/internal/vcs/repo"
	"github.com/tonguyenducmanh/devcli/internal/vcs/worktree"
)

// CheckoutOptions điều khiển lệnh checkout.
type CheckoutOptions struct {
	// Target là nhánh, tag hoặc biểu thức commit cần chuyển tới.
	Target string
	// CreateBranch tạo nhánh mới từ điểm xuất phát rồi chuyển sang đó.
	CreateBranch string
	// Force bỏ qua kiểm tra thay đổi chưa lưu.
	Force bool
	// Detach tách HEAD khỏi nhánh, chỉ dùng cho commit/tag.
	Detach bool
}

// Checkout chuyển HEAD tới nhánh hoặc commit khác.
func Checkout(r *repo.Repo, opts CheckoutOptions) error {
	if opts.CreateBranch != "" {
		return CreateBranch(r, CreateBranchOptions{
			Name:       opts.CreateBranch,
			StartPoint: opts.Target,
			Switch:     true,
		})
	}
	if opts.Target == "" {
		return fmt.Errorf("thiếu đối số: cần chỉ định nhánh hoặc commit cần chuyển tới")
	}

	// Nếu đích là một nhánh cục bộ thì chuyển HEAD theo dạng symbolic.
	if r.BranchExists(opts.Target) && !opts.Detach {
		return switchToBranch(r, opts.Target)
	}

	h, err := resolveCommitish(r, opts.Target)
	if err != nil {
		return err
	}
	tree, err := r.CommitTree(h)
	if err != nil {
		return err
	}
	if !opts.Force {
		if err := ensureCleanForSwitch(r, opts.Target); err != nil {
			return err
		}
	}
	if err := r.ResetWorktreeTo(tree); err != nil {
		return err
	}
	oldHead, _ := r.Head()
	if err := r.SetHeadDetached(h); err != nil {
		return err
	}
	desc := opts.Target
	if _, err := object.ParseHash(h.String()); err == nil {
		desc = h.Short(8)
	}
	return r.AppendReflog("HEAD", oldHead, h, "checkout: chuyển tới "+desc)
}

// RestoreOptions điều khiển lệnh khôi phục file.
type RestoreOptions struct {
	// Source là nơi lấy nội dung: mặc định lấy từ index.
	Source string
	// Staged chỉ cập nhật index mà không đụng worktree.
	Staged bool
	// Worktree chỉ cập nhật worktree.
	Worktree bool
	// SourceCommit khi lấy nội dung từ một commit cụ thể.
	SourceCommit object.Hash
	// Paths là danh sách file cần khôi phục.
	Paths []string
}

// Restore khôi phục file từ index, từ HEAD hoặc từ một commit chỉ định.
func Restore(r *repo.Repo, opts RestoreOptions) error {
	if len(opts.Paths) == 0 {
		return fmt.Errorf("thiếu đường dẫn: cần chỉ định file cần khôi phục")
	}
	// Mặc định lệnh chỉ đụng vào cây làm việc: lấy nội dung đang ở vùng chuẩn
	// bị rồi ghi đè lên đĩa, giống git. Cờ --staged mới đổi vùng chuẩn bị, cờ
	// --worktree thì chỉ định rõ để làm đúng một vùng.
	updateIndex := opts.Staged
	updateWorktree := !opts.Staged

	// Nguồn nội dung: commit chỉ định, HEAD, hoặc index.
	sourceTree := object.ZeroHash
	useCommit := opts.Source != "" || opts.SourceCommit != object.ZeroHash
	// Gỡ thay đổi đã stage tức là đưa vùng chuẩn bị về đúng HEAD, giống git.
	// Lấy từ chính vùng chuẩn bị sẽ khiến lệnh chép lại nội dung đang có nên
	// không gỡ được gì cả.
	if opts.Staged && !opts.Worktree && !useCommit {
		opts.Source = "HEAD"
		useCommit = true
	}
	if useCommit {
		h := opts.SourceCommit
		if h.IsZero() {
			src := opts.Source
			if src == "" {
				src = "HEAD"
			}
			var err error
			h, err = resolveCommitish(r, src)
			if err != nil {
				return err
			}
		}
		// HEAD chưa có commit thì cây rỗng: mọi tệp trong index đều phải
		// được gỡ, đúng như git restore --staged trên nhánh chưa có mốc nào.
		if !h.IsZero() {
			tree, terr := r.CommitTree(h)
			if terr != nil {
				return terr
			}
			sourceTree = tree
		}
	}
	nodes := map[string]repo.TreeNode{}
	if useCommit {
		n, ferr := r.Flatten(sourceTree)
		if ferr != nil {
			return ferr
		}
		nodes = n
	}

	st, err := r.Status()
	if err != nil {
		return err
	}
	for _, p := range opts.Paths {
		matched := false
		for _, e := range st.Entries {
			if !matchPathPattern(e.Path, p) {
				continue
			}
			matched = true
			hash := e.IndexHash
			mode := e.IndexMode
			if useCommit {
				n, ok := nodes[e.Path]
				if !ok {
					// File không tồn tại trong commit nguồn: xóa khỏi index.
					if updateIndex {
						r.Index.Remove(e.Path)
					}
					if updateWorktree {
						_ = worktree.RemoveFile(r.WorkPath(e.Path), r.Root)
					}
					continue
				}
				hash, mode = n.Hash, n.Mode
			}
			if hash.IsZero() {
				continue
			}
			if updateIndex {
				r.Index.Remove(e.Path)
				if err := r.StageFile(e.Path); err != nil {
					return err
				}
				// Ghi đè hash/chế độ nếu lấy từ commit.
				if e2 := r.Index.Get(e.Path); e2 != nil && useCommit {
					e2.Hash = hash
					e2.Mode = mode
				}
			}
			if updateWorktree {
				data, err := r.Objects.ReadBlob(hash)
				if err != nil {
					return err
				}
				if err := worktree.WriteFileSymlinkAware(r.WorkPath(e.Path), mode, data); err != nil {
					return fmt.Errorf("ghi %s: %w", e.Path, err)
				}
			}
		}
		if !matched {
			// Có thể là file chưa từng được theo dõi, kiểm tra trực tiếp trên đĩa.
			if existsInWorktree(r, p) {
				if updateIndex {
					if err := r.StageFile(p); err != nil {
						return err
					}
				}
				matched = true
			}
		}
		if !matched {
			return fmt.Errorf("không khớp với đường dẫn nào được chỉ định: %s", p)
		}
	}
	if updateIndex {
		return r.SaveIndex()
	}
	return nil
}

// ResetOptions điều khiển lệnh reset.
type ResetOptions struct {
	// Target là điểm đích của reset.
	Target string
	// Mode: soft, mixed (mặc định), hard, hoặc merge.
	Mode string
	// Paths giới hạn reset vào các file cụ thể (chỉ dùng với mixed).
	Paths []string
}

// Reset đặt lại HEAD về một commit với các chế độ khác nhau.
func Reset(r *repo.Repo, opts ResetOptions) error {
	if opts.Target == "" {
		opts.Target = "HEAD"
	}
	target, err := resolveCommitish(r, opts.Target)
	if err != nil {
		return err
	}

	// Khi có đường dẫn, chỉ đụng index cho các file đó.
	if len(opts.Paths) > 0 {
		return resetPaths(r, opts.Paths, target)
	}

	tree, err := r.CommitTree(target)
	if err != nil {
		return err
	}
	if _, err := r.Head(); err != nil {
		return err
	}

	switch opts.Mode {
	case "soft":
		// Chỉ di chuyển con trỏ HEAD, giữ nguyên index và worktree.
		if err := r.UpdateHeadCommit(target, "reset --soft: "+opts.Target); err != nil {
			return err
		}
	case "mixed", "":
		// Di chuyển HEAD và nạp lại index, giữ nguyên worktree.
		if err := r.UpdateHeadCommit(target, "reset --mixed: "+opts.Target); err != nil {
			return err
		}
		if err := r.WriteIndexFromTree(tree); err != nil {
			return err
		}
	case "hard":
		// Di chuyển HEAD, đặt cả index và worktree về đích.
		if err := r.UpdateHeadCommit(target, "reset --hard: "+opts.Target); err != nil {
			return err
		}
		if err := r.ResetWorktreeTo(tree); err != nil {
			return err
		}
	default:
		return fmt.Errorf("chế độ reset không hợp lệ: %s", opts.Mode)
	}
	return nil
}

// resetPaths gỡ các file cụ thể khỏi index, đưa về trạng thái của commit đích.
func resetPaths(r *repo.Repo, paths []string, target object.Hash) error {
	tree, err := r.CommitTree(target)
	if err != nil {
		return err
	}
	nodes, err := r.Flatten(tree)
	if err != nil {
		return err
	}
	st, err := r.Status()
	if err != nil {
		return err
	}
	matched := false
	for _, e := range st.Entries {
		if !matchesAnyPath(e.Path, paths) {
			continue
		}
		matched = true
		n, ok := nodes[e.Path]
		if !ok {
			r.Index.Remove(e.Path)
			continue
		}
		r.Index.Remove(e.Path)
		r.Index.Add(indexEntry(n))
	}
	if !matched {
		return fmt.Errorf("không khớp với đường dẫn nào được chỉ định")
	}
	return r.SaveIndex()
}

// Remove xóa file khỏi index và (tuỳ chọn) khỏi worktree.
func Remove(r *repo.Repo, paths []string, removeWorktree bool, cached bool) error {
	st, err := r.Status()
	if err != nil {
		return err
	}
	matched := false
	for _, e := range st.Entries {
		if !matchesAnyPath(e.Path, paths) {
			continue
		}
		matched = true
		if cached {
			r.Index.Remove(e.Path)
			continue
		}
		if removeWorktree {
			r.Index.Remove(e.Path)
			if err := worktree.RemoveFile(r.WorkPath(e.Path), r.Root); err != nil {
				return err
			}
		}
	}
	if !matched {
		return fmt.Errorf("không khớp với đường dẫn nào được chỉ định")
	}
	if !removeWorktree {
		// Chỉ gỡ khỏi index, giữ file trên đĩa.
		for _, p := range paths {
			if _, err := os.Lstat(r.WorkPath(p)); err == nil {
				r.Index.Remove(p)
			}
		}
	}
	return r.SaveIndex()
}

// Move đổi tên hoặc di chuyển file, đồng thời cập nhật index.
func Move(r *repo.Repo, from, to string) error {
	if !existsInWorktree(r, from) {
		return fmt.Errorf("không tìm thấy %s", from)
	}
	if err := os.MkdirAll(dirOf(r.WorkPath(to)), 0o755); err != nil {
		return err
	}
	if err := os.Rename(r.WorkPath(from), r.WorkPath(to)); err != nil {
		return err
	}

	// Nếu file nguồn đang được theo dõi thì cập nhật index theo đường dẫn mới.
	if e := r.Index.Get(from); e != nil {
		r.Index.Remove(from)
		r.Index.Remove(to)
		if err := r.StageFile(to); err != nil {
			return err
		}
	}
	// Nếu nguồn chỉ tồn tại trong HEAD (đã bị xoá trên đĩa) thì cần gỡ ở index.
	r.Index.Remove(from)
	if err := r.StageFile(to); err != nil {
		return err
	}
	return r.SaveIndex()
}
