// Package env cung cấp các hàm quản lý môi trường và đường dẫn toàn cục của ứng dụng.
// Các thành phần khác trong mã nguồn (CLI, VCS, Log...) có thể dùng gói này để
// nhất quán về vị trí lưu trữ dữ liệu.
package env

import (
	"os"
	"path/filepath"
)

// GlobalDir trả về đường dẫn tới thư mục cấu hình toàn cục của người dùng (thường là ~/.tm).
// Hàm này được dùng chung cho toàn bộ chương trình để lưu trữ cấu hình, lịch sử và cache.
func GlobalDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".tm"), nil
}
