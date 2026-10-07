package merge_test

import (
	"github.com/tonguyenducmanh/devcli/internal/vcs/merge"
	"strings"
	"testing"
)

// lines tách nội dung đa dòng thành danh sách dòng.
func lines(s string) []string {
	if s == "" {
		return nil
	}
	s = strings.TrimSuffix(s, "\n")
	return strings.Split(s, "\n")
}

// text nối lại danh sách dòng thành nội dung có xuống dòng cuối.
func text(l []string) string {
	if len(l) == 0 {
		return ""
	}
	return strings.Join(l, "\n") + "\n"
}

func TestMerge3BasicCases(t *testing.T) {
	cases := []struct {
		name                     string
		base, ours, theirs, want string
	}{
		{
			name:   "một phía không đổi",
			base:   "a\nb\nc\n",
			ours:   "a\nb\nc\n",
			theirs: "a\nX\nc\n",
			want:   "a\nX\nc\n",
		},
		{
			name:   "hai phía sửa hai chỗ khác nhau",
			base:   "a\nb\nc\nd\ne\n",
			ours:   "a\nB\nc\nd\ne\n",
			theirs: "a\nb\nc\nd\nE\n",
			want:   "a\nB\nc\nd\nE\n",
		},
		{
			name:   "hai phía sửa cùng một chỗ giống nhau",
			base:   "a\nb\nc\n",
			ours:   "a\nX\nc\n",
			theirs: "a\nX\nc\n",
			want:   "a\nX\nc\n",
		},
		{
			name:   "chèn thêm ở hai vị trí khác nhau",
			base:   "a\nc\n",
			ours:   "a\nb\nc\n",
			theirs: "a\nc\nd\n",
			want:   "a\nb\nc\nd\n",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := merge.Merge3(lines(tc.base), lines(tc.ours), lines(tc.theirs), "ours", "theirs")
			if got.Conflict {
				t.Fatalf("không mong đợi xung đột, nhận được:\n%s", text(got.Lines))
			}
			if text(got.Lines) != tc.want {
				t.Fatalf("kết quả sai:\n--- mong đợi ---\n%s--- nhận được ---\n%s", tc.want, text(got.Lines))
			}
		})
	}
}

func TestMerge3ReportsConflict(t *testing.T) {
	base := lines("a\nb\nc\nd\ne\n")
	ours := lines("a\nB-cua-ta\nc\nd\ne\n")
	theirs := lines("a\nB-cua-han\nc\nd\ne\n")

	got := merge.Merge3(base, ours, theirs, "nhanh", "tinh_nay")
	if !got.Conflict {
		t.Fatalf("phải báo xung đột, nhận được:\n%s", text(got.Lines))
	}
	// Kiểm tra dấu báo xung đột có đủ ba vùng.
	out := text(got.Lines)
	for _, marker := range []string{"<<<<<<< nhanh", "||||||| base", "=======", ">>>>>>> tinh_nay"} {
		if !strings.Contains(out, marker) {
			t.Fatalf("thiếu dấu %q trong kết quả:\n%s", marker, out)
		}
	}
	// Hàm phát hiện phải trả về false khi nội dung còn dấu xung đột.
	if merge.CleanConflictMarkers(got.Lines) {
		t.Fatalf("hàm phát hiện dấu xung đột hoạt động sai")
	}
	if !merge.CleanConflictMarkers(lines("a\nb\n")) {
		t.Fatalf("nội dung sạch phải được coi là không có xung đột")
	}
}

func TestMerge3EditsSameSide(t *testing.T) {
	base := lines("a\nb\nc\n")
	ours := lines("a\nb\nc\nd\n")
	theirs := lines("a\nb\nc\nd\n")
	got := merge.Merge3(base, ours, theirs, "o", "t")
	if got.Conflict {
		t.Fatalf("hai phía giống nhau thì không có xung đột")
	}
	if text(got.Lines) != "a\nb\nc\nd\n" {
		t.Fatalf("sai kết quả: %q", text(got.Lines))
	}
}

func TestMerge3EmptyFile(t *testing.T) {
	// File bên phía chúng ta rỗng, bên kia có nội dung.
	got := merge.Merge3(nil, nil, lines("x\ny\n"), "o", "t")
	if got.Conflict || text(got.Lines) != "x\ny\n" {
		t.Fatalf("sai kết quả: conflict=%v, lines=%q", got.Conflict, text(got.Lines))
	}
}
