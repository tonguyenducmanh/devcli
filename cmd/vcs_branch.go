package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tonguyenducmanh/devcli/internal/vcs/ops"
	"github.com/tonguyenducmanh/devcli/internal/vcs/repo"
)

// vcsBranchCmd quản lý nhánh.
var vcsBranchCmd = &cobra.Command{
	Use:     "branch",
	Aliases: []string{"br"},
	Short:   "Liệt kê, tạo hoặc xóa nhánh",
	Long: `Quản lý các nhánh cục bộ.

  td vcs branch                  liệt kê các nhánh
  td vcs branch -d ten           xóa nhánh đã hợp nhất
  td vcs branch -D ten           xóa nhánh bất kể trạng thái
  td vcs branch ten              tạo nhánh tại HEAD
  td vcs branch ten main         tạo nhánh từ nhánh main
  td vcs branch -m cũ mới        đổi tên nhánh`,
	Args: cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := openRepo(cmd)
		if err != nil {
			return err
		}
		list, _ := cmd.Flags().GetBool("list")
		deleteFlag, _ := cmd.Flags().GetBool("delete")
		forceDel, _ := cmd.Flags().GetBool("force")
		rename, _ := cmd.Flags().GetBool("move")
		withHash, _ := cmd.Flags().GetBool("verbose")

		switch {
		case deleteFlag || forceDel:
			if len(args) == 0 {
				return exitError("cần chỉ định tên nhánh cần xóa")
			}
			results, err := ops.DeleteBranches(r, args, forceDel)
			if err != nil {
				return err
			}
			var failed int
			for _, res := range results {
				if res.Deleted {
					printLine("Đã xóa nhánh %s", res.Name)
					continue
				}
				failed++
				printLine("không xóa được %s: %s", res.Name, res.Reason)
			}
			if failed > 0 {
				return exitError("%d nhánh không được xóa", failed)
			}
			return nil

		case rename:
			if len(args) != 2 {
				return exitError("cần chỉ định tên cũ và tên mới")
			}
			swap, _ := cmd.Flags().GetBool("switch")
			if err := ops.RenameBranch(r, args[0], args[1], swap); err != nil {
				return err
			}
			printLine("Đã đổi tên %s thành %s", args[0], args[1])
			return nil

		case len(args) == 1 && !list:
			// Một đối số mà không có cờ: tạo nhánh mới từ HEAD.
			if err := createBranchCmd(cmd, r, args[0], "HEAD"); err != nil {
				return err
			}
			return nil

		case len(args) == 2 && !list:
			// Hai đối số: tạo nhánh mới từ một nhánh hoặc commit cho trước.
			if err := createBranchCmd(cmd, r, args[0], args[1]); err != nil {
				return err
			}
			return nil

		default:
			return listBranches(r, withHash)
		}
	},
}

// createBranchCmd tạo nhánh mới từ một điểm xuất phát cho trước,
// có tuỳ chọn chuyển thẳng sang nhánh vừa tạo.
func createBranchCmd(cmd *cobra.Command, r *repo.Repo, name, startPoint string) error {
	track, _ := cmd.Flags().GetBool("track")
	opts := ops.CreateBranchOptions{
		Name:       name,
		StartPoint: startPoint,
		Track:      track,
	}
	if err := ops.CreateBranch(r, opts); err != nil {
		return err
	}
	printLine("Đã tạo nhánh %s từ %s", name, startPoint)
	return nil
}

// listBranches in danh sách nhánh kèm hash và trạng thái.
func listBranches(r *repo.Repo, withHash bool) error {
	branches, err := ops.ListBranches(r)
	if err != nil {
		return err
	}
	if len(branches) == 0 {
		printLine("Chưa có nhánh nào.")
		return nil
	}
	for _, b := range branches {
		marker := "  "
		if b.Current {
			marker = "* "
		}
		line := fmt.Sprintf("%s%s", marker, b.Name)
		if withHash {
			line += " " + b.Hash.Short(8)
		}
		// Thông tin theo dõi chỉ hiện khi nhánh có thiết lập.
		if b.Upstream != "" {
			line += fmt.Sprintf(" [%s: đi trước %d, đi sau %d]", b.Upstream, b.Ahead, b.Behind)
		}
		printLine("%s", line)
	}
	return nil
}

