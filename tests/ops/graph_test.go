package ops_test

import (
	"strings"
	"testing"

	"github.com/tonguyenducmanh/devcli/internal/vcs/object"
	"github.com/tonguyenducmanh/devcli/internal/vcs/ops"
)

// fakeCommit là một commit giả với tên đã biết, để dựng lịch sử thử mà không cần
// kho thật. Mã băm tính từ chính tên nên hai tên khác nhau không bao giờ trùng.
type fakeCommit struct {
	name    string
	parents []string
}

// buildLog chuyển danh sách commit giả thành LogEntry mà ops dùng được.
func buildLog(commits []fakeCommit) []ops.LogEntry {
	out := make([]ops.LogEntry, 0, len(commits))
	for _, c := range commits {
		entry := ops.LogEntry{
			Hash:    hashOf(c.name),
			Short:   c.name,
			Summary: c.name,
		}
		for _, p := range c.parents {
			entry.Parents = append(entry.Parents, hashOf(p))
		}
		out = append(out, entry)
	}
	return out
}

// hashOf tính mã băm ổn định cho một tên commit giả.
func hashOf(name string) object.Hash {
	return object.ComputeHash(object.TypeCommit, []byte(name))
}

// renderGraph dựng biểu đồ rồi trả về từng dòng để so sánh được bằng mắt.
func renderGraph(commits []fakeCommit) []string {
	entries := buildLog(commits)
	ops.AssignGraphLanes(entries)
	lines := make([]string, 0, len(entries))
	for _, e := range entries {
		lines = append(lines, strings.TrimRight(e.GraphPrefix+e.Summary, " "))
	}
	return lines
}

// TestGraphLinearHistory bảo đảm lịch sử thẳng vẽ thành một làn duy nhất.
func TestGraphLinearHistory(t *testing.T) {
	got := strings.Join(renderGraph([]fakeCommit{
		{"c3", []string{"c2"}},
		{"c2", []string{"c1"}},
		{"c1", nil},
	}), "\n")
	// Mỗi commit đều nằm ở làn của nó nên dấu * xuất hiện ở mọi dòng, đúng
	// như `git log --graph` trên một nhánh không có nhánh phụ.
	want := "* c3\n* c2\n* c1"
	if got != want {
		t.Errorf("lịch sử thẳng phải vẽ một làn:\n--- nhận\n%s\n--- cần\n%s", got, want)
	}
}

// TestGraphBranchAndMerge bảo đảm nhánh tách và hợp nhất hiện ra thành hai làn.
//
// Đây là phần khó nhất của biểu đồ. Nhánh phụ phải nằm ở cột thứ hai suốt từ
// commit hợp nhất tới commit gốc, rồi hai làn phải gặp nhau ở commit gốc.
func TestGraphBranchAndMerge(t *testing.T) {
	lines := renderGraph([]fakeCommit{
		{"merge", []string{"main2", "side"}},
		{"main2", []string{"main1"}},
		{"side", []string{"side1"}},
		{"main1", []string{"root"}},
		{"side1", []string{"root"}},
		{"root", nil},
	})
	for i, l := range lines {
		t.Logf("dòng %d: %q", i, l)
	}

	// Dòng đầu được đệm khoảng trắng cho cùng bề rộng với các dòng dưới.
	want := []string{
		"*   merge",
		"* | main2",
		"| * side",
		"* | main1",
		"| * side1",
		"*   root",
	}
	if strings.Join(lines, "\n") != strings.Join(want, "\n") {
		t.Errorf("biểu đồ hợp nhất sai:\n--- nhận\n%s\n--- cần\n%s",
			strings.Join(lines, "\n"), strings.Join(want, "\n"))
	}
}

// TestGraphManyBranches bảo đảm làn không bị dồn khi hợp nhất nhiều nhánh cùng lúc.
func TestGraphManyBranches(t *testing.T) {
	lines := renderGraph([]fakeCommit{
		{"octopus", []string{"m1", "s1", "t1"}},
		{"m1", []string{"root"}},
		{"s1", []string{"root"}},
		{"t1", []string{"root"}},
		{"root", nil},
	})
	for i, l := range lines {
		t.Logf("dòng %d: %q", i, l)
	}

	want := []string{
		"*     octopus",
		"* | | m1",
		"| * | s1",
		"| | * t1",
		"*     root",
	}
	if strings.Join(lines, "\n") != strings.Join(want, "\n") {
		t.Errorf("biểu đồ hợp nhất ba nhánh sai:\n--- nhận\n%s\n--- cần\n%s",
			strings.Join(lines, "\n"), strings.Join(want, "\n"))
	}
}

