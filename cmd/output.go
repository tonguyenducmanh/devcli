package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Giá trị màu ANSI được dùng khi bật tô màu.
const (
	colorReset  = "\x1b[0m"
	colorRed    = "\x1b[31m"
	colorGreen  = "\x1b[32m"
	colorYellow = "\x1b[33m"
	colorCyan   = "\x1b[36m"
)

// colorEnabled cho biết output có được tô màu hay không.
// Kết quả được tính một lần trong PersistentPreRun của lệnh gốc.
var colorEnabled = false

// printLine in một dòng ra stdout.
func printLine(format string, args ...any) {
	fmt.Fprintf(os.Stdout, format+"\n", args...)
}

// printErr in ra stderr, dùng cho thông báo lỗi và cảnh báo.
func printErr(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}

// exitError trả về lỗi đã định dạng sẵn để in ra và thoát với mã 1.
func exitError(format string, args ...any) error {
	return fmt.Errorf(format, args...)
}

// isTerminal báo chuẩn ra có hỗ trợ tô màu hay không.
// Gán TERM=dumb hoặc NO_COLOR trong môi trường sẽ tắt tô màu.
func isTerminal() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	switch os.Getenv("TERM") {
	case "", "dumb":
		return false
	}
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

// colorize bọc một chuỗi trong mã màu khi tô màu đang bật.
func colorize(color, s string) string {
	if !colorEnabled || s == "" {
		return s
	}
	return color + s + colorReset
}

// ─── Bộ kiểm tra số lượng đối số, thông báo bằng tiếng Việt ───

// exactArgs yêu cầu đúng n đối số.
func exactArgs(n int, usage string) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) != n {
			return exitError("cần đúng %d đối số (%s), nhận được %d", n, usage, len(args))
		}
		return nil
	}
}

// maximumArgs cho phép tối đa n đối số.
func maximumArgs(n int, usage string) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) > n {
			return exitError("cho phép tối đa %d đối số (%s), nhận được %d", n, usage, len(args))
		}
		return nil
	}
}

// minimumArgs yêu cầu ít nhất n đối số.
func minimumArgs(n int, usage string) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) < n {
			return exitError("cần ít nhất %d đối số (%s), nhận được %d", n, usage, len(args))
		}
		return nil
	}
}

// arbitraryArgs chấp nhận mọi số đối số, kiểm tra thêm sẽ do lệnh tự làm.
func arbitraryArgs(cmd *cobra.Command, args []string) error { return nil }

// noArgsArg từ chối mọi đối số.
func noArgsArg(cmd *cobra.Command, args []string) error {
	if len(args) > 0 {
		return exitError("lệnh này không nhận đối số, nhưng thấy: %v", args)
	}
	return nil
}

// rangeArgs chấp nhận số đối số trong khoảng [min, max].
func rangeArgs(min, max int, usage string) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) < min || len(args) > max {
			return exitError("cần %d đến %d đối số (%s), nhận được %d", min, max, usage, len(args))
		}
		return nil
	}
}
