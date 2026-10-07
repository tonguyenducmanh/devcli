package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tonguyenducmanh/devcli/internal/vcs/ops"
)

// vcsStatusCmd hiển thị trạng thái hiện tại của kho mã nguồn.
var vcsStatusCmd = &cobra.Command{
	Use:     "status",
	Aliases: []string{"st"},
	Short:   "Hiển thị trạng thái thay đổi của cây làm việc",
	Long: `Hiển thị ba vùng chính:
  thay đổi đã stage   sẽ được ghi vào commit tới
  thay đổi chưa stage  đã sửa trên đĩa nhưng chưa đưa vào stage
  file chưa theo dõi  file mới xuất hiện, chưa được td quản lý`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := openRepo(cmd)
		if err != nil {
			return err
		}
		st, err := r.Status()
		if err != nil {
			return err
		}

		// Dòng đầu tiên mô tả nhánh và trạng thái theo dõi.
		if st.Detached {
			if !st.HeadHash.IsZero() {
				printLine("Trên HEAD tách rời tại %s", st.HeadHash.Short(8))
			} else {
				printLine("Trên HEAD tách rời (chưa có commit nào)")
			}
		} else if st.Branch != "" {
			printLine("Trên nhánh %s", st.Branch)
		} else {
			printLine("Chưa có nhánh nào (chạy `td vcs commit` để tạo)")
		}
		if st.HasUpstream {
			line := fmt.Sprintf("Theo dõi: đi trước %d, đi sau %d", st.Ahead, st.Behind)
			if st.Ahead == 0 && st.Behind == 0 {
				line = "Đã đồng bộ với nhánh theo dõi"
			}
			printLine("%s", line)
		}

		staged := st.Staged()
		if len(staged) > 0 {
			printLine("\nThay đổi đã stage:")
			for _, e := range staged {
				printLine("  %s  %s", statusChar(e.IndexStatus), e.Path)
			}
		}

		unstaged := st.Unstaged()
		if len(unstaged) > 0 {
			printLine("\nThay đổi chưa stage:")
			for _, e := range unstaged {
				printLine("  %s  %s", statusChar(e.WorkStatus), e.Path)
			}
		}

		untracked := st.Untracked()
		if len(untracked) > 0 {
			printLine("\nFile chưa được theo dõi:")
			for _, e := range untracked {
				printLine("  %s  %s", "?", e.Path)
			}
		}

		if len(st.Conflicts) > 0 {
			printLine("\nXung đột cần giải quyết:")
			for _, p := range st.Conflicts {
				printLine("  %s  %s", "U", p)
			}
			printLine("\nSau khi sửa, chạy `td vcs add <tệp>` rồi `td vcs commit`.")
		}

		if st.IsClean() && len(st.Conflicts) == 0 {
			printLine("\nCây làm việc sạch, không có gì cần commit.")
		}
		return nil
	},
}

// statusChar trả về ký hiệu trạng thái dạng chữ để dễ đọc khi không có màu.
func statusChar(s byte) string {
	switch s {
	case 'A':
		return "thêm"
	case 'M':
		return "sửa"
	case 'D':
		return "xoá"
	case '?':
		return "mới"
	default:
		return "  "
	}
}

// vcsAddCmd đưa thay đổi vào vùng stage.
var vcsAddCmd = &cobra.Command{
	Use:   "add [tệp...]",
	Short: "Đưa thay đổi vào vùng chuẩn bị commit",
	Long: `Đưa nội dung mới của tệp vào staging area.

  td vcs add .              thêm mọi thay đổi
  td vcs add main.go        thêm một tệp
  td vcs add --update .     chỉ cập nhật tệp đã được theo dõi
  td vcs add "docs/*.md"    thêm theo mẫu đường dẫn`,
	Args: cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := openRepo(cmd)
		if err != nil {
			return err
		}
		all, _ := cmd.Flags().GetBool("all")
		update, _ := cmd.Flags().GetBool("update")

		opts := ops.AddOptions{All: all, Update: update, Paths: args}
		if err := ops.Add(r, opts); err != nil {
			return err
		}
		if verboseEnabled(cmd) {
			st, err := r.Status()
			if err != nil {
				return err
			}
			for _, e := range st.Staged() {
				printLine("đã stage %s", e.Path)
			}
		}
		return nil
	},
}

