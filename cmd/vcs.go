package cmd

import (
	"github.com/spf13/cobra"
)

// newVCSCmd tạo nhóm lệnh quản lý phiên bản.
// Các lệnh con nằm trong thư mục cmd/vcs nên dễ bổ sung lệnh mới.
func newVCSCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vcs",
		Short: "Quản lý phiên bản mã nguồn cục bộ",
		Long: `Nhóm lệnh quản lý phiên bản mã nguồn cục bộ.
Toàn bộ dữ liệu của nhóm này nằm trong thư mục .tdx cạnh dự án.

Các lệnh thường dùng:
  td vcs init                    khởi tạo kho mã nguồn
  td vcs status                  xem trạng thái thay đổi
  td vcs add .                   đưa thay đổi vào vùng stage
  td vcs commit -m "..."         tạo commit
  td vcs log                     xem lịch sử`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	// Cờ -C dùng chung cho mọi lệnh trong nhóm.
	cmd.PersistentFlags().StringP("dir", "C", "", "chạy lệnh tại thư mục khác")

	cmd.AddCommand(
		vcsInitCmd,
		vcsStatusCmd,
		vcsAddCmd,
		vcsCommitCmd,
		vcsLogCmd,
		vcsDiffCmd,
		vcsShowCmd,
		vcsBranchCmd,
		vcsCheckoutCmd,
		vcsSwitchCmd,
		vcsRestoreCmd,
		vcsMergeCmd,
		vcsRebaseCmd,
		vcsCherryPickCmd,
		vcsRevertCmd,
		vcsStashCmd,
		vcsResetCmd,
		vcsTagCmd,
		vcsRmCmd,
		vcsMvCmd,
		vcsReflogCmd,
		vcsHashObjectCmd,
		vcsCatFileCmd,
		vcsFsckCmd,
	)
	return cmd
}
