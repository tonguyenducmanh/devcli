// Package cmd định nghĩa cây lệnh của td bằng cobra.
//
// Cấu trúc lệnh được tổ chức theo nhóm nghiệp vụ. Mỗi nhóm là một lệnh cha
// chứa các lệnh con của nó, giúp dễ mở rộng thêm công cụ mới trong tương lai
// mà không phải sửa cấu trúc chung.
package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/tonguyenducmanh/devcli/internal/vcs/repo"
)

// Phiên bản hiển thị qua lệnh `td version`.
const Version = "0.1.0"

// rootCmd là lệnh gốc của ứng dụng.
var rootCmd = &cobra.Command{
	Use:   "td",
	Short: "td - bộ công cụ dòng lệnh cá nhân",
	Long: `td là bộ công cụ dòng lệnh cá nhân, tự quản lý toàn bộ dữ liệu của nó.

Mỗi nhóm công cụ là một lệnh con của td, ví dụ:
  td vcs ...     quản lý phiên bản mã nguồn cục bộ
  td note ...    (dự phòng) ghi chú nhanh
  td task ...    (dự phòng) quản lý công việc

Dữ liệu của td được lưu trong thư mục .tdx cạnh dự án.`,
	SilenceUsage:  true,
	SilenceErrors: true,
	// Khi không có lệnh con nào được gọi thì in phần trợ giúp.
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			return fmt.Errorf("lệnh không tồn tại: %q", args[0])
		}
		return cmd.Help()
	},
}

// Execute chạy lệnh gốc, trả về lỗi nếu có.
func Execute() error {
	return rootCmd.Execute()
}

// Nhóm lệnh của từng công cụ, đăng ký trong init() để tự mở rộng.
var toolGroups []*cobra.Command

// registerToolGroup đăng ký một nhóm công cụ vào lệnh gốc.
func registerToolGroup(group *cobra.Command) {
	toolGroups = append(toolGroups, group)
	rootCmd.AddCommand(group)
}

func init() {
	rootCmd.SetVersionTemplate("td phiên bản {{.Version}}\n")
	rootCmd.Version = Version

	// Cờ toàn cục áp dụng cho mọi lệnh.
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "in thêm thông tin chi tiết")
	rootCmd.PersistentFlags().Bool("no-color", false, "tắt màu trong output")

	// Đăng ký các nhóm công cụ.
	registerToolGroup(newVCSCmd())
	registerToolGroup(newConfigCmd())
	registerToolGroup(newVersionCmd())
}

// verboseEnabled báo cờ -v có đang bật không.
func verboseEnabled(cmd *cobra.Command) bool {
	v, _ := cmd.Flags().GetBool("verbose")
	return v
}

// exitError trả về lỗi đã định dạng sẵn để in ra và thoát với mã 1.
func exitError(format string, args ...any) error {
	return fmt.Errorf(format, args...)
}

// errRepoNotFound trả về thông báo gợi ý khi lệnh cần repo nhưng không tìm thấy.
func errRepoNotFound(err error) error {
	if errors.Is(err, repo.ErrNotRepo) {
		return fmt.Errorf("chưa có kho td nào ở đây, hãy chạy `td vcs init` để khởi tạo")
	}
	return err
}

// printLine in một dòng ra stdout.
func printLine(format string, args ...any) {
	fmt.Fprintf(os.Stdout, format+"\n", args...)
}