// vcsCommitCmd tạo commit mới.
var vcsCommitCmd = &cobra.Command{
	Use:   "commit",
	Short: "Ghi lại các thay đổi đã stage thành một commit",
	Long: `Tạo commit từ nội dung staging area.

  td vcs commit -m "tin nhắn ngắn"
  td vcs commit -m "tiêu đề" -m "mô tả chi tiết"
  td vcs commit --amend -m "sửa lại commit vừa tạo"
  td vcs commit -a          stage mọi thay đổi rồi commit luôn`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := openRepo(cmd)
		if err != nil {
			return err
		}
		// Nhiều cờ -m được nối lại với nhau bằng một dòng trống.
		messages, _ := cmd.Flags().GetStringArray("message")
		amend, _ := cmd.Flags().GetBool("amend")
		allowEmpty, _ := cmd.Flags().GetBool("allow-empty")
		all, _ := cmd.Flags().GetBool("all")

		msg := strings.Join(messages, "\n\n")
		opts := ops.CommitOptions{
			Message:    msg,
			Amend:      amend,
			AllowEmpty: allowEmpty,
			All:        all,
		}
		h, err := ops.Commit(r, opts)
		if err != nil {
			return err
		}
		printLine("[%s %s] %s", shortBranchOf(r), h.Short(8), firstLineOf(msg))
		return nil
	},
}

// firstLineOf lấy dòng đầu tiên của một message.
func firstLineOf(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}

// shortBranchOf trả về tên nhánh hiện tại, hoặc "HEAD" nếu đang tách rời.
func shortBranchOf(r interface {
	CurrentBranch() (string, error)
}) string {
	b, err := r.CurrentBranch()
	if err != nil || b == "" {
		return "HEAD"
	}
	return b
}

// vcsLogCmd xem lịch sử commit.
var vcsLogCmd = &cobra.Command{
	Use:     "log",
	Aliases: []string{"lg"},
	Short:   "Xem lịch sử commit",
	Long: `Xem lịch sử commit từ HEAD đi ngược lại.

  td vcs log
  td vcs log --oneline
  td vcs log -n 5 --patch
  td vcs log --all
  td vcs log -- cmd/            chỉ xem các commit có sửa thư mục cmd/`,
	Args: cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := openRepo(cmd)
		if err != nil {
			return err
		}
		opts, err := logOptionsFromFlags(cmd, args)
		if err != nil {
			return err
		}
		entries, err := ops.Log(r, opts)
		if err != nil {
			return err
		}
		if len(entries) == 0 {
			printLine("Chưa có commit nào.")
			return nil
		}
		printLogEntries(entries, opts)
		return nil
	},
}

// logOptionsFromFlags gom cờ của lệnh log thành cấu hình.
func logOptionsFromFlags(cmd *cobra.Command, args []string) (ops.LogOptions, error) {
	opts := ops.LogOptions{Revision: "HEAD"}
	opts.Max, _ = cmd.Flags().GetInt("max-count")
	opts.Oneline, _ = cmd.Flags().GetBool("oneline")
	opts.All, _ = cmd.Flags().GetBool("all")
	opts.Patch, _ = cmd.Flags().GetBool("patch")
	opts.Stat, _ = cmd.Flags().GetBool("stat")
	opts.Reverse, _ = cmd.Flags().GetBool("reverse")
	// Đường dẫn lọc nằm sau dấu "--" nếu có.
	if i := indexOf(args, "--"); i >= 0 {
		opts.Paths = args[i+1:]
	}
	return opts, nil
}

// indexOf trả về vị trí phần tử đầu tiên có giá trị cho trước.
func indexOf(list []string, want string) int {
	for i, v := range list {
		if v == want {
			return i
		}
	}
	return -1
}

// printLogEntries in danh sách commit ra stdout.
func printLogEntries(entries []ops.LogEntry, opts ops.LogOptions) {
	for _, e := range entries {
		if opts.Oneline {
			line := e.Short + " " + e.Summary
			if len(e.Refs) > 0 {
				line += " (" + strings.Join(e.Refs, ", ") + ")"
			}
			printLine("%s", line)
			continue
		}
		if len(e.Refs) > 0 {
			printLine("commit %s (%s)", e.Short, strings.Join(e.Refs, ", "))
		} else {
			printLine("commit %s", e.Short)
		}
		printLine("Tác giả: %s <%s>", e.Author, e.Email)
		printLine("Ngày:   %s", e.When.Format("2006-01-02 15:04:05"))
		if e.Body != "" {
			printLine("\n%s", indent(e.Body, "    "))
		} else {
			printLine("\n    %s", e.Summary)
		}
		for _, l := range e.StatLines {
			printLine("%s", l)
		}
		if opts.Patch && e.Patch != "" {
			printLine("\n%s", strings.TrimRight(e.Patch, "\n"))
		}
		printLine("")
	}
}

