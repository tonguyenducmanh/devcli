package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/tonguyenducmanh/devcli/cmd"
)

// File này kiểm tra phần trợ giúp của cây lệnh: tiếng Việt, đúng hành vi khi
// gõ lệnh không kèm tham số, và không còn chỗ nào để lộ nhãn tiếng Anh của
// cobra.

// chayLenh chạy cây lệnh trong bộ nhớ rồi trả về output và lỗi phát sinh.
//
// Phần trợ giúp do cobra in ra, còn phần thông báo của lệnh do hàm printLine
// ghi thẳng ra os.Stdout. Vì vậy phải đổi os.Stdutdown đi một ống thật rồi đọc
// lại, nếu không sẽ chỉ bắt được một nửa.
func chayLenh(t *testing.T, args ...string) (string, error) {
	t.Helper()

	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	cu := os.Stdout
	os.Stdout = writer

	defer func() {
		os.Stdout = cu
	}()

	root := cmd.Root()
	resetFlags(root)
	root.SetOut(writer)
	root.SetErr(writer)
	root.SetArgs(args)

	loi := root.Execute()

	writer.Close()
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, reader); err != nil {
		t.Fatal(err)
	}
	reader.Close()

	return buf.String(), loi
}

// resetFlags trả mọi cờ về giá trị mặc định trước khi chạy lệnh kế tiếp.
//
// Cây lệnh chỉ có một bản dùng chung cho cả gói kiểm thử, mà cobra không tự
// đặt lại cờ sau mỗi lần Execute. Không reset thì cờ --help đã bật ở lần chạy
// trước sẽ khiến lần sau in trợ giúp ngay, và kiểm thử sẽ báo động lung tung.
func resetFlags(root *cobra.Command) {
	var duyet func(c *cobra.Command)
	duyet = func(c *cobra.Command) {
		ve := func(f *pflag.Flag) {
			_ = f.Value.Set(f.DefValue)
			f.Changed = false
		}
		c.Flags().VisitAll(ve)
		c.PersistentFlags().VisitAll(ve)
		for _, sub := range c.Commands() {
			duyet(sub)
		}
	}
	duyet(root)
}

// chayLenhOK chạy lệnh rồi trả về output, đồng thời đòi lệnh phải thành công.
func chayLenhOK(t *testing.T, args ...string) string {
	t.Helper()
	dong, loi := chayLenh(t, args...)
	if loi != nil {
		t.Fatalf("%v phải chạy thành công, gặp lỗi: %v", args, loi)
	}
	return dong
}

// nhanTiengAnh là các nhãn do cobra in ra trong phần trợ giúp. Không nhãn nào
// chứa chữ cái tiếng Việt có dấu, nên chỉ cần so khớp nguyên văn.
var nhanTiengAnh = []string{
	"help for", "version for", "Help about any command",
	"Usage:", "Aliases:", "Examples:", "Available Commands:",
	"Additional Commands:", "Flags:", "Global Flags:",
	"Additional help topics:", "for more information about a command",
	"unknown command", "unknown flag",
}

// TestTroGiopKhongConTiengAnh bảo đảm không chỗ nào của trợ giúp còn sót nhãn
// tiếng Anh của cobra.
//
// Nếu bỏ qua kiểm thử này thì việc dịch phần trợ giúp rất dễ vỡ: chỉ cần ai
// đó thêm một lệnh mới, hoặc cobra đổi cách dựng cờ, là tiếng Anh quay lại mà
// không ai hay.
func TestTroGiopKhongConTiengAnh(t *testing.T) {
	var duyet func(c *cobra.Command)
	duyet = func(c *cobra.Command) {
		path := strings.Fields(c.CommandPath())

		dong := chayLenhOK(t, append(path, "--help")...)
		for _, nhan := range nhanTiengAnh {
			if strings.Contains(dong, nhan) {
				t.Errorf("trợ giúp của %q chứa nhãn tiếng Anh %q", c.CommandPath(), nhan)
			}
		}

		for _, sub := range c.Commands() {
			duyet(sub)
		}
	}
	duyet(cmd.Root())
}

// TestMoTaCoBangTiengViet bảo đảm mọi cờ, kể cả cờ do cobra tự sinh, đều có mô
// tả tiếng Việt.
func TestMoTaCoBangTiengViet(t *testing.T) {
	var duyet func(c *cobra.Command)
	duyet = func(c *cobra.Command) {
		c.InitDefaultHelpFlag()
		c.Flags().VisitAll(func(f *pflag.Flag) {
			if !strings.ContainsAny(f.Usage, chuViet) {
				t.Errorf("cờ --%s của %q mô tả bằng tiếng Anh: %q", f.Name, c.CommandPath(), f.Usage)
			}
		})
		for _, sub := range c.Commands() {
			duyet(sub)
		}
	}
	duyet(cmd.Root())
}

