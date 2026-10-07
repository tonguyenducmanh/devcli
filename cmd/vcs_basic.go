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
	Long: `In trạng thái hiện tại của cây làm việc theo ba nhóm:

  đã stage       nội dung sẽ được ghi vào commit kế tiếp
  chưa stage     đã sửa trên đĩa nhưng chưa đưa vào vùng chuẩn bị
  chưa theo dõi  tệp mới xuất hiện, td chưa quản lý

Ký hiệu đầu mỗi dòng cho biết thao tác: thêm, sửa, xoá hoặc mới.

Dòng đầu tiên cho biết đang ở nhánh nào, HEAD có đang tách rời không, và nhánh
đó đi trước hay đi sau nhánh theo dõi bao nhiêu commit.`,
	Example: `  td vcs status
  td vcs status -C thư-mục-khác
  td vcs st`,
	Args: noArgsArg,
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
	Long: `Đưa nội dung hiện tại của tệp vào vùng chuẩn bị, nơi nội dung được
chốt lại cho commit kế tiếp.

Không có đối số thì chỉ cập nhật các tệp đã được theo dõi. Có thể truyền
đường dẫn cụ thể, một thư mục, hoặc mẫu có dấu * và ?.

Tệp bị xoá khỏi đĩa cũng được gỡ khỏi vùng chuẩn bị. Tệp chưa theo dõi là
tệp td chưa quản lý, thêm vào .tdxignore nếu muốn bỏ qua vĩnh viễn.`,
	Example: `  td vcs add .              thêm mọi thay đổi
  td vcs add main.go        thêm một tệp
  td vcs add --update .     chỉ cập nhật tệp đã được theo dõi
  td vcs add "docs/*.md"    thêm theo mẫu đường dẫn`,
	Args: arbitraryArgs,
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
		if verboseOn(cmd) {
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
	Long: `Ghi lại nội dung vùng chuẩn bị thành một mốc có tên trong lịch sử.

Mỗi commit lưu cây nội dung đầy đủ, tác giả và thời điểm, đồng thời trỏ tới
commit cha nên tạo thành một chuỗi lịch sử. Dùng --amend để viết lại commit
vừa tạo thay vì tạo mốc mới.

Thông điệp phải truyền bằng cờ -m, lặp lại -m để tách tiêu đề và phần mô tả
chi tiết thành hai đoạn.`,
	Example: `  td vcs commit -m "tin nhắn ngắn"
  td vcs commit -m "tiêu đề" -m "mô tả chi tiết"
  td vcs commit --amend -m "sửa lại commit vừa tạo"
  td vcs commit -a          stage mọi thay đổi rồi commit luôn`,
	Args: noArgsArg,
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
	Long: `Duyệt lịch sử commit từ một điểm bắt đầu đi ngược về các commit cha.

Mặc định bắt đầu từ HEAD. Sau dấu hai gạch ngang là danh sách tệp, khi đó chỉ
những commit có thay đổi tệp đó mới được hiển thị.

Cột đầu là mã băm ngắn, kèm các tham chiếu đang trỏ tới commit đó; HEAD được
đánh dấu bằng HEAD -> để phân biệt với các nhánh khác.`,
	Example: `  td vcs log
  td vcs log --oneline
  td vcs log -n 5 --patch
  td vcs log --all
  td vcs log -- cmd/            chỉ xem các commit có sửa thư mục cmd/`,
	Args: arbitraryArgs,
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
	Long: `In thông tin đầy đủ của một commit: mã băm, các tham chiếu trỏ tới nó,
tác giả, thời điểm, nội dung thông điệp và danh sách tệp bị thay đổi.

Mặc định lấy HEAD. Kèm -p để in luôn nội dung khác biệt của từng tệp.`,
	Example: `  td vcs show
  td vcs show HEAD~2
  td vcs show abc1234`,
	Args: maximumArgs(1, "[commit]"),
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
	Short: "Xem nhật ký di chuyển của HEAD hoặc một tham chiếu",
	Long: `Mỗi lần một tham chiếu dịch chuyển sẽ được ghi lại kèm mã băm cũ,
mã băm mới và lý do. Lệnh in danh sách từ mục mới nhất trở về.

Mục cũ vẫn nằm trong kho nên có thể quay lại bằng cách trỏ một tham chiếu
tới mã băm tương ứng, ví dụ: td vcs switch abc1234`,
	Example: `  td vcs reflog
  td vcs reflog main
  td vcs reflog refs/tags/v1.0.0`,
	Args: maximumArgs(1, "[ref]"),
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
	Long: `Xoá tệp khỏi vùng chuẩn bị và khỏi cây làm việc.

Thay đổi được ghi ở lần commit kế tiếp, tệp vẫn còn trong lịch sử. Kèm
--cached để chỉ gỡ khỏi vùng chuẩn bị mà giữ nguyên tệp trên đĩa.

Tệp đã xoá có thể đưa lại bằng lệnh restore trước khi commit.`,
	Example: `  td vcs rm tệp-cũ.txt
  td vcs rm thư-mục/tệp.txt
  td vcs rm --cached tệp-vẫn-giữ.txt`,
	Args: minimumArgs(1, "<tệp>..."),
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
	Short: "Đổi tên hoặc di chuyển một tệp đã được theo dõi",
	Long: `Di chuyển hoặc đổi tên một tệp trong cây làm việc và cập nhật vùng chuẩn
bị theo đường dẫn mới.

Thư mục đích không cần tồn tại trước, được tạo tự động.`,
	Example: `  td vcs mv tên-cũ.txt tên-mới.txt
  td vcs mv tệp.txt thư-mục/tệp.txt`,
	Args: exactArgs(2, "<nguồn> <đích>"),
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := openRepo(cmd)
		if err != nil {
			return err
		}
		return ops.Move(r, args[0], args[1])
	},
}

func init() {
	// Mỗi lệnh tự có -v riêng. Cờ toàn cục đã bị bỏ có chủ đích, xem cmd/flags.go.
	addVerboseFlag(vcsStatusCmd, "liệt kê cả các tệp chưa được theo dõi")
	addVerboseFlag(vcsAddCmd, "in ra từng tệp đã đưa vào vùng chuẩn bị")
	addVerboseFlag(vcsLogCmd, "in cả nội dung thay đổi của từng commit")
	addVerboseFlag(vcsRestoreCmd, "in ra từng tệp đã khôi phục")

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
