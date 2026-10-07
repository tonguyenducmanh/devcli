package cmd

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tonguyenducmanh/devcli/internal/env"
)

var useCmd = &cobra.Command{
	Use:     "use [tên nhóm]",
	GroupID: groupGeneral,
	Short:   "Đặt nhóm công cụ mặc định",
	Long: `Đặt một nhóm công cụ làm mặc định để tiết kiệm thời gian gõ lệnh.

Ví dụ, thay vì lúc nào cũng phải gõ 'td vcs status' hay 'td vcs commit', bạn
chỉ cần đặt nhóm mặc định là 'vcs' bằng lệnh 'td use vcs'. Từ đó, mọi lệnh gọi
không thuộc nhóm nào khác (như 'td status', 'td commit') sẽ tự động được
chuyển hướng sang nhóm 'vcs'.

Để tắt tính năng này và quay về trạng thái bình thường, hãy chạy 'td use'
không kèm theo đối số nào.`,
	Example: `  td use vcs       đặt nhóm 'vcs' làm mặc định
  td status        lệnh này giờ sẽ tương đương với 'td vcs status'
  td use           gỡ bỏ nhóm mặc định, mọi lệnh phải ghi rõ nhóm`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		p, err := defaultGroupPath()
		if err != nil {
			return err
		}
		if len(args) == 0 {
			if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
				return err
			}
			printLine("Đã gỡ thiết lập nhóm mặc định")
			return nil
		}

		target := args[0]
		found := false
		for _, g := range toolGroups {
			if g.Name() == target || hasAlias(g, target) {
				found = true
				break
			}
		}
		if !found {
			return exitError("không tìm thấy nhóm công cụ %q", target)
		}

		if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(p, []byte(target), 0644); err != nil {
			return err
		}
		printLine("Đã đặt nhóm mặc định là '%s'", target)
		return nil
	},
}

func hasAlias(cmd *cobra.Command, name string) bool {
	for _, a := range cmd.Aliases {
		if a == name {
			return true
		}
	}
	return false
}

// defaultGroupPath trả về đường dẫn tới tệp lưu nhóm mặc định.
func defaultGroupPath() (string, error) {
	dir, err := env.GlobalDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "default_group"), nil
}

func getDefaultGroup() string {
	p, err := defaultGroupPath()
	if err != nil {
		return ""
	}
	b, err := os.ReadFile(p)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

func init() {
	rootCmd.AddCommand(useCmd)
}