// TestGraphSeparateBranches bảo đảm hai nhánh độc lập nằm cạnh nhau.
//
// Đây là trường hợp `git log --all`: HEAD ở một nhánh, nhánh kia chưa hợp nhất vào
// thì hai làn phải chạy song song cho tới khi gặp nhau ở commit chung.
func TestGraphSeparateBranches(t *testing.T) {
	lines := renderGraph([]fakeCommit{
		{"topic", []string{"base"}},
		{"main1", []string{"base"}},
		{"base", nil},
	})
	for i, l := range lines {
		t.Logf("dòng %d: %q", i, l)
	}
	if len(lines) != 3 {
		t.Fatalf("phải có ba dòng, nhận %d", len(lines))
	}
	if lines[0] != "*   topic" {
		t.Errorf("nhánh độc lập phải nằm ở làn đầu và được đệm, nhận %q", lines[0])
	}
	if !strings.HasPrefix(lines[1], "| * ") {
		t.Errorf("nhánh chính phải ở làn đầu, nhận %q", lines[1])
	}
	// Hai nhánh gặp nhau ở base, nên chỉ còn một làn và không còn dấu gạch nào.
	if !strings.HasPrefix(lines[2], "* ") || strings.Contains(lines[2], "|") {
		t.Errorf("hai nhánh phải quy về một làn ở base, nhận %q", lines[2])
	}
	// Dòng base được đệm cho cùng bề rộng hai làn như các dòng trên.
	if lines[2] != "*   base" {
		t.Errorf("dòng base phải là %q, nhận %q", "*   base", lines[2])
	}
}

// TestGraphPrefixWidth bảo đảm mọi dòng cùng bề rộng tiền tố.
//
// Tiện ích VS Code ghép tiền tố này vào nhãn của cây, lệch cột thì nhánh nhìn
// không ra đường thẳng.
func TestGraphPrefixWidth(t *testing.T) {
	entries := buildLog([]fakeCommit{
		{"merge", []string{"m", "s"}},
		{"m", []string{"root"}},
		{"s", []string{"root"}},
		{"root", nil},
	})
	ops.AssignGraphLanes(entries)

	// Đo cả khoảng trắng đệm ở cuối, vì bề rộng đó quyết định các cột có
	// thẳng hàng hay không.
	width := -1
	for _, e := range entries {
		w := len(e.GraphPrefix)
		if width == -1 {
			width = w
		}
		if w != width {
			t.Errorf("tiền tố của %s rộng %d ký tự, khác dòng đầu (%d)", e.Summary, w, width)
		}
		t.Logf("%-6s tiền tố %q", e.Summary, e.GraphPrefix)
	}
	// Biểu đồ thử này có hai làn nên tiền tố rộng bốn ký tự.
	if width != 4 {
		t.Errorf("tiền tố hai làn phải rộng 4 ký tự, nhận %d", width)
	}
}

// TestGraphBounds bảo đảm các trường hợp biên không làm hỏng.
func TestGraphBounds(t *testing.T) {
	ops.AssignGraphLanes(nil)              // danh sách rỗng không được lỗi
	ops.AssignGraphLanes([]ops.LogEntry{}) // cũng vậy

	single := strings.Join(renderGraph([]fakeCommit{{"only", nil}}), "\n")
	if single != "* only" {
		t.Errorf("lịch sử một commit phải là %q, nhận %q", "* only", single)
	}
}

// TestGraphCommitWithoutParents bảo đảm commit không có cha không làm lỗi lát cắt.
func TestGraphCommitWithoutParents(t *testing.T) {
	// Cắt lát cha từ vị trí 1 trên lát rỗng sẽ lỗi, và commit cuối cùng của
	// mọi nhánh đều rơi vào trường hợp đó.
	entries := buildLog([]fakeCommit{{"a", nil}, {"b", nil}, {"c", nil}})
	ops.AssignGraphLanes(entries)
	for _, e := range entries {
		if e.GraphPrefix == "" {
			t.Errorf("commit %s không được mất tiền tố", e.Summary)
		}
	}
}
