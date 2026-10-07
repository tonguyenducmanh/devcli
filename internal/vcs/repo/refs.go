package repo

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/tonguyenducmanh/devcli/internal/vcs/index"
	"github.com/tonguyenducmanh/devcli/internal/vcs/object"
	"github.com/tonguyenducmanh/devcli/internal/vcs/worktree"
)

// Hàm wrapper bổ sung cho reflog, gom logic đọc/ghi về một chỗ.
func (r *Repo) reflogPath(ref string) string {
	return r.StatePath("logs/" + ref)
}

// AppendReflog ghi một dòng reflog cho ref.
func (r *Repo) AppendReflog(ref string, old, new object.Hash, msg string) error {
	return r.Refs.AppendReflog(ref, old, new, msg)
}

// UpdateBranch cập nhật một nhánh cục bộ kèm reflog.
func (r *Repo) UpdateBranch(name string, h object.Hash, msg string) error {
	return r.UpdateRefWithLog(headsDir+"/"+name, h, msg)
}

// BranchExists báo xem nhánh cục bộ đã tồn tại chưa.
func (r *Repo) BranchExists(name string) bool {
	return r.Refs.Exists(headsDir + "/" + name)
}

// BranchHash trả về hash commit đầu của nhánh.
func (r *Repo) BranchHash(name string) (object.Hash, error) {
	return r.Refs.Resolve(headsDir + "/" + name)
}

// CreateBranch tạo nhánh mới trỏ tới một commit.
func (r *Repo) CreateBranch(name, startPoint string, hash object.Hash) error {
	if r.BranchExists(name) {
		return fmt.Errorf("nhánh %s đã tồn tại", name)
	}
	return r.UpdateBranch(name, hash, fmt.Sprintf("branch: tạo từ %s", startPoint))
}

// DeleteBranch xóa nhánh, bắt buộc không được xóa nhánh đang ở.
func (r *Repo) DeleteBranch(name string, force bool) error {
	if name == "" {
		return fmt.Errorf("không xóa được nhánh hiện tại")
	}
	if !r.BranchExists(name) {
		return fmt.Errorf("không tìm thấy nhánh %s", name)
	}
	return r.Refs.Remove(headsDir + "/" + name)
}

// RenameBranch đổi tên nhánh cục bộ.
func (r *Repo) RenameBranch(old, newName string) error {
	h, err := r.BranchHash(old)
	if err != nil {
		return fmt.Errorf("không tìm thấy nhánh %s", old)
	}
	if r.BranchExists(newName) {
		return fmt.Errorf("nhánh %s đã tồn tại", newName)
	}
	if err := r.DeleteBranch(old, true); err != nil {
		return err
	}
	if err := r.UpdateBranch(newName, h, "branch: đổi tên"); err != nil {
		return err
	}
	// Nếu đang đứng trên nhánh cũ thì cập nhật HEAD.
	headRef, err := r.ReadHeadRef()
	if err == nil && headRef == headsDir+"/"+old {
		if err := r.SetHeadSymlink(headsDir + "/" + newName); err != nil {
			return err
		}
	}
	// Chuyển luôn cấu hình theo dõi nếu có.
	if r.Config.GetString("branch."+old+".merge", "") != "" {
		r.Config.Unset("branch." + old + ".remote")
		r.Config.Unset("branch." + old + ".merge")
		r.Config.Set("branch."+newName+".remote", r.Config.GetString("branch."+newName+".remote", "local"))
		r.Config.Set("branch."+newName+".merge", r.Config.GetString("branch."+newName+".merge", headsDir+"/"+newName))
		_ = r.Config.Save()
	}
	return nil
}

// TagInfo là thông tin một tag.
type TagInfo struct {
	Name   string
	Hash   object.Hash
	Target object.Hash
	// Annotated cho biết đây là tag có chú thích.
	Annotated bool
	Message   string
}

