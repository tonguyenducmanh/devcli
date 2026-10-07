package repo

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/tonguyenducmanh/devcli/internal/vcs/index"
	"github.com/tonguyenducmanh/devcli/internal/vcs/object"
	"github.com/tonguyenducmanh/devcli/internal/vcs/worktree"
)

// TreeNode là một node khi duyệt phẳng một tree.
type TreeNode struct {
	Path string
	Mode object.FileMode
	Hash object.Hash
}

// Flatten duyệt đệ quy một tree thành danh sách file phẳng (đường dẫn đầy đủ).
func (r *Repo) Flatten(h object.Hash) (map[string]TreeNode, error) {
	out := map[string]TreeNode{}
	if h.IsZero() {
		return out, nil
	}
	var walk func(hash object.Hash, prefix string) error
	walk = func(hash object.Hash, prefix string) error {
		tree, err := r.Objects.ReadTree(hash)
		if err != nil {
			return err
		}
		for _, e := range tree.Entries {
			p := e.Name
			if prefix != "" {
				p = prefix + "/" + e.Name
			}
			if e.Mode.IsTree() {
				if err := walk(e.Hash, p); err != nil {
					return err
				}
				continue
			}
			out[p] = TreeNode{Path: p, Mode: e.Mode, Hash: e.Hash}
		}
		return nil
	}
	if err := walk(h, ""); err != nil {
		return nil, err
	}
	return out, nil
}

