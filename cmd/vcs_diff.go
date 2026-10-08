package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tonguyenducmanh/devcli/internal/vcs/ops"
)

// vcsDiffCmd hiển thị khác biệt giữa các vùng.
var vcsDiffCmd = &cobra.Command{
	Use:     "diff [phạm vi] [-- tệp...]",
	Aliases: []string{"df"},
	Short:   "Hiển thị khác biệt giữa các phiên bản",
	Long: `Hiển thị khác biệt giữa hai vùng bất kỳ.

Mặc định so sánh vùng đã stage với cây làm việc, tức là những sửa đổi chưa
được stage. Dùng --staged để so sánh HEAD với vùng đã stage.

Tham số phạm vi nhận một tên nhánh, một mã băm, hoặc hai mã băm nối bằng hai
dấu chấm để so sánh trực tiếp, ba dấu chấm để so sánh từ điểm chung gần nhất.

Sau dấu hai gạch ngang là danh sách tệp cần lọc.`,
	Example: `  td vcs diff                       đã stage so với cây làm việc
  td vcs diff --staged              HEAD so với vùng đã stage
  td vcs diff main                  commit hiện tại so với main
  td vcs diff main..feature         so sánh hai nhánh
  td vcs diff main...feature        so sánh từ điểm chung gần nhất
  td vcs diff --stat HEAD~1         chỉ xem thống kê thay đổi`,
	Args: arbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := openRepo(cmd)
		if err != nil {
			return err
		}
		staged, _ := cmd.Flags().GetBool("staged")
		stat, _ := cmd.Flags().GetBool("stat")
		nameOnly, _ := cmd.Flags().GetBool("name-only")
		context, _ := cmd.Flags().GetInt("unified")
		color, _ := cmd.Flags().GetString("color")

		opts := ops.DiffOptions{
			Staged:   staged,
			Stat:     stat,
			NameOnly: nameOnly,
			Context:  context,
		}
		// Đường dẫn lọc nằm sau dấu "--".
		opts.Paths, args = splitPathsAtDash(cmd, args)
		// Tham số còn lại là phạm vi commit cần so sánh.
		if len(args) > 0 {
			opts.Revision = args[0]
		}
		if len(args) > 1 {
			return exitError("chỉ chấp nhận một phạm vi, ví dụ: td vcs diff main..feature")
		}

		diffs, err := ops.Diff(r, opts)
		if err != nil {
			return err
		}
		if len(diffs) == 0 {
			printLine("Không có khác biệt nào.")
			return nil
		}
		printDiffs(diffs, opts, color)
		return nil
	},
}

// printDiffs in kết quả diff ra stdout theo định dạng yêu cầu.
func printDiffs(diffs []ops.FileDiff, opts ops.DiffOptions, colorMode string) {
	// Cờ --color của lệnh ghi đè quyết định tô màu ở mức toàn cục.
	switch colorMode {
	case "always":
		colorEnabled = true
	case "never":
		colorEnabled = false
	}

	for _, d := range diffs {
		if opts.NameOnly {
			printLine("%s", d.Path)
			continue
		}
		if opts.Stat {
			marker := diffStatusMarker(d.Status)
			if d.Binary {
				printLine(" %s %s | file nhị phân thay đổi", marker, d.Path)
				continue
			}
			printLine(" %s %s | %d dòng thêm, %d dòng xoá", marker, d.Path, d.Added, d.Del)
			continue
		}
		printLine("%s %s", diffStatusLabel(d.Status), d.Path)
		if d.Binary {
			printLine("    (file nhị phân, không hiển thị nội dung)")
			printLine("")
			continue
		}
		for _, line := range d.Lines {
			fmt.Fprint(os.Stdout, colorizeDiffLine(line))
			fmt.Fprintln(os.Stdout)
		}
		printLine("")
	}
}

// diffStatusMarker trả về ký hiệu một chữ cho trạng thái.
func diffStatusMarker(s byte) string {
	switch s {
	case 'A':
		return "A"
	case 'D':
		return "D"
	default:
		return "M"
	}
}

// diffStatusLabel trả về nhãn đầy đủ cho dòng tiêu đề file.
func diffStatusLabel(s byte) string {
	switch s {
	case 'A':
		return "Thêm    :"
	case 'D':
		return "Xoá     :"
	default:
		return "Sửa     :"
	}
}

// colorizeDiffLine thêm màu cho một dòng diff: xanh cho dòng thêm,
// đỏ cho dòng xoá, xanh dương cho dòng tiêu đề hunk, giữ nguyên dòng ngữ cảnh.
func colorizeDiffLine(line string) string {
	switch {
	case strings.HasPrefix(line, "+"):
		return colorize(colorGreen, line)
	case strings.HasPrefix(line, "-"):
		return colorize(colorRed, line)
	case strings.HasPrefix(line, "@@"):
		return colorize(colorCyan, line)
	default:
		return line
	}
}

func init() {
	vcsDiffCmd.Flags().BoolP("staged", "c", false, "so sánh HEAD với vùng đã stage")
	vcsDiffCmd.Flags().Bool("stat", false, "chỉ hiển thị thống kê thay đổi")
	vcsDiffCmd.Flags().Bool("name-only", false, "chỉ hiển thị tên tệp thay đổi")
	vcsDiffCmd.Flags().IntP("unified", "U", 3, "số dòng ngữ cảnh quanh thay đổi")
	vcsDiffCmd.Flags().String("color", "auto", "màu output: auto, always, never")
}
