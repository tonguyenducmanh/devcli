package repo

import (
	"os"
	"sort"

	"github.com/tonguyenducmanh/devcli/internal/vcs/diff"
	"github.com/tonguyenducmanh/devcli/internal/vcs/index"
	"github.com/tonguyenducmanh/devcli/internal/vcs/object"
	"github.com/tonguyenducmanh/devcli/internal/vcs/storage"
	"github.com/tonguyenducmanh/devcli/internal/vcs/worktree"
)

// FileStatus là trạng thái của một file giữa ba vùng:
// HEAD, index (staging) và worktree.
type FileStatus struct {
	Path        string
	IndexStatus byte // byte 0 = so với HEAD: 'A', 'M', 'D', ' '
	WorkStatus  byte // byte 0 = so với index: 'M', 'D', '?', ' '
	// Nội dung dùng để hiển thị chi tiết.
	IndexMode object.FileMode
	WorkMode  object.FileMode
	IndexHash object.Hash
	WorkHash  object.Hash
	// IsUntracked đánh dấu file mới chưa được theo dõi.
	IsUntracked bool
}

// Status là toàn bộ trạng thái của repo.
type Status struct {
	Branch    string
	Detached  bool
	HeadHash  object.Hash
	Entries   []FileStatus
	Conflicts []string
	// Ahead/Behind so với upstream nếu nhánh có theo dõi.
	HasUpstream bool
	Ahead       int
	Behind      int
}

// IsClean báo xem repo không có thay đổi nào.
func (s *Status) IsClean() bool {
	for _, e := range s.Entries {
		if e.IsUntracked || e.IndexStatus != ' ' || e.WorkStatus != ' ' {
			return false
		}
	}
	return true
}

// Entry tra cứu trạng thái của một đường dẫn cụ thể.
func (s *Status) Entry(path string) (FileStatus, bool) {
	for _, e := range s.Entries {
		if e.Path == path {
			return e, true
		}
	}
	return FileStatus{}, false
}

// IndexStatusOf trả về trạng thái của đường dẫn so với HEAD.
func (s *Status) IndexStatusOf(path string) byte {
	if e, ok := s.Entry(path); ok {
		return e.IndexStatus
	}
	return ' '
}

// Staged trả về các file đã stage.
func (s *Status) Staged() []FileStatus {
	return s.filter(func(e FileStatus) bool { return e.IndexStatus != ' ' })
}

// Unstaged trả về các file đã sửa nhưng chưa stage.
func (s *Status) Unstaged() []FileStatus {
	return s.filter(func(e FileStatus) bool { return e.WorkStatus != ' ' })
}

// Untracked trả về các file chưa được theo dõi.
func (s *Status) Untracked() []FileStatus {
	return s.filter(func(e FileStatus) bool { return e.IsUntracked })
}

func (s *Status) filter(fn func(FileStatus) bool) []FileStatus {
	var out []FileStatus
	for _, e := range s.Entries {
		if fn(e) {
			out = append(out, e)
		}
	}
	return out
}

