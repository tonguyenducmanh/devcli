package ops

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/tonguyenducmanh/devcli/internal/vcs/index"
	"github.com/tonguyenducmanh/devcli/internal/vcs/object"
	"github.com/tonguyenducmanh/devcli/internal/vcs/repo"
)

// globMatchPath khớp đường dẫn với mẫu glob có * và **.
func globMatchPath(pattern, path string) bool {
	// Tách mẫu thành các đoạn theo dấu "/".
	return matchSegments(splitPattern(pattern), splitPattern(path))
}

// splitPattern tách chuỗi theo dấu "/".
func splitPattern(s string) []string {
	var out []string
	cur := ""
	for i := 0; i < len(s); i++ {
		if s[i] == '/' {
			if cur != "" {
				out = append(out, cur)
			}
			cur = ""
			continue
		}
		cur += string(s[i])
	}
	if cur != "" {
		out = append(out, cur)
	}
	return out
}

// matchSegments khớp hai danh sách đoạn, hỗ trợ đoạn "**" khớp nhiều cấp.
func matchSegments(pat, seg []string) bool {
	if len(pat) == 0 {
		return len(seg) == 0
	}
	if pat[0] == "**" {
		// Thử khớp phần còn lại với mọi vị trí bắt đầu trong seg.
		for i := 0; i <= len(seg); i++ {
			if matchSegments(pat[1:], seg[i:]) {
				return true
			}
		}
		return false
	}
	if len(seg) == 0 {
		return false
	}
	if !matchOneSegment(pat[0], seg[0]) {
		return false
	}
	return matchSegments(pat[1:], seg[1:])
}

// matchOneSegment khớp một đoạn đường dẫn với mẫu chứa * và ?.
// Dấu * không vượt qua ranh giới đoạn vì mỗi đoạn đã tách riêng.
func matchOneSegment(pat, seg string) bool {
	p, s := 0, 0
	star, backtrack := -1, 0
	for s < len(seg) {
		switch {
		case p < len(pat) && (pat[p] == '?' || pat[p] == seg[s]):
			p++
			s++
		case p < len(pat) && pat[p] == '*':
			star = p
			backtrack = s
			p++
		case star >= 0:
			p = star + 1
			backtrack++
			s = backtrack
		default:
			return false
		}
	}
	for p < len(pat) && pat[p] == '*' {
		p++
	}
	return p == len(pat)
}

// osStat kiểm tra một file trong worktree có tồn tại không.
func osStat(r *repo.Repo, rel string) (os.FileInfo, error) {
	return os.Lstat(r.WorkPath(rel))
}

// existsInWorktree báo xem đường dẫn có tồn tại trong worktree không.
func existsInWorktree(r *repo.Repo, rel string) bool {
	_, err := os.Lstat(r.WorkPath(rel))
	return err == nil
}

// matchesAnyPath kiểm tra một đường dẫn có khớp với danh sách mẫu nào không.
func matchesAnyPath(path string, patterns []string) bool {
	for _, p := range patterns {
		if matchPathPattern(path, p) {
			return true
		}
	}
	return false
}

// matchPathPattern khớp đường dẫn với một mẫu từ dòng lệnh.
// Hỗ trợ khớp chính xác, khớp theo thư mục và mẫu glob.
func matchPathPattern(path, pattern string) bool {
	if pattern == "" || pattern == "." {
		return true
	}
	pattern = strings.TrimPrefix(filepath.ToSlash(filepath.Clean(pattern)), "./")
	pattern = strings.TrimSuffix(pattern, "/")
	if path == pattern {
		return true
	}
	// Mẫu là thư mục: mọi file bên trong đều khớp.
	if strings.HasPrefix(path, pattern+"/") {
		return true
	}
	// Khớp theo mẫu glob.
	if strings.ContainsAny(pattern, "*?[") {
		return globMatchPath(pattern, path)
	}
	return false
}

// resolveCommitish biến một tên rút gọn (HEAD~2, nhánh, tag, hash) thành hash commit.
func resolveCommitish(r *repo.Repo, name string) (object.Hash, error) {
	return r.ResolveRev(name)
}

// parseHashString chuyển chuỗi hex thành hash.
func parseHashString(s string) (object.Hash, error) { return object.ParseHash(s) }

// dirOf trả về thư mục chứa một đường dẫn.
func dirOf(p string) string { return filepath.Dir(p) }

// indexEntry chuyển một node của cây thành entry index.
func indexEntry(n repo.TreeNode) index.Entry {
	return index.Entry{Mode: n.Mode, Hash: n.Hash, Name: n.Path}
}

// firstLine lấy dòng đầu tiên của một chuỗi.
func firstLine(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}
