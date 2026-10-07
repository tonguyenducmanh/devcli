package worktree

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// Ignore quy tắc bỏ qua file, hỗ trợ cú pháp gitignore cơ bản:
// dấu # (chú thích), ! (phủ định), / (neo tới gốc), * và **.
type Ignore struct {
	rules []ignoreRule
	// base là thư mục chứa file ignore, dùng để tính đường dẫn tương đối.
	base string
}

type ignoreRule struct {
	pattern  string
	negate   bool
	dirOnly  bool
	anchored bool
	// base là thư mục chứa tệp ignore sinh ra quy tắc này. Quy tắc chỉ áp dụng
	// cho những đường dẫn nằm bên trong base, đúng như gitignore lồng nhau.
	base string
	// source là đường dẫn tệp ignore, dùng để nạp lại mà không nhân bản quy
	// tắc. Rỗng với quy tắc nạp từ reader không gắn với tệp nào.
	source string
}

// NewIgnore tạo bộ quy tắc rỗng.
func NewIgnore(base string) *Ignore { return &Ignore{base: base} }

// IgnoreFileName là tên tệp ignore đặt trong thư mục bất kỳ, kể cả thư mục con.
const IgnoreFileName = ".tdxignore"

// AddFile nạp các quy tắc từ một tệp ignore.
//
// Nạp lại cùng một tệp sẽ thay thế các quy tắc cũ chứ không nối thêm, để quét
// cây làm việc nhiều lần vẫn không sinh quy tắc trùng.
func (ig *Ignore) AddFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return ig.addFrom(f, filepath.Dir(path), path)
}

// addFrom đọc quy tắc từ reader rồi nối vào cuối danh sách.
func (ig *Ignore) addFrom(r interface {
	Read([]byte) (int, error)
}, dir, source string) error {
	ig.dropSource(source)

	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		ig.addPattern(line, dir, source)
	}
	return sc.Err()
}

// dropSource xoá các quy tắc cũ đến từ một tệp ignore.
func (ig *Ignore) dropSource(source string) {
	if source == "" {
		return
	}
	kept := ig.rules[:0]
	for _, r := range ig.rules {
		if r.source != source {
			kept = append(kept, r)
		}
	}
	ig.rules = kept
}

// AddReader nạp quy tắc từ một reader, dir là thư mục chứa nguồn quy tắc.
func (ig *Ignore) AddReader(r interface {
	Read([]byte) (int, error)
}, dir string) error {
	return ig.addFrom(r, dir, "")
}

func (ig *Ignore) addPattern(line, dir, source string) {
	r := ignoreRule{base: dir, source: source}
	if strings.HasPrefix(line, "!") {
		r.negate = true
		line = line[1:]
	}
	if strings.HasSuffix(line, "/") {
		r.dirOnly = true
		line = strings.TrimSuffix(line, "/")
	}
	// Neo tới thư mục chứa file ignore khi có dấu gạch chéo ở đầu hoặc giữa mẫu.
	if strings.HasPrefix(line, "/") {
		r.anchored = true
		line = strings.TrimPrefix(line, "/")
	} else if strings.Contains(line, "/") {
		r.anchored = true
	}
	if line == "" {
		return
	}
	r.pattern = line
	ig.rules = append(ig.rules, r)
}

// Matches báo xem đường dẫn tương đối tới repoRoot có bị bỏ qua hay không.
// Quy tắc sau ghi đè quy tắc trước, phù hợp với gitignore.
func (ig *Ignore) Matches(repoRoot, rel string) bool {
	if rel == "" || rel == "." {
		return false
	}
	rel = filepath.ToSlash(rel)
	ignored := false
	for _, r := range ig.rules {
		// Quy tắc từ tệp ignore nằm trong thư mục con chỉ có tác dụng bên
		// trong thư mục đó. Ở đây cắt tiền tố thư mục ra, rồi mẫu được so với
		// phần đường dẫn còn lại.
		sub, ok := trimBasePrefix(repoRoot, r.base, rel)
		if !ok || !r.match(sub) {
			continue
		}
		if r.negate {
			ignored = false
			continue
		}
		ignored = true
	}
	return ignored
}

