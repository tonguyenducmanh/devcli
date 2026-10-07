package ops

import (
	"fmt"
	"sort"
	"strings"

	"github.com/tonguyenducmanh/devcli/internal/vcs/diff"
	"github.com/tonguyenducmanh/devcli/internal/vcs/object"
	"github.com/tonguyenducmanh/devcli/internal/vcs/repo"
)

// DiffOptions điều khiển lệnh diff.
type DiffOptions struct {
	// Revision là commit hoặc phạm vi để so sánh.
	Revision string
	// Staged so sánh index với HEAD.
	Staged bool
	// Paths lọc theo đường dẫn.
	Paths []string
	// Stat hiển thị thống kê thay vì nội dung diff.
	Stat bool
	// NameOnly chỉ hiển thị tên file.
	NameOnly bool
	// Context là số dòng ngữ cảnh quanh thay đổi.
	Context int
}

// FileDiff là kết quả diff của một file.
type FileDiff struct {
	Path   string
	Status byte
	Binary bool
	Old    []byte
	New    []byte
	Lines  []string
	Added  int
	Del    int
}

// Diff tính khác biệt theo nhiều chế độ khác nhau.
func Diff(r *repo.Repo, opts DiffOptions) ([]FileDiff, error) {
	if opts.Context == 0 {
		opts.Context = 3
	}

	var contents map[string]*repo.ContentDiff
	var err error
	switch {
	case opts.Staged:
		// So sánh HEAD với index.
		head, herr := r.Head()
		if herr != nil {
			return nil, herr
		}
		headTree := object.ZeroHash
		if !head.IsZero() {
			headTree, err = r.CommitTree(head)
			if err != nil {
				return nil, err
			}
		}
		contents, err = r.IndexVsTree(headTree)
		if err != nil {
			return nil, err
		}
	case opts.Revision != "":
		contents, err = revisionDiff(r, opts.Revision)
		if err != nil {
			return nil, err
		}
	default:
		// So sánh index với worktree.
		contents, err = r.WorktreeVsIndex()
		if err != nil {
			return nil, err
		}
	}

	paths := make([]string, 0, len(contents))
	for p := range contents {
		if len(opts.Paths) > 0 && !matchesAnyPath(p, opts.Paths) {
			continue
		}
		paths = append(paths, p)
	}
	sort.Strings(paths)

	out := make([]FileDiff, 0, len(paths))
	for _, p := range paths {
		c := contents[p]
		fd := FileDiff{Path: p, Status: c.Status, Binary: c.Binary, Old: c.OldData, New: c.NewData}
		if !c.Binary {
			fd.Added, fd.Del = diff.Stat(c.OldLines, c.NewLines)
			fd.Lines = buildDiffLines(c.OldLines, c.NewLines, opts.Context)
		}
		out = append(out, fd)
	}
	return out, nil
}

// revisionDiff xử lý các dạng so sánh: "A..B", "A...B" hoặc một commit đơn lẻ.
func revisionDiff(r *repo.Repo, rev string) (map[string]*repo.ContentDiff, error) {
	// Dạng "A..B": so sánh hai commit trực tiếp.
	if i := strings.Index(rev, ".."); i >= 0 {
		left := rev[:i]
		right := rev[i+2:]
		dotdot := !strings.HasPrefix(rev[i:], "...")
		// "A...B" dùng merge-base làm phía bên trái.
		leftHash, err := r.ResolveRev(left)
		if err != nil {
			return nil, err
		}
		rightHash, err := r.ResolveRev(right)
		if err != nil {
			return nil, err
		}
		if !dotdot {
			base, err := r.MergeBase(leftHash, rightHash)
			if err != nil {
				return nil, err
			}
			leftHash = base
		}
		l, err := r.CommitTree(leftHash)
		if err != nil {
			return nil, err
		}
		rr, err := r.CommitTree(rightHash)
		if err != nil {
			return nil, err
		}
		return r.DiffTreesContent(l, rr)
	}

	// Một commit đơn lẻ: so sánh với phụ huynh của nó.
	h, err := r.ResolveRev(rev)
	if err != nil {
		return nil, err
	}
	c, err := r.Objects.ReadCommit(h)
	if err != nil {
		return nil, err
	}
	base := object.ZeroHash
	if len(c.Parents) > 0 {
		base, err = r.CommitTree(c.Parents[0])
		if err != nil {
			return nil, err
		}
	}
	return r.DiffTreesContent(base, c.Tree)
}

// buildDiffLines dựng các dòng diff có tiền tố cho hiển thị.
func buildDiffLines(oldLines, newLines []string, context int) []string {
	hunks := diff.HunksWithContext(oldLines, newLines, context)
	var out []string
	for _, h := range hunks {
		out = append(out, fmt.Sprintf("@@ -%d,%d +%d,%d @@", h.AStart+1, h.ACount, h.BStart+1, h.BCount))
		out = append(out, h.Lines...)
	}
	return out
}
