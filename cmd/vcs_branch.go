package cmd

import (
	"fmt"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tonguyenducmanh/devcli/internal/vcs/object"
	"github.com/tonguyenducmanh/devcli/internal/vcs/ops"
	"github.com/tonguyenducmanh/devcli/internal/vcs/repo"
)

// vcsBranchCmd quản lý nhánh.
//
// Các cờ bám theo git branch, trừ những cái liên quan tới nhánh từ xa vì td
// không có khái niệm remote.
var vcsBranchCmd = &cobra.Command{
	Use:     "branch",
	Aliases: []string{"br"},
	Short:   "Liệt kê, tạo, sao chép, đổi tên hoặc xoá các nhánh cục bộ",
	Long: `Liệt kê, tạo, sao chép, đổi tên và xoá các nhánh cục bộ.

Nhánh chỉ là một con trỏ trỏ tới một commit nên tạo nhánh không tốn chi phí sao
chép. Dấu * đánh dấu nhánh đang đứng.

Xoá nhánh chỉ thành công nếu nhánh đó đã được hợp nhất vào nhánh hiện tại; dùng
-D để bỏ qua kiểm tra này. Không thể xoá nhánh đang đứng.

Không có tham số thì in danh sách. Một tham số là tạo nhánh mới tại HEAD, hai
tham số là tạo nhánh mới từ một điểm xuất phát cho trước.

Cờ liệt kê và cờ lọc dùng chung được, ví dụ --merged cùng --sort. Tham số
đưa vào không phải thao tác thì được hiểu là mẫu lọc tên nhánh, giống git.`,
	Example: `  td vcs branch                  liệt kê các nhánh
  td vcs branch -v               liệt kê kèm mã băm và tiêu đề commit
  td vcs branch -vv              như trên, thêm cả nhánh đang theo dõi
  td vcs branch --show-current   in tên nhánh đang đứng
  td vcs branch -l 'tinh-*'      liệt kê các nhánh khớp mẫu
  td vcs branch --merged main    chỉ liệt kê nhánh đã hợp nhất vào main
  td vcs branch -d ten           xoá nhánh đã hợp nhất
  td vcs branch -D ten           xoá nhánh bất kể trạng thái
  td vcs branch ten              tạo nhánh và chuyển sang đó
  td vcs branch ten main         tạo nhánh từ nhánh main
  td vcs branch -m cũ mới        đổi tên nhánh
  td vcs branch -c ten bản-sao   sao chép nhánh ten thành bản-sao`,
	Args: arbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := openRepo(cmd)
		if err != nil {
			return err
		}

		switch {
		case wantsDelete(cmd):
			return runBranchDelete(cmd, r, args)
		case wantsRename(cmd):
			return runBranchRename(cmd, r, args)
		case wantsCopy(cmd):
			return runBranchCopy(cmd, r, args)
		case len(args) == 1 && !wantsListOnly(cmd):
			return runBranchCreate(cmd, r, args[0], "HEAD")
		case len(args) == 2 && !wantsListOnly(cmd):
			return runBranchCreate(cmd, r, args[0], args[1])
		default:
			return runBranchList(cmd, r, args)
		}
	},
}

// ─── Cờ của branch ─────────────────────────────────────────────────

