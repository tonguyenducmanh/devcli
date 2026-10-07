package ops

import (
	"fmt"
	"strings"
	"time"

	"github.com/tonguyenducmanh/devcli/internal/vcs/diff"
	"github.com/tonguyenducmanh/devcli/internal/vcs/object"
	"github.com/tonguyenducmanh/devcli/internal/vcs/repo"
)

// LogOptions điều khiển hiển thị lịch sử commit.
type LogOptions struct {
	// Revision là điểm bắt đầu, mặc định HEAD.
	Revision string
	// Paths lọc commit theo đường dẫn.
	Paths []string
	// Max giới hạn số dòng, 0 = không giới hạn.
	Max int
	// Oneline hiển thị mỗi commit trên một dòng.
	Oneline bool
	// All duyệt lịch sử từ mọi nhánh và tag.
	All bool
	// Patch kèm nội dung diff của từng commit.
	Patch bool
	// Stat kèm thống kê số dòng thay đổi.
	Stat bool
	// ShortHash dùng hash ngắn 7 ký tự thay vì 12.
	ShortHash bool
	// FollowHistory bám theo lịch sử khi lọc theo đường dẫn.
	FollowHistory bool
	// Reverse hiển thị ngược thứ tự thời gian.
	Reverse bool
	// Graph hiển thị sơ đồ nhánh dạng ASCII.
	Graph bool
}

// LogEntry là một dòng log đã định dạng sẵn.
type LogEntry struct {
	Hash        object.Hash
	Short       string
	Author      string
	Email       string
	When        time.Time
	Summary     string
	Refs        []string
	GraphPrefix string
	Body        string
	StatLines   []string
	Patch       string
}

// Log dựng danh sách lịch sử commit đã định dạng.
func Log(r *repo.Repo, opts LogOptions) ([]LogEntry, error) {
	repoOpts := repo.LogOptions{
		Max:   opts.Max,
		Paths: opts.Paths,
	}
	var raw []repo.LogEntry
	var err error
	if opts.All {
		raw, err = r.LogAllRefs(repoOpts)
	} else {
		start := opts.Revision
		if start == "" {
			start = "HEAD"
		}
		h, rerr := resolveCommitish(r, start)
		if rerr != nil {
			return nil, rerr
		}
		raw, err = r.Log(h, repoOpts)
	}
	if err != nil {
		return nil, err
	}
	if opts.Reverse {
		for i, j := 0, len(raw)-1; i < j; i, j = i+1, j-1 {
			raw[i], raw[j] = raw[j], raw[i]
		}
	}

	out := make([]LogEntry, 0, len(raw))
	for _, e := range raw {
		le := LogEntry{
			Hash:    e.Hash,
			Short:   e.Hash.Short(hashLen(opts.ShortHash)),
			Author:  e.Commit.Author.Name,
			Email:   e.Commit.Author.Email,
			When:    e.Commit.Author.When,
			Summary: e.Commit.Summary(),
			Body:    bodyOf(e.Commit.Message),
		}
		if refs, err := r.RefsContaining(e.Hash); err == nil {
			le.Refs = refs
		}
		// Chỉ commit đang trỏ bởi HEAD mới được đánh dấu "HEAD ->",
		// đồng thời bỏ nhãn trùng lặp của chính nhánh đó.
		if cur, _ := r.CurrentBranch(); cur != "" {
			if head, _ := r.Head(); head == e.Hash && containsString(le.Refs, cur) {
				le.Refs = replaceFirst(le.Refs, cur, "HEAD -> "+cur)
			}
		}
		if opts.Stat || opts.Patch {
			if err := fillDiffInfo(r, &le, e.Hash, opts.Patch); err != nil {
				return nil, err
			}
		}
		out = append(out, le)
	}
	return out, nil
}

