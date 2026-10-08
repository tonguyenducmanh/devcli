package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tonguyenducmanh/devcli/internal/vcs/object"
	"github.com/tonguyenducmanh/devcli/internal/vcs/ops"
	"github.com/tonguyenducmanh/devcli/internal/vcs/repo"
)

// vcsMergeCmd hợp nhất hai nhánh.
var vcsMergeCmd = &cobra.Command{
	Use:     "merge <nhánh>",
	Aliases: []string{"mg"},
	Short:   "Hợp nhất một nhánh vào nhánh hiện tại",
	Long: `Kết hợp lịch sử của một nhánh khác vào nhánh đang đứng.

Khi nhánh đích nằm ngay sau nhánh hiện tại, con trỏ chỉ dịch thẳng sang đó mà
không tạo mốc mới. Trong trường hợp lệch nhánh, tm so từng tệp và hợp nhất nội
dung ba phía; tệp không thể tự động hợp nhất sẽ được đánh dấu xung đột.

Gặp xung đột thì lệnh dừng lại và ghi lại trạng thái, dùng --continue để hoàn
tất hoặc --abort để huỷ.`,
	Example: `  tm vcs merge main            hợp nhất nhánh main
  tm vcs merge --no-ff main    luôn tạo commit merge
  tm vcs merge --abort         huỷ lần merge đang dở dang
  tm vcs merge --continue      hoàn tất sau khi giải quyết xung đột`,
	Args: maximumArgs(1, "<nhánh|commit>"),
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := openRepo(cmd)
		if err != nil {
			return err
		}
		abort, _ := cmd.Flags().GetBool("abort")
		continueFlag, _ := cmd.Flags().GetBool("continue")
		noFF, _ := cmd.Flags().GetBool("no-ff")
		ffOnly, _ := cmd.Flags().GetBool("ff-only")
		squash, _ := cmd.Flags().GetBool("squash")
		message, _ := cmd.Flags().GetString("message")

		opts := ops.MergeOptions{
			NoFF:     noFF,
			FFOnly:   ffOnly,
			Message:  message,
			Abort:    abort,
			Continue: continueFlag,
			Squash:   squash,
		}
		if len(args) == 1 {
			opts.Branch = args[0]
		}
		if !abort && !continueFlag && opts.Branch == "" {
			return exitError("cần chỉ định nhánh cần hợp nhất")
		}

		res, err := ops.Merge(r, opts)
		if err != nil {
			return err
		}
		printMergeResult(r, res)
		return nil
	},
}

// printMergeResult in thông báo tương ứng với kết quả merge.
func printMergeResult(r *repo.Repo, res *ops.MergeResult) {
	switch {
	case res.AlreadyUpToDate:
		printLine("Đã cập nhật, không có gì để hợp nhất.")
	case res.FastForward:
		printLine("Đã cập nhật nhanh (fast-forward).")
	case len(res.Conflicts) > 0:
		printLine("Đã hợp nhất với %d file xung đột:", len(res.Conflicts))
		for _, p := range res.Conflicts {
			printLine("  %s", p)
		}
		printLine("\nGiải quyết xong rồi chạy `tm vcs add <tệp>` và `tm vcs commit`.")
	case !res.MergeCommit.IsZero():
		printLine("Đã tạo commit hợp nhất %s", res.MergeCommit.Short(8))
	default:
		printLine("Đã hợp nhất vào vùng stage, chạy `tm vcs commit` để ghi lại.")
	}
}

// vcsRebaseCmd di chuyển commit lên một điểm khác.
var vcsRebaseCmd = &cobra.Command{
	Use:     "rebase [đích]",
	Aliases: []string{"rb"},
	Short:   "Di chuyển các commit hiện tại lên trên một điểm khác",
	Long: `Áp dụng lại các commit của một nhánh lên trên một điểm khác.

Mỗi commit được áp dụng lại bằng cách hợp nhất nội dung với trạng thái hiện
tại, nên lịch sử trở nên gọn và tuyến tính thay vì nhiều nhánh song song.

Có thể rebase một nhánh khác bằng cách truyền cả hai tham số: điểm đích trước,
tên nhánh sau. Dùng --onto khi muốn điểm đích khác điểm gốc.`,
	Example: `  tm vcs rebase main            đưa commit hiện tại lên trên main
  tm vcs rebase <đích> <nhánh>   rebase một nhánh khác lên trên đích
  tm vcs rebase --onto <đích>   chỉ định điểm đích khác
  tm vcs rebase --continue      tiếp tục sau khi giải quyết xung đột
  tm vcs rebase --abort         quay lại trạng thái trước rebase`,
	Args: maximumArgs(2, "[điểm-đích] [nhánh]"),
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := openRepo(cmd)
		if err != nil {
			return err
		}
		abort, _ := cmd.Flags().GetBool("abort")
		continueFlag, _ := cmd.Flags().GetBool("continue")
		skip, _ := cmd.Flags().GetBool("skip")
		onto, _ := cmd.Flags().GetString("onto")
		branch, _ := cmd.Flags().GetString("branch")

		opts := ops.RebaseOptions{
			Onto:     onto,
			Branch:   branch,
			Abort:    abort,
			Continue: continueFlag,
			Skip:     skip,
		}
		// Một đối số là điểm đích, hai đối số là đích rồi tới nhánh cần rebase.
		if len(args) == 1 {
			opts.Upstream = args[0]
		}
		if len(args) == 2 {
			opts.Upstream = args[0]
			opts.Branch = args[1]
		}
		if !abort && !continueFlag && !skip && opts.Upstream == "" && branch == "" {
			return exitError("cần chỉ định nhánh hoặc commit đích để rebase lên")
		}

		res, err := ops.Rebase(r, opts)
		if err != nil {
			return err
		}
		printRebaseResult(res)
		return nil
	},
}