// vcsCheckoutCmd chuyển nhánh hoặc commit.
var vcsCheckoutCmd = &cobra.Command{
	Use:     "checkout",
	Aliases: []string{"co"},
	Short:   "Chuyển sang nhánh hoặc commit khác",
	Long: `Chuyển HEAD sang nhánh, tag hoặc commit khác và cập nhật cây làm việc.

  td vcs checkout main           chuyển sang nhánh main
  td vcs checkout -b moi         tạo nhánh moi rồi chuyển sang đó
  td vcs checkout abc1234        chuyển tới một commit cụ thể (HEAD tách rời)
  td vcs checkout --detach main  chuyển tới commit của main`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := openRepo(cmd)
		if err != nil {
			return err
		}
		create, _ := cmd.Flags().GetString("branch")
		detach, _ := cmd.Flags().GetBool("detach")
		force, _ := cmd.Flags().GetBool("force")

		target := ""
		if len(args) == 1 {
			target = args[0]
		}
		if target == "" && create == "" {
			return exitError("cần chỉ định nhánh hoặc commit cần chuyển tới")
		}
		opts := ops.CheckoutOptions{
			Target:       target,
			CreateBranch: create,
			Force:        force,
			Detach:       detach,
		}
		if err := ops.Checkout(r, opts); err != nil {
			return err
		}
		// Thông báo nơi đang đứu sau khi chuyển.
		branch, _ := r.CurrentBranch()
		if branch == "" {
			head, _ := r.Head()
			printLine("Đang ở trạng thái tách rời tại %s", head.Short(8))
			return nil
		}
		printLine("Đã chuyển sang nhánh %s", branch)
		return nil
	},
}

// vcsSwitchCmd là cách viết ngắn của checkout cho việc chỉ chuyển nhánh.
var vcsSwitchCmd = &cobra.Command{
	Use:     "switch <nhánh>",
	Aliases: []string{"sw"},
	Short:   "Chuyển sang nhánh khác",
	Long: `Chuyển nhánh đang làm việc. Tương đương ` + "`td vcs checkout <nhánh>`" + `.

  td vcs switch main
  td vcs switch -c moi        tạo nhánh moi rồi chuyển sang đó`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := openRepo(cmd)
		if err != nil {
			return err
		}
		create, _ := cmd.Flags().GetString("create")
		force, _ := cmd.Flags().GetBool("discard-changes")

		if create != "" {
			return runSwitchCreate(cmd, r, create)
		}
		if len(args) != 1 {
			return exitError("cần chỉ định tên nhánh cần chuyển tới")
		}
		if err := ops.Checkout(r, ops.CheckoutOptions{Target: args[0], Force: force}); err != nil {
			return err
		}
		printLine("Đã chuyển sang nhánh %s", args[0])
		return nil
	},
}

// runSwitchCreate tạo nhánh mới rồi chuyển sang đó.
func runSwitchCreate(cmd *cobra.Command, r *repo.Repo, name string) error {
	opts := ops.CreateBranchOptions{Name: name, StartPoint: "HEAD", Switch: true}
	if err := ops.CreateBranch(r, opts); err != nil {
		return err
	}
	printLine("Đã tạo và chuyển sang nhánh %s", name)
	return nil
}

