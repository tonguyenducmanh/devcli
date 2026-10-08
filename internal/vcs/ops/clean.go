package ops

import (
	"fmt"

	"github.com/tonguyenducmanh/devcli/internal/vcs/repo"
	"github.com/tonguyenducmanh/devcli/internal/vcs/worktree"
)

// CleanOptions điều khiển lệnh xoá tệp chưa được theo dõi.
type CleanOptions struct {
	// Paths giới hạn vào các tệp này, rỗng nghĩa là xoá tất cả.
	Paths []string
	// DryRun chỉ liệt kê mà không xoá.
	DryRun bool
}

// CleanResult là kết quả cho một đường dẫn được yêu cầu.
type CleanResult struct {
	Path string
	// Removed là những tệp thực sự bị xoá khỏi đĩa, hoặc sẽ bị xoá nếu DryRun.
	Removed []string
	// Skipped là những tệp không xoá được, kèm lý do.
	Skipped []string
}

// Clean xoá tệp chưa được tm theo dõi khỏi cây làm việc.
//
// Chỉ tệp chưa từng được đưa vào vùng chuẩn bị mới bị xoá. Tệp đã được theo dõi
// thì thuộc về lịch sử nên lệnh này không đụng tới, muốn bỏ theo dõi thì dùng
// `tm vcs rm`. Nhờ vậy lệnh không bao giờ làm mất thứ còn cứu được trong kho.
//
// Không có đối số thì xoá mọi tệp chưa theo dõi. Đối số là tệp, thư mục hoặc mẫu
// có dấu * và ?, khớp với cách `tm vcs add` hiểu đường dẫn.
func Clean(r *repo.Repo, opts CleanOptions) ([]CleanResult, error) {
	st, err := r.Status()
	if err != nil {
		return nil, err
	}

	var results []CleanResult
	for _, e := range st.Untracked() {
		// Không có đường dẫn nào nghĩa là áp dụng cho mọi tệp chưa theo dõi.
		if len(opts.Paths) > 0 && !matchesAnyPath(e.Path, opts.Paths) {
			continue
		}
		result := CleanResult{Path: e.Path}
		if !opts.DryRun {
			if err := worktree.RemoveFile(r.WorkPath(e.Path), r.Root); err != nil {
				result.Skipped = append(result.Skipped, fmt.Sprintf("%s: %v", e.Path, err))
				results = append(results, result)
				continue
			}
		}
		result.Removed = append(result.Removed, e.Path)
		results = append(results, result)
	}
	return results, nil
}

// CountCleanable đếm số tệp chưa theo dõi mà lệnh sẽ đụng tới.
//
// Lệnh dùng nó để báo trước còn bao nhiêu tệp trước khi xoá, và để chặn trường
// hợp xoá nhầm hết sạch tệp trong một kho lớn.
func CountCleanable(r *repo.Repo, paths []string) (int, error) {
	results, err := Clean(r, CleanOptions{Paths: paths, DryRun: true})
	if err != nil {
		return 0, err
	}
	total := 0
	for _, result := range results {
		total += len(result.Removed)
	}
	return total, nil
}
