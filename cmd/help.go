package cmd

import (
	"fmt"
	"runtime"
	"slices"
	"strings"

	"github.com/spf13/cobra"
)

// Nhãn tiếng Việt thay cho nhãn tiếng Anh của cobra.
const (
	helpCmdUsage    = "hiển thị trợ giúp của một lệnh"
	helpFlagUsage   = "hiển thị phần trợ giúp của lệnh này"
	versionFlagHelp = "hiển thị phiên bản rồi thoát"
)

// usageTemplate là khuôn trợ giúp tiếng Việt, thay cho khuôn mặc định của
// cobra. Giữ nguyên cấu trúc và điều kiện của khuôn gốc, chỉ dịch nhãn.
const usageTemplate = `Cách dùng:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [lệnh]{{end}}{{if gt (len .Aliases) 0}}

Tên gọi khác:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Ví dụ:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}{{$cmds := .Commands}}{{if eq (len .Groups) 0}}

Các lệnh có sẵn:{{range $cmds}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{else}}{{range $group := .Groups}}

{{.Title}}{{range $cmds}}{{if (and (eq .GroupID $group.ID) (or .IsAvailableCommand (eq .Name "help")))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if not .AllChildCommandsHaveGroup}}

Lệnh khác:{{range $cmds}}{{if (and (eq .GroupID "") (or .IsAvailableCommand (eq .Name "help")))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Cờ:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Cờ toàn cục:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Chủ đề trợ giúp khác:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Xem chi tiết một lệnh bằng: {{.CommandPath}} [lệnh] --help{{end}}
`

// helpTemplate là khuôn phần trợ giúp đầy đủ, thay cho khuôn của cobra.
const helpTemplate = `{{with (or .Long .Short)}}{{. | trimTrailingWhitespaces}}

{{end}}{{if or .Runnable .HasSubCommands}}{{.UsageString}}{{end}}`

// contactLines là các dòng thông tin liên hệ in ở cuối output.
var contactLines = buildContactLines()

// buildContactLines gom tên tác giả và nơi phát hành thành các dòng sẵn sàng
// để in. Biến toàn cục được gắn lúc biên dịch nên hàm này chạy một lần khi
// khởi động, sau khi ldflags đã gán giá trị.
func buildContactLines() []string {
	lines := make([]string, 0, 2)
	if Author != "" {
		lines = append(lines, "Tác giả: "+Author)
	}
	if RepoURL != "" {
		lines = append(lines, "Mã nguồn: "+RepoURL)
	}
	return lines
}

// setupHelp đưa toàn bộ cây lệnh về dùng mô tả tiếng Việt.
//
// Cobra tự sinh cờ --help, lệnh `help` và khuôn trợ giúp với nhãn tiếng Anh.
// Hàm này ghi đè lại từng thứ một, gọi trong init() sau khi cây lệnh đã dựng
// xong.
//
// Cờ --help và --version được tự khai báo thay vì gọi InitDefaultHelpFlag của
// cobra: hàm đó ép nối cờ toàn cục của lệnh cha vào lệnh con, khiến một lệnh
// con nào đó khai báo trùng tên cờ sẽ hỏng lúc khởi động.
func setupHelp() {
	var walk func(cmd *cobra.Command)
	walk = func(cmd *cobra.Command) {
		cmd.SetUsageTemplate(usageTemplate)
		cmd.SetHelpTemplate(helpTemplate)

		if cmd.Flags().Lookup("help") == nil {
			cmd.Flags().BoolP("help", "h", false, helpFlagUsage)
		}

		for _, sub := range cmd.Commands() {
			walk(sub)
		}
	}
	walk(rootCmd)

	// Cờ --version chỉ có ở lệnh gốc, và không lấy chữ viết tắt vì -v đã thuộc
	// về cờ --verbose.
	if rootCmd.Flags().Lookup("version") == nil {
		rootCmd.Flags().Bool("version", false, versionFlagHelp)
	}

	// Trợ giúp của lệnh gốc kết thúc bằng tên tác giả và nơi phát hành. Ghép
	// thẳng vào khuôn vì nội dung này cố định suốt vòng đời tiến trình, không
	// cần tính lúc in.
	if len(contactLines) > 0 {
		rootCmd.SetUsageTemplate(usageTemplate + "\n" + strings.Join(contactLines, "\n") + "\n")
	}

	// Lệnh `help` do ta tự tạo nên không nằm trong vòng duyệt phía trên,
	// phải khai báo cờ --help riêng cho nó.
	helpCmd := newHelpCmd()
	if helpCmd.Flags().Lookup("help") == nil {
		helpCmd.Flags().BoolP("help", "h", false, helpFlagUsage)
	}
	rootCmd.SetHelpCommand(helpCmd)
}

