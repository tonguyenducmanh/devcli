// Package config đọc/ghi file cấu hình của td theo định dạng
// các mục [tên-mục] và các cặp khoá = giá trị.
package config

import (
	"bufio"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

// Config là cấu hình dạng khóa phẳng, ví dụ "user.name".
type Config struct {
	path    string
	entries []entry // giữ thứ tự để ghi lại file mà không làm mờ file cấu hình
}

type entry struct {
	section string
	key     string
	value   string
}

// New tạo config rỗng gắn với một đường dẫn file.
func New(path string) *Config { return &Config{path: path} }

// Load đọc file cấu hình từ đường dẫn cho trước.
// Nếu file không tồn tại thì trả về config rỗng, không phải lỗi.
func Load(path string) (*Config, error) {
	c := &Config{path: path}
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return c, nil
		}
		return nil, err
	}
	defer f.Close()

	section := ""
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = strings.TrimSpace(line[1 : len(line)-1])
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		c.entries = append(c.entries, entry{
			section: section,
			key:     strings.TrimSpace(key),
			value:   unquote(strings.TrimSpace(value)),
		})
	}
	return c, sc.Err()
}

// unquote bỏ dấu nháy bao ngoài nếu có.
func unquote(s string) string {
	if len(s) >= 2 && (s[0] == '"' && s[len(s)-1] == '"') {
		return s[1 : len(s)-1]
	}
	return s
}

func quote(s string) string {
	if s == "" || strings.ContainsAny(s, " \t#;") {
		return `"` + s + `"`
	}
	return s
}

// normalize chuyển "Section.Key" thành khoá phẳng "section.key" không phân biệt hoa thường.
func normalize(key string) (section, name string) {
	s, n, ok := strings.Cut(key, ".")
	if !ok {
		return "", strings.ToLower(key)
	}
	return strings.ToLower(s), strings.ToLower(n)
}

// Get trả về giá trị của khóa, ok=false nếu không có.
func (c *Config) Get(key string) (string, bool) {
	section, name := normalize(key)
	for i := len(c.entries) - 1; i >= 0; i-- {
		e := c.entries[i]
		if strings.ToLower(e.section) == section && strings.ToLower(e.key) == name {
			return e.value, true
		}
	}
	return "", false
}

// GetString trả về giá trị hoặc giá trị mặc định khi thiếu.
func (c *Config) GetString(key, def string) string {
	if v, ok := c.Get(key); ok {
		return v
	}
	return def
}

// GetBool chuyển giá trị thành bool.
func (c *Config) GetBool(key string, def bool) bool {
	v, ok := c.Get(key)
	if !ok {
		return def
	}
	switch strings.ToLower(v) {
	case "true", "yes", "1", "on":
		return true
	case "false", "no", "0", "off":
		return false
	}
	return def
}

// GetInt chuyển giá trị thành số nguyên.
func (c *Config) GetInt(key string, def int) int {
	v, ok := c.Get(key)
	if !ok {
		return def
	}
	var n int
	if _, err := fmt.Sscanf(v, "%d", &n); err == nil {
		return n
	}
	return def
}

// Set đặt hoặc cập nhật một khóa.
func (c *Config) Set(key, value string) {
	section, name := normalize(key)
	for i := range c.entries {
		e := &c.entries[i]
		if strings.ToLower(e.section) == section && strings.ToLower(e.key) == name {
			e.value = value
			return
		}
	}
	c.entries = append(c.entries, entry{section: section, key: name, value: value})
}

// Unset xóa một khóa nếu có.
func (c *Config) Unset(key string) {
	section, name := normalize(key)
	out := c.entries[:0]
	for _, e := range c.entries {
		if strings.ToLower(e.section) == section && strings.ToLower(e.key) == name {
			continue
		}
		out = append(out, e)
	}
	c.entries = out
}

// Keys liệt kê toàn bộ khóa dạng "section.key".
func (c *Config) Keys() []string {
	out := make([]string, 0, len(c.entries))
	for _, e := range c.entries {
		if e.section == "" {
			out = append(out, e.key)
			continue
		}
		out = append(out, e.section+"."+e.key)
	}
	return out
}

// Path trả về đường dẫn file cấu hình.
func (c *Config) Path() string { return c.path }

// Save ghi lại cấu hình xuống file, chỉ tạo thư mục khi cần.
func (c *Config) Save() error {
	if err := os.MkdirAll(filepath.Dir(c.path), 0o755); err != nil {
		return err
	}
	// Gom theo section để giữ định dạng dễ đọc.
	var b strings.Builder
	curSection := "\x00" // giá trị đặc biệt: chưa xuất hiện section nào
	for _, e := range c.entries {
		if e.section != curSection {
			curSection = e.section
			if curSection != "" {
				b.WriteString("[" + curSection + "]\n")
			}
		}
		b.WriteString("\t" + e.key + " = " + quote(e.value) + "\n")
	}
	return os.WriteFile(c.path, []byte(b.String()), 0o644)
}

// GlobalPath trả về vị trí file cấu hình toàn cục của td.
func GlobalPath() (string, error) {
	if v := os.Getenv("TD_CONFIG"); v != "" {
		return v, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	// Ưu tiên đường dẫn theo XDG nếu có.
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "td", "config"), nil
	}
	return filepath.Join(home, ".config", "td", "config"), nil
}

// Mở cấu hình hệ thống: cấu hình toàn cục đè lên cấu hình trong repo.
func Open(gitDir string) (*Config, error) {
	if p, err := GlobalPath(); err == nil {
		if global, err := Load(p); err == nil && len(global.entries) > 0 {
			return global, nil
		}
	}
	return Load(filepath.Join(gitDir, "config"))
}

// ResolveIdentity trả về tên và email tác giả lấy từ cấu hình.
// Khi thiếu, suy ra từ thông tin tài khoản của hệ điều hành,
// nếu không có nữa thì dùng giá trị mặc định của td.
func (c *Config) ResolveIdentity() (name, email string) {
	name = c.GetString("user.name", "")
	email = c.GetString("user.email", "")

	if name == "" || email == "" {
		accName, accHost := accountInfo()
		if name == "" {
			name = accName
		}
		if email == "" {
			email = accName + "@" + accHost
		}
	}
	if name == "" {
		name = "td"
	}
	if email == "" {
		email = "td@localhost"
	}
	return name, email
}

// accountInfo lấy tên tài khoản và tên máy chủ từ hệ điều hành.
// Trả về chuỗi rỗng nếu không lấy được thông tin.
func accountInfo() (accUser, host string) {
	if u, err := user.Current(); err == nil {
		accUser = strings.TrimSpace(u.Username)
		if accUser == "" {
			accUser = strings.TrimSpace(u.Name)
		}
	}
	if h, err := os.Hostname(); err == nil {
		host = strings.TrimSpace(h)
	}
	return accUser, host
}
