package cmd

import (
	"github.com/spf13/cobra"

	"github.com/tonguyenducmanh/devcli/internal/sys"
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

			return sys.Ls(cmd.OutOrStdout(), dir, all, long)
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
			return sys.Head(cmd.OutOrStdout(), args[0], lines)
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
			return sys.Tail(cmd.OutOrStdout(), args[0], lines)
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
			return sys.Cat(cmd.OutOrStdout(), args)
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
			return sys.RmEmpty(cmd.OutOrStdout(), dir)
		},
	}

	pwdCmd := &cobra.Command{
		Use:     "pwd",
		Short:   "In đường dẫn thư mục hiện tại",
		Long:    `In ra đường dẫn tuyệt đối của thư mục làm việc hiện tại.`,
		Example: `  td sys pwd`,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return sys.Pwd(cmd.OutOrStdout())
		},
	}

	mkdirCmd := &cobra.Command{
		Use:   "mkdir [thư mục...]",
		Short: "Tạo thư mục mới",
		Long:  `Tạo một hoặc nhiều thư mục mới. Có thể sử dụng cờ -p để tạo đệ quy các thư mục cha nếu chưa tồn tại.`,
		Example: `  td sys mkdir testdir
  td sys mkdir -p a/b/c`,
		Args: arbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return exitError("cần truyền tên thư mục")
			}
			p, _ := cmd.Flags().GetBool("parents")
			return sys.Mkdir(cmd.OutOrStdout(), args, p)
		},
	}
	mkdirCmd.Flags().BoolP("parents", "p", false, "tạo đệ quy các thư mục cha")

	touchCmd := &cobra.Command{
		Use:     "touch [tệp...]",
		Short:   "Tạo tệp trống hoặc cập nhật thời gian",
		Long:    `Tạo một tệp tin trống nếu chưa tồn tại, hoặc cập nhật thời gian sửa đổi nếu đã tồn tại.`,
		Example: `  td sys touch file.txt`,
		Args:    arbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return exitError("cần truyền tên tệp")
			}
			return sys.Touch(cmd.OutOrStdout(), args)
		},
	}

	rmCmd := &cobra.Command{
		Use:   "rm [tệp...]",
		Short: "Xoá tệp hoặc thư mục",
		Long:  `Xoá một hoặc nhiều tệp. Sử dụng cờ -r để xoá thư mục đệ quy.`,
		Example: `  td sys rm file.txt
  td sys rm -rf dir`,
		Args: arbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return exitError("cần truyền đường dẫn")
			}
			r, _ := cmd.Flags().GetBool("recursive")
			f, _ := cmd.Flags().GetBool("force")
			return sys.Rm(cmd.OutOrStdout(), args, r, f)
		},
	}
	rmCmd.Flags().BoolP("recursive", "r", false, "xoá đệ quy thư mục")
	rmCmd.Flags().BoolP("force", "f", false, "bỏ qua các lỗi không tồn tại")

	cpCmd := &cobra.Command{
		Use:   "cp [nguồn...] [đích]",
		Short: "Sao chép tệp hoặc thư mục",
		Long:  `Sao chép các tệp tin hoặc thư mục từ nguồn đến đích.`,
		Example: `  td sys cp file1.txt file2.txt
  td sys cp -r dir1 dir2`,
		Args: arbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return exitError("cần ít nhất 2 đường dẫn (nguồn và đích)")
			}
			sources := args[:len(args)-1]
			target := args[len(args)-1]
			r, _ := cmd.Flags().GetBool("recursive")
			return sys.Cp(cmd.OutOrStdout(), sources, target, r)
		},
	}
	cpCmd.Flags().BoolP("recursive", "r", false, "sao chép đệ quy thư mục")

	mvCmd := &cobra.Command{
		Use:   "mv [nguồn...] [đích]",
		Short: "Di chuyển hoặc đổi tên tệp",
		Long:  `Di chuyển hoặc đổi tên các tệp tin, thư mục từ nguồn đến đích.`,
		Example: `  td sys mv file1.txt file2.txt
  td sys mv file1.txt dir/`,
		Args: arbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return exitError("cần ít nhất 2 đường dẫn (nguồn và đích)")
			}
			sources := args[:len(args)-1]
			target := args[len(args)-1]
			return sys.Mv(cmd.OutOrStdout(), sources, target)
		},
	}

	wcCmd := &cobra.Command{
		Use:   "wc [tệp...]",
		Short: "Đếm số dòng, từ, ký tự",
		Long:  `Đếm và in ra số dòng, số từ và số byte của các tệp tin.`,
		Example: `  td sys wc file.txt
  td sys wc -l file.txt`,
		Args: arbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return exitError("cần truyền đường dẫn tệp")
			}
			l, _ := cmd.Flags().GetBool("lines")
			wo, _ := cmd.Flags().GetBool("words")
			c, _ := cmd.Flags().GetBool("bytes")
			return sys.Wc(cmd.OutOrStdout(), args, l, wo, c)
		},
	}
	wcCmd.Flags().BoolP("lines", "l", false, "chỉ in số dòng")
	wcCmd.Flags().BoolP("words", "w", false, "chỉ in số từ")
	wcCmd.Flags().BoolP("bytes", "c", false, "chỉ in số byte")

	grepCmd := &cobra.Command{
		Use:     "grep [mẫu] [tệp...]",
		Short:   "Tìm kiếm văn bản trong tệp",
		Long:    `Tìm kiếm các chuỗi văn bản khớp với biểu thức chính quy trong các tệp tin.`,
		Example: `  td sys grep "hello" file.txt`,
		Args:    arbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return exitError("cần truyền mẫu và ít nhất 1 tệp")
			}
			pattern := args[0]
			files := args[1:]
			return sys.Grep(cmd.OutOrStdout(), pattern, files)
		},
	}

	sysCmd.AddCommand(lsCmd, headCmd, tailCmd, catCmd, rmemptyCmd, pwdCmd, mkdirCmd, touchCmd, rmCmd, cpCmd, mvCmd, wcCmd, grepCmd)

	return sysCmd
}
