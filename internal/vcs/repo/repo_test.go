package repo

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tonguyenducmanh/devcli/internal/vcs/object"
)

// newTestRepo tạo một repo tạm cho mỗi phép kiểm tra.
func newTestRepo(t *testing.T) *Repo {
	t.Helper()
	dir := t.TempDir()
	// Ghi danh tính cố định để kết quả không phụ thuộc máy đang chạy.
	r, err := Init(dir, "main")
	if err != nil {
		t.Fatalf("không tạo được repo: %v", err)
	}
	r.Config.Set("user.name", "Người Kiểm Thử")
	r.Config.Set("user.email", "test@example.com")
	if err := r.Config.Save(); err != nil {
		t.Fatalf("không lưu được cấu hình: %v", err)
	}
	return r
}

// writeFile tạo một file trong thư mục làm việc.
func writeFile(t *testing.T, r *Repo, rel, content string) {
	t.Helper()
	abs := r.WorkPath(rel)
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		t.Fatalf("không tạo được thư mục: %v", err)
	}
	if err := os.WriteFile(abs, []byte(content), 0o644); err != nil {
		t.Fatalf("không ghi được file: %v", err)
	}
}

// readFile đọc nội dung file trong thư mục làm việc.
func readFile(t *testing.T, r *Repo, rel string) string {
	t.Helper()
	data, err := os.ReadFile(r.WorkPath(rel))
	if err != nil {
		t.Fatalf("không đọc được file: %v", err)
	}
	return string(data)
}

// commitAll stage mọi thay đổi rồi tạo commit với message cho trước.
func commitAll(t *testing.T, r *Repo, msg string) object.Hash {
	t.Helper()
	if err := StageAllWorktree(r); err != nil {
		t.Fatalf("không stage được: %v", err)
	}
	tree, err := r.TreeFromIndex()
	if err != nil {
		t.Fatalf("không dựng được tree: %v", err)
	}
	head, _ := r.Head()
	var parents []object.Hash
	if !head.IsZero() {
		parents = append(parents, head)
	}
	id := r.Identity()
	c := &object.Commit{
		Tree:      tree,
		Parents:   parents,
		Author:    id,
		Committer: id,
		Message:   msg + "\n",
	}
	h, err := r.Objects.WriteCommit(c)
	if err != nil {
		t.Fatalf("không ghi được commit: %v", err)
	}
	if err := r.UpdateHeadCommit(h, "commit: "+msg); err != nil {
		t.Fatalf("không cập nhật HEAD: %v", err)
	}
	return h
}

// StageAllWorktree đưa toàn bộ nội dung đĩa vào index.
func StageAllWorktree(r *Repo) error {
	nodes, err := r.ScanWorktree()
	if err != nil {
		return err
	}
	for _, n := range nodes {
		if err := r.StageFile(n.Path); err != nil {
			return err
		}
	}
	return r.SaveIndex()
}

func TestInitVaMoRepo(t *testing.T) {
	r := newTestRepo(t)
	// Mở lại từ thư mục con vẫn phải tìm thấy repo.
	child := r.WorkPath("sub/dir")
	if err := os.MkdirAll(child, 0o755); err != nil {
		t.Fatal(err)
	}
	got, err := Open(child)
	if err != nil {
		t.Fatalf("không mở được repo từ thư mục con: %v", err)
	}
	if got.Root != r.Root {
		t.Fatalf("sai thư mục gốc: mong đợi %s, nhận %s", r.Root, got.Root)
	}
	if branch, _ := got.CurrentBranch(); branch != "main" {
		t.Fatalf("sai nhánh hiện tại: %s", branch)
	}
}

func TestMoRepoONoiKhongPhaiRepo(t *testing.T) {
	dir := t.TempDir()
	if _, err := Open(dir); err == nil {
		t.Fatalf("phải báo lỗi khi thư mục không phải repo")
	}
}