// indent thụt lề mọi dòng của văn bản.
func indent(s, prefix string) string {
	lines := strings.Split(s, "\n")
	for i := range lines {
		lines[i] = prefix + lines[i]
	}
	return strings.Join(lines, "\n")
}

// vcsShowCmd hiển thị chi tiết một commit.
var vcsShowCmd = &cobra.Command{
	Use:   "show [commit]",
	Short: "Hiển thị chi tiết của một commit",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := openRepo(cmd)
		if err != nil {
			return err
		}
		rev := "HEAD"
		if len(args) == 1 {
			rev = args[0]
		}
		patch, _ := cmd.Flags().GetBool("patch")
		out, err := ops.Show(r, rev, patch)
		if err != nil {
			return err
		}
		fmt.Fprint(os.Stdout, out)
		return nil
	},
}

// vcsReflogCmd xem nhật ký thay đổi của ref.
var vcsReflogCmd = &cobra.Command{
	Use:   "reflog [ref]",
	Short: "Xem nhật ký di chuyển của HEAD hoặc một ref",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := openRepo(cmd)
		if err != nil {
			return err
		}
		name := "HEAD"
		if len(args) == 1 {
			name = args[0]
		}
		refName, err := r.ReflogRefName(name)
		if err != nil {
			return err
		}
		entries, err := r.ReadReflog(refName)
		if err != nil {
			return exitError("chưa có reflog cho %s", name)
		}
		if len(entries) == 0 {
			printLine("Reflog %s còn trống.", name)
			return nil
		}
		// In từ mới nhất về cũ.
		for i := len(entries) - 1; i >= 0; i-- {
			e := entries[i]
			printLine("%s %s@{%d}: %s", e.New.Short(8), name, len(entries)-1-i, e.Message)
		}
		return nil
	},
}

// vcsRmCmd xóa file khỏi kho mã nguồn.
var vcsRmCmd = &cobra.Command{
	Use:   "rm <tệp>...",
	Short: "Gỡ tệp khỏi theo dõi và khỏi cây làm việc",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := openRepo(cmd)
		if err != nil {
			return err
		}
		cached, _ := cmd.Flags().GetBool("cached")
		return ops.Remove(r, args, !cached, cached)
	},
}

// vcsMvCmd đổi tên hoặc di chuyển file.
var vcsMvCmd = &cobra.Command{
	Use:   "mv <nguồn> <đích>",
	Short: "Đổi tên hoặc di chuyển một file đã được theo dõi",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := openRepo(cmd)
		if err != nil {
			return err
		}
		return ops.Move(r, args[0], args[1])
	},
}

func init() {
	// Cờ riêng cho add.
	vcsAddCmd.Flags().BoolP("all", "A", false, "đưa mọi thay đổi vào stage")
	vcsAddCmd.Flags().BoolP("update", "u", false, "chỉ cập nhật tệp đã được theo dõi")

	// Cờ riêng cho commit.
	vcsCommitCmd.Flags().StringArrayP("message", "m", nil, "nội dung commit, lặp lại để thêm đoạn mô tả")
	vcsCommitCmd.Flags().Bool("amend", false, "sửa lại commit vừa tạo")
	vcsCommitCmd.Flags().Bool("allow-empty", false, "cho phép tạo commit dù không có thay đổi")
	vcsCommitCmd.Flags().BoolP("all", "a", false, "đưa mọi thay đổi vào stage trước khi commit")

	// Cờ riêng cho log.
	vcsLogCmd.Flags().IntP("max-count", "n", 0, "giới hạn số commit hiển thị")
	vcsLogCmd.Flags().BoolP("oneline", "l", false, "mỗi commit trên một dòng")
	vcsLogCmd.Flags().BoolP("all", "a", false, "duyệt lịch sử của mọi nhánh và tag")
	vcsLogCmd.Flags().BoolP("patch", "p", false, "kèm nội dung diff của từng commit")
	vcsLogCmd.Flags().Bool("stat", false, "kèm thống kê số dòng thay đổi")
	vcsLogCmd.Flags().Bool("reverse", false, "in ngược thứ tự thời gian")

	// Cờ riêng cho show.
	vcsShowCmd.Flags().BoolP("patch", "p", true, "kèm nội dung diff")

	// Cờ riêng cho rm.
	vcsRmCmd.Flags().Bool("cached", false, "chỉ gỡ khỏi theo dõi, giữ file trên đĩa")
}
