// Package diff tính khác biệt giữa hai chuỗi nội dung theo thuật toán Myers,
// phục vụ cho hiển thị diff và hỗ trợ three-way merge.
package diff

import (
	"fmt"
	"strings"
)

// Op là một thao tác trong kết quả diff.
type Op uint8

// Các loại thao tác.
const (
	OpEqual Op = iota
	OpInsert
	OpDelete
)

// Change mô tả một dải thay đổi: A[AStart:AEnd] bị thay bằng B[BStart:BEnd].
// Với OpInsert, AStart == AEnd (chèn thêm), với OpDelete, BStart == BEnd.
type Change struct {
	Op           Op
	AStart, AEnd int
	BStart, BEnd int
	A, B         []string
}

// DiffLines tính danh sách thay đổi giữa a và b bằng thuật toán Myers O(ND).
func DiffLines(a, b []string) []Change {
	// Bỏ qua phần chung ở đầu để thu nhỏ bài toán cho thuật toán Myers.
	start := 0
	for start < len(a) && start < len(b) && a[start] == b[start] {
		start++
	}
	// Bỏ luôn phần chung ở cuối vì cả hai chuỗi đều giống nhau ở đó.
	endA, endB := len(a), len(b)
	for endA > start && endB > start && a[endA-1] == b[endB-1] {
		endA--
		endB--
	}

	// Chạy Myers trên phần còn lại, kết quả trả về vị trí tương đối.
	script := myers(a[start:endA], b[start:endB])

	// Dịch vị trí tương đối về vị trí tuyệt đối trong hai chuỗi gốc.
	var changes []Change

	// Phần chung ở đầu được biểu diễn lại để danh sách thay đổi
	// phủ đầy đủ từ vị trí 0, thuận tiện cho việc áp dụng tuần tự.
	if start > 0 {
		changes = append(changes, Change{
			Op:     OpEqual,
			AStart: 0, AEnd: start,
			BStart: 0, BEnd: start,
			A: a[0:start], B: b[0:start],
		})
	}
	// posA và posB là vị trí đã xét xong trên mỗi chuỗi.
	// Nhờ vậy các thay đổi luôn tiến đều, kể cả khi phần chèn được
	// đặt ngay sau một phần xoá tại cùng vị trí.
	posA, posB := start, start
	for _, op := range script {
		switch op.op {
		case OpEqual:
			endA := start + op.a2
			endB := start + op.b2
			changes = append(changes, Change{
				Op: OpEqual, AStart: posA, AEnd: endA,
				BStart: posB, BEnd: endB,
				A: a[posA:endA], B: b[posB:endB],
			})
			posA, posB = endA, endB
		case OpDelete:
			endA := start + op.a2
			changes = append(changes, Change{
				Op: OpDelete, AStart: posA, AEnd: endA,
				BStart: posB, BEnd: posB,
				A: a[posA:endA],
			})
			posA = endA
		case OpInsert:
			endB := start + op.b2
			changes = append(changes, Change{
				Op: OpInsert, AStart: posA, AEnd: posA,
				BStart: posB, BEnd: endB,
				B: b[posB:endB],
			})
			posB = endB
		}
	}

	// Phần chung ở cuối cũng được thêm vào để phủ hết hai chuỗi.
	if endA < len(a) {
		changes = append(changes, Change{
			Op:     OpEqual,
			AStart: endA, AEnd: len(a),
			BStart: endB, BEnd: len(b),
			A: a[endA:], B: b[endB:],
		})
	}
	return changes
}

// split là một cặp vị trí trong lcs.
type split struct {
	op     Op
	a1, a2 int
	b1, b2 int
}