// Status tính trạng thái hiện tại của repo.
func (r *Repo) Status() (*Status, error) {
	branch, err := r.CurrentBranch()
	if err != nil {
		return nil, err
	}
	head, err := r.Head()
	if err != nil {
		return nil, err
	}
	st := &Status{Branch: branch, Detached: branch == "", HeadHash: head}
	if branch == "" && !head.IsZero() {
		st.Detached = true
	}

	// Cây của HEAD (có thể rỗng nếu chưa có commit).
	headTree := object.ZeroHash
	if !head.IsZero() {
		headTree, err = r.CommitTree(head)
		if err != nil {
			return nil, err
		}
	}
	headNodes, err := r.Flatten(headTree)
	if err != nil {
		return nil, err
	}

	// Cây của index, bỏ qua các entry xung đột.
	indexNodes := map[string]TreeNode{}
	for _, e := range r.Index.Entries() {
		if e.Stage != index.StageNormal {
			continue
		}
		indexNodes[e.Name] = IndexTreeNode(e)
	}

	// File thực tế trên đĩa.
	workNodes, err := r.ScanWorktree()
	if err != nil {
		return nil, err
	}

	allPaths := map[string]bool{}
	for p := range headNodes {
		allPaths[p] = true
	}
	for p := range indexNodes {
		allPaths[p] = true
	}
	for _, n := range workNodes {
		allPaths[n.Path] = true
	}
	paths := make([]string, 0, len(allPaths))
	for p := range allPaths {
		paths = append(paths, p)
	}
	sort.Strings(paths)

	seen := map[string]bool{}
	for _, p := range paths {
		if seen[p] {
			continue
		}
		seen[p] = true
		h, hasH := headNodes[p]
		i, hasI := indexNodes[p]
		var w TreeNode
		hasW := false
		for _, n := range workNodes {
			if n.Path == p {
				w = n
				hasW = true
				break
			}
		}

		fs := FileStatus{Path: p, IndexMode: i.Mode, IndexHash: i.Hash, WorkMode: w.Mode}

		switch {
		case !hasH && hasI:
			fs.IndexStatus = 'A' // file mới được stage
		case hasH && !hasI:
			fs.IndexStatus = 'D' // file bị xóa trong index
		case hasH && hasI && (h.Hash != i.Hash || h.Mode != i.Mode):
			fs.IndexStatus = 'M' // nội dung khác HEAD
		default:
			fs.IndexStatus = ' '
		}

		switch {
		case hasI && !hasW:
			fs.WorkStatus = 'D'
		case hasI && hasW:
			// So sánh nội dung trên đĩa với blob trong index.
			wh, err := r.blobHashOfWorktree(p, w.Mode)
			if err != nil {
				return nil, err
			}
			fs.WorkHash = wh
			if wh != i.Hash || w.Mode != i.Mode {
				fs.WorkStatus = 'M'
			} else {
				fs.WorkStatus = ' '
			}
		case !hasI && hasW:
			fs.IsUntracked = true
			fs.WorkStatus = '?'
			if wh, err := r.blobHashOfWorktree(p, w.Mode); err == nil {
				fs.WorkHash = wh
			}
		default:
			fs.WorkStatus = ' '
		}
		st.Entries = append(st.Entries, fs)
	}

	st.Conflicts = r.Index.Conflicts()
	if info, err := r.UpstreamInfo(); err == nil && info != nil {
		st.HasUpstream = true
		st.Ahead, st.Behind = r.AheadBehind(head, info.hash)
	}
	return st, nil
}

// blobHashOfWorktree tính hash blob cho nội dung file trên đĩa mà không ghi vào store.
func (r *Repo) blobHashOfWorktree(rel string, mode object.FileMode) (object.Hash, error) {
	abs := r.WorkPath(rel)
	data, err := worktree.ReadFileSymlinkAware(abs, mode)
	if err != nil {
		return object.ZeroHash, err
	}
	return object.ComputeHash(object.TypeBlob, data), nil
}

// DiffWorktreeIndex tính diff nội dung giữa index và worktree cho một file.
func (r *Repo) DiffWorktreeIndex(rel string) (oldLines, newLines []string, binaryFlag bool, err error) {
	e := r.Index.Get(rel)
	if e == nil {
		data, rerr := readWorktreeFile(r, rel)
		if rerr != nil {
			return nil, nil, false, rerr
		}
		if diff.IsBinary(data) {
			return nil, nil, true, nil
		}
		return nil, splitLines(string(data)), false, nil
	}
	oldData, err := r.Objects.ReadBlob(e.Hash)
	if err != nil {
		return nil, nil, false, err
	}
	newData, err := readWorktreeFile(r, rel)
	if err != nil {
		return nil, nil, false, err
	}
	if diff.IsBinary(oldData) || diff.IsBinary(newData) {
		return nil, nil, true, nil
	}
	return splitLines(string(oldData)), splitLines(string(newData)), false, nil
}