// printRebaseResult in thông báo kết quả rebase.
func printRebaseResult(res *ops.RebaseResult) {
	if len(res.Conflicts) > 0 {
		printLine("Rebase dừng ở %d file xung đột:", len(res.Conflicts))
		for _, p := range res.Conflicts {
			printLine("  %s", p)
		}
		printLine("\nGiải quyết rồi chạy `tm vcs rebase --continue`, hoặc `tm vcs rebase --abort` để hủy.")
		return
	}
	if res.Skipped > 0 {
		printLine("Đã bỏ qua %d commit.", res.Skipped)
	}
	if res.Applied > 0 {
		printLine("Đã áp dụng lại %d commit.", res.Applied)
	}
	if res.Done {
		printLine("Rebase hoàn tất.")
	}
}

// vcsCherryPickCmd áp dụng một commit từ nhánh khác.
var vcsCherryPickCmd = &cobra.Command{
	Use:     "cherry-pick <commit>...",
	Aliases: []string{"cp"},
	Short:   "Áp dụng thay đổi của một commit cụ thể",
	Long: `Chép thay đổi của một hoặc nhiều commit từ nhánh khác vào nhánh đang đứng.

Mỗi commit được áp dụng như một lần hợp nhất ba phía, nên thay đổi được giữ
nguyên dù nhánh nguồn đã tiến xa. Truyền nhiều mã băm để áp dụng theo đúng
thứ tự đã cho.`,
	Example: `  tm vcs cherry-pick abc1234
  tm vcs cherry-pick abc1234 def5678
  tm vcs cherry-pick --no-commit abc1234   chỉ áp dụng vào vùng stage
  tm vcs cherry-pick --abort                huỷ khi đang giải quyết xung đột`,
	Args: arbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := openRepo(cmd)
		if err != nil {
			return err
		}
		noCommit, _ := cmd.Flags().GetBool("no-commit")
		abort, _ := cmd.Flags().GetBool("abort")
		continueFlag, _ := cmd.Flags().GetBool("continue")
		skip, _ := cmd.Flags().GetBool("skip")

		// Gộp cờ vào một tham số duy nhất cho giao diện gọn gàng.
		mode := ""
		if abort {
			mode = "abort"
		}
		if continueFlag {
			mode = "continue"
		}
		if skip {
			mode = "skip"
		}

		opts := ops.CherryPickOptions{NoCommit: noCommit}
		switch mode {
		case "abort":
			opts.Abort = true
		case "continue":
			opts.Continue = true
		case "skip":
			opts.Skip = true
		}
		if mode == "" {
			commits, err := resolveCommitList(r, args)
			if err != nil {
				return err
			}
			opts.Commits = commits
		}

		res, err := ops.CherryPick(r, opts)
		if err != nil {
			return err
		}
		if abort {
			printLine("Đã huỷ cherry-pick.")
			return nil
		}
		if skip {
			printLine("Đã bỏ qua commit hiện tại.")
			return nil
		}
		printPickResult("cherry-pick", res.Commits, res.Conflicts, res.Skipped, noCommit)
		return nil
	},
}

