package cmd

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/tonguyenducmanh/devcli/internal/vcs/config"
)

// newConfigCmd tạo nhóm lệnh cấu hình.
func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Xem và chỉnh sửa cấu hình của td",
		Long: `Cấu hình được lưu ở hai nơi: file toàn cục và file trong từng repo.
Cấu hình toàn cục được đọc trước, nên nên đặt user.name và user.email ở đó.

  td config --global user.name "Tên của tôi"
  td config user.email "ten@example.com"
  td config --list`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfig(cmd, args)
		},
	}
	cmd.PersistentFlags().StringP("dir", "C", "", "chạy lệnh tại thư mục khác")
	cmd.Flags().BoolP("global", "g", false, "áp dụng cho toàn bộ máy thay vì repo hiện tại")
	cmd.Flags().BoolP("list", "l", false, "liệt kê toàn bộ cấu hình")
	cmd.Flags().Bool("unset", false, "xóa một khóa cấu hình")
	return cmd
}

// runConfig thực thi thao tác đọc hoặc ghi cấu hình.
func runConfig(cmd *cobra.Command, args []string) error {
	global, _ := cmd.Flags().GetBool("global")
	listFlag, _ := cmd.Flags().GetBool("list")
	unsetFlag, _ := cmd.Flags().GetBool("unset")

	// Cấu hình toàn cục không cần mở repo.
	if global {
		path, err := config.GlobalPath()
		if err != nil {
			return err
		}
		cfg, err := config.Load(path)
		if err != nil {
			return err
		}
		return applyConfig(cmd, cfg, listFlag, unsetFlag, args, path)
	}

	// Cấu hình trong repo.
	r, err := openRepo(cmd)
	if err != nil {
		return err
	}
	return applyConfig(cmd, r.Config, listFlag, unsetFlag, args, r.Config.Path())
}

// applyConfig áp dụng thao tác lên một đối tượng cấu hình.
func applyConfig(cmd *cobra.Command, cfg *config.Config, listFlag, unsetFlag bool, args []string, path string) error {
	switch {
	case listFlag || len(args) == 0:
		for _, k := range cfg.Keys() {
			v, _ := cfg.Get(k)
			printLine("%s=%s", k, v)
		}
		printLine("\n(nguồn: %s)", path)
		return nil

	case unsetFlag:
		if len(args) != 1 {
			return exitError("cần chỉ định đúng một khóa cần xóa")
		}
		cfg.Unset(args[0])
		return cfg.Save()

	default:
		// Chấp nhận cả hai dạng: `config user.name="Tên"` và
		// `config user.name "Tên có dấu cách"`.
		key, value, ok := splitConfigArg(args)
		if !ok {
			// Không có giá trị mới, nghĩa là chỉ đọc.
			if _, found := cfg.Get(key); !found {
				return exitError("không tìm thấy khóa cấu hình %s", key)
			}
			v, _ := cfg.Get(key)
			printLine("%s", v)
			return nil
		}
		cfg.Set(key, value)
		if err := cfg.Save(); err != nil {
			return err
		}
		if verboseEnabled(cmd) {
			printLine("đã lưu %s vào %s", key, path)
		}
		return nil
	}
}

// splitConfigArg tách đối số của lệnh config thành khóa và giá trị.
// Ưu tiên dấu "=" nếu có, nếu không thì lấy tham số đầu làm khóa
// và nối phần còn lại làm giá trị để giữ được dấu cách.
func splitConfigArg(args []string) (key, value string, ok bool) {
	joined := strings.Join(args, " ")
	if k, v, found := strings.Cut(joined, "="); found {
		return strings.TrimSpace(k), strings.TrimSpace(v), true
	}
	if len(args) == 0 {
		return "", "", false
	}
	if len(args) == 1 {
		return args[0], "", false
	}
	return args[0], strings.Join(args[1:], " "), true
}