// fillDiffInfo bổ sung thống kê hoặc nội dung diff cho một entry log.
func fillDiffInfo(r *repo.Repo, le *LogEntry, h object.Hash, withPatch bool) error {
	c, err := r.Objects.ReadCommit(h)
	if err != nil {
		return err
	}
	parent := object.ZeroHash
	if len(c.Parents) > 0 {
		parent, _ = r.CommitTree(c.Parents[0])
	}
	diffs, err := r.DiffTreesContent(parent, c.Tree)
	if err != nil {
		return err
	}
	paths := make([]string, 0, len(diffs))
	for p := range diffs {
		paths = append(paths, p)
	}
	sortStrings(paths)

	totalAdd, totalDel := 0, 0
	for _, p := range paths {
		d := diffs[p]
		if d.Binary {
			continue
		}
		add, del := diff.Stat(d.OldLines, d.NewLines)
		totalAdd += add
		totalDel += del
		le.StatLines = append(le.StatLines, fmt.Sprintf(" %s | %d +%d -%d", p, countLines(d.OldLines)+add, add, del))
		if withPatch {
			if body := diff.FormatUnified(d.OldLines, d.NewLines, 3); body != "" {
				le.Patch += fmt.Sprintf("--- a/%s\n+++ b/%s\n%s", p, p, body)
			}
		}
	}
	if len(le.StatLines) > 0 {
		le.StatLines = append([]string{fmt.Sprintf(" %d file thay đổi, %d dòng thêm, %d dòng xoá", len(paths), totalAdd, totalDel)}, le.StatLines...)
	}
	return nil
}

// countLines đếm số dòng của một nội dung.
func countLines(lines []string) int { return len(lines) }

// hashLen trả về độ dài hash muốn hiển thị.
func hashLen(short bool) int {
	if short {
		return 7
	}
	return 12
}

// bodyOf lấy phần thân của message sau dòng tiêu đề.
func bodyOf(msg string) string {
	lines := strings.SplitN(strings.TrimSpace(msg), "\n", 2)
	if len(lines) < 2 {
		return ""
	}
	return strings.TrimSpace(lines[1])
}

// containsString kiểm tra một chuỗi có nằm trong danh sách hay không.
func containsString(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

// replaceFirst thay phần tử đầu tiên khớp old bằng new, trả về bản sao.
func replaceFirst(list []string, old, new string) []string {
	out := make([]string, 0, len(list))
	done := false
	for _, v := range list {
		if !done && v == old {
			out = append(out, new)
			done = true
			continue
		}
		out = append(out, v)
	}
	return out
}

// showOptions điều khiển lệnh show.
type showOptions struct {
	patch bool
}

// Show hiển thị chi tiết một commit: message, thay đổi và nội dung diff.
func Show(r *repo.Repo, rev string, withPatch bool) (string, error) {
	h, err := resolveCommitish(r, rev)
	if err != nil {
		return "", err
	}
	c, err := r.Objects.ReadCommit(h)
	if err != nil {
		return "", err
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "commit %s\n", h.String())
	if refs, err := r.RefsContaining(h); err == nil && len(refs) > 0 {
		fmt.Fprintf(&sb, "Refs:   %s\n", strings.Join(refs, ", "))
	}
	if len(c.Parents) > 1 {
		var ps []string
		for _, p := range c.Parents {
			ps = append(ps, p.Short(8))
		}
		fmt.Fprintf(&sb, "Merge:  %s\n", strings.Join(ps, " "))
	}
	fmt.Fprintf(&sb, "Tác giả: %s <%s>\n", c.Author.Name, c.Author.Email)
	fmt.Fprintf(&sb, "Ngày:   %s\n", c.Author.When.Format(time.RFC3339))
	fmt.Fprintf(&sb, "\n    %s\n", indentMessage(c.Message))

	parent := object.ZeroHash
	if len(c.Parents) > 0 {
		parent, _ = r.CommitTree(c.Parents[0])
	}
	diffs, err := r.DiffTreesContent(parent, c.Tree)
	if err != nil {
		return "", err
	}
	paths := make([]string, 0, len(diffs))
	for p := range diffs {
		paths = append(paths, p)
	}
	sortStrings(paths)
	for _, p := range paths {
		d := diffs[p]
		fmt.Fprintf(&sb, "\n %s %s\n", statusLabel(d.Status), p)
		if withPatch && !d.Binary {
			sb.WriteString(diff.FormatUnified(d.OldLines, d.NewLines, 3))
		}
	}
	return sb.String(), nil
}

// indentMessage thụt lề toàn bộ dòng của nội dung commit.
func indentMessage(msg string) string {
	msg = strings.TrimSpace(msg)
	lines := strings.Split(msg, "\n")
	for i := range lines {
		lines[i] = "    " + lines[i]
	}
	return strings.Join(lines, "\n")
}

// statusLabel chuyển mã trạng thái sang nhãn hiển thị.
func statusLabel(s byte) string {
	switch s {
	case 'A':
		return "thêm   "
	case 'D':
		return "xoá    "
	default:
		return "sửa    "
	}
}

// sortStrings sắp xếp chuỗi tăng dần.
func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j] < s[j-1]; j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}
