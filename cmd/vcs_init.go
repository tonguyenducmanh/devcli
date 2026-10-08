package cmd

import (
	"github.com/spf13/cobra"

	"github.com/tonguyenducmanh/devcli/internal/vcs/ops"
	"github.com/tonguyenducmanh/devcli/internal/vcs/repo"
)

// vcsInitCmd khởi tạo kho mã nguồn mới.
var vcsInitCmd = &cobra.Command{
	Use:   "init [thư mục]",
	Short: "Khởi tạo kho mã nguồn td trong thư mục cho trước",
	Long: `Tạo thư mục .tdx trong thư mục cho trước để bắt đầu theo dõi phiên bản.

Tham số thư mục không bắt buộc, mặc định là thư mục hiện tại. Lệnh sẽ báo lỗi
nếu thư mục đó đã có kho, để tránh ghi đè dữ liệu đang có.`,
	Example: `  # Tạo kho trong thư mục hiện tại với nhánh main
  td vcs init

  # Tạo kho trong một thư mục khác với tên nhánh khác
  td vcs init du-an-cua-toi --initial-branch=develop`,
	Args: maximumArgs(1, "[thư mục]"),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := "."
		// -C áp dụng cho mọi lệnh nên init cũng phải nghe nó, nếu không thì
		// `td vcs init -C thu-muc-khac` lại khởi tạo nhầm ở thư mục hiện tại.
		if c, err := cmd.Flags().GetString("dir"); err == nil && c != "" {
			dir = c
		}
		if len(args) == 1 {
			dir = args[0]
		}
		branch, err := cmd.Flags().GetString("initial-branch")
		if err != nil {
			return err
		}
		if branch == "" {
			return exitError("tên nhánh khởi tạo không được để trống")
		}

		r, err := repo.Init(dir, branch)
		if err != nil {
			return err
		}
		printLine("Đã khởi tạo kho td tại %s", r.Root)
		printLine("Nhánh khởi tạo: %s", branch)
		printLine("Bước tiếp theo: td vcs add . && td vcs commit -m \"tin nhắn\"")
		return nil
	},
}

func init() {
	addVerboseFlag(vcsCatFileCmd, "in ra dạng đã rút gọn cho dễ đọc")

	vcsInitCmd.Flags().String("initial-branch", "main", "tên nhánh khởi tạo")
}

// vcsHashObjectCmd in ra mã băm của một tệp.
var vcsHashObjectCmd = &cobra.Command{
	Use:   "hash-object <tệp>...",
	Short: "Tính và in mã băm của nội dung tệp",
	Long: `Đọc nội dung tệp trên đĩa rồi in mã băm tương ứng.

Lệnh không ghi gì vào kho, chỉ cho biết mã băm mà td sẽ dùng nếu tệp đó được
đưa vào kho. Hai tệp có cùng nội dung sẽ cho cùng một mã băm.`,
	Example: `  td vcs hash-object main.go
  td vcs hash-object tệp-một tệp-hai`,
	Args: minimumArgs(1, "<tệp>..."),
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
	Use:   "cat-file <loại> <mã-băm>",
	Short: "In nội dung của một object (blob, tree, commit, tag)",
	Long: `Đọc một object trong kho và kiểm tra loại của nó rồi in nội dung.

Loại hợp lệ: blob, tree, commit, tag. Nếu loại truyền vào không khớp với loại
thật của object thì lệnh báo lỗi. Kèm cờ -v để in chi tiết: với blob là nội
dung đầy đủ, với tree là từng entry kèm chế độ và mã băm, với commit và tag là
toàn bộ phần thô.`,
	Example: `  # Xác minh loại và in nội dung
  td vcs cat-file blob a1b2c3d4

  # In toàn bộ nội dung thô của commit HEAD
  td vcs cat-file commit HEAD -v`,
	Args: exactArgs(2, "<loại> <mã-băm>"),
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := openRepo(cmd)
		if err != nil {
			return err
		}
		return ops.CatFile(r, args[0], args[1], verboseOn(cmd))
	},
}

// vcsFsckCmd kiểm tra tính toàn vẹn của kho dữ liệu.
var vcsFsckCmd = &cobra.Command{
	Use:   "fsck",
	Short: "Kiểm tra tính toàn vẹn của kho và các tham chiếu",
	Long: `Duyệt toàn bộ kho kiểm tra hai điều:

  - Mọi object trong kho có đọc được không (nén zlib không hỏng, header hợp lệ).
  - Mọi tham chiếu và HEAD có trỏ tới một object tồn tại không.

Lệnh trả về mã thoát khác 0 nếu phát hiện vấn đề.`,
	Example: `  td vcs fsck
  td vcs -C du-an-khac fsck`,
	Args: noArgsArg,
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
			printLine("Không phát hiện vấn đề nào (%d object, %d tham chiếu)",
				report.ObjectCount, report.RefCount)
			return nil
		}
		for _, p := range report.Problems {
			printErr("vấn đề: %s", p)
		}
		return exitError("phát hiện %d vấn đề", len(report.Problems))
	},
}