// trimBasePrefix cắt tiền tố thư mục chứa tệp ignore ra khỏi đường dẫn, trả về
// phần còn lại.
//
// Trả về false nếu đường dẫn không nằm trong thư mục đó, tức là quy tắc không
// có tác dụng với đường dẫn này. Chính thư mục chứa tệp ignore cũng không được
// coi là nằm trong đó, vì thư mục đó do tệp ignore ở cấp trên quyết định.
func trimBasePrefix(repoRoot, base, rel string) (string, bool) {
	prefix := filepath.ToSlash(base)
	if root := filepath.ToSlash(repoRoot); strings.HasPrefix(prefix, root) {
		prefix = strings.TrimPrefix(prefix, root)
	}
	prefix = strings.Trim(prefix, "/")
	if prefix == "" || prefix == "." {
		return rel, true
	}
	if rel == prefix {
		return "", false
	}
	if after, ok := strings.CutPrefix(rel, prefix+"/"); ok {
		return after, true
	}
	return "", false
}

// match kiểm tra một quy tắc có khớp với đường dẫn hay không.
func (r ignoreRule) match(rel string) bool {
	// Quy tắc chỉ áp dụng cho thư mục: khớp khi chính đường dẫn là thư mục
	// đó, hoặc khi nằm bên trong một thư mục khớp mẫu.
	if r.dirOnly {
		if r.matchOne(rel) {
			return true
		}
		for dir := filepath.Dir(rel); dir != "." && dir != "/"; dir = filepath.Dir(dir) {
			if r.matchOne(dir) {
				return true
			}
		}
		return false
	}
	return r.matchOne(rel)
}

// matchOne khớp đường dẫn với mẫu của quy tắc, xử lý cả trường hợp neo.
func (r ignoreRule) matchOne(rel string) bool {
	if r.anchored {
		return matchPattern(r.pattern, rel)
	}
	// Quy tắc không neo khớp với bất kỳ cấp thư mục nào.
	if matchPattern(r.pattern, filepath.Base(rel)) {
		return true
	}
	// Hoặc khớp khi đường dẫn nằm ngay dưới thư mục khớp mẫu.
	for dir := filepath.Dir(rel); dir != "." && dir != "/"; dir = filepath.Dir(dir) {
		if matchPattern(r.pattern, dir+"/"+filepath.Base(rel)) {
			return true
		}
	}
	return false
}

// matchPattern khớp một mẫu glob với chuỗi đường dẫn.
// Dấu * không vượt qua dấu "/", nên "src/*.go" chỉ khớp tệp nằm trực tiếp
// trong thư mục src, còn "**" mới khớp được nhiều cấp thư mục.
func matchPattern(pattern, s string) bool {
	if pattern == "**" {
		return true
	}
	// Nếu mẫu chứa "**", thử các cách ghép phổ biến.
	if strings.Contains(pattern, "**/") {
		rest := strings.TrimPrefix(pattern, "**/")
		if matchPattern(rest, s) {
			return true
		}
		if i := strings.Index(s, "/"); i >= 0 && matchPattern(pattern, s[i+1:]) {
			return true
		}
		return false
	}
	return globMatch(pattern, s)
}

// globMatch khớp glob từng ký tự, hỗ trợ * và ?.
// Dấu * không được vượt qua ranh giới thư mục.
func globMatch(pattern, s string) bool {
	p, i := 0, 0
	star, backtrack := -1, 0
	for i < len(s) {
		switch {
		case p < len(pattern) && pattern[p] == '?' && s[i] != '/':
			p++
			i++
		case p < len(pattern) && pattern[p] == s[i]:
			p++
			i++
		case p < len(pattern) && pattern[p] == '*':
			// Khi không thể khớp tiếp, lùi lại vị trí ngay sau dấu *.
			star = p
			backtrack = i
			p++
		case star >= 0 && s[backtrack] != '/':
			p = star + 1
			backtrack++
			i = backtrack
		default:
			return false
		}
	}
	for p < len(pattern) && pattern[p] == '*' {
		p++
	}
	return p == len(pattern)
}

// Len trả về số quy tắc đang có.
func (ig *Ignore) Len() int { return len(ig.rules) }