func init() {
	// Cờ thao tác, theo đúng nghĩa của git branch.
	vcsBranchCmd.Flags().BoolP("list", "l", false, "liệt kê tên nhánh, có thể kèm mẫu lọc")
	vcsBranchCmd.Flags().BoolP("delete", "d", false, "xoá nhánh đã hợp nhất")
	vcsBranchCmd.Flags().BoolP("force", "D", false, "xoá nhánh bất kể đã hợp nhất hay chưa")
	vcsBranchCmd.Flags().BoolP("move", "m", false, "đổi tên nhánh")
	vcsBranchCmd.Flags().BoolP("copy", "c", false, "sao chép nhánh")
	vcsBranchCmd.Flags().BoolP("track", "t", false, "ghi nhớ nhánh theo dõi cho nhánh mới")
	vcsBranchCmd.Flags().Bool("show-current", false, "chỉ in tên nhánh đang đứng")
	vcsBranchCmd.Flags().BoolP("quiet", "q", false, "chỉ báo lỗi, không in thông báo thành công")
	vcsBranchCmd.Flags().BoolP("create-force", "f", false, "ép tạo, ghi đè nhánh đã có")

	// -v là cờ đếm chứ không phải cờ bật tắt: dùng hai lần thì in thêm nhánh
	// đang theo dõi, giống git. CountP tự đặt NoOptDefVal là "+1" để mỗi lần
	// gặp lại cờ thì tăng thêm một, đừng đặt lại thành "1".
	vcsBranchCmd.Flags().CountP("verbose", "v", "liệt kê kèm mã băm và tiêu đề commit, dùng hai lần thì in thêm nhánh đang theo dõi")

	// Cờ lọc danh sách.
	vcsBranchCmd.Flags().String("merged", "", "chỉ in nhánh đã hợp nhất vào commit cho trước")
	vcsBranchCmd.Flags().String("no-merged", "", "chỉ in nhánh chưa hợp nhất vào commit cho trước")
	vcsBranchCmd.Flags().String("contains", "", "chỉ in nhánh có chứa commit cho trước")
	vcsBranchCmd.Flags().String("points-at", "", "chỉ in nhánh trỏ tới đúng object cho trước")
	vcsBranchCmd.Flags().String("sort", "", "sắp xếp theo khoá: name, -name, committerdate, -committerdate")
	vcsBranchCmd.Flags().String("abbrev", "", "số ký tự của mã băm rút gọn, mặc định 8, đặt 0 để in đầy đủ")
	vcsBranchCmd.Flags().String("format", "", "định dạng từng dòng: %s tên, %h mã băm ngắn, %H mã băm đầy đủ, %d tiêu đề")

	// Cờ của checkout và switch.
	vcsCheckoutCmd.Flags().StringP("branch", "b", "", "tạo nhánh mới rồi chuyển sang đó")
	vcsCheckoutCmd.Flags().Bool("detach", false, "chuyển tới một commit, không gắn với nhánh")
	vcsCheckoutCmd.Flags().Bool("force", false, "ghi đè các thay đổi chưa lưu trong cây làm việc")
	vcsSwitchCmd.Flags().StringP("create", "c", "", "tạo nhánh mới rồi chuyển sang đó")
	vcsSwitchCmd.Flags().Bool("discard-changes", false, "bỏ qua các thay đổi chưa lưu")

	// Cờ của restore.
	vcsRestoreCmd.Flags().BoolP("staged", "S", false, "chỉ thay đổi vùng stage, giữ nguyên cây làm việc")
}

// ─── Thao tác của branch ───────────────────────────────────────────

// wantsDelete báo lệnh đang được yêu cầu xoá nhánh hay không.
func wantsDelete(cmd *cobra.Command) bool {
	d, _ := cmd.Flags().GetBool("delete")
	f, _ := cmd.Flags().GetBool("force")
	return d || f
}

// wantsRename báo lệnh đang được yêu cầu đổi tên nhánh hay không.
func wantsRename(cmd *cobra.Command) bool {
	m, _ := cmd.Flags().GetBool("move")
	return m
}

// wantsCopy báo lệnh đang được yêu cầu sao chép nhánh hay không.
func wantsCopy(cmd *cobra.Command) bool {
	c, _ := cmd.Flags().GetBool("copy")
	return c
}

// wantsListOnly báo lệnh đang ở chế độ chỉ liệt kê, không tạo nhánh mới.
func wantsListOnly(cmd *cobra.Command) bool {
	l, _ := cmd.Flags().GetBool("list")
	s, _ := cmd.Flags().GetBool("show-current")
	return l || s
}

