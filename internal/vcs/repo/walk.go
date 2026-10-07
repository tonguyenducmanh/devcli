package repo

import (
	"sort"

	"github.com/tonguyenducmanh/devcli/internal/vcs/object"
	"github.com/tonguyenducmanh/devcli/internal/vcs/storage"
)

// LogEntry là một mắt xích trong lịch sử commit.
type LogEntry struct {
	Commit *object.Commit
	Hash   object.Hash
}

// LogOptions điều khiển cách dựng lịch sử.
type LogOptions struct {
	// Max giới hạn số commit trả về (0 = không giới hạn).
	Max int
	// Paths lọc theo đường dẫn có thay đổi.
	Paths []string
	// Skip giới hạn số commit bỏ qua từ đầu, hữu ích cho phân trang.
	Skip int
	// First dừng ở commit đầu tiên tìm thấy hợp lệ (dùng với Paths).
	First bool
}

// Log dựng danh sách commit từ một điểm bắt đầu theo thứ tự thời gian giảm dần.
func (r *Repo) Log(start object.Hash, opts LogOptions) ([]LogEntry, error) {
	if start.IsZero() {
		return nil, nil
	}
	var out []LogEntry
	seen := map[object.Hash]bool{}
	// Hàng đợi BFS để duyệt lịch sử theo thứ tự gần xa.
	queue := []object.Hash{start}
	skipped := 0

	for len(queue) > 0 {
		h := queue[0]
		queue = queue[1:]
		if seen[h] {
			continue
		}
		seen[h] = true

		c, err := r.Objects.ReadCommit(h)
		if err != nil {
			continue
		}

		if len(opts.Paths) > 0 {
			match, err := r.commitTouchesPaths(h, opts.Paths)
			if err != nil {
				return nil, err
			}
			if !match {
				// Bỏ qua commit này nhưng vẫn đi tiếp các phụ huynh.
				queue = append(queue, c.Parents...)
				continue
			}
		}

		if skipped < opts.Skip {
			skipped++
			queue = append(queue, c.Parents...)
			continue
		}
		out = append(out, LogEntry{Commit: c, Hash: h})
		if opts.Max > 0 && len(out) >= opts.Max {
			break
		}
		if opts.First && len(opts.Paths) > 0 {
			break
		}
		queue = append(queue, c.Parents...)
	}
	return out, nil
}

// commitTouchesPaths kiểm tra một commit có thay đổi bất kỳ đường dẫn nào không.
// Với commit gốc, mọi file trong tree đều được coi là thay đổi.
func (r *Repo) commitTouchesPaths(h object.Hash, paths []string) (bool, error) {
	c, err := r.Objects.ReadCommit(h)
	if err != nil {
		return false, err
	}
	var parentTree object.Hash
	if len(c.Parents) > 0 {
		parentTree, err = r.CommitTree(c.Parents[0])
		if err != nil {
			return false, err
		}
	}
	changes, err := r.DiffTrees(parentTree, c.Tree)
	if err != nil {
		return false, err
	}
	for _, ch := range changes {
		for _, p := range paths {
			if ch.Path == p || matchPathPrefix(ch.Path, p) {
				return true, nil
			}
		}
	}
	return false, nil
}

// matchPathPrefix khớp khi một đường dẫn nằm trong thư mục của đường dẫn kia.
func matchPathPrefix(path, prefix string) bool {
	if prefix == "" {
		return true
	}
	if len(prefix) > len(path) {
		return path == prefix[:len(path)]
	}
	return path[:len(prefix)] == prefix
}

// LogAllRefs gom lịch sử từ mọi nhánh và tag.
func (r *Repo) LogAllRefs(opts LogOptions) ([]LogEntry, error) {
	refs, err := r.Refs.List("refs/")
	if err != nil {
		return nil, err
	}
	var tips []object.Hash
	for _, ref := range refs {
		h, err := r.Refs.Resolve(ref)
		if err != nil {
			continue
		}
		tips = append(tips, h)
	}
	if head, err := r.Head(); err == nil && !head.IsZero() {
		tips = append(tips, head)
	}
	if len(tips) == 0 {
		return nil, nil
	}
	// Gộp kết quả rồi loại trùng theo hash.
	merged := map[object.Hash]LogEntry{}
	for _, t := range tips {
		entries, err := r.Log(t, LogOptions{})
		if err != nil {
			return nil, err
		}
		for _, e := range entries {
			if _, ok := merged[e.Hash]; !ok {
				merged[e.Hash] = e
			}
		}
	}
	out := make([]LogEntry, 0, len(merged))
	for _, e := range merged {
		out = append(out, e)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Commit.Author.When.After(out[j].Commit.Author.When)
	})
	if opts.Max > 0 && len(out) > opts.Max {
		out = out[:opts.Max]
	}
	return out, nil
}

// RefsContaining liệt kê các nhánh và tag chứa một commit.
// Ref nội bộ của td như refs/stash không được đưa vào danh sách này.
func (r *Repo) RefsContaining(h object.Hash) ([]string, error) {
	refs, err := r.Refs.List("refs/")
	if err != nil {
		return nil, err
	}
	var out []string
	for _, ref := range refs {
		if !isPublicRef(ref) {
			continue
		}
		rh, err := r.Refs.Resolve(ref)
		if err != nil || rh.IsZero() {
			continue
		}
		ok, err := r.IsAncestor(h, rh)
		if err != nil {
			continue
		}
		if ok {
			out = append(out, storage.ShortName(ref))
		}
	}
	sort.Strings(out)
	return out, nil
}

// isPublicRef báo xem ref có hiển thị cho người dùng hay không.
// refs/stash là ref riêng của td nên bị ẩn khỏi trang trí của lịch sử.
func isPublicRef(ref string) bool {
	return ref != "refs/stash"
}