// vcsRevertCmd hoàn tác một commit.
var vcsRevertCmd = &cobra.Command{
	Use:     "revert <commit>...",
	Aliases: []string{"rv"},
	Short:   "Hoàn tác thay đổi của một commit",
	Long: `Hoàn tác thay đổi của một commit bằng cách tạo một commit mới.

Lịch sử không bị viết lại nên lệnh này an toàn với nhánh đã chia sẻ. Khi hoàn
tác nhiều commit, chúng được xử lý theo thứ tự ngược: commit mới nhất trước.`,
	Example: `  tm vcs revert abc1234
  tm vcs revert --no-commit abc1234
  tm vcs revert --abort`,
	Args: arbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := openRepo(cmd)
		if err != nil {
			return err
		}
		noCommit, _ := cmd.Flags().GetBool("no-commit")
		abort, _ := cmd.Flags().GetBool("abort")
		continueFlag, _ := cmd.Flags().GetBool("continue")
		skip, _ := cmd.Flags().GetBool("skip")

		mode := ""
		if abort {
			mode = "abort"
		}
		if continueFlag {
			mode = "continue"
		}
		if skip {
			mode = "skip"
		}

		opts := ops.RevertOptions{NoCommit: noCommit}
		switch mode {
		case "abort":
			opts.Abort = true
		case "continue":
			opts.Continue = true
		case "skip":
			opts.Skip = true
		}
		if mode == "" {
			commits, err := resolveCommitList(r, args)
			if err != nil {
				return err
			}
			opts.Commits = commits
		}

		res, err := ops.Revert(r, opts)
		if err != nil {
			return err
		}
		if abort {
			printLine("Đã huỷ revert.")
			return nil
		}
		if skip {
			printLine("Đã bỏ qua commit hiện tại.")
			return nil
		}
		printPickResult("revert", res.Commits, res.Conflicts, res.Skipped, noCommit)
		return nil
	},
}

// resolveCommitList chuyển danh sách tham số thành danh sách hash commit.
func resolveCommitList(r *repo.Repo, args []string) ([]object.Hash, error) {
	if len(args) == 0 {
		return nil, exitError("cần chỉ định ít nhất một commit")
	}
	var out []object.Hash
	for _, a := range args {
		h, err := r.ResolveRev(a)
		if err != nil {
			return nil, err
		}
		out = append(out, h)
	}
	return out, nil
}

// printPickResult in thông báo chung cho cherry-pick và revert.
func printPickResult(name string, commits []object.Hash, conflicts []string, skipped int, noCommit bool) {
	if len(conflicts) > 0 {
		printLine("%s dừng ở %d file xung đột:", name, len(conflicts))
		for _, p := range conflicts {
			printLine("  %s", p)
		}
		printLine("\nGiải quyết xong rồi chạy `tm vcs %s --continue`.", name)
		return
	}
	if skipped > 0 {
		printLine("Đã bỏ qua %d commit.", skipped)
	}
	if len(commits) == 0 {
		if noCommit {
			printLine("Đã áp dụng thay đổi vào vùng stage.")
		} else {
			printLine("Không có gì để làm, nội dung đã đúng.")
		}
		return
	}
	parts := make([]string, 0, len(commits))
	for _, h := range commits {
		parts = append(parts, h.Short(8))
	}
	verb := "Đã áp dụng"
	if name == "revert" {
		verb = "Đã hoàn tác"
	}
	printLine("%s %d commit: %s", verb, len(commits), strings.Join(parts, ", "))
}

// vcsStashCmd quản lý các bản lưu tạm.
var vcsStashCmd = &cobra.Command{
	Use:     "stash",
	Aliases: []string{"st"},
	Short:   "Lưu tạm và khôi phục các thay đổi chưa commit",
	Long: `Cất các thay đổi chưa commit vào một kho tạm rồi đưa cây làm việc
về trạng thái sạch.

Mỗi lần lưu tạo thêm một mục trong danh sách. Dùng -u để cất kèm cả tệp chưa
theo dõi; những tệp này sẽ bị gỡ khỏi đĩa và trở lại khi áp dụng lại.

Số ở đối số chỉ vị trí trong danh sách, tính từ 0 cho mục mới nhất.`,
	Example: `  tm vcs stash                 lưu thay đổi hiện tại
  tm vcs stash -u              kèm cả file chưa được theo dõi
  tm vcs stash list            xem các bản đã lưu
  tm vcs stash apply           áp dụng bản mới nhất, giữ lại trong danh sách
  tm vcs stash pop             áp dụng rồi xóa bản đó
  tm vcs stash drop            xóa một bản
  tm vcs stash clear           xóa toàn bộ`,
	Args: arbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := openRepo(cmd)
		if err != nil {
			return err
		}
		untracked, _ := cmd.Flags().GetBool("include-untracked")
		message, _ := cmd.Flags().GetString("message")
		listFlag, _ := cmd.Flags().GetBool("list")
		apply, _ := cmd.Flags().GetBool("apply")
		pop, _ := cmd.Flags().GetBool("pop")
		drop, _ := cmd.Flags().GetBool("drop")
		clear, _ := cmd.Flags().GetBool("clear")

		// Cho phép viết dạng lệnh con: `tm vcs stash list`.
		if len(args) > 0 {
			switch args[0] {
			case "list", "show":
				listFlag = true
				args = args[1:]
			case "apply":
				apply = true
				args = args[1:]
			case "pop":
				pop = true
				args = args[1:]
			case "drop":
				drop = true
				args = args[1:]
			case "clear":
				clear = true
				args = args[1:]
			}
		}
		// Vị trí stash truyền vào, mặc định là bản mới nhất.
		which := 0
		for _, a := range args {
			if i, err := parseIntArg(a); err == nil {
				which = i
			}
		}

		switch {
		case clear:
			if err := ops.StashClear(r); err != nil {
				return err
			}
			printLine("Đã xóa toàn bộ bản lưu tạm.")
			return nil

		case drop:
			if err := ops.StashDrop(r, which); err != nil {
				return err
			}
			printLine("Đã xóa bản lưu tạm ở vị trí %d.", which)
			return nil

		case apply:
			if err := ops.StashApply(r, which); err != nil {
				return err
			}
			printLine("Đã áp dụng bản lưu tạm ở vị trí %d.", which)
			return nil

		case pop:
			if err := ops.StashPop(r, which); err != nil {
				return err
			}
			printLine("Đã áp dụng và xóa bản lưu tạm ở vị trí %d.", which)
			return nil

		case listFlag:
			entries, err := ops.StashList(r)
			if err != nil {
				return err
			}
			if len(entries) == 0 {
				printLine("Chưa có bản lưu tạm nào.")
				return nil
			}
			for _, e := range entries {
				printLine("stash@{%d}: %s", e.Index, e.Message)
			}
			return nil

		default:
			h, err := ops.Stash(r, message, untracked)
			if err != nil {
				return err
			}
			printLine("Đã lưu tạm thay đổi tại %s", h.Short(8))
			return nil
		}
	},
}