// FlattenSorted trả về cây phẳng đã sắp xếp theo đường dẫn.
func (r *Repo) FlattenSorted(h object.Hash) ([]TreeNode, error) {
	m, err := r.Flatten(h)
	if err != nil {
		return nil, err
	}
	out := make([]TreeNode, 0, len(m))
	for _, v := range m {
		out = append(out, v)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out, nil
}

// WriteTree dựng object tree từ danh sách đường dẫn phẳng.
// Các node được gom theo thư mục cha rồi dựng đệ quy từ dưới lên trên.
func (r *Repo) WriteTree(nodes []TreeNode, dir string) (object.Hash, error) {
	// Tập entry của thư mục đang dựng, khóa là tên entry (tên rỗng cho chính nó).
	type slot struct {
		mode object.FileMode
		hash object.Hash
		// children giữ các node nằm trong thư mục con, để dựng đệ quy.
		children []TreeNode
	}
	entries := map[string]*slot{}

	for _, n := range nodes {
		name := n.Path
		slash := strings.IndexByte(name, '/')
		if slash < 0 {
			// Nằm trực tiếp trong thư mục đang dựng.
			entries[name] = &slot{mode: n.Mode, hash: n.Hash}
			continue
		}
		// Nằm trong một thư mục con: tạo entry thư mục nếu chưa có.
		childName := name[:slash]
		s, ok := entries[childName]
		if !ok {
			s = &slot{mode: object.ModeTree}
			entries[childName] = s
		}
		// Lưu đường dẫn đầy đủ để đệ quy xử lý trong thư mục con.
		s.children = append(s.children, TreeNode{Path: name, Mode: n.Mode, Hash: n.Hash})
	}

	tree := &object.Tree{}
	for name, s := range entries {
		if s.mode.IsTree() {
			// Đệ quy dựng cây của thư mục con.
			sub := make([]TreeNode, 0, len(s.children))
			prefix := name + "/"
			for _, n := range s.children {
				sub = append(sub, TreeNode{
					Path: strings.TrimPrefix(n.Path, prefix),
					Mode: n.Mode,
					Hash: n.Hash,
				})
			}
			h, err := r.WriteTree(sub, dir)
			if err != nil {
				return object.ZeroHash, err
			}
			tree.Entries = append(tree.Entries, object.TreeEntry{
				Mode: object.ModeTree,
				Name: name,
				Hash: h,
			})
			continue
		}
		tree.Entries = append(tree.Entries, object.TreeEntry{
			Mode: s.mode,
			Name: name,
			Hash: s.hash,
		})
	}
	return r.Objects.WriteTree(tree)
}

// TreeFromIndex dựng object tree từ staging area hiện tại.
func (r *Repo) TreeFromIndex() (object.Hash, error) {
	var nodes []TreeNode
	for _, e := range r.Index.Entries() {
		if e.Stage != index.StageNormal {
			continue // bỏ qua entry xung đột khi dựng tree
		}
		nodes = append(nodes, TreeNode{Path: e.Name, Mode: e.Mode, Hash: e.Hash})
	}
	if len(nodes) == 0 {
		// Tree rỗng vẫn cần một hash hợp lệ.
		return r.Objects.WriteTree(&object.Tree{})
	}
	return r.WriteTree(nodes, "")
}

// DiffTrees trả về danh sách thay đổi giữa hai tree ở dạng:
// Status là 'A' (thêm), 'D' (xoá) hoặc 'M' (sửa).
func (r *Repo) DiffTrees(from, to object.Hash) ([]TreeChange, error) {
	oldNodes := map[string]TreeNode{}
	if !from.IsZero() {
		n, err := r.Flatten(from)
		if err != nil {
			return nil, err
		}
		oldNodes = n
	}
	newNodes := map[string]TreeNode{}
	if !to.IsZero() {
		n, err := r.Flatten(to)
		if err != nil {
			return nil, err
		}
		newNodes = n
	}

	var paths []string
	seen := map[string]bool{}
	for p := range oldNodes {
		if !seen[p] {
			seen[p] = true
			paths = append(paths, p)
		}
	}
	for p := range newNodes {
		if !seen[p] {
			seen[p] = true
			paths = append(paths, p)
		}
	}
	sort.Strings(paths)

	var out []TreeChange
	for _, p := range paths {
		o, hasOld := oldNodes[p]
		n, hasNew := newNodes[p]
		switch {
		case hasOld && hasNew:
			if o.Hash != n.Hash || o.Mode != n.Mode {
				out = append(out, TreeChange{Path: p, Status: 'M', Old: o, New: n})
			}
		case hasNew:
			out = append(out, TreeChange{Path: p, Status: 'A', New: n})
		default:
			out = append(out, TreeChange{Path: p, Status: 'D', Old: o})
		}
	}
	return out, nil
}

// TreeChange mô tả một thay đổi giữa hai tree.
type TreeChange struct {
	Path   string
	Status byte // 'A', 'D', 'M'
	Old    TreeNode
	New    TreeNode
}

// CommitTree trả về tree của một commit.
func (r *Repo) CommitTree(h object.Hash) (object.Hash, error) {
	c, err := r.Objects.ReadCommit(h)
	if err != nil {
		return object.ZeroHash, err
	}
	return c.Tree, nil
}

// IsAncestor kiểm tra xem ancestor có nằm trong lịch sử của commit hay không.
// Dùng BFS trên đồ thị phụ huynh, đủ cho quy mô repo cá nhân.
func (r *Repo) IsAncestor(ancestor, commit object.Hash) (bool, error) {
	if ancestor == commit {
		return true, nil
	}
	if ancestor.IsZero() || commit.IsZero() {
		return false, nil
	}
	seen := map[object.Hash]bool{commit: true}
	queue := []object.Hash{commit}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		c, err := r.Objects.ReadCommit(cur)
		if err != nil {
			if errors.Is(err, object.ErrObjectType) {
				continue
			}
			return false, err
		}
		for _, p := range c.Parents {
			if p == ancestor {
				return true, nil
			}
			if !seen[p] {
				seen[p] = true
				queue = append(queue, p)
			}
		}
	}
	return false, nil
}

