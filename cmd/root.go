// Package cmd định nghĩa cây lệnh của tm bằng cobra.
//
// Cấu trúc lệnh được tổ chức theo nhóm nghiệp vụ. Mỗi nhóm là một lệnh cha
// chứa các lệnh con của nó, giúp dễ mở rộng thêm công cụ mới trong tương lai
// mà không phải sửa cấu trúc chung.
package cmd

import (
	"errors"
	"fmt"
	"os"
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
// Nguồn cấu hình nằm trong phần cấu hình của scripts/build_binaries.sh.
var (
	// AppName là tên gọi lệnh trên terminal, ví dụ "tm" trong "tm vcs status".
	AppName = "tm"

	// Version là phiên bản hiển thị qua lệnh `tm version`.
	Version = "0.0.0-dev"

	// Author là tên tác giả, in ở lệnh version và cuối phần trợ giúp.
	Author = "Tô Nguyễn Đức Mạnh"

	// RepoURL là nơi phát hành, in ở cuối phần trợ giúp.
	RepoURL = "github.com/tonguyenducmanh/devcli"
)

// Tên nhóm lệnh, dùng để gom lệnh trong phần trợ giúp.
const (
	groupVersionControl = "version-control"
	groupGeneral        = "general"
	groupSystem         = "system"
)

// rootCmd là lệnh gốc của ứng dụng.
var rootCmd = &cobra.Command{
	Use:   AppName,
	Short: "Bộ công cụ dòng lệnh cá nhân, tất cả gói dưới một lệnh duy nhất",
	Long: fmt.Sprintf(`%s gom mọi công cụ dòng lệnh bạn dùng hằng ngày dưới một lệnh duy nhất.

Mỗi nhóm công cụ là một lệnh con của %s, ví dụ:
  %s vcs ...     quản lý phiên bản mã nguồn cục bộ

Mỗi nhóm tự quản lý toàn bộ dữ liệu của nó trong thư mục riêng cạnh dự án, nên
%s không cần cài thêm hay cấu hình gì cả.`,
		AppName, AppName, AppName, AppName),
	Example: strings.Join([]string{
		"  " + AppName + " vcs init                   khởi tạo kho tại thư mục hiện tại",
		"  " + AppName + " vcs status                 xem các thay đổi chưa commit",
		"  " + AppName + " vcs commit -m \"tin nhắn\"   ghi lại thay đổi",
		"  " + AppName + " config --list              xem cấu hình đang dùng",
		"",
		"  " + AppName + "                            xem phiên bản và danh sách lệnh",
		"  " + AppName + " -v                         xem thêm thông tin môi trường",
		"  " + AppName + " --help                     xem trợ giúp đầy đủ",
	}, "\n"),
	SilenceUsage:  true,
	SilenceErrors: true,
	// Nhận mọi đối số rồi tự báo lỗi. Không khai báo Args thì cobra dùng
	// bộ kiểm tra riêng và in thông báo "unknown command" bằng tiếng Anh.
	Args: arbitraryArgs,
	// Gọi lệnh gốc mà không kèm lệnh con thì in tên, phiên bản và danh sách
	// lệnh, thay vì in cả trang trợ giúp dài.
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			return exitError("không có lệnh nào tên %q, xem danh sách: %s --help", args[0], AppName)
		}
		return runRoot(cmd)
	},
	// Chạy trước mọi lệnh con để quyết định có tô màu output hay không.
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		setupColor()
	},
}

// Root trả về lệnh gốc để các công cụ sinh tài liệu có thể duyệt cây lệnh.
func Root() *cobra.Command { return rootCmd }

// Execute chạy lệnh gốc, trả về lỗi nếu có.
func Execute() error {
	args := os.Args[1:]
	if len(args) > 0 {
		first := args[0]
		if !strings.HasPrefix(first, "-") {
			known := false
			for _, c := range rootCmd.Commands() {
				if c.Name() == first || hasAlias(c, first) {
					known = true
					break
				}
			}
			if !known {
				if def := getDefaultGroup(); def != "" {
					newArgs := append([]string{def}, args...)
					rootCmd.SetArgs(newArgs)
				}
			}
		}
	}
	return rootCmd.Execute()
}

// setupColor quyết định có tô màu output hay không, dựa vào khả năng của
// terminal và biến môi trường NO_COLOR.
func setupColor() {
	colorEnabled = isTerminal()
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

	// Cố ý không đặt cờ toàn cục nào. Mỗi lệnh tự khai báo cờ của riêng nó, và
	// -v mang nghĩa khác nhau tuỳ lệnh: `tm -v` là phiên bản, còn
	// `tm vcs branch -v` là hiện mã băm, đúng như git. Dùng chung một cờ
	// toàn cục sẽ khiến chữ viết tắt bị che trong lệnh con mà không ai hay.
	//
	// Tắt màu thì dùng biến môi trường NO_COLOR hoặc TERM=dumb, xem isTerminal.
	addVerboseFlag(rootCmd, "in thêm thông tin môi trường")

	// Tắt lệnh completion sinh tự động của cobra. Mô tả cờ của lệnh đó viết
	// bằng tiếng Anh, mà dự án này không bán tài liệu hoàn chỉnh nào khác nữa.
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	// Đăng ký các nhóm công cụ.
	registerToolGroup(newVCSCmd())
	registerToolGroup(newSysCmd())
	registerToolGroup(newConfigCmd())
	registerToolGroup(newVersionCmd())

	// Gắn nhóm cho các lệnh để phần trợ giúp gọn gàng hơn.
	rootCmd.AddGroup(
		&cobra.Group{ID: groupGeneral, Title: "Lệnh chung:"},
		&cobra.Group{ID: groupVersionControl, Title: "Quản lý phiên bản:"},
		&cobra.Group{ID: groupSystem, Title: "Tiện ích hệ thống:"},
	)

	// Dịch phần trợ giúp sang tiếng Việt, phải làm sau khi cây lệnh đã đủ.
	setupHelp()
}

// toolGroups lưu các nhóm công cụ đã đăng ký, dùng cho kiểm thử.
var toolGroups []*cobra.Command

// registerToolGroup đăng ký một nhóm công cụ vào lệnh gốc.
func registerToolGroup(group *cobra.Command) {
	toolGroups = append(toolGroups, group)
	rootCmd.AddCommand(group)
}
