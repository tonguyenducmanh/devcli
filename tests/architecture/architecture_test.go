package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/tonguyenducmanh/devcli/cmd"
)

// chuViet là tập chữ cái tiếng Việt có dấu, dùng để chắc chắn mô tả của lệnh
// không phải tiếng Anh rút gọn.
const chuViet = "àáảãạăằắẳẵặâầấẩẫậèéẻẽẹêềếểễệìíỉĩịòóỏõọôồốổỗộơờớởỡợùúủũụưừứửữựỳýỷỹỵđ"

// File này kiểm tra các bất biến kiến trúc của dự án. Chúng không kiểm tra
// hành vi nghiệp vụ mà kiểm tra *hình dạng* của mã nguồn, nhằm ngăn một thay
// đổi vô hại trông ra có vẻ đúng lại làm hỏng nguyên tắc của cả dự án.
//
// Mỗi kiểm thử ở đây đều là một lời hứa rõ ràng với người đọc mã và với các
// bản phát hành sau này: nếu vi phạm thì kiểm thử đỏ.

// moduleRoot trả về thư mục gốc của module.
func moduleRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("không tìm thấy go.mod")
		}
		dir = parent
	}
}

// goFiles liệt kê mọi tệp Go trong module, theo đường dẫn tương đối.
func goFiles(t *testing.T) []string {
	t.Helper()
	root := moduleRoot(t)
	var out []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			// Bỏ qua thư mục sinh tự động và thư mục quản lý phiên bản.
			switch d.Name() {
			case ".git", ".tdx", "vendor":
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(path, ".go") {
			rel, relErr := filepath.Rel(root, path)
			if relErr != nil {
				return relErr
			}
			out = append(out, filepath.ToSlash(rel))
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// TestKhongGoiLenhHeThong bảo đảm mã nguồn không chạy tiến trình ngoài.
//
// td tự chứa toàn bộ chức năng của mình. Nếu một tương lai nào đó thêm
// exec.Command vào đây thì kho mã sẽ phụ thuộc vào công cụ được cài trên máy,
// việc đóng gói và kiểm thử cũng trở nên bấp bênh. Vì vậy tận gốc cấm hẳn.
func TestKhongGoiLenhHeThong(t *testing.T) {
	for _, file := range goFiles(t) {
		if strings.HasSuffix(file, "_test.go") {
			continue // tệp kiểm thử được phép gọi chương trình khác
		}
		data, err := os.ReadFile(filepath.Join(moduleRoot(t), file))
		if err != nil {
			t.Fatal(err)
		}
		for i, line := range strings.Split(string(data), "\n") {
			if strings.Contains(line, `"os/exec"`) || strings.Contains(line, "exec.Command") {
				t.Errorf("%s:%d: mã nguồn không được gọi chương trình ngoài: %s",
					file, i+1, strings.TrimSpace(line))
			}
		}
	}
}

// allowedImports mô tả quan hệ phụ thuộc được phép giữa các tầng.
// Mỗi khoá là một tầng, mỗi giá trị là tập tầng mà tầng đó được phép biết tới.
var allowedImports = map[string][]string{
	"internal/vcs/object":   {},
	"internal/vcs/storage":  {"internal/vcs/object"},
	"internal/vcs/index":    {"internal/vcs/object"},
	"internal/vcs/config":   {},
	"internal/vcs/diff":     {},
	"internal/vcs/merge":    {"internal/vcs/diff"},
	"internal/vcs/worktree": {"internal/vcs/object"},
	"internal/vcs/repo": {
		"internal/vcs/config",
		"internal/vcs/index",
		"internal/vcs/object",
		"internal/vcs/storage",
		"internal/vcs/worktree",
	},
	"internal/vcs/ops": {
		"internal/vcs/config",
		"internal/vcs/diff",
		"internal/vcs/index",
		"internal/vcs/merge",
		"internal/vcs/object",
		"internal/vcs/repo",
		"internal/vcs/storage",
		"internal/vcs/worktree",
	},
	"cmd":                   {"internal/vcs/config", "internal/vcs/object", "internal/vcs/ops", "internal/vcs/repo"},
	"internal/tools/docgen": {"cmd"},
}

// internalPackages trả về tập đường dẫn của các gói thuộc module.
func internalPackages() map[string]bool {
	out := make(map[string]bool, len(allowedImports))
	for p := range allowedImports {
		out[p] = true
	}
	return out
}

// TestKhongPhuThuocCheo giữa các tầng bảo đảm phụ thuộc đi đúng một chiều.
//
// Khi một tầng biết tới tầng thấp hơn quá mức cần thiết, thứ tự phụ thuộc
// sẽ hỗn loạn và việc thay thế một tầng trở nên rất tốn công.
func TestKhongPhuThuocCheo(t *testing.T) {
	known := internalPackages()
	root := moduleRoot(t)

	for _, file := range goFiles(t) {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		dir := pathDir(file)
		allowed, ok := allowedImports[dir]
		if !ok {
			continue // không thuộc tầng có quy tắc, chẳng hạn main
		}
		data, err := os.ReadFile(filepath.Join(root, file))
		if err != nil {
			t.Fatal(err)
		}
		for _, imp := range importLines(string(data)) {
			if !known[imp] {
				continue // gói ngoài module, không kiểm soát
			}
			if !contains(allowed, imp) {
				t.Errorf("%s: tầng %s không được phụ thuộc vào %s", file, dir, imp)
			}
		}
	}
}

// TestBinaryKhongPhuThuocLenhNgoai biên dịch rồi chạy td với PATH rỗng.
//
// Nếu bản nhị phân vẫn chạy được khi không có bất kỳ chương trình nào khác
// trên máy thì nó thật sự tự trị. Đây là kiểm chứng thực thi cho lời hứa của
// TestKhongGoiLenhHeThong.
func TestBinaryKhongPhuThuocLenhNgoai(t *testing.T) {
	root := moduleRoot(t)
	bin := filepath.Join(t.TempDir(), "td")

	// Dùng chính go của phiên chạy kiểm thử để biên dịch.
	goBin, err := exec.LookPath("go")
	if err != nil {
		t.Skipf("không tìm thấy go: %v", err)
	}
	build := exec.Command(goBin, "build", "-o", bin, ".")
	build.Dir = root
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("biên dịch thất bại: %v\n%s", err, out)
	}

	// Chạy với PATH trỏ vào một thư mục rỗng để không có lệnh nào tìm được.
	emptyDir := t.TempDir()
	cmd := exec.Command(bin, "version")
	cmd.Dir = emptyDir
	cmd.Env = append(os.Environ(), "PATH="+emptyDir)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("chạy với PATH rỗng thất bại: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "td phiên bản") {
		t.Fatalf("kết quả không mong đợi: %s", out)
	}
}

// pathDir trả về thư mục chứa một đường dẫn tương đối, luôn có dấu / cuối.
func pathDir(p string) string {
	i := strings.LastIndex(p, "/")
	if i < 0 {
		return ""
	}
	return p[:i]
}

// importLines lấy các đường dẫn gói trong khối import của một tệp.
func importLines(src string) []string {
	var out []string
	inBlock := false
	for _, line := range strings.Split(src, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "import (") {
			inBlock = true
			continue
		}
		if inBlock && line == ")" {
			break
		}
		if inBlock {
			if p, ok := unquote(line); ok {
				out = append(out, p)
			}
			continue
		}
		// Dạng một dòng: import "duong/dan".
		rest, ok := strings.CutPrefix(line, "import ")
		if !ok {
			continue
		}
		if p, ok := unquote(strings.TrimSpace(rest)); ok {
			out = append(out, p)
		}
	}
	return out
}

// unquote bóc dấu nháy quanh một chuỗi nguyên.
func unquote(s string) (string, bool) {
	s = strings.TrimSpace(s)
	if len(s) < 2 || s[0] != '"' || s[len(s)-1] != '"' {
		return "", false
	}
	// Bỏ phần bí danh sau dấu nháy đóng nếu có.
	path := s[1 : len(s)-1]
	if i := strings.Index(path, " "); i >= 0 {
		path = path[:i]
	}
	return path, true
}

// contains báo xem một danh sách chuỗi có chứa giá trị cần tìm hay không.
func contains(list []string, want string) bool {
	for _, v := range list {
		if v == want {
			return true
		}
	}
	return false
}

// TestMoiLenhDeuCoMoTaDayDu bảo đảm không lệnh nào bị bỏ trống phần trợ giúp.
//
// Tài liệu sinh tự động trong docs/cli và phần trợ giúp trên dòng lệnh đều lấy
// từ Short, Long và Example. Thiếu một trong ba thì tài liệu của dự án nghèo
// nàn đi ngay lập tức, nên kiểm tra này giữ chất lượng tài liệu.
func TestMoiLenhDeuCoMoTaDayDu(t *testing.T) {
	var check func(cmd *cobra.Command)
	seen := 0
	check = func(cmd *cobra.Command) {
		if cmd.Name() == "help" {
			return // lệnh dựng sẵn của cobra
		}
		seen++
		if strings.TrimSpace(cmd.Short) == "" {
			t.Errorf("lệnh %q thiếu Short", cmd.CommandPath())
		}
		if strings.TrimSpace(cmd.Long) == "" {
			t.Errorf("lệnh %q thiếu Long, phần trợ giúp sẽ rất sơ sài", cmd.CommandPath())
		}
		if strings.TrimSpace(cmd.Example) == "" {
			t.Errorf("lệnh %q thiếu Example, người dùng không có mẫu để làm theo", cmd.CommandPath())
		}
		if !strings.ContainsAny(cmd.Long, chuViet) {
			t.Errorf("lệnh %q: Long nên viết bằng tiếng Việt", cmd.CommandPath())
		}
		for _, sub := range cmd.Commands() {
			check(sub)
		}
	}
	check(cmd.Root())

	if seen < 20 {
		t.Errorf("chỉ duyệt %d lệnh, số lượng có vẻ bất thường", seen)
	}
}

// TestMoiLenhDeuDatTenDichDung bảo đảm tên tham số ghi rõ ý nghĩa bằng tiếng Việt.
func TestMoiLenhDeuDatTenDichDung(t *testing.T) {
	var check func(cmd *cobra.Command)
	check = func(cmd *cobra.Command) {
		cmd.Flags().VisitAll(func(f *pflag.Flag) {
			if f.Name == "help" {
				return // cờ dựng sẵn của cobra
			}
			if strings.TrimSpace(f.Usage) == "" {
				t.Errorf("cờ --%s của lệnh %q thiếu mô tả", f.Name, cmd.CommandPath())
			}
			if !strings.ContainsAny(f.Usage, chuViet) {
				t.Errorf("cờ --%s của lệnh %q: mô tả nên viết bằng tiếng Việt", f.Name, cmd.CommandPath())
			}
		})
		for _, sub := range cmd.Commands() {
			check(sub)
		}
	}
	check(cmd.Root())
}
