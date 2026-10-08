package main

// Kiểm thử cho phần dòng lệnh: ở đây thứ quan trọng là cách các lệnh đọc
// đối số, vì người dùng chạm vào đó qua đường dẫn và dấu hai gạch ngang.

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/tonguyenducmanh/devcli/cmd"
	"github.com/tonguyenducmanh/devcli/internal/vcs/worktree"
)

// runChạy lệnh trong bộ nhớ rồi trả về toàn bộ output và lỗi phát sinh.
//
// Cùng cách làm với tests/architecture/help_test.go: đổi os.Stdout sang một
// ống thật để bắt được cả phần trợ giúp của cobra lẫn thông báo do printLine
// ghi thẳng ra.
func run(t *testing.T, args ...string) (string, error) {
	t.Helper()

	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	savedOut := os.Stdout
	os.Stdout = writer

	defer func() {
		os.Stdout = savedOut
	}()

	root := cmd.Root()
	resetFlags(root)
	root.SetOut(writer)
	root.SetErr(writer)
	root.SetArgs(args)

	runErr := root.Execute()

	writer.Close()
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(reader); err != nil {
		t.Fatal(err)
	}
	reader.Close()

	return buf.String(), runErr
}

// mustRun chạy lệnh rồi đòi lệnh phải thành công.
func mustRun(t *testing.T, args ...string) string {
	t.Helper()
	out, err := run(t, args...)
	if err != nil {
		t.Fatalf("lệnh %v thất bại: %v\n%s", args, err, out)
	}
	return out
}

// resetFlags đưa mọi cờ về giá trị mặc định, vì cây lệnh dùng chung cho mọi
// lần chạy trong cùng một tiến trình.
//
// Cờ dạng danh sách như -m của commit không dùng Set được: hàm Set của nó
// nối thêm vào danh sách đang có, nên phải xoá cả danh sách rồi đặt lại.
func resetFlags(root *cobra.Command) {
	var walkAll func(c *cobra.Command)
	walkAll = func(c *cobra.Command) {
		apply := func(f *pflag.Flag) {
			if slice, ok := f.Value.(pflag.SliceValue); ok {
				_ = slice.Replace(nil)
			} else {
				_ = f.Value.Set(f.DefValue)
			}
			f.Changed = false
		}
		c.Flags().VisitAll(apply)
		c.PersistentFlags().VisitAll(apply)
		for _, sub := range c.Commands() {
			walkAll(sub)
		}
	}
	walkAll(root)
}

// newRepo tạo kho tm tạm rồi trả về thư mục gốc của nó.
func newRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	mustRun(t, "vcs", "init", "-C", dir)
	return dir
}

