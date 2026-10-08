package cmd

import (
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/tonguyenducmanh/devcli/internal/vcs/repo"
)

// newVCSCmd tạo nhóm lệnh quản lý phiên bản.
// Các lệnh con được khai báo trong các tệp cmd/vcs_*.go nên dễ bổ sung lệnh mới.
func newVCSCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "vcs",
		GroupID: groupVersionControl,
		Short:   "Quản lý phiên bản mã nguồn cục bộ",
		Long: `Nhóm lệnh quản lý phiên bản mã nguồn cục bộ.

Toàn bộ dữ liệu của nhóm này nằm trong thư mục .tdx cạnh dự án, gồm lịch
sử commit, các nhánh, các tag và vùng chuẩn bị.

Mỗi lệnh dưới đây chạy trên kho tìm thấy bằng cách đi lên từ thư mục hiện
tại. Dùng -C để chỉ định thư mục khác.

BỎ QUA TỆP KHÔNG MUỐN THEO DÕI

td không tự đoán tệp nào là tạm, tệp nào là dữ liệu do trình biên dịch sinh
ra. Muốn bỏ qua thì ghi mẫu vào tệp .tdxignore ở gốc dự án, mỗi dòng một mẫu:

  *.log           bỏ qua mọi tệp kết thúc bằng .log, ở mọi cấp thư mục
  build           bỏ qua thư mục build, ở mọi cấp
  /build          chỉ bỏ qua thư mục build nằm ở gốc dự án
  docs/*.tmp      bỏ qua tệp .tmp nằm trong thư mục docs
  **/cache/       bỏ qua thư mục cache, ở mọi cấp
  !giữ-lại.log    phủ định lại quy tắc trước, tệp này lại được theo dõi
  # ghi chú       dòng bắt đầu bằng dấu # là chú thích

Dấu / ở cuối mẫu nói mẫu đó chỉ áp dụng cho thư mục. Dấu * không vượt qua dấu
/, dấu ** vượt được nhiều cấp. Quy tắc ở dưới thắng quy tắc ở trên.

Muốn quy tắc chỉ áp dụng cho riêng máy này thì ghi vào .tdx/info/exclude, tệp
đó nằm trong .tdx nên không được commit. Còn .tdxignore nằm ở gốc dự án nên
có thể commit để cả nhóm cùng dùng.

Quy tắc cũng đặt được trong thư mục con, khi đó nó chỉ áp dụng bên trong thư mục
đó chứ không lan sang nơi khác.

Hai tệp này được td đọc tự động, không cần khai báo ở đâu. Tệp bị bỏ qua sẽ
không xuất hiện trong status và không được add vào vùng chuẩn bị.

Xem các quy tắc đang có trong kho: td vcs ignore`,
		Example: `  # Khởi tạo kho rồi ghi lại thay đổi đầu tiên
  td vcs init
  td vcs add .
  td vcs commit -m "tin nhắn đầu tiên"

  # Xem nhánh hiện tại và lịch sử gọn
  td vcs status
  td vcs log --oneline -n 10

  # Tạo nhánh, làm việc rồi hợp nhất về nhánh chính
  td vcs switch -c tinh-nang
  td vcs commit -am "bổ sung tính năng"
  td vcs switch main
  td vcs merge tinh-nang`,
		// Gọi nhóm lệnh mà không kèm lệnh con thì in danh sách lệnh con,
		// thay vì in cả trang trợ giúp dài. Cờ -v của nhóm này in thêm tình
		// trạng kho mã nguồn hiện tại.
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				return exitError("không có lệnh nào tên %q, xem danh sách: %s vcs --help", args[0], AppName)
			}
			if verboseOn(cmd) {
				printRepoSummary(cmd)
			}
			printCommandList(cmd)
			return nil
		},
	}
	// Cờ của nhóm. Không phải cờ toàn cục, chỉ áp dụng từ vcs trở xuống.
	cmd.PersistentFlags().StringP("dir", "C", "", "chạy lệnh tại thư mục khác")
	cmd.Flags().BoolP("verbose", "v", false, "in thêm tình trạng kho mã nguồn hiện tại")

	cmd.AddCommand(
		vcsInitCmd,
		vcsIgnoreCmd,
		vcsStatusCmd,
		vcsAddCmd,
		vcsCommitCmd,
		vcsLogCmd,
		vcsDiffCmd,
		vcsShowCmd,
		vcsBranchCmd,
		vcsCheckoutCmd,
		vcsSwitchCmd,
		vcsRestoreCmd,
		vcsMergeCmd,
		vcsRebaseCmd,
		vcsCherryPickCmd,
		vcsRevertCmd,
		vcsStashCmd,
		vcsResetCmd,
		vcsTagCmd,
		vcsRmCmd,
		vcsCleanCmd,
		vcsMvCmd,
		vcsReflogCmd,
		vcsHashObjectCmd,
		vcsCatFileCmd,
		vcsFsckCmd,
	)
	return cmd
}

// printRepoSummary in tình trạng kho mã nguồn tìm thấy quanh thư mục hiện tại.
//
// Chạy được cả khi chưa có kho: khi đó chỉ in một dòng gợi ý, vì việc chưa có
// kho không phải lỗi khi người dùng chỉ muốn xem danh sách lệnh.
func printRepoSummary(cmd *cobra.Command) {
	dir, err := cmd.Flags().GetString("dir")
	if err != nil || dir == "" {
		dir = "."
	}

	r, err := repo.Open(dir)
	if err != nil {
		printLine("Chưa có kho mã nguồn ở đây, hãy chạy `%s vcs init` để khởi tạo.", AppName)
		return
	}

	branch, _ := r.CurrentBranch()
	if branch == "" {
		branch = "(chưa có nhánh)"
	}
	branches, _ := r.Branches()
	tags, _ := r.Tags()

	printLine("Kho: %s", r.Root)
	printLine("Dữ liệu: %s", filepath.Join(r.Root, repo.DirName))
	printLine("Nhánh hiện tại: %s", branch)
	printLine("Số nhánh: %d, số tag: %d", len(branches), len(tags))
}