// vcsRestoreCmd khôi phục file từ index hoặc từ commit.
var vcsRestoreCmd = &cobra.Command{
	Use:     "restore <tệp>...",
	Aliases: []string{"rst"},
	Short:   "Khôi phục lại nội dung tệp",
	Long: `Khôi phục tệp về trạng thái trước đó.

  td vcs restore main.go         lấy lại nội dung đang stage
  td vcs restore --staged main.go  gỡ thay đổi đã stage
  td vcs restore --source=HEAD~1 main.go  lấy từ một commit khác`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := openRepo(cmd)
		if err != nil {
			return err
		}
		staged, _ := cmd.Flags().GetBool("staged")
		worktreeOnly, _ := cmd.Flags().GetBool("worktree")
		source, _ := cmd.Flags().GetString("source")

		if staged && worktreeOnly {
			return exitError("không thể dùng đồng thời --staged và --worktree")
		}
		opts := ops.RestoreOptions{
			Source:   source,
			Staged:   staged,
			Worktree: worktreeOnly,
			Paths:    args,
		}
		if err := ops.Restore(r, opts); err != nil {
			return err
		}
		printLine("Đã khôi phục %s", strings.Join(args, ", "))
		return nil
	},
}

// vcsResetCmd đặt lại HEAD theo một commit.
var vcsResetCmd = &cobra.Command{
	Use:     "reset [commit] [tệp...]",
	Aliases: []string{"rs"},
	Short:   "Di chuyển con trỏ HEAD về commit khác",
	Long: `Đặt lại HEAD theo ba chế độ:

  --soft   chỉ di chuyển HEAD, giữ nguyên index và cây làm việc
  --mixed  di chuyển HEAD và nạp lại index (mặc định)
  --hard   di chuyển HEAD, index và cây làm việc

  td vcs reset --soft HEAD~1
  td vcs reset HEAD~1 -- main.go   chỉ gỡ main.go khỏi vùng stage`,
	Args: cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := openRepo(cmd)
		if err != nil {
			return err
		}
		soft, _ := cmd.Flags().GetBool("soft")
		mixed, _ := cmd.Flags().GetBool("mixed")
		hard, _ := cmd.Flags().GetBool("hard")

		// Chỉ cho phép một chế độ tại một thời điểm.
		count := 0
		mode := ""
		if soft {
			count++
			mode = "soft"
		}
		if mixed {
			count++
			mode = "mixed"
		}
		if hard {
			count++
			mode = "hard"
		}
		if count > 1 {
			return exitError("chỉ chọn một trong --soft, --mixed, --hard")
		}

		// Tách phần đích commit và danh sách tệp sau dấu "--".
		opts := ops.ResetOptions{Mode: mode}
		sep := indexOf(args, "--")
		if sep >= 0 {
			opts.Paths = args[sep+1:]
			args = args[:sep]
		}
		if len(args) > 0 {
			opts.Target = args[0]
		}
		if err := ops.Reset(r, opts); err != nil {
			return err
		}
		desc := "mixed"
		if mode != "" {
			desc = mode
		}
		printLine("Đã reset (%s) về %s", desc, opts.Target)
		return nil
	},
}