// MergeBase tìm commit tổ tiên chung gần nhất của hai commit.
func (r *Repo) MergeBase(a, b object.Hash) (object.Hash, error) {
	if a.IsZero() || b.IsZero() {
		return object.ZeroHash, nil
	}
	// Thu thập tập tổ tiên của a rồi BFS từ b theo thứ tự khoảng cách.
	ancestors, err := r.ancestorSet(a)
	if err != nil {
		return object.ZeroHash, err
	}
	seen := map[object.Hash]bool{}
	queue := []object.Hash{b}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		if ancestors[cur] {
			return cur, nil
		}
		if seen[cur] {
			continue
		}
		seen[cur] = true
		c, err := r.Objects.ReadCommit(cur)
		if err != nil {
			if errors.Is(err, object.ErrObjectType) {
				continue
			}
			return object.ZeroHash, err
		}
		queue = append(queue, c.Parents...)
	}
	return object.ZeroHash, nil
}

// ancestorSet gom toàn bộ tổ tiên (bao gồm chính nó) của một commit.
func (r *Repo) ancestorSet(h object.Hash) (map[object.Hash]bool, error) {
	out := map[object.Hash]bool{}
	queue := []object.Hash{h}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		if out[cur] {
			continue
		}
		out[cur] = true
		c, err := r.Objects.ReadCommit(cur)
		if err != nil {
			if errors.Is(err, object.ErrObjectType) {
				continue
			}
			return nil, err
		}
		queue = append(queue, c.Parents...)
	}
	return out, nil
}

// ReadBlobLines đọc blob và tách thành các dòng.
func (r *Repo) ReadBlobLines(h object.Hash) ([]string, error) {
	data, err := r.Objects.ReadBlob(h)
	if err != nil {
		return nil, err
	}
	return splitLines(string(data)), nil
}

// splitLines tách nội dung thành dòng, giữ nguyên quy ước xuống dòng cuối.
func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	s = strings.ReplaceAll(s, "\r\n", "\n")
	if strings.HasSuffix(s, "\n") {
		s = s[:len(s)-1]
	}
	return strings.Split(s, "\n")
}

// ReadIndexBlob đọc nội dung blob của một entry trong index.
func (r *Repo) ReadIndexBlob(e index.Entry) ([]byte, error) {
	if e.Hash.IsZero() {
		return nil, nil
	}
	return r.Objects.ReadBlob(e.Hash)
}

// IndexTreeNode chuyển entry index sang node để so sánh cây.
func IndexTreeNode(e index.Entry) TreeNode {
	return TreeNode{Path: e.Name, Mode: e.Mode, Hash: e.Hash}
}

// ScanWorktree quét thư mục làm việc và trả về danh sách file không bị bỏ qua.
// File trong .tdx luôn bị loại khỏi kết quả.
func (r *Repo) ScanWorktree() ([]TreeNode, error) {
	var out []TreeNode
	err := filepath.WalkDir(r.Root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // bỏ qua thư mục không đọc được
		}
		rel, rerr := r.RelPath(path)
		if rerr != nil {
			return nil
		}
		if rel == DirName || strings.HasPrefix(rel, DirName+"/") {
			return filepath.SkipDir
		}
		if d.IsDir() {
			// Nạp .tdxignore của thư mục này ngay khi bước vào, trước khi xét
			// thư mục con bên trong. Nạp lúc duyệt tới tệp thì thứ tự từ vựng
			// của WalkDir có thể làm một thư mục con nào đó được vào trước khi
			// quy tắc kịp nạp. Quy tắc sâu hơn nạp sau nên thắng, giống git.
			if rel != "." {
				r.loadIgnoreIn(path)
			}
			// Không đi vào thư mục bị bỏ qua.
			if r.Ignore.Matches(r.Root, rel) {
				return filepath.SkipDir
			}
			return nil
		}
		if !d.Type().IsRegular() && d.Type()&os.ModeSymlink == 0 {
			return nil // bỏ qua socket, fifo, device
		}
		if r.Ignore.Matches(r.Root, rel) {
			return nil
		}
		fi, ferr := os.Lstat(path)
		if ferr != nil {
			return nil
		}
		mode := worktree.FileModeFromInfo(fi)
		out = append(out, TreeNode{Path: rel, Mode: mode})
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out, nil
}

