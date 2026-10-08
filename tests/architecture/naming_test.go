package main

// Kiểm thử cho quy tắc đặt tên: tên hàm, biến, kiểu, tệp và thư mục đều bằng
// tiếng Anh, không dấu, không ký tự ngoài ASCII. Chỉ chú thích mới là tiếng Việt.
//
// Lý do cần chặn bằng kiểm thử chứ không chỉ bằng quy ước: tên có dấu vỡ ở nhiều
// nơi. Terminal trên Windows không nhận phần lớn ký tự ngoài ASCII trong đường
// dẫn, script build phải chạy được trên ba nền tảng, và người khác gõ lại đường
// dẫn từ một thông báo lỗi thì phải gõ đúng dấu ở từng ký tự.

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// regexpMustCompile dịch lỗi biên dịch mẫu thành lỗi kiểm thử có thông điệp rõ
// ràng hơn. Mẫu viết trong hàm kiểm thử, lỗi chỉ lộ ra lúc chạy.
func regexpMustCompile(pattern string) *regexp.Regexp {
	return regexp.MustCompile(pattern)
}

// skippedDirs là thư mục không thuộc mã nguồn của kho.
var skippedDirs = map[string]bool{
	".git":         true,
	".tmx":         true,
	"vendor":       true,
	"out":          true,
	"node_modules": true,
}

// junkFiles là tệp rác do hệ điều hành tạo ra. Chúng nằm trong .gitignore và
// không thuộc kho, nhưng vẫn nằm trên đĩa nên phải bỏ qua chứ không đổi tên được.
var junkFiles = map[string]bool{
	".DS_Store": true,
	"Thumbs.db": true,
}

// conventionalRootFiles là tệp ở gốc kho dùng chữ hoa, vì đó là quy ước có sẵn
// của cộng đồng chứ không phải vì phá quy tắc chữ thường.
var conventionalRootFiles = map[string]bool{
	"README.md":       true,
	"AGENTS.md":       true,
	"CHANGELOG.md":    true,
	"CONTRIBUTING.md": true,
	"LICENSE":         true,
}

// vietnameseWords là những từ tiếng Việt hay dùng trong tên tệp.
//
// Đây chỉ là lưới an toàn, không phải cách nhận biết ngôn ngữ: tên viết không dấu
// thì máy không phân biệt được "diff-view" với "khung-so-sanh" nếu không có danh
// sách này. Cố tình chỉ gồm từ dài từ bốn ký tự trở lên và không trùng từ tiếng
// Anh, để không bắt nhầm tên hợp lệ.
var vietnameseWords = map[string]bool{
	"bao": true, "bien": true, "chan": true, "chay": true,
	"che": true, "chuc": true, "chuyen": true, "cung": true,
	"danh": true, "dau": true, "den": true, "dia": true,
	"dinh": true, "doc": true, "dong": true, "duoc": true,
	"ghep": true, "ghim": true, "giua": true, "goi": true,
	"hien": true, "hinh": true, "hoa": true, "huong": true,
	"khoang": true, "khac": true, "kiem": true, "kieu": true,
	"lam": true, "lang": true, "lenh": true, "lich": true,
	"lien": true, "luu": true, "mau": true, "moi": true,
	"ngan": true, "nghiem": true, "nghe": true, "ngoi": true,
	"ngon": true, "nguon": true, "nhanh": true, "nhom": true,
	"nhu": true, "phuc": true, "quyet": true, "ranh": true,
	"rong": true, "sanh": true, "sinh": true, "sua": true,
	"tach": true, "tang": true, "them": true, "thu": true,
	"thuy": true, "tien": true, "toan": true, "trong": true,
	"truoc": true, "tuan": true, "tuong": true, "viet": true,
	"xoa": true,
}

