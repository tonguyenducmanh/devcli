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
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			printLine("td phiên bản %s", Version)
			printLine("nền tảng: %s/%s, %s", runtime.GOOS, runtime.GOARCH, runtime.Version())
			return nil
		},
	}
}
