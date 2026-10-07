package ops_test

import (
	"github.com/tonguyenducmanh/devcli/internal/vcs/ops"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tonguyenducmanh/devcli/internal/vcs/object"
	"github.com/tonguyenducmanh/devcli/internal/vcs/repo"
)

// newRepo tạo repo tạm với danh tính cố định.
func newRepo(t *testing.T) *repo.Repo {
	t.Helper()
	r, err := repo.Init(t.TempDir(), "main")
	if err != nil {
		t.Fatalf("không tạo được repo: %v", err)
	}
	r.Config.Set("user.name", "Người Kiểm Thử")
	r.Config.Set("user.email", "test@example.com")
	if err := r.Config.Save(); err != nil {
		t.Fatal(err)
	}
	return r
}

// write ghi nội dung một file trong thư mục làm việc, tạo thư mục cha nếu thiếu.
func write(t *testing.T, r *repo.Repo, rel, content string) {
	t.Helper()
	abs := r.WorkPath(rel)
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(abs, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// read đọc nội dung một file trong thư mục làm việc.
func read(t *testing.T, r *repo.Repo, rel string) string {
	t.Helper()
	data, err := os.ReadFile(r.WorkPath(rel))
	if err != nil {
		t.Fatalf("không đọc được %s: %v", rel, err)
	}
	return string(data)
}

// commitAll stage mọi thay đổi rồi commit với message cho trước.
func commitAll(t *testing.T, r *repo.Repo, msg string) object.Hash {
	t.Helper()
	h, err := ops.Commit(r, ops.CommitOptions{Message: msg, All: true})
	if err != nil {
		t.Fatalf("commit thất bại: %v", err)
	}
	return h
}

// checkout chuyển nhánh và báo lỗi nếu không thành công.
func checkout(t *testing.T, r *repo.Repo, branch string) {
	t.Helper()
	if err := ops.Checkout(r, ops.CheckoutOptions{Target: branch}); err != nil {
		t.Fatalf("không chuyển được nhánh %s: %v", branch, err)
	}
}

func TestAddAndCommitBasic(t *testing.T) {
	r := newRepo(t)
	write(t, r, "a.txt", "dòng 1\n")
	write(t, r, "src/main.go", "package main\n")

	if err := ops.Add(r, ops.AddOptions{All: true}); err != nil {
		t.Fatal(err)
	}
	h := commitAll(t, r, "commit đầu tiên")

	// Sau commit cây làm việc phải sạch.
	st, err := r.Status()
	if err != nil {
		t.Fatal(err)
	}
	if !st.IsClean() {
		t.Fatalf("cây làm việc phải sạch sau commit: %+v", st.Entries)
	}
	if h.IsZero() {
		t.Fatalf("phải tạo được commit")
	}
}

func TestCommitWithoutMessageErrors(t *testing.T) {
	r := newRepo(t)
	if _, err := ops.Commit(r, ops.CommitOptions{Message: "không có gì"}); err == nil {
		t.Fatalf("phải báo lỗi khi chưa có gì để commit")
	}
	write(t, r, "a.txt", "x\n")
	if _, err := ops.Commit(r, ops.CommitOptions{Message: "không stage"}); err == nil {
		t.Fatalf("phải báo lỗi khi có thay đổi chưa stage")
	}
}

func TestCommitAmend(t *testing.T) {
	r := newRepo(t)
	write(t, r, "a.txt", "nội dung\n")
	c1 := commitAll(t, r, "tin nhắn cũ")

	if _, err := ops.Commit(r, ops.CommitOptions{Message: "tin nhắn mới", Amend: true, AllowEmpty: true}); err != nil {
		t.Fatal(err)
	}
	head, _ := r.Head()
	if head == c1 {
		t.Fatalf("amend phải tạo commit mới")
	}
	entries, err := ops.Log(r, ops.LogOptions{Max: 5})
	if err != nil {
		t.Fatal(err)
	}
	if entries[0].Summary != "tin nhắn mới" {
		t.Fatalf("tin nhắn chưa được cập nhật: %q", entries[0].Summary)
	}
	// Số commit vẫn phải là 1 vì amend thay thế chứ không thêm.
	if len(entries) != 1 {
		t.Fatalf("amend không được tạo thêm commit, nhận %d", len(entries))
	}
}

func TestBranchAndCheckout(t *testing.T) {
	r := newRepo(t)
	write(t, r, "a.txt", "1\n")
	commitAll(t, r, "c1")

	if err := ops.CreateBranch(r, ops.CreateBranchOptions{Name: "tinh-nang", StartPoint: "main"}); err != nil {
		t.Fatal(err)
	}
	checkout(t, r, "tinh-nang")
	write(t, r, "b.txt", "2\n")
	commitAll(t, r, "trên tinh nang")

	checkout(t, r, "main")
	if _, err := os.Stat(r.WorkPath("b.txt")); !os.IsNotExist(err) {
		t.Fatalf("file của nhánh khác không được xuất hiện ở main")
	}

	res, err := ops.Merge(r, ops.MergeOptions{Branch: "tinh-nang"})
	if err != nil {
		t.Fatal(err)
	}
	if !res.FastForward {
		t.Fatalf("merge từ nhánh con phải fast-forward: %+v", res)
	}
	if _, err := os.Stat(r.WorkPath("b.txt")); err != nil {
		t.Fatalf("sau fast-forward phải có file của nhánh khác: %v", err)
	}
}

func TestDeleteBranch(t *testing.T) {
	r := newRepo(t)
	write(t, r, "a.txt", "1\n")
	commitAll(t, r, "c1")
	if err := ops.CreateBranch(r, ops.CreateBranchOptions{Name: "cu", StartPoint: "main"}); err != nil {
		t.Fatal(err)
	}

	// Không được xóa nhánh đang đứng.
	results, err := ops.DeleteBranches(r, []string{"main"}, true)
	if err != nil {
		t.Fatal(err)
	}
	if results[0].Deleted {
		t.Fatalf("không được xóa nhánh hiện tại")
	}

	results, err = ops.DeleteBranches(r, []string{"cu"}, true)
	if err != nil {
		t.Fatal(err)
	}
	if !results[0].Deleted {
		t.Fatalf("phải xóa được nhánh khác: %+v", results[0])
	}
	if r.BranchExists("cu") {
		t.Fatalf("nhánh vẫn còn sau khi xóa")
	}
}

func TestMergeMissingFastForwardErrors(t *testing.T) {
	r := newRepo(t)
	write(t, r, "f.txt", "a\n")
	commitAll(t, r, "c1")

	if _, err := ops.Merge(r, ops.MergeOptions{Branch: "khong-co-that"}); err == nil {
		t.Fatalf("phải báo lỗi khi nhánh không tồn tại")
	}
}

func TestMergeBothDirections(t *testing.T) {
	r := newRepo(t)
	write(t, r, "f.txt", "a\nb\nc\nd\ne\n")
	commitAll(t, r, "c1")

	if err := ops.CreateBranch(r, ops.CreateBranchOptions{Name: "phu", StartPoint: "main"}); err != nil {
		t.Fatal(err)
	}
	checkout(t, r, "phu")
	write(t, r, "f.txt", "a\nB-cua-phu\nc\nd\ne\n")
	commitAll(t, r, "sửa trên nhánh phụ")

	checkout(t, r, "main")
	write(t, r, "f.txt", "a\nb\nc\nd\nE-cua-main\n")
	commitAll(t, r, "sửa trên main")

	res, err := ops.Merge(r, ops.MergeOptions{Branch: "phu"})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Conflicts) != 0 {
		t.Fatalf("sửa hai chỗ khác nhau thì không được xung đột: %v", res.Conflicts)
	}
	want := "a\nB-cua-phu\nc\nd\nE-cua-main\n"
	if got := read(t, r, "f.txt"); got != want {
		t.Fatalf("nội dung sau merge sai:\nmong đợi %q\nnhận  %q", want, got)
	}
}

func TestMergeConflictAndResolution(t *testing.T) {
	r := newRepo(t)
	write(t, r, "f.txt", "a\nb\nc\n")
	commitAll(t, r, "c1")

	if err := ops.CreateBranch(r, ops.CreateBranchOptions{Name: "phu", StartPoint: "main"}); err != nil {
		t.Fatal(err)
	}
	checkout(t, r, "phu")
	write(t, r, "f.txt", "a\nB-cua-phu\nc\n")
	commitAll(t, r, "sửa trên nhánh phụ")

	checkout(t, r, "main")
	write(t, r, "f.txt", "a\nB-cua-main\nc\n")
	commitAll(t, r, "sửa trên main")

	res, err := ops.Merge(r, ops.MergeOptions{Branch: "phu"})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Conflicts) != 1 || res.Conflicts[0] != "f.txt" {
		t.Fatalf("phải báo đúng một xung đột: %+v", res.Conflicts)
	}
	content := read(t, r, "f.txt")
	for _, marker := range []string{"<<<<<<<", "=======", ">>>>>>>"} {
		if !strings.Contains(content, marker) {
			t.Fatalf("thiếu dấu %q trong file xung đột:\n%s", marker, content)
		}
	}

	// Giải quyết thủ công rồi commit để hoàn tất.
	write(t, r, "f.txt", "a\nB-da_lua_chon\nc\n")
	if err := ops.Add(r, ops.AddOptions{All: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := ops.Commit(r, ops.CommitOptions{Message: "giải quyết xung đột"}); err != nil {
		t.Fatal(err)
	}
	if ops.MergeInProgress(r) {
		t.Fatalf("sau khi commit phải không còn trạng thái merge")
	}
}

func TestMergeAbort(t *testing.T) {
	r := newRepo(t)
	write(t, r, "f.txt", "a\nb\nc\n")
	commitAll(t, r, "c1")
	if err := ops.CreateBranch(r, ops.CreateBranchOptions{Name: "phu", StartPoint: "main"}); err != nil {
		t.Fatal(err)
	}
	checkout(t, r, "phu")
	write(t, r, "f.txt", "a\nB\nc\n")
	commitAll(t, r, "sửa")
	checkout(t, r, "main")
	write(t, r, "f.txt", "a\nM\nc\n")
	commitAll(t, r, "sửa")

	if _, err := ops.Merge(r, ops.MergeOptions{Branch: "phu"}); err != nil {
		t.Fatal(err)
	}
	if !ops.MergeInProgress(r) {
		t.Fatalf("phải còn trạng thái merge sau khi gặp xung đột")
	}
	if _, err := ops.Merge(r, ops.MergeOptions{Abort: true}); err != nil {
		t.Fatal(err)
	}
	if got := read(t, r, "f.txt"); got != "a\nM\nc\n" {
		t.Fatalf("abort phải khôi phục nội dung trước merge, nhận %q", got)
	}
}

func TestCherryPick(t *testing.T) {
	r := newRepo(t)
	write(t, r, "f.txt", "1\n")
	commitAll(t, r, "c1")

	if err := ops.CreateBranch(r, ops.CreateBranchOptions{Name: "phu", StartPoint: "main"}); err != nil {
		t.Fatal(err)
	}
	checkout(t, r, "phu")
	write(t, r, "f.txt", "1\n2\n")
	target := commitAll(t, r, "thêm dòng")

	checkout(t, r, "main")
	res, err := ops.CherryPick(r, ops.CherryPickOptions{Commits: []object.Hash{target}})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Commits) != 1 {
		t.Fatalf("phải tạo một commit mới: %+v", res)
	}
	if got := read(t, r, "f.txt"); got != "1\n2\n" {
		t.Fatalf("nội dung sau cherry-pick sai: %q", got)
	}
}

func TestRevert(t *testing.T) {
	r := newRepo(t)
	write(t, r, "f.txt", "1\n")
	commitAll(t, r, "c1")
	write(t, r, "f.txt", "1\n2\n")
	c2 := commitAll(t, r, "thêm dòng")

	if _, err := ops.Revert(r, ops.RevertOptions{Commits: []object.Hash{c2}}); err != nil {
		t.Fatal(err)
	}
	if got := read(t, r, "f.txt"); got != "1\n" {
		t.Fatalf("revert phải xóa nội dung đã thêm, nhận %q", got)
	}
}

func TestStashAndPop(t *testing.T) {
	r := newRepo(t)
	write(t, r, "f.txt", "1\n")
	commitAll(t, r, "c1")
	write(t, r, "f.txt", "1\n2\n")

	h, err := ops.Stash(r, "việc dang dở", false)
	if err != nil {
		t.Fatalf("stash thất bại: %v", err)
	}
	if got := read(t, r, "f.txt"); got != "1\n" {
		t.Fatalf("stash phải đưa file về nội dung cũ, nhận %q", got)
	}
	entries, err := ops.StashList(r)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Hash != h {
		t.Fatalf("danh sách stash sai: %+v", entries)
	}
	if err := ops.StashPop(r, 0); err != nil {
		t.Fatal(err)
	}
	if got := read(t, r, "f.txt"); got != "1\n2\n" {
		t.Fatalf("pop phải khôi phục nội dung đã sửa, nhận %q", got)
	}
}

func TestStashUntrackedFile(t *testing.T) {
	r := newRepo(t)
	write(t, r, "f.txt", "1\n")
	commitAll(t, r, "c1")
	write(t, r, "moi.txt", "file mới\n")

	if _, err := ops.Stash(r, "", true); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(r.WorkPath("moi.txt")); !os.IsNotExist(err) {
		t.Fatalf("file chưa theo dõi phải bị xóa khỏi đĩa sau khi stash")
	}
	if err := ops.StashPop(r, 0); err != nil {
		t.Fatal(err)
	}
	if got := read(t, r, "moi.txt"); got != "file mới\n" {
		t.Fatalf("pop phải khôi phục cả file chưa theo dõi, nhận %q", got)
	}
}

func TestStashWithNoChanges(t *testing.T) {
	r := newRepo(t)
	write(t, r, "f.txt", "1\n")
	commitAll(t, r, "c1")
	if _, err := ops.Stash(r, "", false); err == nil {
		t.Fatalf("phải báo lỗi khi không có gì để lưu tạm")
	}
}

func TestRebaseOntoOtherBranch(t *testing.T) {
	r := newRepo(t)
	write(t, r, "base.txt", "1\n")
	commitAll(t, r, "c1")

	if err := ops.CreateBranch(r, ops.CreateBranchOptions{Name: "phu", StartPoint: "main"}); err != nil {
		t.Fatal(err)
	}
	checkout(t, r, "phu")
	write(t, r, "phu.txt", "nội dung phụ\n")
	commitAll(t, r, "c2")

	checkout(t, r, "main")
	write(t, r, "main.txt", "nội dung main\n")
	commitAll(t, r, "c3")

	// Rebase nhánh phụ lên trên main.
	res, err := ops.Rebase(r, ops.RebaseOptions{Upstream: "main", Branch: "phu"})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Conflicts) > 0 {
		t.Fatalf("rebase không được gây xung đột: %v", res.Conflicts)
	}
	if res.Applied != 1 {
		t.Fatalf("phải áp dụng lại 1 commit, nhận %d", res.Applied)
	}
	// Sau rebase, người dùng được đưa về nhánh ban đầu (main),
	// nên cây làm việc phải chứa nội dung của main.
	if branch, _ := r.CurrentBranch(); branch != "main" {
		t.Fatalf("sau rebase phải quay lại nhánh ban đầu, nhận %q", branch)
	}
	if _, err := os.Stat(r.WorkPath("main.txt")); err != nil {
		t.Fatalf("sau rebase phải có nội dung của main: %v", err)
	}

	// Nội dung riêng của nhánh phụ vẫn phải nằm trong lịch sử nhánh đó.
	otherTip, err := r.BranchHash("phu")
	if err != nil {
		t.Fatal(err)
	}
	otherTree, err := r.CommitTree(otherTip)
	if err != nil {
		t.Fatal(err)
	}
	otherNodes, err := r.Flatten(otherTree)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := otherNodes["phu.txt"]; !ok {
		t.Fatalf("sau rebase nhánh phụ phải giữ tệp riêng của nó")
	}
	if _, ok := otherNodes["main.txt"]; !ok {
		t.Fatalf("sau rebase nhánh phụ phải chứa cả nội dung của main")
	}
}

