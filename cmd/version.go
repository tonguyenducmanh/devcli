package cmd

import (
	"runtime"

	"github.com/spf13/cobra"
)

// newVersionCmd tạo lệnh hiển thị phiên bản.
func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Hiển thị phiên bản của td",
		Long: `In số phiên bản, hệ điều hành và kiến trúc máy đang chạy.

Thông tin này hữu ích khi báo lỗi, giúp biết bản dựng được biên dịch từ
phiên bản nào.`,
		Example: `  td version`,
		Args:    noArgsArg,
		RunE: func(cmd *cobra.Command, args []string) error {
			printLine("td phiên bản %s", Version)
			printLine("nền tảng: %s/%s, %s", runtime.GOOS, runtime.GOARCH, runtime.Version())
			return nil
		},
	}
}
