package cmd

import (
	"github.com/spf13/cobra"
)

// Dự án này không dùng cờ toàn cục. Mỗi lệnh tự khai báo cờ của riêng nó, và
// chữ viết tắt -v mang nghĩa khác nhau tuỳ lệnh, đúng như git:
//
//	td -v              thông tin môi trường
//	td vcs -v          tình trạng kho mã nguồn hiện tại
//	td vcs branch -v   mã băm và tiêu đề của từng nhánh
//
// Lý do không dùng cờ toàn cục: nếu lệnh con khai báo cờ trùng tên hoặc trùng
// chữ viết tắt với cờ ở lệnh cha, pflag sẽ âm thầm bỏ qua cờ của lệnh cha.
// Người dùng thấy cờ toàn cục trong phần trợ giúp nhưng nó chết, và cùng một
// chữ viết tắt mang hai nghĩa tuỳ lệnh.

// verboseOn báo cờ -v của lệnh này có đang bật không.
//
// Lệnh không khai báo cờ -v thì hàm trả về false, nên gọi ở đâu cũng được.
func verboseOn(cmd *cobra.Command) bool {
	f := cmd.Flags().Lookup("verbose")
	if f == nil {
		return false
	}
	v, err := cmd.Flags().GetBool("verbose")
	return err == nil && v
}

// addVerboseFlag khai báo cờ -v cho một lệnh.
//
// desc phải nói đúng việc lệnh đó làm thêm khi bật cờ, vì với mỗi lệnh một
// nghĩa. Ví dụ `td vcs branch -v` hiện mã băm, không phải bật ghi log.
func addVerboseFlag(cmd *cobra.Command, desc string) {
	cmd.Flags().BoolP("verbose", "v", false, desc)
}