// writeFile ghi một tệp trong kho.
func writeFile(t *testing.T, root, rel, content string) {
	t.Helper()
	full := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// commitAll stage rồi commit toàn bộ với nội dung cho trước.
func commitAll(t *testing.T, root, message string) {
	t.Helper()
	mustRun(t, "vcs", "add", "-C", root, ".")
	mustRun(t, "vcs", "commit", "-C", root, "-m", message)
}

// TestDiffPathsAfterDash bảo đảm danh sách tệp sau dấu hai gạch ngang được
// dùng để lọc kết quả diff.
//
// pflag nuốt mất dấu hai gạch ngang nên nếu lệnh tự dò trong args thì tệp bị
// hiểu nhầm là tên phạm vi commit và diff hỏng.
func TestDiffPathsAfterDash(t *testing.T) {
	root := newRepo(t)
	writeFile(t, root, "muon.txt", "ban dau\n")
	writeFile(t, root, "khong.txt", "ban dau\n")
	commitAll(t, root, "c1")

	writeFile(t, root, "muon.txt", "sua\n")
	writeFile(t, root, "khong.txt", "sua\n")
	mustRun(t, "vcs", "add", "-C", root, ".")

	out := mustRun(t, "vcs", "diff", "-C", root, "--staged", "--", "muon.txt")
	if !strings.Contains(out, "muon.txt") {
		t.Fatalf("diff phải chứa tệp được lọc: %s", out)
	}
	if strings.Contains(out, "khong.txt") {
		t.Fatalf("diff không được chứa tệp ngoài danh sách: %s", out)
	}
}

// TestDiffUnstagedPathsAfterDash làm cùng việc cho phần chưa stage.
func TestDiffUnstagedPathsAfterDash(t *testing.T) {
	root := newRepo(t)
	writeFile(t, root, "muon.txt", "ban dau\n")
	writeFile(t, root, "khong.txt", "ban dau\n")
	commitAll(t, root, "c1")

	writeFile(t, root, "muon.txt", "sua\n")
	writeFile(t, root, "khong.txt", "sua\n")

	out := mustRun(t, "vcs", "diff", "-C", root, "--", "khong.txt")
	if !strings.Contains(out, "khong.txt") || strings.Contains(out, "muon.txt") {
		t.Fatalf("diff chưa stage phải chỉ lấy tệp được lọc: %s", out)
	}
}

// TestLogPathsAfterDash bảo đảm log lọc được theo tệp.
func TestLogPathsAfterDash(t *testing.T) {
	root := newRepo(t)
	// c1 chỉ đụng khac.txt, nên nó không được lọt vào log của muon.txt.
	writeFile(t, root, "khac.txt", "ban dau\n")
	commitAll(t, root, "c1")

	writeFile(t, root, "muon.txt", "sua\n")
	commitAll(t, root, "c2")

	out := mustRun(t, "vcs", "log", "-C", root, "--oneline", "--", "muon.txt")
	if !strings.Contains(out, "c2") {
		t.Fatalf("log phải chứa commit sửa tệp được lọc: %s", out)
	}
	if strings.Contains(out, "c1") {
		t.Fatalf("log không được chứa commit không đụng tệp lọc: %s", out)
	}
}

// TestCleanListsBeforeDeleting bảo đảm lệch chỉ mặc định không xoá gì.
//
// Lệnh xoá tệp không thể hỏi lại trong kịch bản, nên phải là hai lần chạy: một
// lần để xem, một lần có -f mới xoá thật.
func TestCleanListsBeforeDeleting(t *testing.T) {
	root := newRepo(t)
	writeFile(t, root, "theo-doi.txt", "còn trong kho\n")
	commitAll(t, root, "c1")
	writeFile(t, root, "rac.txt", "tạm\n")

	dry := mustRun(t, "vcs", "clean", "-C", root)
	if !strings.Contains(dry, "rac.txt") {
		t.Fatalf("lệch chỉ phải liệt kê tệp sẽ bị xoá: %s", dry)
	}
	if _, err := os.Stat(filepath.Join(root, "rac.txt")); err != nil {
		t.Fatalf("chưa có -f thì không được xoá: %v", err)
	}

	mustRun(t, "vcs", "clean", "-f", "-C", root)
	if _, err := os.Stat(filepath.Join(root, "rac.txt")); !os.IsNotExist(err) {
		t.Fatalf("có -f thì phải xoá: %v", err)
	}
	// Tệp đã được theo dõi không được đụng tới.
	if _, err := os.Stat(filepath.Join(root, "theo-doi.txt")); err != nil {
		t.Fatalf("tệp đã theo dõi không được xoá: %v", err)
	}
}

// TestShowFileIndexReadsStagedContent bảo đảm đọc được nội dung trong vùng chuẩn
// bị, kể cả lúc tệp đã khớp với vùng chuẩn bị và không còn khác biệt nào.
//
// Đây là nguồn của phía gốc trong khung so sánh thay đổi trên đĩa của tiện ích
// VS Code. Đọc bằng `show-file` không có --index thì tệp đã khớp sẽ không còn
// khác biệt để dựng lại, và phía gốc rơi về chuỗi rỗng.
func TestShowFileIndexReadsStagedContent(t *testing.T) {
	root := newRepo(t)
	writeFile(t, root, "a.txt", "dòng 1\ndòng 2\n")
	commitAll(t, root, "c1")

	// Sửa trên đĩa rồi khôi phục: tệp khớp lại vùng chuẩn bị.
	writeFile(t, root, "a.txt", "dòng 1\ndòng 3\n")
	mustRun(t, "vcs", "restore", "-C", root, "a.txt")

	if got := mustRun(t, "vcs", "show-file", "--index", "-C", root, "--", "a.txt"); got != "dòng 1\ndòng 2\n" {
		t.Fatalf("vùng chuẩn bị phải trả lại nội dung đã stage, nhận %q", got)
	}

	// Stage nội dung mới thì phải trả nội dung mới.
	writeFile(t, root, "a.txt", "dòng 1\ndòng 4\n")
	mustRun(t, "vcs", "add", "-C", root, "a.txt")
	if got := mustRun(t, "vcs", "show-file", "--index", "-C", root, "--", "a.txt"); got != "dòng 1\ndòng 4\n" {
		t.Fatalf("sau khi stage thì phải trả nội dung mới, nhận %q", got)
	}

	// Vùng chuẩn bị và lịch sử là hai nơi khác nhau, phải cho ra hai kết quả
	// khác nhau trên cùng một tệp.
	head := mustRun(t, "vcs", "show-file", "-C", root, "--", "a.txt")
	if head != "dòng 1\ndòng 2\n" {
		t.Fatalf("HEAD phải là nội dung đã commit, nhận %q", head)
	}

	// Tệp chưa thêm vào thì không có trong vùng chuẩn bị, phải báo lỗi chứ
	// không in ra nội dung rỗng rồi coi như thành công.
	writeFile(t, root, "moi.txt", "tệp mới\n")
	out, err := run(t, "vcs", "show-file", "--index", "-C", root, "--", "moi.txt")
	if err == nil {
		t.Fatalf("tệp chưa thêm vào phải báo lỗi, nhận %q", out)
	}
	if !strings.Contains(err.Error(), "vùng chuẩn bị") {
		t.Errorf("thông báo lỗi phải nói rõ là vùng chuẩn bị, nhận %q", err.Error())
	}

	// Đọc vùng chuẩn bị thì không cần điểm xuất phát, có thừa phải báo lỗi.
	if _, err := run(t, "vcs", "show-file", "--index", "HEAD", "-C", root, "--", "a.txt"); err == nil {
		t.Error("đọc vùng chuẩn bị mà lại đưa thêm điểm xuất phát thì phải báo lỗi")
	}
}

// TestInitCreatesDefaultIgnore bảo đảm init tạo tệp ignore mẫu và không đè lên
// tệp người dùng đã có.
func TestInitCreatesDefaultIgnore(t *testing.T) {
	dir := t.TempDir()
	out := mustRun(t, "vcs", "init", "-C", dir)

	ignore := filepath.Join(dir, ".tmxignore")
	body, err := os.ReadFile(ignore)
	if err != nil {
		t.Fatalf("init phải tạo tệp ignore ở gốc kho: %v", err)
	}
	if !strings.Contains(out, ".tmxignore") {
		t.Errorf("lệnh phải báo là đã tạo tệp ignore: %s", out)
	}

	// Kho đã có sẵn thì init báo lỗi, nên kiểm tra việc không đè bằng cách gọi
	// hàm của tầng nghiệp vụ trực tiếp.
	if err := os.WriteFile(ignore, []byte("# tùy chọn\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	created, err := worktree.WriteDefaultIgnore(dir)
	if err != nil {
		t.Fatal(err)
	}
	if created {
		t.Error("tệp ignore đã tồn tại thì không được ghi đè")
	}
	body, err = os.ReadFile(ignore)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "# tùy chọn\n" {
		t.Errorf("nội dung tệp ignore bị ghi đè: %q", body)
	}
}

// TestInitIgnoreSkipsBuildArtifacts bảo đảm tệp mẫu loại trừ đúng mấy thứ mà
// người dùng không muốn theo dõi.
func TestInitIgnoreSkipsBuildArtifacts(t *testing.T) {
	dir := t.TempDir()
	mustRun(t, "vcs", "init", "-C", dir)
	writeFile(t, dir, "node_modules/dep/index.js", "x\n")
	writeFile(t, dir, "out/app", "x\n")
	writeFile(t, dir, "debug.log", "x\n")
	writeFile(t, dir, "main.go", "package main\n")

	out := mustRun(t, "vcs", "status", "-C", dir)
	for _, gone := range []string{"node_modules", "out/", "debug.log"} {
		if strings.Contains(out, gone) {
			t.Errorf("tệp ignore mẫu phải bỏ qua %s: %s", gone, out)
		}
	}
	if !strings.Contains(out, "main.go") {
		t.Errorf("mã nguồn thật phải còn trong status: %s", out)
	}
}

// TestCleanOnlyNamedPaths bảo đảm lệch chỉ xoá đúng các tệp được nêu tên.
func TestCleanOnlyNamedPaths(t *testing.T) {
	root := newRepo(t)
	writeFile(t, root, "a.txt", "1\n")
	commitAll(t, root, "c1")
	writeFile(t, root, "rac.txt", "tạm\n")
	writeFile(t, root, "rac.log", "tạm\n")

	mustRun(t, "vcs", "clean", "-f", "-C", root, "rac.txt")
	if _, err := os.Stat(filepath.Join(root, "rac.txt")); !os.IsNotExist(err) {
		t.Fatalf("rac.txt phải bị xoá: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "rac.log")); err != nil {
		t.Fatalf("rac.log không nằm trong danh sách nên phải còn lại: %v", err)
	}
}

// TestResetPathsAfterDash bảo đảm reset chỉ gỡ đúng các tệp được nêu tên.
//
// Tệp bị gỡ khỏi vùng chuẩn bị vẫn khác HEAD nên xuất hiện ở nhóm chưa stage,
// đúng như git.
func TestResetPathsAfterDash(t *testing.T) {
	root := newRepo(t)
	writeFile(t, root, "giai.txt", "ban dau\n")
	writeFile(t, root, "khac.txt", "ban dau\n")
	commitAll(t, root, "c1")

	writeFile(t, root, "giai.txt", "sua\n")
	writeFile(t, root, "khac.txt", "sua\n")
	mustRun(t, "vcs", "add", "-C", root, ".")

	mustRun(t, "vcs", "reset", "-C", root, "HEAD", "--", "giai.txt")

	out := mustRun(t, "vcs", "status", "-C", root)
	staged, unstaged, _ := splitSections(out)
	if _, ok := staged["giai.txt"]; ok {
		t.Fatalf("giai.txt phải được gỡ khỏi vùng chuẩn bị: %s", out)
	}
	if _, ok := unstaged["giai.txt"]; !ok {
		t.Fatalf("giai.txt phải còn ở nhóm chưa stage: %s", out)
	}
	if _, ok := staged["khac.txt"]; !ok {
		t.Fatalf("khac.txt không nằm trong danh sách nên phải giữ nguyên: %s", out)
	}
}

// splitSections cắt output của status thành ba nhóm tệp theo tên nhóm.
func splitSections(out string) (staged, unstaged, untracked map[string]bool) {
	groups := map[string]map[string]bool{
		"staged":    {},
		"unstaged":  {},
		"untracked": {},
	}
	group := ""
	for _, line := range strings.Split(out, "\n") {
		switch {
		case strings.HasPrefix(line, "Thay đổi đã stage:"):
			group = "staged"
		case strings.HasPrefix(line, "Thay đổi chưa stage:"):
			group = "unstaged"
		case strings.HasPrefix(line, "File chưa được theo dõi:"):
			group = "untracked"
		case strings.TrimSpace(line) == "":
			group = ""
		case group != "":
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				// Dòng gồm ký hiệu trạng thái rồi đến đường dẫn có thể
				// chứa khoảng trắng, nên ghép phần còn lại lại.
				groups[group][strings.Join(fields[1:], " ")] = true
			}
		}
	}
	return groups["staged"], groups["unstaged"], groups["untracked"]
}

// TestShowFileReadsFromRevision bảo đảm đọc được nội dung tệp ở một điểm lịch sử.
//
// Khung so sánh trong trình soạn thảo gọi lệnh này cho từng phía của khung, nên
// lệnh phải in ra nội dung nguyên văn chứ không phải khác biệt.
func TestShowFileReadsFromRevision(t *testing.T) {
	root := newRepo(t)
	writeFile(t, root, "a.txt", "một\nhai\n")
	writeFile(t, root, "khac.txt", "giữ nguyên\n")
	commitAll(t, root, "c1")
	writeFile(t, root, "a.txt", "một\nhai sửa\n")
	writeFile(t, root, "moi.txt", "tệp mới\n")
	commitAll(t, root, "c2")

	if out := mustRun(t, "vcs", "show-file", "-C", root, "HEAD", "--", "a.txt"); out != "một\nhai sửa\n" {
		t.Fatalf("nội dung ở HEAD phải là bản mới nhất, nhận %q", out)
	}
	if out := mustRun(t, "vcs", "show-file", "-C", root, "HEAD~1", "--", "a.txt"); out != "một\nhai\n" {
		t.Fatalf("nội dung ở commit cũ phải là bản cũ, nhận %q", out)
	}
	// Tệp mà commit cuối không đụng tới vẫn phải đọc được.
	if out := mustRun(t, "vcs", "show-file", "-C", root, "HEAD", "--", "khac.txt"); out != "giữ nguyên\n" {
		t.Fatalf("tệp không đổi vẫn phải đọc được, nhận %q", out)
	}

	// Tệp không có ở điểm đó phải báo lỗi chứ không in ra rỗng.
	if _, err := run(t, "vcs", "show-file", "-C", root, "HEAD~1", "--", "moi.txt"); err == nil {
		t.Fatal("đọc tệp chưa tồn tại ở điểm đó phải báo lỗi")
	}
	// Không nêu tệp nào thì báo lỗi chứ không in ra nội dung rỗng.
	if out, err := run(t, "vcs", "show-file", "-C", root, "HEAD"); err == nil || out != "" {
		t.Fatalf("thiếu danh sách tệp phải báo lỗi, nhận %q", out)
	}
}
