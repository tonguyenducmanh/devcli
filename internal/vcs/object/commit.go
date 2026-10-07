package object

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Identity là thông tin tác giả/người commit.
type Identity struct {
	Name  string
	Email string
	When  time.Time
}

// Format chuyển thông tin thành chuỗi "Tên <email> 1234567890 +0700",
// gồm tên, địa chỉ, mốc thời gian và múi giờ dạng chênh lệch so với UTC.
func (i Identity) Format() string {
	return fmt.Sprintf("%s <%s> %d %s", i.Name, i.Email, i.When.Unix(), formatTZ(i.When))
}

// formatTZ sinh múi giờ dạng +0700 từ độ lệch giờ địa phương.
func formatTZ(t time.Time) string {
	_, off := t.Zone()
	sign := "+"
	if off < 0 {
		sign = "-"
		off = -off
	}
	return fmt.Sprintf("%s%02d%02d", sign, off/3600, (off%3600)/60)
}

// ParseIdentity phân tích ngược một chuỗi do Format tạo ra thành Identity.
func ParseIdentity(s string) (Identity, error) {
	open := strings.LastIndex(s, "<")
	closeIdx := strings.LastIndex(s, ">")
	if open < 0 || closeIdx < open {
		return Identity{}, fmt.Errorf("tác giả không hợp lệ: %q", s)
	}
	id := Identity{
		Name:  strings.TrimSpace(s[:open]),
		Email: s[open+1 : closeIdx],
	}
	rest := strings.TrimSpace(s[closeIdx+1:])
	fields := strings.Fields(rest)
	if len(fields) >= 1 {
		if sec, err := strconv.ParseInt(fields[0], 10, 64); err == nil {
			id.When = time.Unix(sec, 0)
		}
	}
	if len(fields) >= 2 && (fields[1][0] == '+' || fields[1][0] == '-') {
		if off, err := strconv.ParseInt(fields[1], 10, 64); err == nil {
			id.When = id.When.In(time.FixedZone("", int(off)*3600))
		}
	}
	return id, nil
}

// Commit là một object commit, lưu lịch sử thay đổi.
type Commit struct {
	Tree      Hash
	Parents   []Hash
	Author    Identity
	Committer Identity
	Message   string
	// Extra giữ các header mở rộng (ví dụ gpgsig) dạng thô.
	Extra []byte
}

// Encode chuyển commit thành nội dung thô để lưu trữ.
func (c *Commit) Encode() []byte {
	var b bytes.Buffer
	fmt.Fprintf(&b, "tree %s\n", c.Tree)
	for _, p := range c.Parents {
		fmt.Fprintf(&b, "parent %s\n", p)
	}
	fmt.Fprintf(&b, "author %s\n", c.Author.Format())
	fmt.Fprintf(&b, "committer %s\n", c.Committer.Format())
	if len(c.Extra) > 0 {
		b.Write(c.Extra)
		if b.Bytes()[b.Len()-1] != '\n' {
			b.WriteByte('\n')
		}
	}
	b.WriteByte('\n')
	b.WriteString(c.Message)
	return b.Bytes()
}

// DecodeCommit phân tích nội dung thô thành Commit.
func DecodeCommit(data []byte) (*Commit, error) {
	// Tách phần header và message bằng dòng trống đầu tiên.
	idx := bytes.Index(data, []byte("\n\n"))
	var header, msg []byte
	if idx < 0 {
		header, msg = data, nil
	} else {
		header, msg = data[:idx], data[idx+2:]
	}
	c := &Commit{Message: string(msg)}
	lines := strings.Split(string(header), "\n")
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		// Header mở rộng được ghi tiếp dòng bắt đầu bằng khoảng trắng.
		if strings.HasPrefix(line, " ") && len(c.Extra) > 0 {
			c.Extra = append(c.Extra, line[1:]...)
			c.Extra = append(c.Extra, '\n')
			continue
		}
		key, value, ok := strings.Cut(line, " ")
		if !ok {
			continue
		}
		switch key {
		case "tree":
			h, err := ParseHash(value)
			if err != nil {
				return nil, err
			}
			c.Tree = h
		case "parent":
			h, err := ParseHash(value)
			if err != nil {
				return nil, err
			}
			c.Parents = append(c.Parents, h)
		case "author":
			id, err := ParseIdentity(value)
			if err != nil {
				return nil, err
			}
			c.Author = id
		case "committer":
			id, err := ParseIdentity(value)
			if err != nil {
				return nil, err
			}
			c.Committer = id
		default:
			// Giữ lại header lạ để không mất thông tin khi decode/encode lại.
			c.Extra = append(c.Extra, line...)
			c.Extra = append(c.Extra, '\n')
		}
	}
	if c.Tree.IsZero() {
		return nil, errors.New("commit thiếu trường tree")
	}
	return c, nil
}

// Summary trả về dòng tiêu đề đầu tiên của message.
func (c *Commit) Summary() string {
	msg := strings.TrimSpace(c.Message)
	if i := strings.IndexByte(msg, '\n'); i >= 0 {
		msg = msg[:i]
	}
	return msg
}
