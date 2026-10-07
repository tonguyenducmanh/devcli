package storage

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/tonguyenducmanh/devcli/internal/vcs/object"
)

// newStore tạo object store trong thư mục tạm.
func newStore(t *testing.T) *ObjectStore {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "objects")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	return NewObjectStore(dir)
}

func TestGhiVaDocBlob(t *testing.T) {
	s := newStore(t)
	content := []byte("nội dung thử nghiệm\n")

	h, err := s.WriteBlob(content)
	if err != nil {
		t.Fatal(err)
	}
	// Mã băm phải khớp với cách ComputeHash tính trực tiếp.
	if want := object.ComputeHash(object.TypeBlob, content); h != want {
		t.Fatalf("hash không đúng: nhận %s, mong đợi %s", h, want)
	}
	if !s.Has(h) {
		t.Fatalf("object phải tồn tại sau khi ghi")
	}
	typ, data, err := s.Read(h)
	if err != nil {
		t.Fatal(err)
	}
	if typ != object.TypeBlob {
		t.Fatalf("sai loại object: %s", typ)
	}
	if !bytes.Equal(data, content) {
		t.Fatalf("nội dung đọc lại khác: %q", data)
	}
}

func TestGhiHaiLanKhongHongFile(t *testing.T) {
	s := newStore(t)
	content := []byte("abc")
	h1, err := s.WriteBlob(content)
	if err != nil {
		t.Fatal(err)
	}
	h2, err := s.WriteBlob(content)
	if err != nil {
		t.Fatal(err)
	}
	if h1 != h2 {
		t.Fatalf("cùng nội dung phải cho cùng hash")
	}
	// Đếm số object trong kho.
	count := 0
	if err := s.Each(func(object.Hash) error {
		count++
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("kho phải có đúng 1 object, nhận %d", count)
	}
}

func TestDocObjectKhongTonTai(t *testing.T) {
	s := newStore(t)
	if _, _, err := s.Read(object.ComputeHash(object.TypeBlob, []byte("x"))); err == nil {
		t.Fatalf("phải báo lỗi khi đọc object không tồn tại")
	}
}

func TestDocTreeVaCommit(t *testing.T) {
	s := newStore(t)
	blobHash, err := s.WriteBlob([]byte("dữ liệu"))
	if err != nil {
		t.Fatal(err)
	}
	tree := &object.Tree{Entries: []object.TreeEntry{
		{Mode: object.ModeBlob, Name: "a.txt", Hash: blobHash},
	}}
	treeHash, err := s.WriteTree(tree)
	if err != nil {
		t.Fatal(err)
	}
	got, err := s.ReadTree(treeHash)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Entries) != 1 || got.Entries[0].Name != "a.txt" {
		t.Fatalf("cây đọc lại không khớp: %+v", got.Entries)
	}

	id := object.Identity{Name: "Tên", Email: "ten@example.com"}
	c := &object.Commit{
		Tree:      treeHash,
		Parents:   []object.Hash{treeHash},
		Author:    id,
		Committer: id,
		Message:   "tin nhắn thử\n\nphần mô tả\n",
	}
	commitHash, err := s.WriteCommit(c)
	if err != nil {
		t.Fatal(err)
	}
	gotCommit, err := s.ReadCommit(commitHash)
	if err != nil {
		t.Fatal(err)
	}
	if gotCommit.Tree != treeHash {
		t.Fatalf("tree của commit sai")
	}
	if gotCommit.Summary() != "tin nhắn thử" {
		t.Fatalf("tiêu đề commit sai: %q", gotCommit.Summary())
	}
	if gotCommit.Author.Name != "Tên" || gotCommit.Author.Email != "ten@example.com" {
		t.Fatalf("thông tin tác giả sai: %+v", gotCommit.Author)
	}
}

func TestRefDocVaGhi(t *testing.T) {
	dir := t.TempDir()
	r := NewRefStore(dir)
	hash := object.ComputeHash(object.TypeBlob, []byte("x"))

	if _, err := r.Resolve("refs/heads/main"); err != ErrRefNotFound {
		t.Fatalf("phải báo ref chưa tồn tại")
	}
	if err := r.Write("refs/heads/main", hash); err != nil {
		t.Fatal(err)
	}
	got, err := r.Resolve("refs/heads/main")
	if err != nil {
		t.Fatal(err)
	}
	if got != hash {
		t.Fatalf("hash không khớp")
	}

	// Ref gián tiếp HEAD trỏ tới nhánh.
	if err := r.WriteSymbolic("HEAD", "refs/heads/main"); err != nil {
		t.Fatal(err)
	}
	if got, err := r.Resolve("HEAD"); err != nil || got != hash {
		t.Fatalf("HEAD gián tiếp phải phân giải được: %v %v", got, err)
	}

	refs, err := r.List("refs/heads/")
	if err != nil {
		t.Fatal(err)
	}
	if len(refs) != 1 || refs[0] != "refs/heads/main" {
		t.Fatalf("danh sách ref sai: %v", refs)
	}

	// Xóa ref phải dọn luôn thư mục rỗng.
	if err := r.Remove("refs/heads/main"); err != nil {
		t.Fatal(err)
	}
	if r.Exists("refs/heads/main") {
		t.Fatalf("ref phải bị xóa")
	}
	if _, err := os.Stat(filepath.Join(dir, "refs", "heads")); !os.IsNotExist(err) {
		t.Fatalf("thư mục rỗng phải được dọn")
	}
}

func TestReflog(t *testing.T) {
	dir := t.TempDir()
	r := NewRefStore(dir)
	h1 := object.ComputeHash(object.TypeBlob, []byte("1"))
	h2 := object.ComputeHash(object.TypeBlob, []byte("2"))

	if _, err := r.ReadReflog("refs/heads/main"); err != ErrReflogNotFound {
		t.Fatalf("phải báo reflog chưa tồn tại")
	}
	if err := r.AppendReflog("refs/heads/main", h1, h2, "commit: thử"); err != nil {
		t.Fatal(err)
	}
	entries, err := r.ReadReflog("refs/heads/main")
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("phải có 1 dòng reflog, nhận %d", len(entries))
	}
	if entries[0].New != h2 || entries[0].Old != h1 {
		t.Fatalf("nội dung reflog sai: %+v", entries[0])
	}
	if entries[0].Message != "commit: thử" {
		t.Fatalf("message reflog sai: %q", entries[0].Message)
	}
}

func TestObjectRong(t *testing.T) {
	s := newStore(t)
	// Tree rỗng vẫn phải ghi và đọc được.
	h, err := s.WriteTree(&object.Tree{})
	if err != nil {
		t.Fatal(err)
	}
	tree, err := s.ReadTree(h)
	if err != nil {
		t.Fatal(err)
	}
	if len(tree.Entries) != 0 {
		t.Fatalf("tree rỗng phải không có entry")
	}
}