// StageFile đọc file từ worktree, ghi blob và cập nhật index.
func (r *Repo) StageFile(rel string) error {
	abs := r.WorkPath(rel)
	fi, err := os.Lstat(abs)
	if err != nil {
		if os.IsNotExist(err) {
			r.Index.Remove(rel)
			return nil
		}
		return err
	}
	mode := worktree.FileModeFromInfo(fi)
	data, err := worktree.ReadFileSymlinkAware(abs, mode)
	if err != nil {
		return fmt.Errorf("đọc %s: %w", rel, err)
	}
	h, err := r.Objects.WriteBlob(data)
	if err != nil {
		return err
	}
	r.Index.Add(index.Entry{
		CTime: worktree.StatTimes(fi),
		MTime: worktree.StatTimes(fi),
		Mode:  mode,
		Size:  uint32(len(data)),
		Hash:  h,
		Name:  rel,
	})
	return nil
}

// UnstagePath đưa một đường dẫn trở về trạng thái của HEAD trong index.
func (r *Repo) UnstagePath(rel string) error {
	r.Index.Remove(rel)
	head, err := r.Head()
	if err != nil || head.IsZero() {
		return nil
	}
	tree, err := r.CommitTree(head)
	if err != nil {
		return err
	}
	nodes, err := r.Flatten(tree)
	if err != nil {
		return err
	}
	if n, ok := nodes[rel]; ok {
		r.Index.Add(index.Entry{Mode: n.Mode, Hash: n.Hash, Name: rel})
	}
	return nil
}

// ApplyTreeToWorktree ghi nội dung của một tree xuống thư mục làm việc.
// Tham số from là tree cũ để biết cần xóa những file nào.
func (r *Repo) ApplyTreeToWorktree(from, to object.Hash) error {
	changes, err := r.DiffTrees(from, to)
	if err != nil {
		return err
	}
	return r.applyChanges(changes)
}

// applyChanges thi hành danh sách thay đổi lên worktree.
func (r *Repo) applyChanges(changes []TreeChange) error {
	// Xóa trước để tránh xung đột thư mục/file.
	for _, c := range changes {
		if c.Status == 'D' {
			abs := r.WorkPath(c.Path)
			if err := worktree.RemoveFile(abs, r.Root); err != nil {
				return fmt.Errorf("xoá %s: %w", c.Path, err)
			}
		}
	}
	for _, c := range changes {
		if c.Status == 'D' {
			continue
		}
		data, err := r.Objects.ReadBlob(c.New.Hash)
		if err != nil {
			return err
		}
		abs := r.WorkPath(c.Path)
		if err := worktree.WriteFileSymlinkAware(abs, c.New.Mode, data); err != nil {
			return fmt.Errorf("ghi %s: %w", c.Path, err)
		}
	}
	return nil
}

