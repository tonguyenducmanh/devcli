package diff

import (
	"strings"
	"testing"
)

// lines tách nội dung đa dòng thành danh sách dòng.
func lines(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(strings.TrimSuffix(s, "\n"), "\n")
}

// text nối lại danh sách dòng thành nội dung.
func text(l []string) string {
	if len(l) == 0 {
		return ""
	}
	return strings.Join(l, "\n") + "\n"
}

// applyChanges áp dụng kết quả diff lên chuỗi a để dựng lại chuỗi b.
func applyChanges(a []string, changes []Change) []string {
	var out []string
	pos := 0
	for _, c := range changes {
		switch c.Op {
		case OpEqual:
			out = append(out, a[pos:c.AEnd]...)
			pos = c.AEnd
		case OpDelete:
			pos = c.AEnd
		case OpInsert:
			out = append(out, c.B...)
		}
	}
	return append(out, a[pos:]...)
}

func TestDiffLinesDungLuon(t *testing.T) {
	cases := []struct {
		name string
		a, b string
	}{
		{"không đổi", "a\nb\nc\n", "a\nb\nc\n"},
		{"sửa một dòng", "a\nb\nc\n", "a\nB\nc\n"},
		{"thêm dòng ở giữa", "a\nc\n", "a\nb\nc\n"},
		{"xoá dòng ở giữa", "a\nb\nc\n", "a\nc\n"},
		{"xoá toàn bộ", "a\nb\n", ""},
		{"tạo mới toàn bộ", "", "a\nb\n"},
		{"hoán đổi hai dòng", "a\nb\nc\nd\n", "c\nd\na\nb\n"},
		{"sửa liên tiếp nhiều dòng", "1\n2\n3\n4\n5\n", "1\nX\nY\n4\n5\n"},
		{"thêm ở đầu", "b\nc\n", "a\nb\nc\n"},
		{"xoá ở đầu", "a\nb\nc\n", "b\nc\n"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			a := lines(tc.a)
			b := lines(tc.b)
			changes := DiffLines(a, b)
			if got := text(applyChanges(a, changes)); got != tc.b {
				t.Fatalf("áp dụng diff sai:\n--- mong đợi ---\n%q\n--- nhận được ---\n%q", tc.b, got)
			}
			// Các thay đổi phải phủ kín đúng các vị trí khác nhau,
			// không được trùng lặp hay chừa sót.
			posA, posB := 0, 0
			for _, c := range changes {
				if c.AStart != posA {
					t.Fatalf("vị trí A không khớp: mong đợi %d, nhận %d (changes=%+v)", posA, c.AStart, changes)
				}
				if c.BStart != posB {
					t.Fatalf("vị trí B không khớp: mong đợi %d, nhận %d", posB, c.BStart)
				}
				posA, posB = c.AEnd, c.BEnd
			}
			if posA != len(a) || posB != len(b) {
				t.Fatalf("diff chưa phủ hết: A còn %d/%d, B còn %d/%d", posA, len(a), posB, len(b))
			}
		})
	}
}

func TestDiffLinesCacDauRaDauBang(t *testing.T) {
	a := []string{"a", "b", "c", "d", "e"}
	b := []string{"a", "X", "c", "d", "E"}
	changes := DiffLines(a, b)
	var adds, dels int
	for _, c := range changes {
		switch c.Op {
		case OpInsert:
			adds += c.BEnd - c.BStart
		case OpDelete:
			dels += c.AEnd - c.AStart
		}
	}
	if adds != 2 || dels != 2 {
		t.Fatalf("thống kê sai: thêm %d, xoá %d (mong đợi 2 và 2)", adds, dels)
	}
}

func TestHunksWithContext(t *testing.T) {
	a := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}
	b := []string{"1", "2", "3", "4", "X", "6", "7", "8", "9", "Y"}
	hunks := HunksWithContext(a, b, 1)
	// Hai thay đổi cách nhau 4 dòng, với ngữ cảnh 1 dòng thì tách thành 2 hunk.
	if len(hunks) != 2 {
		t.Fatalf("mong đợi 2 hunk, nhận được %d: %+v", len(hunks), hunks)
	}
	// Mỗi hunk gồm 1 dòng ngữ cảnh, 1 dòng xoá và 1 dòng thêm.
	wantA := []int{3, 2}
	wantB := []int{3, 2}
	for i, h := range hunks {
		if h.ACount != wantA[i] || h.BCount != wantB[i] {
			t.Fatalf("hunk %d sai số dòng: A=%d/%d B=%d/%d (%+v)", i, h.ACount, wantA[i], h.BCount, wantB[i], h)
		}
	}
}

func TestHunksGopKhiKhoangCachNho(t *testing.T) {
	a := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}
	// Hai thay đổi cách nhau 2 dòng, ngữ cảnh 1 dòng nên gộp làm một hunk.
	b := []string{"1", "X", "3", "4", "Y", "6", "7", "8", "9", "10"}
	if hunks := HunksWithContext(a, b, 1); len(hunks) != 1 {
		t.Fatalf("khoảng cách ngắn phải gộp thành 1 hunk, nhận %d: %+v", len(hunks), hunks)
	}
	// Khoảng cách lớn hơn 2 lần ngữ cảnh thì phải tách.
	c := []string{"1", "X", "3", "4", "5", "6", "7", "Y", "9", "10"}
	if hunks := HunksWithContext(a, c, 1); len(hunks) != 2 {
		t.Fatalf("khoảng cách lớn phải tách thành 2 hunk, nhận %d: %+v", len(hunks), hunks)
	}
}

func TestHunksKhongCoThayDoi(t *testing.T) {
	a := []string{"a", "b"}
	if hunks := HunksWithContext(a, a, 3); len(hunks) != 0 {
		t.Fatalf("không có thay đổi thì không được có hunk, nhận %d", len(hunks))
	}
}

func TestIsBinary(t *testing.T) {
	if IsBinary([]byte("văn bản thường")) {
		t.Fatalf("văn bản thường không phải file nhị phân")
	}
	if !IsBinary([]byte{'a', 0, 'b'}) {
		t.Fatalf("dữ liệu chứa byte 0 phải được coi là nhị phân")
	}
}

func TestFormatUnified(t *testing.T) {
	a := []string{"1", "2", "3"}
	b := []string{"1", "X", "3"}
	out := FormatUnified(a, b, 1)
	if !strings.Contains(out, "@@") || !strings.Contains(out, "+X") {
		t.Fatalf("diff dạng thống nhất thiếu nội dung:\n%s", out)
	}
}