// TestLenhGocChayThatKhongInTroGiop bảo đảm gõ lệnh gốc không ra trang trợ giúp.
//
// `td` không kèm lệnh con nào thì phải làm được việc gì đó hữu ích, đó là in
// phiên bản và danh sách lệnh. Nếu ai đó đổi lệnh gốc thành in trợ giúp thì
// kiểm thử này đỏ.
func TestLenhGocChayThatKhongInTroGiop(t *testing.T) {
	dong := chayLenhOK(t)

	if strings.Contains(dong, "Cách dùng:") {
		t.Error("gõ lệnh gốc không nên in trợ giúp, hãy in danh sách lệnh")
	}
	for _, can := range []string{"vcs", "config", "version"} {
		if !strings.Contains(dong, can) {
			t.Errorf("danh sách lệnh thiếu %q", can)
		}
	}
}

// TestCoVerboseInThongTinMoiTruong bảo đảm cờ -v làm được việc gì đó.
//
// Cờ -v từng chỉ in trợ giúp nên vô dụng, giờ nó phải in thông tin môi trường.
func TestCoVerboseInThongTinMoiTruong(t *testing.T) {
	dong := chayLenhOK(t, "-v")

	if !strings.Contains(dong, "nền tảng:") {
		t.Errorf("cờ -v chưa in thông tin môi trường, nhận được: %q", dong)
	}
	if strings.Contains(dong, "Cách dùng:") {
		t.Error("cờ -v không nên in trợ giúp")
	}
}

// TestNhomLenhChayThatKhongInTroGiop bảo đảm gõ một nhóm lệnh không ra trợ giúp.
func TestNhomLenhChayThatKhongInTroGiop(t *testing.T) {
	for _, nhom := range []string{"vcs"} {
		dong := chayLenhOK(t, nhom)

		if strings.Contains(dong, "Cách dùng:") {
			t.Errorf("%s không nên in trợ giúp", nhom)
		}
		if !strings.Contains(dong, "Xem chi tiết") {
			t.Errorf("%s phải chỉ dẫn cách xem chi tiết", nhom)
		}
	}
}

// TestTenLenhSaiBaoLoiTiengViet bảo đảm gõ sai tên lệnh thì báo lỗi tiếng Việt
// với mã thoát khác 0, chứ không in help hay báo bằng tiếng Anh.
func TestTenLenhSaiBaoLoiTiengViet(t *testing.T) {
	for _, args := range [][]string{
		{"khongTonTai"},
		{"vcs", "khongTonTai"},
		{"version", "thuaThamSo"},
	} {
		dong, loi := chayLenh(t, args...)

		if loi == nil {
			t.Errorf("%v phải báo lỗi, lại chạy thành công", args)
			continue
		}
		if strings.Contains(dong, "Cách dùng:") {
			t.Errorf("%v không nên in trợ giúp khi sai", args)
		}
		if !strings.ContainsAny(loi.Error(), chuViet) {
			t.Errorf("%v báo lỗi bằng tiếng Anh: %q", args, loi)
		}
	}
}

// TestXemTroGiopChiTietBằngTienViet bảo đảm xem trợ giúp chi tiết vẫn chạy được
// sau khi dịch sang tiếng Việt.
func TestXemTroGiopChiTietBangTienViet(t *testing.T) {
	for _, args := range [][]string{
		{"--help"},
		{"vcs", "--help"},
		{"vcs", "merge", "--help"},
		{"help"},
		{"help", "vcs", "merge"},
	} {
		dong := chayLenhOK(t, args...)

		if !strings.Contains(dong, "Cách dùng:") {
			t.Errorf("%v phải in phần trợ giúp", args)
		}
		if !strings.ContainsAny(dong, chuViet) {
			t.Errorf("%v in trợ giúp không có tiếng Việt: %q", args, dong)
		}
	}
}

