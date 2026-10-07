package ops

import (
	"fmt"

	"github.com/tonguyenducmanh/devcli/internal/vcs/object"
	"github.com/tonguyenducmanh/devcli/internal/vcs/repo"
)

// HashWorktreeFile tính hash blob của một file trên đĩa mà không ghi vào kho.
func HashWorktreeFile(r *repo.Repo, rel string) (object.Hash, error) {
	if !existsInWorktree(r, rel) {
		return object.ZeroHash, fmt.Errorf("không tìm thấy %s", rel)
	}
	abs := r.WorkPath(rel)
	fi, err := lstatPath(abs)
	if err != nil {
		return object.ZeroHash, err
	}
	data, err := readWorktreeData(abs, fileModeFromInfo(fi))
	if err != nil {
		return object.ZeroHash, err
	}
	return object.ComputeHash(object.TypeBlob, data), nil
}

// CatFile in nội dung của một object theo loại được yêu cầu.
func CatFile(r *repo.Repo, kind, rev string, verbose bool) error {
	h, err := resolveCommitish(r, rev)
	if err != nil {
		return err
	}
	actualType, data, err := r.Objects.Read(h)
	if err != nil {
		return err
	}
	if kind != "" && kind != string(actualType) {
		return fmt.Errorf("%s không phải object loại %s", h.Short(12), kind)
	}

	switch actualType {
	case object.TypeBlob:
		// Với blob chỉ hiện hash, nội dung thô khi -v được dùng.
		if !verbose {
			printOut("%s\n", h)
			return nil
		}
		printOut("%s", string(data))
	case object.TypeTree:
		tree, err := object.DecodeTree(data)
		if err != nil {
			return err
		}
		if verbose {
			for _, e := range tree.Entries {
				kind := "blob"
				if e.Mode.IsTree() {
					kind = "tree"
				}
				printOut("%s %s %s\t%s\n", e.Mode, kind, e.Hash, e.Name)
			}
			return nil
		}
		printOut("%s\n", h)
	case object.TypeCommit:
		if verbose {
			printOut("%s", string(data))
			return nil
		}
		printOut("%s\n", h)
	case object.TypeTag:
		if verbose {
			printOut("%s", string(data))
			return nil
		}
		printOut("%s\n", h)
	default:
		return fmt.Errorf("loại object không hỗ trợ: %s", actualType)
	}
	return nil
}

// FsckReport là kết quả kiểm tra toàn vẹn kho dữ liệu.
type FsckReport struct {
	ObjectCount int
	RefCount    int
	Problems    []string
}

// Fsck kiểm tra xem mọi ref có trỏ tới object hợp lệ và object có đọc được không.
func Fsck(r *repo.Repo) (*FsckReport, error) {
	report := &FsckReport{}

	// Kiểm tra tất cả object trong kho có đọc được hay không.
	if err := r.Objects.Each(func(h object.Hash) error {
		report.ObjectCount++
		if _, _, err := r.Objects.Read(h); err != nil {
			report.Problems = append(report.Problems, fmt.Sprintf("object %s không đọc được: %v", h.Short(12), err))
		}
		return nil
	}); err != nil {
		return nil, err
	}

	// Kiểm tra các ref trỏ tới commit hợp lệ.
	refs, err := r.Refs.List("refs/")
	if err != nil {
		return nil, err
	}
	for _, ref := range refs {
		report.RefCount++
		h, err := r.Refs.Resolve(ref)
		if err != nil {
			report.Problems = append(report.Problems, fmt.Sprintf("ref %s không đọc được: %v", ref, err))
			continue
		}
		if !r.Objects.Has(h) {
			report.Problems = append(report.Problems, fmt.Sprintf("ref %s trỏ tới object không tồn tại %s", ref, h.Short(12)))
		}
	}
	if head, err := r.Head(); err == nil && !head.IsZero() {
		if !r.Objects.Has(head) {
			report.Problems = append(report.Problems, fmt.Sprintf("HEAD trỏ tới object không tồn tại %s", head.Short(12)))
		}
	}
	return report, nil
}