// parseIntArg chuyển chuỗi sang số nguyên, báo lỗi nếu không hợp lệ.
func parseIntArg(s string) (int, error) {
	var n int
	if _, err := fmt.Sscanf(s, "%d", &n); err != nil {
		return 0, err
	}
	return n, nil
}

func init() {
	// Cờ của merge.
	vcsMergeCmd.Flags().Bool("abort", false, "huỷ lần merge đang dở dang")
	vcsMergeCmd.Flags().Bool("continue", false, "hoàn tất merge sau khi giải quyết xung đột")
	vcsMergeCmd.Flags().Bool("no-ff", false, "luôn tạo commit merge dù có thể fast-forward")
	vcsMergeCmd.Flags().Bool("ff-only", false, "chỉ cho phép fast-forward")
	vcsMergeCmd.Flags().Bool("squash", false, "gộp thay đổi vào vùng stage mà không tạo commit merge")
	vcsMergeCmd.Flags().StringP("message", "m", "", "nội dung cho commit merge")

	// Cờ của rebase.
	vcsRebaseCmd.Flags().Bool("abort", false, "huỷ rebase và trở về trạng thái trước đó")
	vcsRebaseCmd.Flags().Bool("continue", false, "tiếp tục rebase sau khi giải quyết xung đột")
	vcsRebaseCmd.Flags().Bool("skip", false, "bỏ qua commit hiện tại")
	vcsRebaseCmd.Flags().String("onto", "", "commit đích thay cho điểm gốc")
	vcsRebaseCmd.Flags().String("branch", "", "rebase cho nhánh này thay vì nhánh hiện tại")

	// Cờ dùng chung cho cherry-pick và revert.
	for _, c := range []*cobra.Command{vcsCherryPickCmd, vcsRevertCmd} {
		c.Flags().BoolP("no-commit", "n", false, "chỉ áp dụng thay đổi vào vùng stage")
		c.Flags().Bool("abort", false, "huỷ thao tác đang dở dang")
		c.Flags().Bool("continue", false, "tiếp tục sau khi giải quyết xung đột")
		c.Flags().Bool("skip", false, "bỏ qua commit hiện tại")
	}

	// Cờ của stash.
	vcsStashCmd.Flags().BoolP("include-untracked", "u", false, "kèm cả file chưa được theo dõi")
	vcsStashCmd.Flags().StringP("message", "m", "", "mô tả ngắn cho bản lưu tạm")
	vcsStashCmd.Flags().BoolP("list", "l", false, "liệt kê các bản đã lưu")
	vcsStashCmd.Flags().Bool("apply", false, "áp dụng mà không xóa khỏi danh sách")
	vcsStashCmd.Flags().Bool("pop", false, "áp dụng rồi xóa khỏi danh sách")
	vcsStashCmd.Flags().Bool("drop", false, "xóa một bản lưu tạm")
	vcsStashCmd.Flags().Bool("clear", false, "xóa toàn bộ bản lưu tạm")
}