// DiffTreesContent tính diff nội dung giữa hai tree cho các file đã đổi.
// Trả về map đường dẫn -> (cũ, mới, là file nhị phân).
func (r *Repo) DiffTreesContent(from, to object.Hash) (map[string]*ContentDiff, error) {
	changes, err := r.DiffTrees(from, to)
	if err != nil {
		return nil, err
	}
	out := map[string]*ContentDiff{}
	for _, c := range changes {
		cd := &ContentDiff{Path: c.Path, Status: c.Status}
		if !c.Old.Hash.IsZero() {
			oldData, err := r.Objects.ReadBlob(c.Old.Hash)
			if err != nil {
				return nil, err
			}
			cd.OldData = oldData
		}
		if !c.New.Hash.IsZero() {
			newData, err := r.Objects.ReadBlob(c.New.Hash)
			if err != nil {
				return nil, err
			}
			cd.NewData = newData
		}
		cd.Binary = (diff.IsBinary(cd.OldData) && len(cd.OldData) > 0) ||
			(diff.IsBinary(cd.NewData) && len(cd.NewData) > 0)
		if !cd.Binary {
			cd.OldLines = splitLines(string(cd.OldData))
			cd.NewLines = splitLines(string(cd.NewData))
		}
		out[c.Path] = cd
	}
	return out, nil
}

// ContentDiff là nội dung diff của một file.
type ContentDiff struct {
	Path     string
	Status   byte
	OldData  []byte
	NewData  []byte
	OldLines []string
	NewLines []string
	Binary   bool
}

// UpstreamInfo là thông tin nhánh theo dõi của nhánh hiện tại.
type UpstreamInfo struct {
	Name string
	hash object.Hash
}

// UpstreamInfo đọc cấu hình branch.<tên>.merge của nhánh hiện tại.
func (r *Repo) UpstreamInfo() (*UpstreamInfo, error) {
	branch, err := r.CurrentBranch()
	if err != nil || branch == "" {
		return nil, err
	}
	mergeRef := r.Config.GetString("branch."+branch+".merge", "")
	if mergeRef == "" {
		return nil, nil
	}
	h, err := r.Refs.Resolve(mergeRef)
	if err != nil {
		// Nhánh theo dõi có thể chưa tồn tại, coi như chưa cấu hình.
		return nil, nil
	}
	return &UpstreamInfo{Name: storage.ShortName(mergeRef), hash: h}, nil
}

// SetUpstream ghi nhánh theo dõi cho một nhánh cục bộ.
func (r *Repo) SetUpstream(branch, mergeRef, short string) error {
	r.Config.Set("branch."+branch+".remote", short)
	r.Config.Set("branch."+branch+".merge", mergeRef)
	return r.Config.Save()
}

// UnsetUpstream gỡ cấu hình theo dõi.
func (r *Repo) UnsetUpstream(branch string) error {
	r.Config.Unset("branch." + branch + ".remote")
	r.Config.Unset("branch." + branch + ".merge")
	return r.Config.Save()
}

// AheadBehind đếm số commit tiến/lùi giữa hai điểm.
func (r *Repo) AheadBehind(local, remote object.Hash) (ahead, behind int) {
	if local.IsZero() || remote.IsZero() {
		return 0, 0
	}
	localSet, err := r.ancestorSet(local)
	if err != nil {
		return 0, 0
	}
	remoteSet, err := r.ancestorSet(remote)
	if err != nil {
		return 0, 0
	}
	for h := range localSet {
		if !remoteSet[h] {
			ahead++
		}
	}
	for h := range remoteSet {
		if !localSet[h] {
			behind++
		}
	}
	return ahead, behind
}

// FindCommonBase là wrapper mỏng cho MergeBase.
func (r *Repo) FindCommonBase(a, b object.Hash) (object.Hash, error) { return r.MergeBase(a, b) }

// readWorktreeFile đọc nội dung file trên đĩa, tự nhận diện chế độ file.
func readWorktreeFile(r *Repo, rel string) ([]byte, error) {
	abs := r.WorkPath(rel)
	mode := object.ModeBlob
	if fi, err := os.Lstat(abs); err == nil {
		mode = worktree.FileModeFromInfo(fi)
	}
	return worktree.ReadFileSymlinkAware(abs, mode)
}

// readWorktreeData đọc nội dung file trên đĩa theo chế độ đã biết trước.
func readWorktreeData(r *Repo, rel string, mode object.FileMode) ([]byte, error) {
	return worktree.ReadFileSymlinkAware(r.WorkPath(rel), mode)
}