// shorthandSet đọc cờ viết tắt theo ký hiệu, ví dụ shorthandSet(cmd, "D") là -D.
func shorthandSet(cmd *cobra.Command, shorthand string) bool {
	f := cmd.Flags().ShorthandLookup(shorthand)
	if f == nil {
		return false
	}
	v, err := cmd.Flags().GetBool(f.Name)
	return err == nil && v
}

// runBranchDelete xoá các nhánh nêu trong args.
func runBranchDelete(cmd *cobra.Command, r *repo.Repo, args []string) error {
	if len(args) == 0 {
		return exitError("cần chỉ định tên ít nhất một nhánh cần xoá")
	}
	quiet, _ := cmd.Flags().GetBool("quiet")

	results, err := ops.DeleteBranches(r, args, shorthandSet(cmd, "D"))
	if err != nil {
		return err
	}
	var failed int
	for _, res := range results {
		if !res.Deleted {
			failed++
			printErr("không xoá được %s: %s", res.Name, res.Reason)
			continue
		}
		if !quiet {
			printLine("Đã xoá nhánh %s", res.Name)
		}
	}
	if failed > 0 {
		return exitError("%d nhánh không được xoá", failed)
	}
	return nil
}

// runBranchRename đổi tên một nhánh.
func runBranchRename(cmd *cobra.Command, r *repo.Repo, args []string) error {
	if len(args) != 2 {
		return exitError("cần chỉ định tên cũ và tên mới")
	}
	swap, _ := cmd.Flags().GetBool("switch")
	if err := ops.RenameBranch(r, args[0], args[1], swap); err != nil {
		return err
	}
	quiet, _ := cmd.Flags().GetBool("quiet")
	if !quiet {
		printLine("Đã đổi tên %s thành %s", args[0], args[1])
	}
	return nil
}

// runBranchCopy tạo một nhánh mới trỏ cùng chỗ với nhánh có sẵn.
func runBranchCopy(cmd *cobra.Command, r *repo.Repo, args []string) error {
	if len(args) != 2 {
		return exitError("cần chỉ định tên nhánh nguồn và tên nhánh mới")
	}
	if err := ops.CreateBranch(r, ops.CreateBranchOptions{
		Name:        args[1],
		StartPoint:  args[0],
		ForceCreate: true,
	}); err != nil {
		return err
	}
	quiet, _ := cmd.Flags().GetBool("quiet")
	if !quiet {
		printLine("Đã sao chép %s thành %s", args[0], args[1])
	}
	return nil
}

// ─── Liệt kê nhánh ────────────────────────────────────────────────

// branchListOptions gom các cờ lọc và định dạng của chế độ liệt kê.
type branchListOptions struct {
	patterns []string
	verbose  int
	abbrev   int
	format   string
	sortKey  string
	merged   string
	noMerged string
	contains string
	pointsAt string
	showOnly bool
}

// readBranchListOptions đọc các cờ của chế độ liệt kê từ dòng lệnh.
func readBranchListOptions(cmd *cobra.Command) branchListOptions {
	opts := branchListOptions{}
	opts.verbose, _ = cmd.Flags().GetCount("verbose")
	opts.format, _ = cmd.Flags().GetString("format")
	opts.sortKey, _ = cmd.Flags().GetString("sort")
	opts.merged, _ = cmd.Flags().GetString("merged")
	opts.noMerged, _ = cmd.Flags().GetString("no-merged")
	opts.contains, _ = cmd.Flags().GetString("contains")
	opts.pointsAt, _ = cmd.Flags().GetString("points-at")

	// Mặc định rút gọn còn 8 ký tự cho khớp với phần còn lại của output.
	// --abbrev=0 nghĩa là in mã băm đầy đủ.
	opts.abbrev = 8
	if v, _ := cmd.Flags().GetString("abbrev"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			return opts
		}
		opts.abbrev = n
	}
	return opts
}