// ResetWorktreeTo đặt index và worktree về đúng nội dung của một tree.
// Mọi file trên đĩa khác với tree đích đều bị ghi lại hoặc xoá.
func (r *Repo) ResetWorktreeTo(tree object.Hash) error {
	// Đọc trạng thái hiện tại của cây làm việc để biết cần sửa những gì.
	workNodes, err := r.ScanWorktree()
	if err != nil {
		return err
	}
	current := map[string]TreeNode{}
	for _, n := range workNodes {
		data, err := readWorktreeData(r, n.Path, n.Mode)
		if err != nil {
			// File không đọc được thì coi như không tồn tại để tránh treo.
			continue
		}
		current[n.Path] = TreeNode{
			Path: n.Path,
			Mode: n.Mode,
			Hash: object.ComputeHash(object.TypeBlob, data),
		}
	}

	target := map[string]TreeNode{}
	if !tree.IsZero() {
		target, err = r.Flatten(tree)
		if err != nil {
			return err
		}
	}

	// Lập danh sách thay đổi giữa trạng thái hiện tại và trạng thái đích.
	var changes []TreeChange
	seen := map[string]bool{}
	for p, n := range target {
		seen[p] = true
		if cur, ok := current[p]; ok && cur.Hash == n.Hash && cur.Mode == n.Mode {
			continue
		}
		changes = append(changes, TreeChange{Path: p, Status: 'M', Old: current[p], New: n})
	}
	for p, n := range current {
		if seen[p] {
			continue
		}
		changes = append(changes, TreeChange{Path: p, Status: 'D', Old: n})
	}
	sort.Slice(changes, func(i, j int) bool { return changes[i].Path < changes[j].Path })

	if err := r.applyChanges(changes); err != nil {
		return err
	}

	// Đồng bộ lại index theo tree đích.
	r.Index.Clear()
	if tree.IsZero() {
		return r.SaveIndex()
	}
	paths := make([]string, 0, len(target))
	for p := range target {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	for _, p := range paths {
		n := target[p]
		var size uint32
		if data, err := r.Objects.ReadBlob(n.Hash); err == nil {
			size = uint32(len(data))
		}
		r.Index.Add(index.Entry{Mode: n.Mode, Size: size, Hash: n.Hash, Name: n.Path})
	}
	return r.SaveIndex()
}

// UpdateRefWithLog cập nhật ref kèm reflog.
func (r *Repo) UpdateRefWithLog(ref string, h object.Hash, msg string) error {
	old, _ := r.Refs.Resolve(ref)
	if err := r.Refs.Write(ref, h); err != nil {
		return err
	}
	return r.Refs.AppendReflog(ref, old, h, msg)
}

// loadIgnoreIn nạp tệp ignore nằm trong một thư mục, nếu có.
//
// Không nạp lại tệp ở thư mục gốc vì Open đã lo rồi. Lỗi đọc tệp bị bỏ qua:
// một tệp ignore hỏng không nên làm hỏng luôn việc quét cây làm việc, và tệp
// đó sẽ được báo lỗi khi lệnh ignore in nội dung.
func (r *Repo) loadIgnoreIn(dir string) {
	if filepath.Base(dir) == DirName {
		return
	}
	f := filepath.Join(dir, worktree.IgnoreFileName)
	if _, err := os.Stat(f); err != nil {
		return
	}
	_ = r.Ignore.AddFile(f)
}

// IgnoreFiles trả về các tệp chứa quy tắc bỏ qua, theo đường dẫn tương đối tới
// gốc kho.
//
// Gồm .tdx/info/exclude và mọi tệp .tdxignore từ gốc xuống các thư mục con.
// Kết quả sắp theo thứ tự thư mục nông trước, để đọc ra quy tắc theo đúng thứ
// tự quyết định: quy tắc sâu hơn nằm sau nên thắng.
//
// Thư mục đã bị bỏ qua thì không vào, vì git cũng không đọc tệp ignore bên
// trong thư mục mà chính quy tắc ở cấp trên đã loại bỏ.
func (r *Repo) IgnoreFiles() ([]string, error) {
	var out []string

	err := filepath.WalkDir(r.Root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // bỏ qua thư mục không đọc được
		}
		rel, rerr := r.RelPath(path)
		if rerr != nil {
			return nil
		}
		if rel == DirName || strings.HasPrefix(rel, DirName+"/") {
			return filepath.SkipDir
		}
		if d.IsDir() {
			r.loadIgnoreIn(path)
			if r.Ignore.Matches(r.Root, rel) {
				return filepath.SkipDir
			}
			return nil
		}
		if d.Name() == worktree.IgnoreFileName {
			out = append(out, rel)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.SliceStable(out, func(i, j int) bool {
		di, dj := strings.Count(out[i], "/"), strings.Count(out[j], "/")
		if di != dj {
			return di < dj
		}
		return out[i] < out[j]
	})

	// Tệp exclude của riêng máy để cuối cùng. Nó được nạp trước .tdxignore nên
	// quy tắc chung của dự án có thể phủ lên nó, đọc ra sau cho đúng ý nghĩa.
	if _, err := os.Stat(filepath.Join(r.GitDir, "info", "exclude")); err == nil {
		out = append(out, filepath.ToSlash(filepath.Join(DirName, "info", "exclude")))
	}
	return out, nil
}