// myers chạy thuật toán Myers với backtrack để trả về script thay đổi.
// Mọi vị trí trả về đều tương đối so với a và b.
func myers(a, b []string) []split {
	n, m := len(a), len(b)
	if n == 0 && m == 0 {
		return nil
	}
	if n == 0 {
		return []split{{op: OpInsert, b1: 0, b2: m}}
	}
	if m == 0 {
		return []split{{op: OpDelete, a1: 0, a2: n}}
	}

	// Mảng V theo độ lệch d; V[d+1] là vị trí x hiện tại trên đường d.
	maxD := n + m
	v := make([]int, 2*maxD+2)
	trace := make([][]int, 0, maxD+1)

	offset := maxD + 1
	v[offset+1] = 0
	for d := 0; d <= maxD; d++ {
		snapshot := make([]int, len(v))
		copy(snapshot, v)
		trace = append(trace, snapshot)

		for k := -d; k <= d; k += 2 {
			var x int
			if k == -d || (k != d && v[offset+k-1] < v[offset+k+1]) {
				x = v[offset+k+1] // chèn
			} else {
				x = v[offset+k-1] + 1 // xóa
			}
			y := x - k
			for x < n && y < m && a[x] == b[y] {
				x++
				y++
			}
			v[offset+k] = x
			if x >= n && y >= m {
				return backtrack(trace, offset, n, m)
			}
		}
	}
	return []split{{op: OpDelete, a1: 0, a2: n}, {op: OpInsert, b1: 0, b2: m}}
}

// backtrack truy ngược đường đi để dựng script thay đổi.
// n và m là độ dài của hai chuỗi đã dùng khi chạy Myers.
func backtrack(trace [][]int, offset, n, m int) []split {
	var rev []split
	x, y := n, m
	for d := len(trace) - 1; d > 0; d-- {
		v := trace[d]
		k := x - y
		var prevK int
		if k == -d || (k != d && v[offset+k-1] < v[offset+k+1]) {
			prevK = k + 1
		} else {
			prevK = k - 1
		}
		prevX := v[offset+prevK]
		prevY := prevX - prevK

		// Đoạn giữa prevX..x là phần bằng nhau (snake).
		for x > prevX && y > prevY {
			x--
			y--
			rev = append(rev, split{op: OpEqual, a1: x, a2: x + 1, b1: y, b2: y + 1})
		}
		if x == prevX {
			rev = append(rev, split{op: OpInsert, b1: prevY, b2: y})
		} else {
			rev = append(rev, split{op: OpDelete, a1: prevX, a2: x})
		}
		x, y = prevX, prevY
	}
	for x > 0 && y > 0 {
		x--
		y--
		rev = append(rev, split{op: OpEqual, a1: x, a2: x + 1, b1: y, b2: y + 1})
	}
	for x > 0 {
		x--
		rev = append(rev, split{op: OpDelete, a1: x, a2: x + 1})
	}
	for y > 0 {
		y--
		rev = append(rev, split{op: OpInsert, b1: y, b2: y + 1})
	}
	// Đảo ngược để có thứ tự tăng dần.
	out := make([]split, 0, len(rev))
	for i := len(rev) - 1; i >= 0; i-- {
		out = append(out, rev[i])
	}
	return out
}

// SplitLines tách nội dung thành các dòng, giữ ký tự xuống dòng.
func SplitLines(s string) []string {
	if s == "" {
		return nil
	}
	s = strings.ReplaceAll(s, "\r\n", "\n")
	if strings.HasSuffix(s, "\n") {
		s = s[:len(s)-1]
	}
	return strings.Split(s, "\n")
}

// JoinLines ngược lại, thêm xuống dòng cuối nếu có nội dung.
func JoinLines(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n") + "\n"
}

// Hunk là một khối diff kèm ngữ cảnh.
type Hunk struct {
	AStart, ACount int
	BStart, BCount int
	Lines          []string // dòng đã gắn tiềp tố + - hoặc khoảng trắng
}

// Row là một dòng trong bảng căn dòng hai phiên bản.
type Row struct {
	// Op là loại thao tác: bằng nhau, thêm mới hoặc xoá.
	Op Op
	// A và B là nội dung dòng tương ứng, có thể rỗng với một phía.
	A, B string
	// AIdx và BIdx là vị trí dòng (bắt đầu từ 0) trong hai chuỗi gốc.
	AIdx, BIdx int
}

// alignRows dựng bảng căn dòng từ kết quả diff, mỗi phần tử là một dòng.
func alignRows(a, b []string) []Row {
	var rows []Row
	for _, c := range DiffLines(a, b) {
		switch c.Op {
		case OpEqual:
			for i := range c.A {
				rows = append(rows, Row{
					Op: OpEqual, A: c.A[i], B: c.B[i],
					AIdx: c.AStart + i, BIdx: c.BStart + i,
				})
			}
		case OpDelete:
			for i := range c.A {
				rows = append(rows, Row{
					Op: OpDelete, A: c.A[i],
					AIdx: c.AStart + i, BIdx: c.BStart,
				})
			}
		case OpInsert:
			for i := range c.B {
				rows = append(rows, Row{
					Op: OpInsert, B: c.B[i],
					AIdx: c.AEnd, BIdx: c.BStart + i,
				})
			}
		}
	}
	return rows
}

