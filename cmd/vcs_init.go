package cmd

import (
	"github.com/spf13/cobra"

	"github.com/tonguyenducmanh/devcli/internal/vcs/ops"
	"github.com/tonguyenducmanh/devcli/internal/vcs/repo"
)

// vcsInitCmd khởi tạo kho mã nguồn mới.
var vcsInitCmd = &cobra.Command{
	Use:   "init [thư mục]",
	Short: "Khởi tạo kho mã nguồn td trong thư mục hiện tại",
	Long: `Tạo thư mục .tdx trong thư mục cho trước để bắt đầu theo dõi phiên bản.

Ví dụ:
  td vcs init
  td vcs init --initial-branch=develop`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := "."
		if len(args) == 1 {
			dir = args[0]
		}
		branch, err := cmd.Flags().GetString("initial-branch")
		if err != nil {
			return err
		}

		// Khởi tạo với nhánh mặc định là "main" nếu người dùng không chỉ định.
		r, err := repo.Init(dir, branch)
		if err != nil {
			return err
		}
		printLine("Đã khởi tạo kho mã nguồn td tại %s", r.Root)
		printLine("Nhánh mặc định: %s", branch)
		printLine("Bắt đầu bằng: td vcs add . && td vcs commit -m \"tin nhắn\"")
		return nil
	},
}

func init() {
	vcsInitCmd.Flags().String("initial-branch", "main", "tên nhánh khởi tạo")
}

// vcsHashObjectCmd in ra hash của một tệp, hữu ích để kiểm tra nội dung object.
var vcsHashObjectCmd = &cobra.Command{
	Use:   "hash-object <tệp>...",
	Short: "Tính và in hash của nội dung tệp",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := openRepo(cmd)
		if err != nil {
			return err
		}
		for _, a := range args {
			rel, err := r.RelPath(a)
			if err != nil {
				return err
			}
			h, err := ops.HashWorktreeFile(r, rel)
			if err != nil {
				return err
			}
			printLine("%s  %s", h, rel)
		}
		return nil
	},
}

// vcsCatFileCmd in nội dung của một object trong kho.
var vcsCatFileCmd = &cobra.Command{
	Use:   "cat-file <loại> <hash>",
	Short: "In nội dung của một object (blob, tree, commit)",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := openRepo(cmd)
		if err != nil {
			return err
		}
		return ops.CatFile(r, args[0], args[1], verboseEnabled(cmd))
	},
}

// vcsFsckCmd kiểm tra tính toàn vẹn của kho dữ liệu.
var vcsFsckCmd = &cobra.Command{
	Use:   "fsck",
	Short: "Kiểm tra tính toàn vẹn của kho object và ref",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := openRepo(cmd)
		if err != nil {
			return err
		}
		report, err := ops.Fsck(r)
		if err != nil {
			return err
		}
		if len(report.Problems) == 0 {
			printLine("Không phát hiện vấn đề nào (%d object, %d ref)", report.ObjectCount, report.RefCount)
			return nil
		}
		for _, p := range report.Problems {
			printLine("vấn đề: %s", p)
		}
		return exitError("phát hiện %d vấn đề", len(report.Problems))
	},
}
