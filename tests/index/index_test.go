package index_test

import (
	"github.com/tonguyenducmanh/devcli/internal/vcs/index"
	"os"
	"path/filepath"
	"testing"

	"github.com/tonguyenducmanh/devcli/internal/vcs/object"
)

// fixture là một index kèm đường dẫn tệp của nó.
// Index giữ đường dẫn ở trường nội bộ nên không lấy được từ bên ngoài,
// bọc lại ở đây để kiểm thử mở lại được sau khi lưu.
type fixture struct {
	idx  *index.Index
	path string
}

// newIndex tạo một index rỗng nằm trong thư mục tạm.
func newIndex(t *testing.T) fixture {
	t.Helper()
	path := filepath.Join(t.TempDir(), "index")
	idx, err := index.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	return fixture{idx: idx, path: path}
}

// sampleEntry tạo một entry mẫu với đường dẫn cho trước.
func sampleEntry(name string) index.Entry {
	return index.Entry{
		Mode: object.ModeBlob,
		Size: 42,
		Hash: object.ComputeHash(object.TypeBlob, []byte(name)),
		Name: name,
	}
}

func TestIndexRongVaPhieuMaGhiDoc(t *testing.T) {
	fx := newIndex(t)
	if fx.idx.Len() != 0 {
		t.Fatalf("index mới phải rỗng")
	}
	// Ghi rồi đọc lại phải cho index rỗng.
	if err := fx.idx.Save(); err != nil {
		t.Fatal(err)
	}
	got, err := index.Open(fx.path)
	if err != nil {
		t.Fatal(err)
	}
	if got.Len() != 0 {
		t.Fatalf("đọc lại index rỗng phải không có entry nào")
	}
}

func TestIndexGhiDocNhieuDuongDan(t *testing.T) {
	fx := newIndex(t)
	// Thêm theo thứ tự không sắp xếp để kiểm tra logic sắp xếp khi ghi.
	for _, name := range []string{"z.txt", "a.txt", "dir/b.txt", "dir/a.txt", "m.txt"} {
		fx.idx.Add(sampleEntry(name))
	}
	if err := fx.idx.Save(); err != nil {
		t.Fatal(err)
	}

	got, err := index.Open(fx.path)
	if err != nil {
		t.Fatalf("đọc index lỗi: %v", err)
	}
	want := []string{"a.txt", "dir/a.txt", "dir/b.txt", "m.txt", "z.txt"}
	entries := got.Entries()
	if len(entries) != len(want) {
		t.Fatalf("số lượng entry sai: %d (mong đợi %d)", len(entries), len(want))
	}
	for i, e := range entries {
		if e.Name != want[i] {
			t.Fatalf("thứ tự sai tại %d: %q (mong đợi %q)", i, e.Name, want[i])
		}
		if e.Hash != sampleEntry(e.Name).Hash {
			t.Fatalf("hash của %s không khớp sau khi đọc lại", e.Name)
		}
		if e.Mode != object.ModeBlob {
			t.Fatalf("chế độ file của %s không khớp: %s", e.Name, e.Mode)
		}
		if e.Size != 42 {
			t.Fatalf("kích thước của %s không khớp: %d", e.Name, e.Size)
		}
	}
}

func TestIndexNhieuStageChoMotDuongDan(t *testing.T) {
	fx := newIndex(t)
	fx.idx.Add(sampleEntry("x.txt"))
	// Ghi thêm các stage của xung đột.
	for _, st := range []index.Stage{index.StageBase, index.StageOurs, index.StageTheirs} {
		e := sampleEntry("x.txt")
		e.Stage = st
		e.Hash = object.ComputeHash(object.TypeBlob, []byte{byte(st)})
		fx.idx.Set(e)
	}
	if err := fx.idx.Save(); err != nil {
		t.Fatal(err)
	}

	got, err := index.Open(fx.path)
	if err != nil {
		t.Fatal(err)
	}
	if !got.HasConflicts() {
		t.Fatalf("phải nhận diện được xung đột")
	}
	if names := got.Conflicts(); len(names) != 1 || names[0] != "x.txt" {
		t.Fatalf("danh sách xung đột sai: %v", names)
	}
	stages := got.Stages("x.txt")
	if len(stages) != 4 {
		t.Fatalf("phải có 4 stage cho một đường dẫn, nhận %d", len(stages))
	}

	// Xóa các stage khác 0 sau khi giải quyết xung đột.
	got.RemoveStages("x.txt")
	if got.HasConflicts() {
		t.Fatalf("sau khi xóa stage phải hết xung đột")
	}
	if got.Get("x.txt") == nil {
		t.Fatalf("entry stage 0 phải được giữ lại")
	}
}