func TestCommitVaStatus(t *testing.T) {
	r := newTestRepo(t)
	writeFile(t, r, "a.txt", "nội dung A\n")
	writeFile(t, r, "src/main.go", "package main\n")

	if err := StageAllWorktree(r); err != nil {
		t.Fatal(err)
	}
	st, err := r.Status()
	if err != nil {
		t.Fatal(err)
	}
	if len(st.Staged()) != 2 {
		t.Fatalf("phải có 2 file đã stage, nhận %d", len(st.Staged()))
	}
	if st.IndexStatusOf("a.txt") != 'A' {
		t.Fatalf("file mới phải có trạng thái thêm")
	}

	h := commitAll(t, r, "commit đầu")
	if h.IsZero() {
		t.Fatalf("commit không tạo được")
	}
	st, err = r.Status()
	if err != nil {
		t.Fatal(err)
	}
	if !st.IsClean() {
		t.Fatalf("sau khi commit phải sạch, nhận %v", st.Entries)
	}
}

func TestThayDoiChuaStageKhongLoiVaoCommit(t *testing.T) {
	r := newTestRepo(t)
	writeFile(t, r, "a.txt", "v1\n")
	c1 := commitAll(t, r, "c1")

	// Sửa file trên đĩa nhưng không stage.
	writeFile(t, r, "a.txt", "v2\n")
	st, err := r.Status()
	if err != nil {
		t.Fatal(err)
	}
	if len(st.Unstaged()) != 1 || st.Unstaged()[0].WorkStatus != 'M' {
		t.Fatalf("phải phát hiện file đã sửa chưa stage: %+v", st.Entries)
	}
	if len(st.Staged()) != 0 {
		t.Fatalf("không có gì được stage, nhận %+v", st.Staged())
	}

	// Dựng tree trực tiếp từ index: phải vẫn là nội dung cũ.
	tree, err := r.TreeFromIndex()
	if err != nil {
		t.Fatal(err)
	}
	nodes, err := r.Flatten(tree)
	if err != nil {
		t.Fatal(err)
	}
	c1Nodes, err := r.Flatten(mustCommitTree(t, r, c1))
	if err != nil {
		t.Fatal(err)
	}
	if nodes["a.txt"].Hash != c1Nodes["a.txt"].Hash {
		t.Fatalf("tree từ index phải khớp với HEAD khi chưa stage thay đổi")
	}
}

// mustCommitTree trả về tree của một commit.
func mustCommitTree(t *testing.T, r *Repo, h object.Hash) object.Hash {
	t.Helper()
	tree, err := r.CommitTree(h)
	if err != nil {
		t.Fatal(err)
	}
	return tree
}

func TestIndexVaCayDungPhanCay(t *testing.T) {
	r := newTestRepo(t)
	writeFile(t, r, "docs/huong-dan.md", "# hướng dẫn\n")
	writeFile(t, r, "src/deep/nested/file.go", "package deep\n")
	commitAll(t, r, "c1")

	head, err := r.Head()
	if err != nil || head.IsZero() {
		t.Fatalf("không có HEAD: %v", err)
	}
	tree, err := r.CommitTree(head)
	if err != nil {
		t.Fatal(err)
	}
	nodes, err := r.Flatten(tree)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{"docs/huong-dan.md", "src/deep/nested/file.go"} {
		if _, ok := nodes[p]; !ok {
			t.Fatalf("thiếu %s trong cây phẳng: %+v", p, nodes)
		}
	}
	// Cây trung gian phải tồn tại đúng dạng tree.
	dirHash, err := r.Objects.ReadTree(mustTreeHash(t, r, tree, "src"))
	if err != nil {
		t.Fatalf("đọc cây thư mục con thất bại: %v", err)
	}
	if len(dirHash.Entries) == 0 {
		t.Fatalf("cây thư mục con rỗng")
	}
}

// mustTreeHash trả về hash tree của một đường dẫn con trong cây.
func mustTreeHash(t *testing.T, r *Repo, tree object.Hash, name string) object.Hash {
	t.Helper()
	top, err := r.Objects.ReadTree(tree)
	if err != nil {
		t.Fatal(err)
	}
	e := top.Get(name)
	if e == nil {
		t.Fatalf("không có entry %s", name)
	}
	return e.Hash
}

