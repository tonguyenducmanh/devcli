package repo

import (
	"fmt"

	"github.com/tonguyenducmanh/devcli/internal/vcs/object"
)

// UpdateHeadCommit cập nhật commit mà HEAD trỏ tới.
// Nếu HEAD là ref nhánh thì cập nhật cả ref nhánh và reflog tương ứng.
func (r *Repo) UpdateHeadCommit(h object.Hash, msg string) error {
	headRef, err := r.ReadHeadRef()
	if err != nil {
		return err
	}
	if headRef == "" {
		// HEAD đang tách rời: chỉ cập nhật file HEAD.
		old, _ := r.Head()
		if err := r.SetHeadDetached(h); err != nil {
			return err
		}
		return r.Refs.AppendReflog(headFile, old, h, msg)
	}
	if err := r.UpdateRefWithLog(headRef, h, msg); err != nil {
		return err
	}
	// Ghi thêm vào reflog HEAD để theo dõi được việc cập nhật nhánh.
	old, _ := r.Refs.Resolve(headRef)
	return r.Refs.AppendReflog(headFile, old, h, msg)
}

// UpdateHeadRef chuyển HEAD sang một nhánh khác, ghi reflog cho cả hai.
func (r *Repo) UpdateHeadRef(ref, msg string) error {
	oldHead, _ := r.Head()
	if err := r.SetHeadSymlink(ref); err != nil {
		return err
	}
	newHead, err := r.Refs.Resolve(ref)
	if err != nil {
		// Nhánh mới chưa có commit nào, reflog chỉ ghi giá trị rỗng.
		return r.Refs.AppendReflog(headFile, oldHead, object.ZeroHash, msg)
	}
	return r.Refs.AppendReflog(headFile, oldHead, newHead, msg)
}

// ReadReflog đọc reflog của một ref từ RefStore.
func (r *Repo) ReadReflog(name string) ([]reflogEntry, error) {
	entries, err := r.Refs.ReadReflog(name)
	if err != nil {
		return nil, err
	}
	out := make([]reflogEntry, 0, len(entries))
	for _, e := range entries {
		out = append(out, reflogEntry{Old: e.Old, New: e.New, Author: e.Author, Message: e.Message})
	}
	return out, nil
}

// reflogEntry là bản sao kiểu của storage.ReflogEntry để tránh phụ thuộc chéo.
type reflogEntry = struct {
	Old     object.Hash
	New     object.Hash
	Author  string
	Message string
}

// ReflogRefName trả về tên ref đầy đủ tương ứng với tên rút gọn.
func (r *Repo) ReflogRefName(name string) (string, error) {
	if name == "HEAD" || name == "head" {
		return headFile, nil
	}
	if len(name) >= 5 && name[:5] == "refs/" {
		return name, nil
	}
	if r.Refs.Exists(headsDir + "/" + name) {
		return headsDir + "/" + name, nil
	}
	if r.Refs.Exists(tagsDir + "/" + name) {
		return tagsDir + "/" + name, nil
	}
	return "", fmt.Errorf("không tìm thấy ref %q", name)
}