// newHelpCmd tạo lệnh `help` tiếng Việt thay cho lệnh mặc định của cobra.
func newHelpCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "help [lệnh]",
		Short: helpCmdUsage,
		Long: fmt.Sprintf(`Xem trợ giúp của một lệnh bất kỳ.

Gõ %s help <tên lệnh> để xem mô tả, cú pháp, các ví dụ và các cờ của lệnh đó.
Gõ không kèm tên lệnh thì xem trợ giúp của lệnh gốc.`, AppName),
		Example: strings.Join([]string{
			"  " + AppName + " help vcs          xem trợ giúp của nhóm lệnh vcs",
			"  " + AppName + " help vcs merge    xem trợ giúp của lệnh merge",
		}, "\n"),
		// Nhận đường dẫn lệnh, ví dụ `tm help vcs merge`.
		Args: arbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Không có tên lệnh thì in trợ giúp của lệnh gốc.
			if len(args) == 0 {
				return rootCmd.Help()
			}
			// Find trả về lệnh khớp gần nhất cùng phần đối số còn thừa, nên
			// phải kiểm tra phần thừa thì mới biết người dùng có gõ sai tên
			// không. `tm help vcs merge` thì phần thừa rỗng.
			target, leftover, err := rootCmd.Find(args)
			if err != nil {
				return exitError("không tìm thấy lệnh %q", strings.Join(args, " "))
			}
			if len(leftover) > 0 {
				return exitError("%s không có lệnh con tên %q, xem danh sách: %s --help",
					target.CommandPath(), leftover[0], target.CommandPath())
			}
			return target.Help()
		},
	}
}

// printContact in tên tác giả và nơi phát hành, dùng ở cuối output.
func printContact() {
	if len(contactLines) == 0 {
		return
	}
	printLine("")
	for _, line := range contactLines {
		printLine("%s", line)
	}
}

// printCommandList in tên và mô tả ngắn của các lệnh con, theo từng nhóm.
//
// Dùng khi lệnh gốc hoặc một nhóm lệnh được gọi mà không kèm lệnh con nào.
// Như vậy lệnh đó vẫn làm được việc gì đó hữu ích thay vì chỉ in trợ giúp.
func printCommandList(cmd *cobra.Command) {
	// Gom lệnh con theo nhóm, giữ đúng thứ tự khai báo và bỏ trùng tên nhóm.
	var titles []string
	titleOf := map[string]string{}
	for _, sub := range cmd.Commands() {
		if !sub.IsAvailableCommand() {
			continue
		}
		title := "Lệnh khác:"
		for _, g := range cmd.Groups() {
			if g.ID == sub.GroupID {
				title = g.Title
				break
			}
		}
		titleOf[sub.Name()] = title
		if !slices.Contains(titles, title) {
			titles = append(titles, title)
		}
	}

	// Độ rộng bằng tên dài nhất để các mô tả thẳng hàng.
	width := 0
	for _, sub := range cmd.Commands() {
		if sub.IsAvailableCommand() && len(sub.Name()) > width {
			width = len(sub.Name())
		}
	}

	for _, title := range titles {
		printLine("%s", colorize(colorCyan, title))
		for _, sub := range cmd.Commands() {
			if !sub.IsAvailableCommand() || titleOf[sub.Name()] != title {
				continue
			}
			printLine("  %-*s  %s", width, sub.Name(), sub.Short)
		}
		printLine("")
	}

	printLine("Xem chi tiết một lệnh: %s <tên lệnh> --help", cmd.CommandPath())
	printContact()
}

// runRoot xử lý khi lệnh gốc được gọi mà không kèm lệnh con nào.
//
// Gọi lệnh gốc không có mục đích gì thì in tên, phiên bản và danh sách lệnh.
// Cờ -v bật thêm thông tin môi trường, vốn chẳng có tác dụng gì nếu chỉ in
// trợ giúp.
func runRoot(cmd *cobra.Command) error {
	if verboseOn(cmd) {
		printVersionFull()
		return nil
	}

	printLine("%s %s", AppName, Version)
	printLine("")
	printCommandList(cmd)
	return nil
}

// printVersionFull in thông tin môi trường đầy đủ, dùng khi bật cờ -v.
func printVersionFull() {
	printLine("%s phiên bản %s", AppName, Version)
	printLine("nền tảng: %s/%s, %s", runtime.GOOS, runtime.GOARCH, runtime.Version())
	printContact()
}