// TestLenhConKhongDungLaiTenCoToanCuc bảo đảm không lệnh con nào khai báo cờ
// trùng tên hoặc trùng chữ viết tắt với cờ toàn cục.
//
// Cơ chế của pflag là âm thầm bỏ qua cờ bị che, không báo lỗi. Lệnh con che cờ
// toàn cục thì cờ toàn cục chết lặng lẽ trong lệnh đó, và cùng một chữ viết tắt
// mang hai nghĩa khác nhau tuỳ lệnh. Người dùng không có cách nào biết.
//
// Chỉ cờ do chính lệnh đó khai báo mới bị soi. Sau khi lệnh đã chạy, pflag nối
// cờ toàn cục của các lệnh cha vào chính lệnh con, nên phải loại bỏ những cờ đó
// ra trước.
func TestLenhConKhongDungLaiTenCoToanCuc(t *testing.T) {
	root := cmd.Root()

	var duyet func(c *cobra.Command)
	duyet = func(c *cobra.Command) {
		// Gom cờ toàn cục của mọi lệnh cha.
		tenCha := map[string]bool{}
		chuTatCha := map[string]string{}
		for parent := c.Parent(); parent != nil; parent = parent.Parent() {
			parent.PersistentFlags().VisitAll(func(f *pflag.Flag) {
				tenCha[f.Name] = true
				if f.Shorthand != "" {
					chuTatCha[f.Shorthand] = f.Name
				}
			})
		}

		c.Flags().VisitAll(func(f *pflag.Flag) {
			if tenCha[f.Name] {
				return // cờ này thuộc về lệnh cha, không phải lệnh này
			}
			if ten, ok := chuTatCha[f.Shorthand]; ok && f.Shorthand != "" {
				t.Errorf("lệnh %q dùng chữ viết tắt -%s cho --%s, trùng cờ toàn cục --%s của lệnh cha",
					c.CommandPath(), f.Shorthand, f.Name, ten)
			}
			for parent := c.Parent(); parent != nil; parent = parent.Parent() {
				if parent.PersistentFlags().Lookup(f.Name) != nil {
					t.Errorf("lệnh %q khai báo cờ --%s, trùng với cờ toàn cục của %q",
						c.CommandPath(), f.Name, parent.Name())
				}
			}
		})

		for _, sub := range c.Commands() {
			duyet(sub)
		}
	}
	duyet(root)
}

// TestGiaTriMacDinhKhopCauHinhBuild bảo đảm giá trị mặc định trong Go không
// lệch với cấu hình build.
//
// Script build ghi đè các biến này bằng ldflags, nên khi chạy `go build` thẳng
// mà không qua script thì dùng giá trị mặc định ở Go. Hai nơi phải giống nhau,
// nếu không thì bản dựng tay sẽ mang tên khác bản dựng từ script. Có kiểm thử
// này thì đổi một bên mà quên bên kia sẽ bị chặn.
func TestGiaTriMacDinhKhopCauHinhBuild(t *testing.T) {
	noiDung, err := os.ReadFile(filepath.Join(moduleRoot(t), "scripts", "build_binaries.sh"))
	if err != nil {
		t.Fatal(err)
	}

	canDoi := map[string]string{
		"CMD_NAME": cmd.AppName,
		"REPO_URL": cmd.RepoURL,
		"AUTHOR":   cmd.Author,
	}
	for ten, gt := range canDoi {
		trongScript := giaTriTrongScript(string(noiDung), ten)
		if trongScript == "" {
			t.Errorf("không tìm thấy biến %s trong scripts/build_binaries.sh", ten)
			continue
		}
		if trongScript != gt {
			t.Errorf("%s lệch: script build là %q, mặc định trong Go là %q", ten, trongScript, gt)
		}
	}
}

// giaTriTrongScript đọc giá trị của một biến gán thẳng trong tập lệnh shell,
// bỏ dấu nháy quanh giá trị nếu có.
func giaTriTrongScript(noiDung, ten string) string {
	for _, dong := range strings.Split(noiDung, "\n") {
		rest, ok := strings.CutPrefix(dong, ten+"=")
		if !ok {
			continue
		}
		gt := strings.TrimSpace(rest)
		gt = strings.Trim(gt, `"`)
		return gt
	}
	return ""
}

// TestNhanLienHeDuocInRa bảo đảm tên tác giả và nơi phát hành hiện ở nơi người
// đọc cần tìm.
func TestNhanLienHeDuocInRa(t *testing.T) {
	for _, args := range [][]string{
		{"version"},
		{"-v"},
		{"--help"},
	} {
		dong := chayLenhOK(t, args...)
		if !strings.Contains(dong, cmd.Author) {
			t.Errorf("%v không in tên tác giả %q", args, cmd.Author)
		}
		if !strings.Contains(dong, cmd.RepoURL) {
			t.Errorf("%v không in nơi phát hành %q", args, cmd.RepoURL)
		}
	}
}
