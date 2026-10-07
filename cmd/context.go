package cmd

import (
	"github.com/spf13/cobra"

	"github.com/tonguyenducmanh/devcli/internal/vcs/repo"
)

// openRepo mở repo tại thư mục hiện tại, báo lỗi rõ ràng nếu chưa khởi tạo.
func openRepo(cmd *cobra.Command) (*repo.Repo, error) {
	// Tuỳ chọn -C cho phép chạy lệnh trên một thư mục khác.
	dir, err := cmd.Flags().GetString("dir")
	if err != nil || dir == "" {
		dir = "."
	}
	r, err := repo.Open(dir)
	if err != nil {
		return nil, errRepoNotFound(err)
	}
	return r, nil
}