// vcsTagCmd quản lý tag.
var vcsTagCmd = &cobra.Command{
	Use:   "tag",
	Short: "Liệt kê, tạo hoặc xóa tag",
	Long: `Đánh dấu một điểm trong lịch sử bằng tag.

  td vcs tag                        liệt kê các tag
  td vcs tag v1.0.0                 tạo tag nhẹ
  td vcs tag -a v1.0.0 -m "ghi chú"  tạo tag có chú thích
  td vcs tag -d v1.0.0              xóa tag`,
	Args: cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := openRepo(cmd)
		if err != nil {
			return err
		}
		annotated, _ := cmd.Flags().GetBool("annotate")
		message, _ := cmd.Flags().GetString("message")
		deleteFlag, _ := cmd.Flags().GetBool("delete")
		listFlag, _ := cmd.Flags().GetBool("list")

		if deleteFlag {
			if len(args) == 0 {
				return exitError("cần chỉ định tên tag cần xóa")
			}
			for _, name := range args {
				if err := r.DeleteTag(name); err != nil {
					return err
				}
				printLine("Đã xóa tag %s", name)
			}
			return nil
		}

		if len(args) == 1 && !listFlag {
			// Một đối số: tạo tag.
			name := args[0]
			if annotated && message == "" {
				return exitError("cần nhập chú thích bằng -m khi tạo tag có chú thích")
			}
			h, err := r.Head()
			if err != nil {
				return err
			}
			if h.IsZero() {
				return exitError("chưa có commit nào để gắn tag")
			}
			if annotated {
				if err := r.WriteAnnotatedTag(name, h, message); err != nil {
					return err
				}
			} else {
				if err := r.WriteTag(name, h); err != nil {
					return err
				}
			}
			printLine("Đã tạo tag %s tại %s", name, h.Short(8))
			return nil
		}

		tags, err := r.ListTags()
		if err != nil {
			return err
		}
		if len(tags) == 0 {
			printLine("Chưa có tag nào.")
			return nil
		}
		for _, t := range tags {
			line := t.Name + " " + t.Target.Short(8)
			if t.Annotated {
				line += " (có chú thích) " + t.Message
			}
			printLine("%s", line)
		}
		return nil
	},
}

func init() {
	// Cờ của branch.
	vcsBranchCmd.Flags().BoolP("list", "l", false, "liệt kê các nhánh")
	vcsBranchCmd.Flags().BoolP("delete", "d", false, "xóa nhánh")
	vcsBranchCmd.Flags().BoolP("force", "D", false, "xóa nhánh bất kể đã hợp nhất hay chưa")
	vcsBranchCmd.Flags().BoolP("verbose", "v", false, "kèm hash của từng nhánh")
	vcsBranchCmd.Flags().BoolP("move", "m", false, "đổi tên nhánh")
	vcsBranchCmd.Flags().BoolP("switch", "s", false, "chuyển sang nhánh mới sau khi đổi tên")
	vcsBranchCmd.Flags().BoolP("track", "t", false, "ghi nhớ nhánh theo dõi cho nhánh mới")

	// Cờ của checkout và switch.
	vcsCheckoutCmd.Flags().StringP("branch", "b", "", "tạo nhánh mới rồi chuyển sang đó")
	vcsCheckoutCmd.Flags().Bool("detach", false, "chuyển tới một commit, không gắn với nhánh")
	vcsCheckoutCmd.Flags().Bool("force", false, "ghi đè các thay đổi chưa lưu trong cây làm việc")

	vcsSwitchCmd.Flags().StringP("create", "c", "", "tạo nhánh mới rồi chuyển sang đó")
	vcsSwitchCmd.Flags().Bool("discard-changes", false, "bỏ qua các thay đổi chưa lưu")

	// Cờ của restore.
	vcsRestoreCmd.Flags().BoolP("staged", "S", false, "chỉ thay đổi vùng stage, giữ nguyên cây làm việc")
	vcsRestoreCmd.Flags().BoolP("worktree", "W", false, "chỉ thay đổi cây làm việc")
	vcsRestoreCmd.Flags().StringP("source", "s", "", "lấy nội dung từ một commit thay vì index")

	// Cờ của reset.
	vcsResetCmd.Flags().Bool("soft", false, "chỉ di chuyển HEAD")
	vcsResetCmd.Flags().Bool("mixed", false, "di chuyển HEAD và nạp lại index")
	vcsResetCmd.Flags().Bool("hard", false, "di chuyển HEAD, index và cây làm việc")

	// Cờ của tag.
	vcsTagCmd.Flags().BoolP("annotate", "a", false, "tạo tag có chú thích")
	vcsTagCmd.Flags().StringP("message", "m", "", "nội dung chú thích cho tag")
	vcsTagCmd.Flags().BoolP("delete", "d", false, "xóa tag")
	vcsTagCmd.Flags().BoolP("list", "l", false, "liệt kê các tag")
}