// Tags trả về danh sách tag kèm chi tiết.
func (r *Repo) ListTags() ([]TagInfo, error) {
	names, err := r.Tags()
	if err != nil {
		return nil, err
	}
	out := make([]TagInfo, 0, len(names))
	for _, n := range names {
		full := tagsDir + "/" + n
		h, err := r.Refs.Resolve(full)
		if err != nil {
			continue
		}
		info := TagInfo{Name: n, Hash: h, Target: h}
		// Tag có chú thích lưu nội dung dạng "object <hash>\ntype commit\n..."
		if data, err := os.ReadFile(r.reflogPath(full)); err == nil && len(data) > 0 {
			_ = data
		}
		if raw, err := r.readTagObject(h); err == nil {
			info.Annotated = true
			info.Target = raw.target
			info.Message = raw.message
		}
		out = append(out, info)
	}
	return out, nil
}

// tagObject là nội dung của tag có chú thích.
type tagObject struct {
	target  object.Hash
	message string
}

// readTagObject đọc object tag nếu tồn tại.
func (r *Repo) readTagObject(h object.Hash) (tagObject, error) {
	t, data, err := r.Objects.Read(h)
	if err != nil {
		return tagObject{}, err
	}
	if t != object.TypeTag {
		return tagObject{}, fmt.Errorf("không phải tag")
	}
	s := string(data)
	out := tagObject{}
	for _, line := range strings.Split(s, "\n") {
		if strings.HasPrefix(line, "object ") {
			if th, err := object.ParseHash(strings.TrimPrefix(line, "object ")); err == nil {
				out.target = th
			}
		}
		if strings.HasPrefix(line, "tag ") {
			continue
		}
	}
	// Message nằm sau dòng trống đầu tiên.
	if i := strings.Index(s, "\n\n"); i >= 0 {
		out.message = strings.TrimSpace(s[i+2:])
	}
	return out, nil
}

// WriteTag tạo tag nhẹ trỏ thẳng tới commit.
func (r *Repo) WriteTag(name string, target object.Hash) error {
	return r.Refs.Write(tagsDir+"/"+name, target)
}

// WriteAnnotatedTag tạo tag có chú thích và ghi object tag.
func (r *Repo) WriteAnnotatedTag(name string, target object.Hash, message string) error {
	id := r.Identity()
	var sb strings.Builder
	fmt.Fprintf(&sb, "object %s\n", target)
	fmt.Fprintf(&sb, "type commit\n")
	fmt.Fprintf(&sb, "tag %s\n", name)
	fmt.Fprintf(&sb, "tagger %s\n", id.Format())
	sb.WriteString("\n")
	sb.WriteString(message)
	if !strings.HasSuffix(message, "\n") {
		sb.WriteString("\n")
	}
	h, err := r.Objects.Write(object.TypeTag, []byte(sb.String()))
	if err != nil {
		return err
	}
	if err := r.Refs.Write(tagsDir+"/"+name, h); err != nil {
		return err
	}
	return r.Refs.AppendReflog(tagsDir+"/"+name, object.ZeroHash, h, "tag: "+message)
}

// DeleteTag xóa một tag.
func (r *Repo) DeleteTag(name string) error {
	if !r.Refs.Exists(tagsDir + "/" + name) {
		return fmt.Errorf("không tìm thấy tag %s", name)
	}
	return r.Refs.Remove(tagsDir + "/" + name)
}

// WriteIndexFromTree nạp toàn bộ nội dung của một tree vào index.
func (r *Repo) WriteIndexFromTree(tree object.Hash) error {
	r.Index.Clear()
	if tree.IsZero() {
		return r.SaveIndex()
	}
	nodes, err := r.Flatten(tree)
	if err != nil {
		return err
	}
	paths := make([]string, 0, len(nodes))
	for p := range nodes {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	for _, p := range paths {
		n := nodes[p]
		var size uint32
		if data, err := r.Objects.ReadBlob(n.Hash); err == nil {
			size = uint32(len(data))
		}
		// Cố gắng lấy thông tin stat từ file trên đĩa để tối ưu sau này.
		e := index.Entry{Mode: n.Mode, Size: size, Hash: n.Hash, Name: p}
		if fi, err := os.Lstat(r.WorkPath(p)); err == nil {
			e.CTime = worktree.StatTimes(fi)
			e.MTime = worktree.StatTimes(fi)
		}
		r.Index.Add(e)
	}
	return r.SaveIndex()
}