func TestIndexLayVaXoa(t *testing.T) {
	fx := newIndex(t)
	fx.idx.Add(sampleEntry("a.txt"))
	fx.idx.Add(sampleEntry("b.txt"))

	if fx.idx.Get("a.txt") == nil {
		t.Fatalf("phải lấy được entry theo tên")
	}
	fx.idx.Remove("a.txt")
	if fx.idx.Get("a.txt") != nil {
		t.Fatalf("entry đã xóa không được tìm thấy")
	}
	if paths := fx.idx.Paths(); len(paths) != 1 || paths[0] != "b.txt" {
		t.Fatalf("danh sách đường dẫn sai: %v", paths)
	}

	fx.idx.Clear()
	if fx.idx.Len() != 0 {
		t.Fatalf("clear phải làm index rỗng")
	}
}

func TestIndexCapNhatEntryCu(t *testing.T) {
	fx := newIndex(t)
	e := sampleEntry("a.txt")
	fx.idx.Add(e)

	// Stage file với nội dung mới phải thay thế entry cũ, không nhân bản.
	e2 := sampleEntry("a.txt")
	e2.Hash = object.ComputeHash(object.TypeBlob, []byte("nội dung mới"))
	e2.Size = 99
	fx.idx.Add(e2)

	entries := fx.idx.Entries()
	if len(entries) != 1 {
		t.Fatalf("phải chỉ có 1 entry, nhận %d", len(entries))
	}
	if entries[0].Hash != e2.Hash || entries[0].Size != 99 {
		t.Fatalf("entry chưa được cập nhật: %+v", entries[0])
	}
}

func TestIndexCheDoFileKhacNhau(t *testing.T) {
	fx := newIndex(t)
	for name, mode := range map[string]object.FileMode{
		"bin":        object.ModeExec,
		"link":       object.ModeSymlink,
		"van-ban.md": object.ModeBlob,
	} {
		e := sampleEntry(name)
		e.Mode = mode
		fx.idx.Add(e)
	}
	if err := fx.idx.Save(); err != nil {
		t.Fatal(err)
	}
	got, err := index.Open(fx.path)
	if err != nil {
		t.Fatal(err)
	}
	modes := map[string]object.FileMode{}
	for _, e := range got.Entries() {
		modes[e.Name] = e.Mode
	}
	if modes["bin"] != object.ModeExec {
		t.Fatalf("chế độ file thực thi sai: %s", modes["bin"])
	}
	if modes["link"] != object.ModeSymlink {
		t.Fatalf("chế độ liên kết tượng trưng sai: %s", modes["link"])
	}
	if modes["van-ban.md"] != object.ModeBlob {
		t.Fatalf("chế độ file thường sai: %s", modes["van-ban.md"])
	}
}

func TestIndexFileHong(t *testing.T) {
	path := filepath.Join(t.TempDir(), "index")
	// Không có chữ ký DIRC thì phải báo lỗi.
	if err := os.WriteFile(path, []byte("XXXX\x00\x00\x00\x02\x00\x00\x00\x00"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := index.Open(path); err == nil {
		t.Fatalf("phải báo lỗi với file index không hợp lệ")
	}
}

func TestIndexTenRatDai(t *testing.T) {
	// Đường dẫn dài hơn giới hạn 0xFFF phải được cắt bớt ở cờ nhưng
	// vẫn không làm hỏng việc đọc lại các đường dẫn bình thường.
	fx := newIndex(t)
	long := ""
	for len(long) < 500 {
		long += "thư-mục-con/"
	}
	fx.idx.Add(sampleEntry(long + "tên-tệp-rất-dài.txt"))
	if err := fx.idx.Save(); err != nil {
		t.Fatal(err)
	}
	got, err := index.Open(fx.path)
	if err != nil {
		t.Fatalf("đọc index với đường dẫn dài phải không lỗi: %v", err)
	}
	if got.Len() != 1 {
		t.Fatalf("phải có 1 entry, nhận %d", got.Len())
	}
}

func TestIndexWalk(t *testing.T) {
	fx := newIndex(t)
	fx.idx.Add(sampleEntry("a.txt"))
	fx.idx.Add(sampleEntry("b.txt"))

	seen := map[string]bool{}
	fx.idx.Walk(func(e index.Entry) bool {
		seen[e.Name] = true
		return true
	})
	if len(seen) != 2 {
		t.Fatalf("duyệt index thiếu entry: %v", seen)
	}

	// Dừng sớm khi hàm trả về false.
	count := 0
	fx.idx.Walk(func(index.Entry) bool {
		count++
		return false
	})
	if count != 1 {
		t.Fatalf("phải dừng sau phần tử đầu tiên, đếm %d", count)
	}
}