func TestMergeBaseVaIsAncestor(t *testing.T) {
	r := newTestRepo(t)
	writeFile(t, r, "a.txt", "1\n")
	c1 := commitAll(t, r, "c1")

	writeFile(t, r, "b.txt", "2\n")
	c2 := commitAll(t, r, "c2")

	if ok, err := r.IsAncestor(c1, c2); err != nil || !ok {
		t.Fatalf("c1 phải là tổ tiên của c2: %v %v", ok, err)
	}
	if ok, err := r.IsAncestor(c2, c1); err != nil || ok {
		t.Fatalf("c2 không được là tổ tiên của c1")
	}
	base, err := r.MergeBase(c1, c2)
	if err != nil {
		t.Fatal(err)
	}
	if base != c1 {
		t.Fatalf("điểm chung phải là c1, nhận %s", base.Short(8))
	}
}

func TestResolveRev(t *testing.T) {
	r := newTestRepo(t)
	writeFile(t, r, "a.txt", "1\n")
	c1 := commitAll(t, r, "c1")
	writeFile(t, r, "b.txt", "2\n")
	c2 := commitAll(t, r, "c2")

	cases := map[string]object.Hash{
		"HEAD":      c2,
		"HEAD~1":    c1,
		"HEAD^":     c1,
		"main":      c2,
		"main~1":    c1,
		c2.Short(8): c2,
	}
	for name, want := range cases {
		got, err := r.ResolveRev(name)
		if err != nil {
			t.Fatalf("phân giải %q lỗi: %v", name, err)
		}
		if got != want {
			t.Fatalf("%q phải ra %s, nhận %s", name, want.Short(8), got.Short(8))
		}
	}
	if _, err := r.ResolveRev("khong-ton-tai"); err == nil {
		t.Fatalf("tham chiếu không tồn tại phải báo lỗi")
	}
}

func TestBranchesVaTags(t *testing.T) {
	r := newTestRepo(t)
	writeFile(t, r, "a.txt", "1\n")
	c1 := commitAll(t, r, "c1")

	if err := r.CreateBranch("tinh-nang", "main", c1); err != nil {
		t.Fatalf("không tạo được nhánh: %v", err)
	}
	branches, err := r.Branches()
	if err != nil {
		t.Fatal(err)
	}
	if len(branches) != 2 {
		t.Fatalf("phải có 2 nhánh, nhận %v", branches)
	}

	if err := r.WriteTag("v1", c1); err != nil {
		t.Fatal(err)
	}
	tags, err := r.Tags()
	if err != nil {
		t.Fatal(err)
	}
	if len(tags) != 1 || tags[0] != "v1" {
		t.Fatalf("sai danh sách tag: %v", tags)
	}

	// Xóa nhánh rồi kiểm tra không còn tồn tại.
	if err := r.DeleteBranch("tinh-nang", true); err != nil {
		t.Fatal(err)
	}
	if r.BranchExists("tinh-nang") {
		t.Fatalf("nhánh vẫn còn sau khi xóa")
	}
}

func TestIgnoreRules(t *testing.T) {
	r := newTestRepo(t)
	writeFile(t, r, ".tdxignore", "build/\n*.log\n")
	// Nạp lại quy tắc vì repo đang mở từ trước khi có file ignore.
	r2, err := Open(r.Root)
	if err != nil {
		t.Fatal(err)
	}
	cases := map[string]bool{
		"build/out.o":  true,
		"debug.log":    true,
		"src/main.go":  false,
		"keep.log.txt": false,
	}
	for path, want := range cases {
		if got := r2.Ignore.Matches(r2.Root, path); got != want {
			t.Fatalf("%s: mong đợi bỏ qua=%v, nhận %v", path, want, got)
		}
	}
}

func TestIgnoredFilesKhongVaIndex(t *testing.T) {
	r := newTestRepo(t)
	writeFile(t, r, ".tdxignore", "*.log\n")
	r2, err := Open(r.Root)
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, r2, "keep.txt", "giữ\n")
	writeFile(t, r2, "huongdan.log", "bỏ qua\n")

	if err := StageAllWorktree(r2); err != nil {
		t.Fatal(err)
	}
	if r2.Index.Get("huongdan.log") != nil {
		t.Fatalf("file bị bỏ qua không được vào index")
	}
	if r2.Index.Get("keep.txt") == nil {
		t.Fatalf("file thường phải vào index")
	}
}
