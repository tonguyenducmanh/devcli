package config_test

import (
	"github.com/tonguyenducmanh/devcli/internal/vcs/config"
	"os"
	"path/filepath"
	"testing"
)

func newConfig(t *testing.T) (*config.Config, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config")
	return config.New(path), path
}

func TestSaveAndLoadRoundTrip(t *testing.T) {
	c, path := newConfig(t)
	c.Set("user.name", "Tên Có Dấu Cách")
	c.Set("user.email", "ten@example.com")
	c.Set("core.editor", "vim")
	if err := c.Save(); err != nil {
		t.Fatal(err)
	}

	// File phải tồn tại sau khi lưu.
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}

	loaded, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if v, _ := loaded.Get("user.name"); v != "Tên Có Dấu Cách" {
		t.Fatalf("giá trị có dấu cách phải được giữ nguyên: %q", v)
	}
	if v, _ := loaded.Get("user.email"); v != "ten@example.com" {
		t.Fatalf("sai giá trị: %q", v)
	}
	if v, _ := loaded.Get("core.editor"); v != "vim" {
		t.Fatalf("sai giá trị: %q", v)
	}
}

func TestCaseInsensitiveKeys(t *testing.T) {
	c, path := newConfig(t)
	c.Set("User.Name", "Tên")
	if err := c.Save(); err != nil {
		t.Fatal(err)
	}
	loaded, _ := config.Load(path)
	// Tên section và khoá không phân biệt hoa thường.
	if v, ok := loaded.Get("user.name"); !ok || v != "Tên" {
		t.Fatalf("phải tìm thấy khoá bất kể hoa thường: %q %v", v, ok)
	}
}

func TestSetResetAndUnset(t *testing.T) {
	c, _ := newConfig(t)
	c.Set("user.name", "Một")
	c.Set("user.name", "Hai")
	// Đặt lại phải ghi đè chứ không tạo bản trùng.
	if v, _ := c.Get("user.name"); v != "Hai" {
		t.Fatalf("đặt lại phải ghi đè: %q", v)
	}
	keys := c.Keys()
	if len(keys) != 1 {
		t.Fatalf("phải chỉ có 1 khoá, nhận %v", keys)
	}

	c.Unset("user.name")
	if _, ok := c.Get("user.name"); ok {
		t.Fatalf("khoá đã xoá không được tìm thấy")
	}
}

func TestDefaultValues(t *testing.T) {
	c, _ := newConfig(t)
	if got := c.GetString("khong.co", "mặc định"); got != "mặc định" {
		t.Fatalf("phải trả về giá trị mặc định: %q", got)
	}
	if !c.GetBool("khong.co", true) {
		t.Fatalf("bool mặc định phải được giữ")
	}
	if c.GetInt("khong.co", 7) != 7 {
		t.Fatalf("số mặc định phải được giữ")
	}

	c.Set("co.dung", "true")
	c.Set("co.sai", "false")
	c.Set("co.khongDoc", "không-phải-boolean")
	c.Set("co.so", "42")
	if !c.GetBool("co.dung", false) {
		t.Fatalf("giá trị true phải được hiểu đúng")
	}
	if c.GetBool("co.sai", true) {
		t.Fatalf("giá trị false phải được hiểu đúng")
	}
	// Giá trị không phải boolean thì trả về mặc định.
	if !c.GetBool("co.khongDoc", true) {
		t.Fatalf("giá trị không hợp lệ phải trả về mặc định")
	}
	if c.GetInt("co.so", 0) != 42 {
		t.Fatalf("phải đọc được số: %d", c.GetInt("co.so", 0))
	}
}

func TestReadMissingFileErrors(t *testing.T) {
	// Đọc file không tồn tại phải cho cấu hình rỗng chứ không phải lỗi.
	c, err := config.Load(filepath.Join(t.TempDir(), "khong-co.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if len(c.Keys()) != 0 {
		t.Fatalf("cấu hình phải rỗng")
	}
}

func TestSkipCommentsAndBlankLines(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config")
	content := `# dòng chú thích
; kiểu chú thích khác

[core]
	repositoryformatversion = 0
	# chú thích trong section
	filemode = true
[user]
	name = "Tên"
	email = ten@example.com
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	c, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if v, _ := c.Get("core.filemode"); v != "true" {
		t.Fatalf("sai giá trị: %q", v)
	}
	if v, _ := c.Get("user.name"); v != "Tên" {
		t.Fatalf("phải bỏ dấu nháy kép: %q", v)
	}
	if v, _ := c.Get("user.email"); v != "ten@example.com" {
		t.Fatalf("sai giá trị: %q", v)
	}
	// repositoryformatversion chỉ có khoá, giá trị rỗng.
	if v, _ := c.Get("core.repositoryformatversion"); v != "0" {
		t.Fatalf("sai giá trị: %q", v)
	}
}

func TestGlobalPathFollowsEnv(t *testing.T) {
	// Đường dẫn cấu hình toàn cục phải tôn trọng biến môi trường TM_CONFIG.
	t.Setenv("TM_CONFIG", "/tmp/khung-kiem-thu/config")
	got, err := config.GlobalPath()
	if err != nil {
		t.Fatal(err)
	}
	if got != "/tmp/khung-kiem-thu/config" {
		t.Fatalf("sai đường dẫn: %q", got)
	}
}

func TestResolveIdentity(t *testing.T) {
	c, _ := newConfig(t)
	c.Set("user.name", "Tên Trong Cấu Hình")
	c.Set("user.email", "hocdanh@example.com")

	name, email := c.ResolveIdentity()
	if name != "Tên Trong Cấu Hình" || email != "hocdanh@example.com" {
		t.Fatalf("phải ưu tiên cấu hình: %q / %q", name, email)
	}
}