// runBranchList in danh sách nhánh theo các cờ lọc.
func runBranchList(cmd *cobra.Command, r *repo.Repo, args []string) error {
	opts := readBranchListOptions(cmd)
	opts.patterns = args

	if only, _ := cmd.Flags().GetBool("show-current"); only {
		name, err := r.CurrentBranch()
		if err != nil {
			return err
		}
		if name != "" {
			printLine("%s", name)
		}
		return nil
	}

	branches, err := ops.ListBranches(r)
	if err != nil {
		return err
	}

	filtered, err := filterBranches(r, branches, opts)
	if err != nil {
		return err
	}
	if len(filtered) == 0 {
		printLine("Không có nhánh nào khớp.")
		return nil
	}

	// Nhánh đang đứng luôn đứng đầu, y hệt git, bất kể sắp xếp thế nào.
	currentName, _ := r.CurrentBranch()
	sortBranches(filtered, currentName, opts.sortKey)

	for _, b := range filtered {
		printLine("%s", formatBranch(b, currentName, opts))
	}
	return nil
}

// filterBranches áp dụng toàn bộ cờ lọc lên danh sách nhánh.
func filterBranches(r *repo.Repo, branches []ops.Branch, opts branchListOptions) ([]ops.Branch, error) {
	var out []ops.Branch

	// Mỗi điều kiện lọc theo lịch sử cần phân giải tham chiếu một lần rồi
	// dùng lại cho mọi nhánh, nên phân giải trước rồi mới lọc.
	var mergedHash, noMergedHash, containsHash object.Hash
	var err error
	if opts.merged != "" {
		if mergedHash, err = r.ResolveRev(opts.merged); err != nil {
			return nil, err
		}
	}
	if opts.noMerged != "" {
		if noMergedHash, err = r.ResolveRev(opts.noMerged); err != nil {
			return nil, err
		}
	}
	if opts.contains != "" {
		if containsHash, err = r.ResolveRev(opts.contains); err != nil {
			return nil, err
		}
	}
	var pointsAtHash object.Hash
	if opts.pointsAt != "" {
		if pointsAtHash, err = r.ResolveRev(opts.pointsAt); err != nil {
			return nil, err
		}
	}

	for _, b := range branches {
		if !matchPatterns(b.Name, opts.patterns) {
			continue
		}
		if pointsAtHash != (object.Hash{}) && b.Hash != pointsAtHash {
			continue
		}
		if mergedHash != (object.Hash{}) {
			ok, err := r.IsAncestor(mergedHash, b.Hash)
			if err != nil || !ok {
				continue
			}
		}
		if noMergedHash != (object.Hash{}) {
			ok, err := r.IsAncestor(noMergedHash, b.Hash)
			if err == nil && ok {
				continue
			}
		}
		if containsHash != (object.Hash{}) {
			ok, err := r.IsAncestor(containsHash, b.Hash)
			if err != nil || !ok {
				continue
			}
		}
		out = append(out, b)
	}
	return out, nil
}

// matchPatterns kiểm tra tên nhánh có khớp ít nhất một mẫu nào không.
// Không có mẫu thì mọi nhánh đều khớp.
func matchPatterns(name string, patterns []string) bool {
	if len(patterns) == 0 {
		return true
	}
	for _, p := range patterns {
		if matched, err := path.Match(p, name); err == nil && matched {
			return true
		}
	}
	return false
}

