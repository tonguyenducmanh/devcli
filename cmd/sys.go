package cmd

import (
	"github.com/spf13/cobra"

	"github.com/tonguyenducmanh/devcli/internal/coreutils"
)

func newSysCmd() *cobra.Command {
	sysCmd := &cobra.Command{
		Use:     "sys",
		GroupID: groupSystem,
		Short:   "Các tiện ích hệ thống (ls, cat, head, tail)",
		Long:    `Nhóm lệnh chứa các tiện ích thao tác với tệp tin và hệ thống, mô phỏng lại các lệnh quen thuộc trên Linux.`,
		Example: `  td sys ls
  td sys cat file.txt
  td sys head -n 5 file.txt`,
	}

	lsCmd := &cobra.Command{
		Use:   "ls [thư mục]",
		Short: "Liệt kê các tệp tin trong thư mục",
		Long: `Liệt kê các tệp tin và thư mục con trong thư mục được chỉ định.
Nếu không truyền thư mục, mặc định sẽ liệt kê thư mục hiện tại.`,
		Example: `  td sys ls
  td sys ls -l
  td sys ls -a /tmp`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := "."
			if len(args) > 0 {
				dir = args[0]
			}
			all, _ := cmd.Flags().GetBool("all")
			long, _ := cmd.Flags().GetBool("long")

			return coreutils.Ls(cmd.OutOrStdout(), dir, all, long)
		},
	}
	lsCmd.Flags().BoolP("all", "a", false, "hiển thị cả tệp ẩn")
	lsCmd.Flags().BoolP("long", "l", false, "hiển thị chi tiết (quyền, kích thước, ngày giờ)")

	headCmd := &cobra.Command{
		Use:   "head [tệp]",
		Short: "In N dòng đầu tiên của tệp",
		Long: `Đọc và in ra N dòng đầu tiên của một tệp văn bản.
Mặc định in 10 dòng đầu tiên.`,
		Example: `  td sys head file.txt
  td sys head -n 5 file.txt`,
		Args: arbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return exitError("cần truyền đường dẫn tệp")
			}
			lines, _ := cmd.Flags().GetInt("lines")
			return coreutils.Head(cmd.OutOrStdout(), args[0], lines)
		},
	}
	headCmd.Flags().IntP("lines", "n", 10, "số dòng cần in")

	tailCmd := &cobra.Command{
		Use:   "tail [tệp]",
		Short: "In N dòng cuối cùng của tệp",
		Long: `Đọc và in ra N dòng cuối cùng của một tệp văn bản.
Mặc định in 10 dòng cuối cùng.`,
		Example: `  td sys tail file.txt
  td sys tail -n 5 file.txt`,
		Args: arbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return exitError("cần truyền đường dẫn tệp")
			}
			lines, _ := cmd.Flags().GetInt("lines")
			return coreutils.Tail(cmd.OutOrStdout(), args[0], lines)
		},
	}
	tailCmd.Flags().IntP("lines", "n", 10, "số dòng cần in")

	catCmd := &cobra.Command{
		Use:   "cat [tệp...]",
		Short: "In toàn bộ nội dung của tệp",
		Long: `Đọc và in ra toàn bộ nội dung của một hoặc nhiều tệp văn bản.
Nếu không truyền tệp nào hoặc truyền '-' hệ thống sẽ đọc từ đầu vào chuẩn (stdin).`,
		Example: `  td sys cat file.txt
  td sys cat file1.txt file2.txt`,
		Args: arbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return coreutils.Cat(cmd.OutOrStdout(), args)
		},
	}

	rmemptyCmd := &cobra.Command{
		Use:   "rmempty [thư mục]",
		Short: "Xoá các thư mục rỗng đệ quy",
		Long: `Tìm và xoá tất cả các thư mục rỗng bên trong thư mục được chỉ định.
Quá trình này được thực hiện đệ quy (xoá thư mục con rỗng, sau đó nếu thư mục cha rỗng thì xoá tiếp).
Mặc định sẽ quét thư mục hiện tại.`,
		Example: `  td sys rmempty
  td sys rmempty /tmp/test`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := "."
			if len(args) > 0 {
				dir = args[0]
			}
			return coreutils.RmEmpty(cmd.OutOrStdout(), dir)
		},
	}

	sysCmd.AddCommand(lsCmd, headCmd, tailCmd, catCmd, rmemptyCmd)

	return sysCmd
}
