package cmd

import (
	"github.com/spf13/cobra"
)

// newVCSCmd tạo nhóm lệnh quản lý phiên bản.
// Các lệnh con được khai báo trong các tệp cmd/vcs_*.go nên dễ bổ sung lệnh mới.
func newVCSCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "vcs",
		GroupID: groupVersionControl,
		Short:   "Quản lý phiên bản mã nguồn cục bộ",
		Long: `Nhóm lệnh quản lý phiên bản mã nguồn cục bộ.

Toàn bộ dữ liệu của nhóm này nằm trong thư mục .tdx cạnh dự án, gồm lịch
sử commit, các nhánh, các tag và vùng chuẩn bị.

Mỗi lệnh dưới đây chạy trên kho tìm thấy bằng cách đi lên từ thư mục hiện
tại. Dùng -C để chỉ định thư mục khác.`,
		Example: `  # Khởi tạo kho rồi ghi lại thay đổi đầu tiên
  td vcs init
  td vcs add .
  td vcs commit -m "tin nhắn đầu tiên"

  # Xem nhánh hiện tại và lịch sử gọn
  td vcs status
  td vcs log --oneline -n 10

  # Tạo nhánh, làm việc rồi hợp nhất về nhánh chính
  td vcs switch -c tinh-nang
  td vcs commit -am "bổ sung tính năng"
  td vcs switch main
  td vcs merge tinh-nang`,
		// Gọi nhóm lệnh mà không kèm lệnh con thì in danh sách lệnh con,
		// thay vì in cả trang trợ giúp dài.
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				return exitError("không có lệnh nào tên %q, xem danh sách: %s vcs --help", args[0], AppName)
			}
			printCommandList(cmd)
			return nil
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
