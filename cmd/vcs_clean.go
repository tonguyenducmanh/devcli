package cmd

import (
	"github.com/spf13/cobra"

	"github.com/tonguyenducmanh/devcli/internal/vcs/ops"
)

// vcsCleanCmd xoá tệp chưa được tm theo dõi.
var vcsCleanCmd = &cobra.Command{
	Use:     "clean [tệp...]",
	Aliases: []string{"cln"},
	Short:   "Xoá tệp chưa được theo dõi",
	Long: `Xoá khỏi cây làm việc những tệp tm chưa từng theo dõi.

Tệp đã được đưa vào vùng chuẩn bị thì thuộc về lịch sử nên lệnh này không đụng
tới, muốn bỏ theo dõi thì dùng ` + "`tm vcs rm`" + `. Nhờ vậy lệnh không bao giờ
xoá mất thứ còn cứu được trong kho.

Đối số là tệp, thư mục hoặc mẫu có dấu * và ?, khớp với cách ` + "`tm vcs add`" + `
hiểu đường dẫn. Không có đối số thì áp dụng cho mọi tệp chưa theo dõi.

Lệnh không hỏi lại: không có cờ -f thì chỉ liệt kê những gì sẽ bị xoá, đọc xong
xác nhận trong danh sách rồi chạy lại với -f mới xoá thật.`,
	Example: `  tm vcs clean              liệt kê các tệp sẽ bị xoá
  tm vcs clean -f           xoá mọi tệp chưa được theo dõi
  tm vcs clean -f build/    xoá tệp chưa theo dõi trong thư mục build
  tm vcs clean -f "*.log"   xoá mọi tệp kết thúc bằng .log`,
	Args: arbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := openRepo(cmd)
		if err != nil {
			return err
		}
		force, _ := cmd.Flags().GetBool("force")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if force && dryRun {
			return exitError("không thể dùng đồng thời --force và --dry-run")
		}

		opts := ops.CleanOptions{Paths: args, DryRun: !force}
		if dryRun {
			opts.DryRun = true
		}

		results, err := ops.Clean(r, opts)
		if err != nil {
			return err
		}

		var removed, skipped []string
		for _, res := range results {
			removed = append(removed, res.Removed...)
			skipped = append(skipped, res.Skipped...)
		}
		if len(removed) == 0 && len(skipped) == 0 {
			printLine("Không có tệp chưa được theo dõi nào khớp.")
			return nil
		}
		for _, note := range skipped {
			printErr("không xoá được %s", note)
		}
		if opts.DryRun {
			for _, p := range removed {
				printLine("sẽ xoá %s", p)
			}
			printLine("\nThêm -f nếu thực sự muốn xoá %d tệp này.", len(removed))
			return nil
		}
		for _, p := range removed {
			printLine("đã xoá %s", p)
		}
		printLine("\nĐã xoá %d tệp chưa được theo dõi.", len(removed))
		return nil
	},
}

func init() {
	vcsCleanCmd.Flags().BoolP("force", "f", false, "xoá thật thay vì chỉ liệt kê")
	vcsCleanCmd.Flags().Bool("dry-run", false, "chỉ liệt kê, giống khi không kèm -f")
}