func TestRebaseKeepsBranchName(t *testing.T) {
	r := newRepo(t)
	write(t, r, "a.txt", "1\n")
	commitAll(t, r, "c1")
	write(t, r, "b.txt", "2\n")
	c2 := commitAll(t, r, "c2")

	// Rebase lên chính HEAD hiện tại không làm thay đổi gì.
	res, err := ops.Rebase(r, ops.RebaseOptions{Upstream: "HEAD"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Applied != 0 {
		t.Fatalf("không nên áp dụng lại commit nào, nhận %d", res.Applied)
	}
	head, _ := r.Head()
	if head != c2 {
		t.Fatalf("HEAD không được đổi")
	}
}

func TestRebaseAndAbort(t *testing.T) {
	r := newRepo(t)
	write(t, r, "gia.txt", "1\n2\n3\n")
	commitAll(t, r, "c1")

	if err := ops.CreateBranch(r, ops.CreateBranchOptions{Name: "phu", StartPoint: "main"}); err != nil {
		t.Fatal(err)
	}
	checkout(t, r, "phu")
	// Sửa cùng dòng với main để tạo xung đột khi rebase.
	write(t, r, "gia.txt", "1\n2\ncu-aa-phu\n")
	commitAll(t, r, "trên phu")

	checkout(t, r, "main")
	write(t, r, "gia.txt", "1\n2\ncu-aa-main\n")
	commitAll(t, r, "trên main")
	origHead, _ := r.Head()

	// Rebase nhánh phụ lên main: xung đột nên rebase phải dừng lại.
	res, err := ops.Rebase(r, ops.RebaseOptions{Upstream: "main", Branch: "phu"})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Conflicts) == 0 {
		t.Fatalf("phải báo xung đột, nhận %+v", res)
	}
	if !ops.RebaseInProgress(r) {
		t.Fatalf("rebase dừng giữa chừng thì phải còn trạng thái")
	}

	// Huỷ rebase phải đưa mọi thứ về trạng thái trước đó.
	if _, err := ops.Rebase(r, ops.RebaseOptions{Abort: true}); err != nil {
		t.Fatalf("abort thất bại: %v", err)
	}
	if ops.RebaseInProgress(r) {
		t.Fatalf("abort phải xóa trạng thái rebase")
	}
	if got := read(t, r, "gia.txt"); got != "1\n2\ncu-aa-main\n" {
		t.Fatalf("abort phải khôi phục nội dung trước rebase, nhận %q", got)
	}
	head, _ := r.Head()
	if head != origHead {
		t.Fatalf("abort phải đưa HEAD về đúng commit trước khi rebase")
	}
}

func TestResetModes(t *testing.T) {
	r := newRepo(t)
	write(t, r, "f.txt", "1\n")
	commitAll(t, r, "c1")
	write(t, r, "f.txt", "2\n")
	c2 := commitAll(t, r, "c2")

	// Hard reset đưa cả nội dung đĩa về c1.
	if err := ops.Reset(r, ops.ResetOptions{Target: "HEAD~1", Mode: "hard"}); err != nil {
		t.Fatal(err)
	}
	if got := read(t, r, "f.txt"); got != "1\n" {
		t.Fatalf("hard reset phải khôi phục nội dung đĩa, nhận %q", got)
	}
	head, _ := r.Head()
	if head == c2 {
		t.Fatalf("hard reset phải di chuyển HEAD")
	}

	// Soft reset chỉ di chuyển con trỏ, giữ nguyên index.
	if err := ops.Reset(r, ops.ResetOptions{Target: "HEAD", Mode: "soft"}); err != nil {
		t.Fatal(err)
	}
}

func TestTag(t *testing.T) {
	r := newRepo(t)
	write(t, r, "f.txt", "1\n")
	c1 := commitAll(t, r, "c1")

	if err := r.WriteTag("v1", c1); err != nil {
		t.Fatal(err)
	}
	if err := r.WriteAnnotatedTag("v2", c1, "ghi chú"); err != nil {
		t.Fatal(err)
	}
	tags, err := r.ListTags()
	if err != nil {
		t.Fatal(err)
	}
	if len(tags) != 2 {
		t.Fatalf("phải có 2 tag, nhận %d", len(tags))
	}
	if !tags[1].Annotated {
		t.Fatalf("tag có chú thích phải được nhận diện")
	}
	if err := r.DeleteTag("v1"); err != nil {
		t.Fatal(err)
	}
	if r.Refs.Exists("refs/tags/v1") {
		t.Fatalf("tag phải bị xóa")
	}
}

func TestLogAndRefFilter(t *testing.T) {
	r := newRepo(t)
	write(t, r, "a.txt", "1\n")
	commitAll(t, r, "sửa a")
	write(t, r, "b.txt", "1\n")
	commitAll(t, r, "thêm b")

	all, err := ops.Log(r, ops.LogOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 2 {
		t.Fatalf("phải có 2 commit, nhận %d", len(all))
	}
	if all[0].Summary != "thêm b" {
		t.Fatalf("thứ tự log sai: %q", all[0].Summary)
	}

	// Chỉ lọc theo một đường dẫn.
	filtered, err := ops.Log(r, ops.LogOptions{Paths: []string{"a.txt"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(filtered) != 1 || filtered[0].Summary != "sửa a" {
		t.Fatalf("lọc theo đường dẫn sai: %+v", filtered)
	}
}

func TestDiffThreeWay(t *testing.T) {
	r := newRepo(t)
	write(t, r, "f.txt", "1\n2\n")
	commitAll(t, r, "c1")

	// Sửa trên đĩa: diff mặc định phải thấy thay đổi.
	write(t, r, "f.txt", "1\nX\n")
	diffs, err := ops.Diff(r, ops.DiffOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(diffs) != 1 || diffs[0].Path != "f.txt" {
		t.Fatalf("diff mặc định phải thấy thay đổi chưa stage: %+v", diffs)
	}

	// Sau khi stage: diff mặc định rỗng, diff --staged có nội dung.
	if err := ops.Add(r, ops.AddOptions{All: true}); err != nil {
		t.Fatal(err)
	}
	diffs, err = ops.Diff(r, ops.DiffOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(diffs) != 0 {
		t.Fatalf("sau khi stage, diff mặc định phải rỗng: %+v", diffs)
	}
	diffs, err = ops.Diff(r, ops.DiffOptions{Staged: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(diffs) != 1 || diffs[0].Added != 1 || diffs[0].Del != 1 {
		t.Fatalf("diff staged sai: %+v", diffs)
	}
}

func TestRemoveAndMove(t *testing.T) {
	r := newRepo(t)
	write(t, r, "cu.txt", "1\n")
	write(t, r, "di.txt", "2\n")
	commitAll(t, r, "c1")

	if err := ops.Move(r, "cu.txt", "thu-muc/moi.txt"); err != nil {
		t.Fatalf("move thất bại: %v", err)
	}
	if _, err := os.Stat(r.WorkPath("thu-muc/moi.txt")); err != nil {
		t.Fatalf("file sau move phải tồn tại: %v", err)
	}

	if err := ops.Remove(r, []string{"di.txt"}, true, false); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(r.WorkPath("di.txt")); !os.IsNotExist(err) {
		t.Fatalf("file sau rm phải bị xóa khỏi đĩa")
	}
}

func TestRestoreFromCommit(t *testing.T) {
	r := newRepo(t)
	write(t, r, "f.txt", "ban dau\n")
	commitAll(t, r, "c1")
	write(t, r, "f.txt", "sau\n")
	commitAll(t, r, "c2")

	if err := ops.Restore(r, ops.RestoreOptions{Source: "HEAD~1", Paths: []string{"f.txt"}}); err != nil {
		t.Fatal(err)
	}
	if got := read(t, r, "f.txt"); got != "ban dau\n" {
		t.Fatalf("restore phải lấy nội dung từ commit cũ, nhận %q", got)
	}
}

func TestFsckCleanRepo(t *testing.T) {
	r := newRepo(t)
	write(t, r, "f.txt", "1\n")
	commitAll(t, r, "c1")

	report, err := ops.Fsck(r)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Problems) != 0 {
		t.Fatalf("kho sạch không được có vấn đề: %v", report.Problems)
	}
	if report.ObjectCount == 0 || report.RefCount == 0 {
		t.Fatalf("số lượng object và ref phải lớn hơn 0: %+v", report)
	}
}
