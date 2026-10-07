// Package cmd định nghĩa cây lệnh của td bằng cobra.
//
// Cấu trúc lệnh được tổ chức theo nhóm nghiệp vụ. Mỗi nhóm là một lệnh cha
// chứa các lệnh con của nó, giúp dễ mở rộng thêm công cụ mới trong tương lai
// mà không phải sửa cấu trúc chung.
package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tonguyenducmanh/devcli/internal/vcs/repo"
)

// Các biến toàn cục của ứng dụng.
//
// Đều khai báo bằng `var` để script build ghi đè được giá trị lúc biên dịch
// bằng cờ -ldflags "-X ...cmd.<Tên>=<giá trị>". Giá trị dưới đây là giá trị
// mặc định khi chạy thẳng bằng `go build` mà không qua script.
//
// Nguồn cấu hình nằm trong scripts/build.conf.
var (
	// AppName là tên gọi lệnh trên terminal, ví dụ "td" trong "td vcs status".
	AppName = "td"

	// Version là phiên bản hiển thị qua lệnh `td version`.
	Version = "0.0.0-dev"

	// RepoURL là nơi phát hành, dùng để in gợi ý khi báo lỗi.
	// Để rỗng nghĩa là không in gợi ý.
	RepoURL = ""
)

// Tên nhóm lệnh, dùng để gom lệnh trong phần trợ giúp.
const (
	groupVersionControl = "version-control"
	groupGeneral        = "general"
)

// rootCmd là lệnh gốc của ứng dụng.
var rootCmd = &cobra.Command{
	Use:   AppName,
	Short: AppName + " - bộ công cụ dòng lệnh cá nhân",
	Long: fmt.Sprintf(`%s là bộ công cụ dòng lệnh cá nhân, tự quản lý toàn bộ dữ liệu của nó.

Mỗi nhóm công cụ là một lệnh con của %s, ví dụ:
  %s vcs ...     quản lý phiên bản mã nguồn cục bộ

Dữ liệu của %s được lưu trong thư mục .tdx cạnh dự án.`,
		AppName, AppName, AppName, AppName),
	Example: strings.Join([]string{
		"  " + AppName + " vcs init                   khởi tạo kho tại thư mục hiện tại",
		"  " + AppName + " vcs status                 xem các thay đổi chưa commit",
		"  " + AppName + " vcs commit -m \"tin nhắn\"   ghi lại thay đổi",
		"  " + AppName + " config --list              xem cấu hình đang dùng",
		"  " + AppName + " --help                     xem toàn bộ lệnh",
	}, "\n"),
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
		return fmt.Errorf("chưa có kho %s nào ở đây, hãy chạy `%s vcs init` để khởi tạo", AppName, AppName)
	}
	return err
}

func init() {
	rootCmd.SetVersionTemplate(AppName + " phiên bản {{.Version}}\n")
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