// repoFilePaths liệt kê mọi tệp trong kho, theo đường dẫn tương đối có dấu gạch
// chéo.
func repoFilePaths(t *testing.T) []string {
	t.Helper()
	root := moduleRoot(t)
	var out []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if skippedDirs[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		if junkFiles[d.Name()] {
			return nil
		}
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return relErr
		}
		out = append(out, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// splitName tách một tên tệp thành các từ, theo dấu gạch nối, gạch dưới và dấu
// chấm.
//
// Tách theo dấu chấm vì phần mở rộng không mang nghĩa: `nhanh.test.ts` phải ra
// "nhanh" chứ không phải một từ lạ không ai đoán được.
func splitName(name string) []string {
	return strings.Fields(strings.NewReplacer("-", " ", "_", " ", ".", " ").Replace(name))
}

// isGeneratedPath báo đường dẫn có nằm trong thư mục do chương trình sinh ra.
//
// `docs/agents/cli/` lấy tên tệp từ tên lệnh, nên tên ở đó phản ánh cây lệnh chứ
// không phải lựa chọn của người viết. Kiểm tra tên ở đó sẽ báo nhầm.
func isGeneratedPath(rel string) bool {
	return strings.HasPrefix(rel, "docs/agents/cli/")
}

// unique giữ lại mỗi phần tử đúng một lần, vẫn theo thứ tự, để thông báo lỗi đọc
// được thay vì lặp lại cùng một tệp.
func unique(items []string) []string {
	seen := make(map[string]bool, len(items))
	out := make([]string, 0, len(items))
	for _, item := range items {
		if !seen[item] {
			seen[item] = true
			out = append(out, item)
		}
	}
	sort.Strings(out)
	return out
}

// sortedKeys trả về các khoá của bản đồ theo thứ tự, để in ra thông báo lỗi mà
// không phụ thuộc thứ tự ngẫu nhiên của bảng băm.
func sortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// TestNamesAreAscii chặn ký tự ngoài ASCII trong tên tệp và tên thư mục.
//
// Đây là phần chắc chắn nhất của quy tắc đặt tên: tên có dấu tiếng Việt vỡ trên
// Windows và trong kịch bản shell, nên phải chặn lúc đặt tên chứ không đợi tới
// lúc hỏng.
func TestNamesAreAscii(t *testing.T) {
	var problems []string
	for _, rel := range repoFilePaths(t) {
		if isGeneratedPath(rel) {
			continue
		}
		for _, segment := range strings.Split(rel, "/") {
			for _, r := range segment {
				if r > 127 {
					problems = append(problems, rel)
					break
				}
			}
		}
	}
	if len(problems) > 0 {
		t.Errorf("tên tệp và tên thư mục phải dùng ký tự ASCII, không dấu:\n  %s",
			strings.Join(unique(problems), "\n  "))
	}
}

// TestNamesAreEnglish chặn từ tiếng Việt trong tên tệp và tên thư mục.
//
// Chú thích thì tiếng Việt, còn tên tệp thì tiếng Anh để dùng được ở mọi nơi. Kiểm
// thử này bắt phần tên viết không dấu, và chỉ bắt được từ có trong danh sách nên
// phần còn lại vẫn cần đọc tên khi review.
func TestNamesAreEnglish(t *testing.T) {
	var problems []string
	for _, rel := range repoFilePaths(t) {
		if isGeneratedPath(rel) {
			continue
		}
		for _, segment := range strings.Split(rel, "/") {
			for _, word := range splitName(segment) {
				if vietnameseWords[word] {
					problems = append(problems, rel+" (từ \""+word+"\")")
					break
				}
			}
		}
	}
	if len(problems) > 0 {
		t.Errorf("tên tệp và tên thư mục phải bằng tiếng Anh:\n  %s",
			strings.Join(unique(problems), "\n  "))
	}
}

// TestRootNamesAreLowercase chặn chữ hoa ngoài các tệp quen thuộc ở gốc kho.
//
// Tệp ở gốc kho là nơi mọi người nhìn vào đầu tiên, nên quy ước chữ thường ở đây
// giữ cho danh sách tệp trông như một danh sách lệnh Go.
func TestRootNamesAreLowercase(t *testing.T) {
	var problems []string
	for _, rel := range repoFilePaths(t) {
		if strings.Contains(rel, "/") {
			continue
		}
		if conventionalRootFiles[rel] {
			continue
		}
		if rel != strings.ToLower(rel) {
			problems = append(problems, rel)
		}
	}
	if len(problems) > 0 {
		t.Errorf("tệp ở gốc kho phải viết chữ thường, trừ %s:\n  %s",
			strings.Join(sortedKeys(conventionalRootFiles), ", "),
			strings.Join(unique(problems), "\n  "))
	}
}

// TestIdentifiersAreEnglish chặn chữ có dấu trong tên hàm, biến và kiểu.
//
// Chỉ nhìn phần khai báo chứ không nhìn toàn bộ tệp, vì chuỗi trong mã thì được
// phép chứa tiếng Việt cho tới khi phần di chuyển sang tiếng Anh xong: tên tệp và
// tệp kiểm thử còn chứa nội dung tiếng Việt.
func TestIdentifiersAreEnglish(t *testing.T) {
	// Dòng khai báo trong Go: func, type, var, const.
	goDecl := regexpMustCompile(`^\s*(?:func\s+(?:\([^)]*\)\s*)?|type\s+|var\s+|const\s+)([A-Za-z_][^ \t(=;]*)`)
	// Dòng khai báo trong TypeScript.
	tsDecl := regexpMustCompile(`^\s*(?:export\s+)?(?:const|let|var|function|class|interface|type|enum)\s+([A-Za-z_][^ \t(=;]*)`)

	files := append(goFiles(t), tsFiles(t)...)
	var problems []string
	for _, rel := range files {
		body, err := os.ReadFile(filepath.Join(moduleRoot(t), rel))
		if err != nil {
			t.Fatalf("đọc %s: %v", rel, err)
		}
		pattern := goDecl
		if strings.HasSuffix(rel, ".ts") {
			pattern = tsDecl
		}
		for i, line := range strings.Split(string(body), "\n") {
			match := pattern.FindStringSubmatch(line)
			if match == nil {
				continue
			}
			for _, r := range match[1] {
				if r > 127 {
					problems = append(problems, fmt.Sprintf("%s:%d  %s", rel, i+1, match[1]))
					break
				}
			}
		}
	}
	if len(problems) > 0 {
		t.Errorf("tên hàm, biến và kiểu phải bằng tiếng Anh không dấu:\n  %s",
			strings.Join(unique(problems), "\n  "))
	}
}

// tsFiles liệt kê mọi tệp TypeScript của tiện ích VS Code, theo đường dẫn tương
// đối có dấu gạch chéo.
//
// Tiện ích không thuộc module Go nên `goFiles` không thấy, và nó cũng phải tuân
// theo cùng quy tắc đặt tên.
func tsFiles(t *testing.T) []string {
	t.Helper()
	root := moduleRoot(t)
	var out []string
	err := filepath.WalkDir(filepath.Join(root, "editors", "vscode", "src"), func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if skippedDirs[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".ts") {
			return nil
		}
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return relErr
		}
		out = append(out, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}
