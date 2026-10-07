// Package merge thực hiện hợp nhất ba phiên bản (base, ours, theirs) theo
// thuật toán diff3: mỗi nhánh được chia thành các dải thay đổi so với base,
// rồi ghép các dải lại với nhau. Những dải hai bên cùng sửa mà cho ra nội dung
// khác nhau sẽ được đánh dấu xung đột.
package merge

import (
	"fmt"
	"strings"

	"github.com/tonguyenducmanh/devcli/internal/vcs/diff"
)

// Result là kết quả hợp nhất.
type Result struct {
	// Lines là nội dung đã hợp nhất, có dấu báo xung đột nếu còn xung đột.
	Lines []string
	// Conflict cho biết còn xung đột chưa được giải quyết.
	Conflict bool
}

// changeOp là một dải thay đổi của một nhánh so với base.
// Dải [Start,End) trong base bị thay bằng Repl.
type changeOp struct {
	Start, End int
	Repl       []string
}

// Merge3 hợp nhất nội dung của base, ours và theirs.
// labelOurs và labelTheirs dùng cho nhãn trong dấu báo xung đột.
func Merge3(base, ours, theirs []string, labelOurs, labelTheirs string) Result {
	// Các trường hợp dễ xử lý trước, tránh chạy thuật toán diff không cần thiết.
	if equalLines(ours, theirs) {
		return Result{Lines: ours}
	}
	if equalLines(base, ours) {
		// Phía chúng ta không đổi, lấy thay đổi từ phía kia.
		return Result{Lines: theirs}
	}
	if equalLines(base, theirs) {
		// Phía kia không đổi, giữ nguyên phía chúng ta.
		return Result{Lines: ours}
	}

	ourOps := changeOps(base, ours)
	theirOps := changeOps(base, theirs)

	out := make([]string, 0, len(base))
	conflict := false

	// pos là vị trí đang xét trong base.
	pos := 0
	oi, ti := 0, 0

	for {
		nextO := len(base) + 1
		if oi < len(ourOps) {
			nextO = ourOps[oi].Start
		}
		nextT := len(base) + 1
		if ti < len(theirOps) {
			nextT = theirOps[ti].Start
		}
		next := nextO
		if nextT < next {
			next = nextT
		}

		// Không còn dải thay đổi nào, chép nốt phần cuối của base.
		if next > len(base) {
			out = append(out, base[pos:]...)
			break
		}
		// Chép đoạn base mà cả hai nhánh đều giữ nguyên.
		if next > pos {
			out = append(out, base[pos:next]...)
			pos = next
		}

		// Gom tất cả dải thay đổi bắt đầu tại vị trí hiện tại,
		// đồng thời mở rộng vùng ảnh hưởng tới hết dải dài nhất.
		var ourGroup, theirGroup []changeOp
		end := pos
		for oi < len(ourOps) && ourOps[oi].Start <= pos {
			ourGroup = append(ourGroup, ourOps[oi])
			if ourOps[oi].End > end {
				end = ourOps[oi].End
			}
			oi++
		}
		for ti < len(theirOps) && theirOps[ti].Start <= pos {
			theirGroup = append(theirGroup, theirOps[ti])
			if theirOps[ti].End > end {
				end = theirOps[ti].End
			}
			ti++
		}
		if end < pos {
			end = pos
		}

		ourText := applyOps(base, ourGroup, pos, end)
		theirText := applyOps(base, theirGroup, pos, end)
		baseText := base[pos:end]

		switch {
		case equalLines(ourText, theirText):
			// Hai nhánh cho ra cùng một kết quả, chấp nhận luôn.
			out = append(out, ourText...)
		case equalLines(baseText, ourText):
			// Chỉ phía kia thay đổi.
			out = append(out, theirText...)
		case equalLines(baseText, theirText):
			// Chỉ phía chúng ta thay đổi.
			out = append(out, ourText...)
		default:
			// Hai bên sửa cùng một chỗ theo hai cách khác nhau.
			conflict = true
			out = append(out, fmt.Sprintf("<<<<<<< %s", labelOurs))
			out = append(out, ourText...)
			out = append(out, "||||||| base")
			out = append(out, baseText...)
			out = append(out, "=======")
			out = append(out, theirText...)
			out = append(out, fmt.Sprintf(">>>>>>> %s", labelTheirs))
		}
		pos = end
	}
	return Result{Lines: out, Conflict: conflict}
}

// changeOps chuyển kết quả diff thành danh sách dải thay đổi đã gộp
// (xoá liền kề chèn thêm thành một dải thay thế duy nhất).
func changeOps(base, other []string) []changeOp {
	var out []changeOp
	for _, c := range diff.DiffLines(base, other) {
		switch c.Op {
		case diff.OpEqual:
			continue
		case diff.OpDelete:
			// Gộp với dải chèn ngay sau để thành một lần thay thế.
			if len(out) > 0 && out[len(out)-1].End == c.AStart && len(out[len(out)-1].Repl) == 0 {
				out[len(out)-1].End = c.AEnd
				continue
			}
			out = append(out, changeOp{Start: c.AStart, End: c.AEnd})
		case diff.OpInsert:
			// Chèn tại vị trí AStart: gộp vào dải liền trước nếu có.
			if len(out) > 0 && out[len(out)-1].End == c.AStart {
				cur := &out[len(out)-1]
				cur.Repl = append(cur.Repl, c.B...)
				if c.AEnd > cur.End {
					cur.End = c.AEnd
				}
				continue
			}
			out = append(out, changeOp{Start: c.AStart, End: c.AEnd, Repl: c.B})
		}
	}
	return out
}

// applyOps dựng nội dung của đoạn base[s:e) sau khi áp các dải thay đổi.
// Các dải trong ops phải đã được lọc để nằm gọn trong [s,e).
func applyOps(base []string, ops []changeOp, s, e int) []string {
	var out []string
	pos := s
	for _, op := range ops {
		if op.Start > pos {
			out = append(out, base[pos:op.Start]...)
		}
		out = append(out, op.Repl...)
		if op.End > pos {
			pos = op.End
		}
	}
	if pos < e {
		out = append(out, base[pos:e]...)
	}
	return out
}

// equalLines so sánh hai danh sách dòng.
func equalLines(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// CleanConflictMarkers kiểm tra nội dung còn dấu xung đột hay không.
func CleanConflictMarkers(lines []string) bool {
	for _, l := range lines {
		t := strings.TrimSpace(l)
		if strings.HasPrefix(t, "<<<<<<<") || strings.HasPrefix(t, ">>>>>>>") ||
			t == "=======" || strings.HasPrefix(t, "|||||||") {
			return false
		}
	}
	return true
}
