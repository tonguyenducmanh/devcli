package ops

// Dựng biểu đồ nhánh cho lịch sử commit.
//
// Mỗi commit nằm trong một "làn" (lane), mỗi làn là một cột hai ký tự. Commit
// nối vào làn đang chờ chính nó, còn các làn khác vẽ đường thẳng đứng. Khi một
// commit có nhiều cha thì mỗi cha sau cha đầu nhận một làn riêng, nên nhánh tách
// và hợp nhất hiện ra ngay trong lịch sử.
//
// Thuật toán dựa vào điều kiện thứ tự: danh sách commit phải luôn đứng sau con
// trước cha. `Repo.Log` duyệt bằng hàng đợi BFS nên điều kiện này có sẵn.
//
// Bề rộng tiền tố được chuẩn hoá về cùng số làn của biểu đồ, vì tiện ích VS
// Code ghép tiền tố này vào nhãn của cây; lệch cột thì nhánh nhìn không thành
// đường thẳng.

import "github.com/tonguyenducmanh/devcli/internal/vcs/object"

// Ký tự vẽ làn. Mỗi làn rộng đúng hai ký tự để các cột thẳng hàng.
const (
	graphCommit  byte = '*'
	graphLane    byte = '|'
	graphSpacing byte = ' '
)

// graphCellWidth là bề rộng của một làn.
const graphCellWidth = 2

// AssignGraphLanes gán cho mỗi commit tiền tố làn để dựng biểu đồ.
//
// Hàm ghi vào trường `GraphPrefix` của từng entry, nên phải gọi sau khi đã có
// danh sách commit.
func AssignGraphLanes(entries []LogEntry) {
	// lanes[i] là mã băm mà làn thứ i đang chờ. Mã băm rỗng nghĩa là làn đó
	// trống, được dùng lại cho nhánh mới tách ra.
	lanes := make([]object.Hash, 0, 4)
	// rows lưu ô vẽ của từng commit theo thứ tự, chuẩn hoá sau khi biết tổng
	// số làn.
	rows := make([][]byte, 0, len(entries))

	for i := range entries {
		entry := &entries[i]
		parents := entry.parents()

		lane := laneOf(lanes, entry.Hash)
		if lane < 0 {
			lane = freeLane(&lanes)
		}
		// Các làn khác cũng đang chờ chính commit này thì đã hợp nhất vào đây,
		// nên trả chúng về trống trước khi vẽ. Vẽ trước sẽ để lại dấu gạch
		// thừa ở những làn không còn dẫn tới đâu.
		releaseLanes(lanes, entry.Hash, lane)
		rows = append(rows, renderLanes(lanes, lane))

		// Cha đầu tiên nối tiếp ngay trong làn của commit này, vì đó là nhánh
		// chính mà lịch sử đi tiếp.
		if len(parents) > 0 {
			lanes[lane] = parents[0]
		} else {
			lanes[lane] = object.ZeroHash
		}
		// Các cha còn lại mỗi người một làn: đó là các nhánh đã hợp nhất vào.
		for _, parent := range splitParents(parents) {
			slot := freeLane(&lanes)
			lanes[slot] = parent
		}
	}

	width := maxLanes(rows)
	for i := range entries {
		entries[i].GraphPrefix = string(padRow(rows[i], width))
	}
}

// maxLanes trả về số làn lớn nhất trong toàn bộ biểu đồ.
func maxLanes(rows [][]byte) int {
	max := 0
	for _, row := range rows {
		if n := len(row) / graphCellWidth; n > max {
			max = n
		}
	}
	return max
}

// padRow nối thêm khoảng trắng cho tới khi đủ `cells` làn, để mọi dòng cùng
// bề rộng.
func padRow(row []byte, cells int) []byte {
	for len(row)/graphCellWidth < cells {
		row = append(row, graphSpacing, graphSpacing)
	}
	return row
}

// parents trả về danh sách mã băm cha của commit.
func (e LogEntry) parents() []object.Hash {
	return e.Parents
}

// splitParents trả về các cha từ thứ hai trở đi, tức là những nhánh đã hợp nhất vào.
//
// Tách riêng để chỗ gọi không phải tự cắt: lát cha rỗng cắt từ vị trí 1 sẽ lỗi,
// và commit cuối cùng của mọi nhánh đều rơi vào trường hợp đó.
func splitParents(parents []object.Hash) []object.Hash {
	if len(parents) <= 1 {
		return nil
	}
	return parents[1:]
}

// laneOf tìm làn đang chờ chính commit này, trả -1 nếu chưa làn nào chờ nó.
func laneOf(lanes []object.Hash, hash object.Hash) int {
	for i, lane := range lanes {
		if !lane.IsZero() && lane == hash {
			return i
		}
	}
	return -1
}

// releaseLanes trả các làn đang chờ `hash` trở lại thành trống, trừ làn `keep`.
//
// Nhiều làn có thể cùng chờ một mã băm khi hai nhánh cùng dẫn về một commit. Khi
// commit đó hiện ra ở làn `keep`, các làn còn lại đã quy về đó nên phải trống,
// nếu không sẽ vẽ thừa một đường thẳng đứng kéo dài xuống dưới.
func releaseLanes(lanes []object.Hash, hash object.Hash, keep int) {
	for i, lane := range lanes {
		if i != keep && lane == hash {
			lanes[i] = object.ZeroHash
		}
	}
}

// freeLane trả về làn trống đầu tiên, mở thêm làn mới khi đã đầy.
//
// Nhận con trỏ vì khi phải thêm làn thì `append` sẽ cấp phát mảng mới; truyền
// slice theo giá trị thì phần tử mới không nằm trong slice của người gọi và làn đó
// sẽ mất ngay ở lượt sau.
//
// Khi hai làn cùng chờ một mã băm thì commit đó sẽ nối vào làn trước, nên làn sau
// tự động trống và được dùng lại: đó là cách biểu diễn hai nhánh hợp nhất lại.
func freeLane(lanes *[]object.Hash) int {
	for i, lane := range *lanes {
		if lane.IsZero() {
			return i
		}
	}
	*lanes = append(*lanes, object.ZeroHash)
	return len(*lanes) - 1
}

// renderLanes dựng phần đầu dòng cho một commit đang nằm ở làn `at`.
func renderLanes(lanes []object.Hash, at int) []byte {
	out := make([]byte, 0, len(lanes)*graphCellWidth)
	for i, lane := range lanes {
		switch {
		case i == at:
			out = append(out, graphCommit, graphSpacing)
		case lane.IsZero():
			out = append(out, graphSpacing, graphSpacing)
		default:
			out = append(out, graphLane, graphSpacing)
		}
	}
	return out
}