// HunksWithContext gom các thay đổi thành hunk kèm số dòng ngữ cảnh.
func HunksWithContext(a, b []string, context int) []Hunk {
	return hunksWithContext(a, b, context)
}

// span là một khoảng trên bảng căn dòng, gồm chỉ số bắt đầu và kết thúc (không bao gồm).
type span struct {
	start, end int
}

// changeSpans gom các dòng thay đổi thành từng cụm liên tiếp, chưa thêm ngữ cảnh.
func changeSpans(rows []Row) []span {
	var out []span
	for i := 0; i < len(rows); {
		if rows[i].Op == OpEqual {
			i++
			continue
		}
		start := i
		for i < len(rows) && rows[i].Op != OpEqual {
			i++
		}
		out = append(out, span{start: start, end: i})
	}
	return out
}

// mergeSpans gộp hai cụm thay đổi khi khoảng cách giữa chúng không vượt quá
// ngưỡng cho trước, mặc định gấp đôi số dòng ngữ cảnh để output gọn.
func mergeSpans(spans []span, maxGap int) []span {
	var out []span
	for _, s := range spans {
		if len(out) == 0 {
			out = append(out, s)
			continue
		}
		last := &out[len(out)-1]
		// Số dòng bằng nhau nằm giữa hai cụm.
		gap := s.start - last.end
		if gap <= maxGap {
			last.end = s.end
			continue
		}
		out = append(out, s)
	}
	return out
}

// hunksWithContext là phần hiện thực, thao tác trên bảng căn dòng.
func hunksWithContext(a, b []string, context int) []Hunk {
	if context < 0 {
		context = 0
	}
	rows := alignRows(a, b)
	spans := mergeSpans(changeSpans(rows), 2*context)

	hunks := make([]Hunk, 0, len(spans))
	for _, s := range spans {
		// Mở rộng cụm thay đổi về hai bên bằng số dòng ngữ cảnh.
		start := s.start
		lead := 0
		for start > 0 && lead < context && rows[start-1].Op == OpEqual {
			start--
			lead++
		}
		end := s.end
		trail := 0
		for end < len(rows) && trail < context && rows[end].Op == OpEqual {
			end++
			trail++
		}

		h := Hunk{
			AStart: rows[start].AIdx,
			BStart: rows[start].BIdx,
		}
		for _, r := range rows[start:end] {
			switch r.Op {
			case OpEqual:
				h.ACount++
				h.BCount++
				h.Lines = append(h.Lines, " "+r.A)
			case OpDelete:
				h.ACount++
				h.Lines = append(h.Lines, "-"+r.A)
			case OpInsert:
				h.BCount++
				h.Lines = append(h.Lines, "+"+r.B)
			}
		}
		hunks = append(hunks, h)
	}
	return hunks
}

// IsBinary đoán xem nội dung có phải file nhị phân hay không.
func IsBinary(data []byte) bool {
	limit := len(data)
	if limit > 8000 {
		limit = 8000
	}
	for i := 0; i < limit; i++ {
		if data[i] == 0 {
			return true
		}
	}
	return false
}

// Stat tổng hợp số dòng thêm/xóa giữa hai nội dung.
func Stat(a, b []string) (added, removed int) {
	for _, c := range DiffLines(a, b) {
		switch c.Op {
		case OpInsert:
			added += c.BEnd - c.BStart
		case OpDelete:
			removed += c.AEnd - c.AStart
		}
	}
	return
}

// FormatUnified dựng diff thống nhất dạng văn bản, dùng cho log --patch.
func FormatUnified(a, b []string, context int) string {
	hunks := hunksWithContext(a, b, context)
	if len(hunks) == 0 {
		return ""
	}
	var sb strings.Builder
	for _, h := range hunks {
		fmt.Fprintf(&sb, "@@ -%d,%d +%d,%d @@\n", h.AStart+1, h.ACount, h.BStart+1, h.BCount)
		for _, l := range h.Lines {
			sb.WriteString(l)
			sb.WriteString("\n")
		}
	}
	return sb.String()
}
