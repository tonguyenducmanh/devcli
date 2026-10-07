package object_test

import (
	"github.com/tonguyenducmanh/devcli/internal/vcs/object"
	"testing"
)

func TestComputeHashKhopChuanGit(t *testing.T) {
	// Giá trị này là mã băm ổn định của blob chứa "hello\n",
	// dùng để chắc chắn thuật toán băm không bị thay đổi ngoài ý muốn.
	got := object.ComputeHash(object.TypeBlob, []byte("hello\n"))
	const want = "ce013625030ba8dba906f756967f9e9ca394464a"
	if got.String() != want {
		t.Fatalf("mã băm blob không đúng:\nnhận  %s\nmong đợi %s", got, want)
	}

	// Tree rỗng có hash 4b825dc642cb6eb9a060e54bf8d69288fbee4904.
	empty := object.ComputeHash(object.TypeTree, (&object.Tree{}).Encode())
	const wantEmpty = "4b825dc642cb6eb9a060e54bf8d69288fbee4904"
	if empty.String() != wantEmpty {
		t.Fatalf("mã băm tree rỗng không đúng:\nnhận  %s\nmong đợi %s", empty, wantEmpty)
	}
}

func TestParseHashVaHienThi(t *testing.T) {
	h := object.ComputeHash(object.TypeBlob, []byte("hello\n"))
	parsed, err := object.ParseHash(h.String())
	if err != nil {
		t.Fatal(err)
	}
	if parsed != h {
		t.Fatalf("parse hash sai")
	}
	if h.Short(7) != h.String()[:7] {
		t.Fatalf("hiển thị ngắn sai")
	}
	if _, err := object.ParseHash("abc"); err == nil {
		t.Fatalf("phải báo lỗi với hash quá ngắn")
	}
	if !object.ZeroHash.IsZero() {
		t.Fatalf("hash rỗng phải báo IsZero")
	}
}

func TestTreeSapXepVaMaHoa(t *testing.T) {
	// Entry thư mục "src" phải được so sánh như "src/" nên
	// "src.txt" phải đứng trước "src" theo quy tắc so sánh có dấu "/" hậu tố.
	tree := &object.Tree{Entries: []object.TreeEntry{
		{Mode: object.ModeBlob, Name: "src.txt"},
		{Mode: object.ModeBlob, Name: "README.md"},
		{Mode: object.ModeTree, Name: "src"},
		{Mode: object.ModeBlob, Name: "a.go"},
	}}
	tree.Sort()
	want := []string{"README.md", "a.go", "src.txt", "src"}
	for i, e := range tree.Entries {
		if e.Name != want[i] {
			t.Fatalf("thứ tự sai tại vị trí %d: %q (mong đợi %q)", i, e.Name, want[i])
		}
	}

	// Giải mã lại phải cho cùng kết quả.
	decoded, err := object.DecodeTree(tree.Encode())
	if err != nil {
		t.Fatal(err)
	}
	if len(decoded.Entries) != 4 || decoded.Entries[0].Name != "README.md" {
		t.Fatalf("giải mã tree sai: %+v", decoded.Entries)
	}
	// Hash của tree phải không đổi sau vòng mã hóa rồi giải mã.
	if object.ComputeHash(object.TypeTree, tree.Encode()) != object.ComputeHash(object.TypeTree, decoded.Encode()) {
		t.Fatalf("hash tree phải giữ nguyên qua vòng mã hóa")
	}
}

func TestTreeUpsertVaRemove(t *testing.T) {
	tree := &object.Tree{}
	tree.Upsert(object.TreeEntry{Mode: object.ModeBlob, Name: "b.txt", Hash: object.ComputeHash(object.TypeBlob, []byte("b"))})
	tree.Upsert(object.TreeEntry{Mode: object.ModeBlob, Name: "a.txt", Hash: object.ComputeHash(object.TypeBlob, []byte("a"))})
	if len(tree.Entries) != 2 || tree.Entries[0].Name != "a.txt" {
		t.Fatalf("upsert phải giữ thứ tự: %+v", tree.Entries)
	}
	if tree.Get("a.txt") == nil {
		t.Fatalf("phải tìm thấy entry vừa thêm")
	}
	tree.Remove("a.txt")
	if tree.Get("a.txt") != nil || len(tree.Entries) != 1 {
		t.Fatalf("remove không hoạt động")
	}
}

func TestCommitEncodeDecode(t *testing.T) {
	id := object.Identity{Name: "Nguyễn Văn A", Email: "a@example.com"}
	treeHash := object.ComputeHash(object.TypeTree, (&object.Tree{}).Encode())
	parent := object.ComputeHash(object.TypeBlob, []byte("p"))

	c := &object.Commit{
		Tree:      treeHash,
		Parents:   []object.Hash{parent},
		Author:    id,
		Committer: id,
		Message:   "tiêu đề\n\nnội dung chi tiết\ndòng hai\n",
	}
	data := c.Encode()

	// Định dạng phải là header rồi dòng trống rồi message.
	wantHeader := "tree " + treeHash.String() + "\n"
	if len(data) < len(wantHeader) || string(data[:len(wantHeader)]) != wantHeader {
		t.Fatalf("header commit sai: %q", string(data[:40]))
	}

	got, err := object.DecodeCommit(data)
	if err != nil {
		t.Fatal(err)
	}
	if got.Tree != treeHash {
		t.Fatalf("tree không khớp sau khi giải mã")
	}
	if len(got.Parents) != 1 || got.Parents[0] != parent {
		t.Fatalf("phụ huynh không khớp: %+v", got.Parents)
	}
	if got.Summary() != "tiêu đề" {
		t.Fatalf("tiêu đề sai: %q", got.Summary())
	}
	if got.Message != c.Message {
		t.Fatalf("message không khớp: %q", got.Message)
	}
	// Giải mã rồi mã hóa lại phải cho kết quả giống hệt.
	if string(got.Encode()) != string(data) {
		t.Fatalf("mã hóa lại không ổn định")
	}
}

func TestCommitThieuTreeBaoLoi(t *testing.T) {
	if _, err := object.DecodeCommit([]byte("author A <a@b> 1 +0700\n\nmsg\n")); err == nil {
		t.Fatalf("phải báo lỗi khi commit không có tree")
	}
}

func TestParseIdentity(t *testing.T) {
	id, err := object.ParseIdentity("Tên Nguyễn <ten@example.com> 1700000000 +0700")
	if err != nil {
		t.Fatal(err)
	}
	if id.Name != "Tên Nguyễn" || id.Email != "ten@example.com" {
		t.Fatalf("thông tin sai: %+v", id)
	}
	if id.When.Unix() != 1700000000 {
		t.Fatalf("thời điểm sai: %d", id.When.Unix())
	}
	if _, err := object.ParseIdentity("không có dấu ngoặc"); err == nil {
		t.Fatalf("phải báo lỗi với chuỗi không hợp lệ")
	}
}

func TestFileMode(t *testing.T) {
	if !object.ModeTree.IsTree() {
		t.Fatalf("ModeTree phải là thư mục")
	}
	if !object.ModeExec.IsExec() {
		t.Fatalf("ModeExec phải là file thực thi")
	}
	if object.ModeBlob.IsTree() {
		t.Fatalf("ModeBlob không phải thư mục")
	}
}