// sortBranches sắp xếp danh sách nhánh theo khoá đã cho.
// Nhánh đang đứng được ghim lên đầu, giống git.
func sortBranches(branches []ops.Branch, currentName, key string) {
	desc := strings.HasPrefix(key, "-")
	key = strings.TrimPrefix(key, "-")

	less := func(i, j int) bool { return branches[i].Name < branches[j].Name }
	switch key {
	case "committerdate":
		less = func(i, j int) bool {
			return branchDate(branches[i]) < branchDate(branches[j])
		}
	case "refname", "":
		// Mặc định là theo tên, đã nằm ở trên.
	default:
		// Khoá lạ thì giữ nguyên thứ tự tên, không đoán bừa.
	}

	sort.SliceStable(branches, less)
	if desc {
		for i, j := 0, len(branches)-1; i < j; i, j = i+1, j-1 {
			branches[i], branches[j] = branches[j], branches[i]
		}
	}

	// Đưa nhánh đang đứng lên đầu sau khi đã sắp xếp phần còn lại.
	for i, b := range branches {
		if b.Name == currentName && i > 0 {
			copy(branches[1:i+1], branches[0:i])
			branches[0] = b
			break
		}
	}
}

// branchDate lấy thời điểm commit cuối cùng của nhánh, dùng cho --sort.
func branchDate(b ops.Branch) int64 {
	if b.CommittedAt.IsZero() {
		return 0
	}
	return b.CommittedAt.Unix()
}

// shortHash rút gọn mã băm, số 0 nghĩa là in đầy đủ.
//
// Không dùng Hash.Short vì hàm đó hiểu số 0 là "dùng mặc định 7", còn ở đây 0
// là ý rõ ràng của người dùng: in mã băm đầy đủ.
func shortHash(h object.Hash, n int) string {
	if n <= 0 {
		return h.String()
	}
	return h.Short(n)
}

// formatBranch dựng một dòng output cho nhánh, theo các cờ định dạng.
func formatBranch(b ops.Branch, currentName string, opts branchListOptions) string {
	marker := "  "
	if b.Name == currentName {
		marker = "* "
	}

	// --format thay thế toàn bộ, kể cả dấu đánh dấu, giống git.
	if opts.format != "" {
		return marker + strings.NewReplacer(
			"%s", b.Name,
			"%h", shortHash(b.Hash, opts.abbrev),
			"%H", b.Hash.String(),
			"%d", b.Subject,
		).Replace(opts.format)
	}

	line := marker + b.Name
	if opts.verbose > 0 {
		line += " " + shortHash(b.Hash, opts.abbrev)
		if b.Subject != "" {
			line += " " + b.Subject
		}
	}
	// Nhánh theo dõi chỉ hiện từ -vv trở lên, giống git.
	if opts.verbose > 1 && b.Upstream != "" {
		line += fmt.Sprintf(" [%s: đi trước %d, đi sau %d]", b.Upstream, b.Ahead, b.Behind)
	}
	return line
}

// runBranchCreate tạo nhánh mới rồi chuyển sang đó.
func runBranchCreate(cmd *cobra.Command, r *repo.Repo, name, startPoint string) error {
	track, _ := cmd.Flags().GetBool("track")
	force := shorthandSet(cmd, "f")
	quiet, _ := cmd.Flags().GetBool("quiet")

	if err := ops.CreateBranch(r, ops.CreateBranchOptions{
		Name:        name,
		StartPoint:  startPoint,
		Switch:      true,
		Track:       track,
		ForceCreate: force,
	}); err != nil {
		return err
	}
	if !quiet {
		printLine("Đã tạo nhánh %s từ %s", name, startPoint)
	}
	return nil
}

