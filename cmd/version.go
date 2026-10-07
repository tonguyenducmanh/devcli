package cmd

import (
	"github.com/spf13/cobra"
)

// newVersionCmd tạo lệnh hiển thị phiên bản.
func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "version",
		GroupID: groupGeneral,
		Short:   "Hiển thị phiên bản và thông tin môi trường",
		Long: `In số phiên bản, tên tác giả, hệ điều hành và kiến trúc máy đang chạy.

Thông tin này hữu ích khi báo lỗi, giúp biết bản dựng được biên dịch từ
phiên bản nào.`,
		Example: "  " + AppName + " version",
		Args:    noArgsArg,
		RunE: func(cmd *cobra.Command, args []string) error {
			printVersionFull()
			return nil
		},
	}
}
