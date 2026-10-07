// Package cmd định nghĩa cây lệnh của td bằng cobra.
//
// Cấu trúc lệnh được tổ chức theo nhóm nghiệp vụ. Mỗi nhóm là một lệnh cha
// chứa các lệnh con của nó, giúp dễ mở rộng thêm công cụ mới trong tương lai
// mà không phải sửa cấu trúc chung.
package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tonguyenducmanh/devcli/internal/vcs/repo"
)

// Version là phiên bản hiển thị qua lệnh `td version`.
//
// Khai báo bằng var (không phải const) để script build ghi đè được giá trị
// này bằng cờ -ldflags "-X ...cmd.Version=<số phiên bản>" lúc biên dịch.
var Version = "0.1.0"

// Tên nhóm lệnh, dùng để gom lệnh trong phần trợ giúp.
const (
	groupVersionControl = "version-control"
	groupGeneral        = "general"
)

// rootCmd là lệnh gốc của ứng dụng.
var rootCmd = &cobra.Command{
	Use:   "td",
	Short: "td - bộ công cụ dòng lệnh cá nhân",
	Long: `td là bộ công cụ dòng lệnh cá nhân, tự quản lý toàn bộ dữ liệu của nó.

Mỗi nhóm công cụ là một lệnh con của td, ví dụ:
  td vcs ...     quản lý phiên bản mã nguồn cục bộ

Dữ liệu của td được lưu trong thư mục .tdx cạnh dự án.`,
	Example: `  td vcs init                     khởi tạo kho tại thư mục hiện tại
  td vcs status                   xem các thay đổi chưa commit
  td vcs commit -m "tin nhắn"     ghi lại thay đổi
  td config --list                xem cấu hình đang dùng
  td --help                       xem toàn bộ lệnh`,
	SilenceUsage:  true,
	SilenceErrors: true,
	// Khi không có lệnh con nào được gọi thì in phần trợ giúp.
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			return exitError("lệnh không tồn tại: %q", args[0])
		}
		return cmd.Help()
	},
	// Chạy trước mọi lệnh con để quyết định có tô màu output hay không.
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		setupColor(cmd)
	},
}

// Root trả về lệnh gốc để các công cụ sinh tài liệu có thể duyệt cây lệnh.
func Root() *cobra.Command { return rootCmd }

// Execute chạy lệnh gốc, trả về lỗi nếu có.
func Execute() error {
	return rootCmd.Execute()
}

// setupColor quyết định có tô màu output dựa trên cờ và khả năng của terminal.
func setupColor(cmd *cobra.Command) {
	noColor, _ := cmd.Flags().GetBool("no-color")
	colorEnabled = !noColor && isTerminal()
}

// verboseEnabled báo cờ -v có đang bật không.
func verboseEnabled(cmd *cobra.Command) bool {
	v, _ := cmd.Flags().GetBool("verbose")
	return v
}

// errRepoNotFound trả về thông báo gợi ý khi lệnh cần repo nhưng không tìm thấy.
func errRepoNotFound(err error) error {
	if errors.Is(err, repo.ErrNotRepo) {
		return fmt.Errorf("chưa có kho td nào ở đây, hãy chạy `td vcs init` để khởi tạo")
	}
	return err
}

func init() {
	rootCmd.SetVersionTemplate("td phiên bản {{.Version}}\n")
	rootCmd.Version = Version

	// Cờ toàn cục áp dụng cho mọi lệnh.
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "in thêm thông tin chi tiết")
	rootCmd.PersistentFlags().Bool("no-color", false, "tắt màu trong output")

	// Ẩn lệnh sinh tự động của cobra vì không dùng đến.
	rootCmd.CompletionOptions.HiddenDefaultCmd = true

	// Đăng ký các nhóm công cụ.
	registerToolGroup(newVCSCmd())
	registerToolGroup(newConfigCmd())
	registerToolGroup(newVersionCmd())

	// Gắn nhóm cho các lệnh để phần trợ giúp gọn gàng hơn.
	rootCmd.AddGroup(
		&cobra.Group{ID: groupGeneral, Title: "Lệnh chung"},
		&cobra.Group{ID: groupVersionControl, Title: "Quản lý phiên bản"},
	)
}

// toolGroups lưu các nhóm công cụ đã đăng ký, dùng cho kiểm thử.
var toolGroups []*cobra.Command

// registerToolGroup đăng ký một nhóm công cụ vào lệnh gốc.
func registerToolGroup(group *cobra.Command) {
	toolGroups = append(toolGroups, group)
	rootCmd.AddCommand(group)
}
