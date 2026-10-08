package cmd

import (
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tonguyenducmanh/devcli/internal/vcs/repo"
)

// vcsIgnoreCmd in ra các quy tắc bỏ qua tệp của kho.
//
// Lệnh này chỉ đọc, không sửa gì cả. Nhiệm vụ của nó là chỉ ra quy tắc đang có
// và quy tắc đó đến từ tệp nào, vì quy tắc nằm rải rác ở nhiều tệp thì rất
// khó hình dung tại sao một tệp lại bị bỏ qua.
var vcsIgnoreCmd = &cobra.Command{
	Use:   "ignore",
	Short: "In các quy tắc bỏ qua tệp đang có trong kho",
	Long: `In nội dung mọi tệp chứa quy tắc bỏ qua tệp của kho.

Mỗi tệp in ra kèm đường dẫn để biết quy tắc đó có hiệu lực ở đâu. Tệp ở thư mục
nông in trước, quy tắc ở thư mục sâu hơn in sau và thắng, đúng như khi thực thi.

Nếu kho chưa có quy tắc nào thì lệnh in luôn phần trợ giúp này.

Tệp ignore nằm ở đâu:

  .tmxignore         ở gốc dự án hoặc trong bất kỳ thư mục con nào. Quy tắc chỉ
                     áp dụng bên trong thư mục chứa nó, và không lan sang thư mục
                     khác. Nên commit tệp này để cả nhóm cùng dùng.
  .tmx/info/exclude  riêng cho máy này, nằm trong .tmx nên không được commit.

Cách viết mẫu:

  *.log           bỏ qua mọi tệp kết thúc bằng .log, ở mọi cấp thư mục
  build           bỏ qua thư mục build, ở mọi cấp
  /build          chỉ bỏ qua thư mục build nằm ở gốc dự án
  docs/*.tmp      bỏ qua tệp .tmp nằm trong thư mục docs
  **/cache/       bỏ qua thư mục cache, ở mọi cấp
  !giữ-lại.log    phủ định lại quy tắc trước, tệp này lại được theo dõi
  # ghi chú       dòng bắt đầu bằng dấu # là chú thích

Dấu / ở cuối mẫu nói mẫu đó chỉ áp dụng cho thư mục. Dấu * không vượt qua dấu
/, dấu ** vượt được nhiều cấp. Quy tắc ở dưới thắng quy tắc ở trên.

Hai tệp này được tm đọc tự động, không cần khai báo ở đâu. Tệp bị bỏ qua sẽ
không xuất hiện trong status và không được add vào vùng chuẩn bị.`,
	Example: `  tm vcs ignore          in các quy tắc bỏ qua đang có
  tm vcs ignore --help   xem cách viết mẫu bỏ qua tệp và thư mục`,
	Args: noArgsArg,
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := openRepo(cmd)
		if err != nil {
			return err
		}

		files, err := r.IgnoreFiles()
		if err != nil {
			return err
		}

		// Chỉ đếm dòng thật sự là quy tắc. Dòng trắng và dòng chú thích không có
		// tác dụng, nên một tệp toàn chú thích vẫn phải coi là chưa có quy tắc.
		bodies := make(map[string][]string, len(files))
		total := 0
		for _, f := range files {
			lines := readIgnoreFile(r, f)
			bodies[f] = lines
			total += countRules(lines)
		}

		if total == 0 {
			printLine("Kho chưa có quy tắc bỏ qua tệp nào.")
			printLine("")
			return cmd.Help()
		}

		for _, f := range files {
			printLine("%s", colorize(colorCyan, f))
			if len(bodies[f]) == 0 {
				printLine("  (rỗng)")
				printLine("")
				continue
			}
			for _, line := range bodies[f] {
				printLine("  %s", line)
			}
			printLine("")
		}
		printLine("Tổng cộng %d quy tắc trong %d tệp.", total, len(files))
		return nil
	},
}

// readIgnoreFile đọc một tệp ignore theo đường dẫn tương đối, bỏ khoảng trắng
// thừa ở hai đầu dòng và bỏ dòng trắng.
//
// Tệp đọc lỗi thì coi như rỗng, vì lệnh này chỉ để xem, không nên vì một tệp hỏng
// mà không in ra gì cả.
func readIgnoreFile(r *repo.Repo, rel string) []string {
	data, err := os.ReadFile(r.WorkPath(rel))
	if err != nil {
		return nil
	}
	var out []string
	for _, line := range strings.Split(string(data), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			out = append(out, line)
		}
	}
	return out
}

// countRules đếm số dòng là quy tắc thật, bỏ qua dòng chú thích.
func countRules(lines []string) int {
	var n int
	for _, line := range lines {
		if !strings.HasPrefix(line, "#") {
			n++
		}
	}
	return n
}
