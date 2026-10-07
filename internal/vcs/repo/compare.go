package repo

import (
	"os"

	"github.com/tonguyenducmanh/devcli/internal/vcs/diff"
	"github.com/tonguyenducmanh/devcli/internal/vcs/index"
	"github.com/tonguyenducmanh/devcli/internal/vcs/object"
	"github.com/tonguyenducmanh/devcli/internal/vcs/worktree"
)

// IndexVsTree so sánh nội dung index với một tree để biết thay đổi đã stage.
func (r *Repo) IndexVsTree(tree object.Hash) (map[string]*ContentDiff, error) {
	baseNodes, err := r.Flatten(tree)
	if err != nil {
		return nil, err
	}
	idxNodes := map[string]TreeNode{}
	for _, e := range r.Index.Entries() {
		if e.Stage != index.StageNormal {
			continue
		}
		idxNodes[e.Name] = IndexTreeNode(e)
	}
	return r.buildContentDiff(baseNodes, idxNodes, "index")
}

// WorktreeVsIndex so sánh nội dung trên đĩa với index để biết thay đổi chưa stage.
func (r *Repo) WorktreeVsIndex() (map[string]*ContentDiff, error) {
	idxNodes := map[string]TreeNode{}
	for _, e := range r.Index.Entries() {
		if e.Stage != index.StageNormal {
			continue
		}
		idxNodes[e.Name] = IndexTreeNode(e)
	}
	workNodes, err := r.ScanWorktree()
	if err != nil {
		return nil, err
	}
	workMap := map[string]TreeNode{}
	for _, n := range workNodes {
		workMap[n.Path] = n
	}
	return r.buildContentDiff(idxNodes, workMap, "worktree")
}

// buildContentDiff dựng nội dung diff từ hai tập node.
// side cho biết phía b nhân với nguồn nào để đọc dữ liệu cho đúng.
func (r *Repo) buildContentDiff(a, b map[string]TreeNode, side string) (map[string]*ContentDiff, error) {
	all := map[string]bool{}
	for p := range a {
		all[p] = true
	}
	for p := range b {
		all[p] = true
	}
	out := map[string]*ContentDiff{}
	for p := range all {
		old, hasOld := a[p]
		nw, hasNew := b[p]
		cd := &ContentDiff{Path: p}
		switch {
		case hasOld && hasNew:
			cd.Status = 'M'
		case hasNew:
			cd.Status = 'A'
		default:
			cd.Status = 'D'
		}
		if hasOld {
			data, err := r.Objects.ReadBlob(old.Hash)
			if err != nil {
				return nil, err
			}
			cd.OldData = data
		}
		if hasNew {
			if side == "worktree" {
				// File trên đĩa chưa có object nên phải đọc trực tiếp.
				abs := r.WorkPath(p)
				fi, err := os.Lstat(abs)
				if err != nil {
					continue // file biến mất trong lúc quét
				}
				data, err := worktree.ReadFileSymlinkAware(abs, worktree.FileModeFromInfo(fi))
				if err != nil {
					continue
				}
				cd.NewData = data
			} else {
				data, err := r.Objects.ReadBlob(nw.Hash)
				if err != nil {
					return nil, err
				}
				cd.NewData = data
			}
		}
		// Nội dung hai phía giống nhau thì không có gì để báo.
		if hasOld && hasNew && string(cd.OldData) == string(cd.NewData) {
			continue
		}
		cd.Binary = (len(cd.OldData) > 0 && diff.IsBinary(cd.OldData)) ||
			(len(cd.NewData) > 0 && diff.IsBinary(cd.NewData))
		if !cd.Binary {
			cd.OldLines = splitLines(string(cd.OldData))
			cd.NewLines = splitLines(string(cd.NewData))
		}
		out[p] = cd
	}
	return out, nil
}