// vcsCheckoutCmd chuyển nhánh hoặc commit.
var vcsCheckoutCmd = &cobra.Command{
	Use:     "checkout",
	Aliases: []string{"co"},
	Short:   "Chuyển sang nhánh hoặc commit khác",
	Long: `Di chuyển con trỏ HEAD sang nhánh, tag hoặc commit khác, đồng thời đưa vùng
chuẩn bị và cây làm việc về đúng nội dung của đích.

Khi chuyển sang một commit cụ thể, HEAD trở nên tách rời khỏi nhánh. Các thay
đổi chưa lưu sẽ không bị ghi đè, lệnh báo lỗi để bạn xử lý trước.`,
	Example: `  td vcs checkout main           chuyển sang nhánh main
  td vcs checkout -b moi         tạo nhánh moi rồi chuyển sang đó
  td vcs checkout abc1234        chuyển tới một commit cụ thể (HEAD tách rời)
  td vcs checkout --detach main  chuyển tới commit của main`,
	Args: maximumArgs(1, "[nhánh|commit]"),
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
	Long: `Chuyển nhánh đang làm việc. Lệnh rút gọn của checkout dành cho trường hợp
chỉ cần đổi nhánh, không cần thêm tuỳ chọn nào khác.`,
	Example: `  td vcs switch main
  td vcs switch -c moi        tạo nhánh moi rồi chuyển sang đó`,
	Args: maximumArgs(1, "<nhánh>"),
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
	Long: `Đưa lại nội dung tệp từ một nguồn khác về cây làm việc hoặc vùng chuẩn bị.

Mặc định lấy nội dung đang có trong vùng chuẩn bị, tức là huỷ các sửa đổi chưa
stage. Dùng --staged để chỉ gỡ khỏi vùng chuẩn bị mà giữ nguyên cây làm việc,
và --source để lấy từ một commit bất kỳ.`,
	Example: `  td vcs restore main.go         lấy lại nội dung đang stage
  td vcs restore --staged main.go  gỡ thay đổi đã stage
  td vcs restore --source=HEAD~1 main.go  lấy từ một commit khác`,
	Args: minimumArgs(1, "<tệp>..."),
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
	Long: `Di chuyển con trỏ HEAD về một commit khác, kèm mức độ áp dụng cho vùng chuẩn bị
và cây làm việc.

  --soft   chỉ di chuyển HEAD, giữ nguyên vùng chuẩn bị và cây làm việc
  --mixed  di chuyển HEAD và nạp lại vùng chuẩn bị (mặc định)
  --hard   di chuyển HEAD, vùng chuẩn bị và cây làm việc

Chế độ --hard ghi đè mọi thay đổi chưa lưu, dùng cẩn thận.

Nếu có danh sách tệp sau dấu hai gạch ngang, lệnh chỉ gỡ những tệp đó khỏi vùng
chuẩn bị, không động tới con trỏ HEAD.`,
	Example: `  td vcs reset --soft HEAD~1
  td vcs reset HEAD~1 -- main.go   chỉ gỡ main.go khỏi vùng stage`,
	Args: arbitraryArgs,
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
	Long: `Đánh dấu một điểm trong lịch sử bằng tên dễ nhớ, ví dụ theo phiên bản.

Tag nhẹ chỉ là một con trỏ trỏ tới commit. Tag có chú thích (-a kèm -m) tạo
thêm một object riêng nên lưu được lời giải thích, ai đó và thời điểm tạo.

Một tham số là tạo tag tại HEAD, không có tham số thì in danh sách.

Dùng -q khi cần đọc danh sách bằng kịch bản: lệnh sẽ in ra rỗng thay vì in
thông báo "Chưa có tag nào", nên không phải lọc bỏ câu văn bản.`,
	Example: `  td vcs tag                        liệt kê các tag
  td vcs tag v1.0.0                 tạo tag nhẹ
  td vcs tag -a v1.0.0 -m "ghi chú"  tạo tag có chú thích
  td vcs tag -d v1.0.0              xóa tag`,
	Args: arbitraryArgs,
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
			// Với -q thì im lặng hoàn toàn, để kịch bản đọc được danh sách
			// rỗng mà không phải lọc bỏ câu thông báo.
			quiet, _ := cmd.Flags().GetBool("quiet")
			if !quiet {
				printLine("Chưa có tag nào.")
			}
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
	// Cờ của restore, phần còn lại.
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
	vcsTagCmd.Flags().BoolP("quiet", "q", false, "liệt kê và im lặng khi không có tag nào, để dùng trong kịch bản")
}
