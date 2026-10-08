// Package repo cung cấp lớp trừu tượng cao cấp cho một kho mã nguồn của tm:
// mở repo, đọc/ghi cây, index, ref và trạng thái HEAD.
package repo

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tonguyenducmanh/devcli/internal/vcs/config"
	"github.com/tonguyenducmanh/devcli/internal/vcs/index"
	"github.com/tonguyenducmanh/devcli/internal/vcs/object"
	"github.com/tonguyenducmanh/devcli/internal/vcs/storage"
	"github.com/tonguyenducmanh/devcli/internal/vcs/worktree"
)

// Tên thư mục lưu trữ dữ liệu của tm.
const DirName = ".tmx"

// Các tên thư mục con.
const (
	objectsDir = "objects"
	refsDir    = "refs"
	headsDir   = "refs/heads"
	tagsDir    = "refs/tags"
	indexFile  = "index"
	headFile   = "HEAD"
)

// ErrNotRepo báo lỗi khi thư mục hiện tại không thuộc repo nào.
var ErrNotRepo = errors.New("không phải repo tm (chạy `tm vcs init` trước)")

// ErrNotFound báo lỗi khi không tìm thấy đối tượng được yêu cầu.
var ErrNotFound = errors.New("không tìm thấy")

// Repo là một kho mã nguồn.
type Repo struct {
	// Root là thư mục gốc làm việc.
	Root string
	// GitDir là thư mục lưu dữ liệu nội bộ.
	GitDir string

	Objects *storage.ObjectStore
	Refs    *storage.RefStore
	Config  *config.Config
	Index   *index.Index
	Ignore  *worktree.Ignore
}

// Open mở repo từ một đường dẫn, tìm .tmx từ thư mục đó đi lên trên.
func Open(start string) (*Repo, error) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return nil, err
	}
	for {
		gitDir := filepath.Join(dir, DirName)
		if fi, err := os.Stat(gitDir); err == nil && fi.IsDir() {
			return openAt(dir, gitDir)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return nil, ErrNotRepo
		}
		dir = parent
	}
}

// OpenAt mở repo tại đúng một thư mục gốc cho trước.
func OpenAt(root string) (*Repo, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	gitDir := filepath.Join(abs, DirName)
	if fi, err := os.Stat(gitDir); err != nil || !fi.IsDir() {
		return nil, ErrNotRepo
	}
	return openAt(abs, gitDir)
}

// FindRoot trả về thư mục gốc chứa .tmx, hoặc rỗng nếu không có.
func FindRoot(start string) string {
	dir, err := filepath.Abs(start)
	if err != nil {
		return ""
	}
	for {
		if fi, err := os.Stat(filepath.Join(dir, DirName)); err == nil && fi.IsDir() {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

func openAt(root, gitDir string) (*Repo, error) {
	r := &Repo{
		Root:    root,
		GitDir:  gitDir,
		Objects: storage.NewObjectStore(filepath.Join(gitDir, objectsDir)),
		Refs:    storage.NewRefStore(gitDir),
	}
	cfg, err := config.Open(gitDir)
	if err != nil {
		return nil, err
	}
	r.Config = cfg

	idx, err := index.Open(filepath.Join(gitDir, indexFile))
	if err != nil {
		return nil, err
	}
	r.Index = idx

	// Nạp các quy tắc bỏ qua: .tmx/info/exclude và .tmxignore ở gốc repo.
	ig := worktree.NewIgnore(root)
	if err := ig.AddFile(filepath.Join(gitDir, "info", "exclude")); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	if err := ig.AddFile(filepath.Join(root, ".tmxignore")); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	r.Ignore = ig
	return r, nil
}

// Init tạo một repo mới tại thư mục cho trước.
func Init(root string, defaultBranch string) (*Repo, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(abs, 0o755); err != nil {
		return nil, err
	}
	gitDir := filepath.Join(abs, DirName)
	if fi, err := os.Stat(gitDir); err == nil && fi.IsDir() {
		return nil, fmt.Errorf("repo đã tồn tại tại %s", abs)
	}
	for _, d := range []string{objectsDir, refsDir, "info", "branches"} {
		if err := os.MkdirAll(filepath.Join(gitDir, filepath.FromSlash(d)), 0o755); err != nil {
			return nil, err
		}
	}
	r, err := openAt(abs, gitDir)
	if err != nil {
		return nil, err
	}
	// HEAD trỏ tới nhánh mặc định.
	if err := r.Refs.WriteSymbolic(headFile, headsDir+"/"+defaultBranch); err != nil {
		return nil, err
	}
	// Lưu nhánh mặc định trong config để các lệnh sau dùng lại.
	r.Config.Set("tm.defaultbranch", defaultBranch)
	if err := r.Config.Save(); err != nil {
		return nil, err
	}
	// Tạo tệp ignore mẫu ở gốc kho nếu chưa có, để kho mới khởi động mà không
	// phải lần theo tệp build, log hay cache của công cụ khác.
	if _, err := worktree.WriteDefaultIgnore(abs); err != nil {
		return nil, err
	}
	// Tạo .tmx/info/exclude rỗng nếu chưa có để người dùng chỉnh sửa.
	excludePath := filepath.Join(gitDir, "info", "exclude")
	if _, err := os.Stat(excludePath); os.IsNotExist(err) {
		if err := os.WriteFile(excludePath, []byte("# các mẫu file bị bỏ qua, mỗi dòng một mẫu\n"), 0o644); err != nil {
			return nil, err
		}
	}
	// Ghi reflog ban đầu cho HEAD.
	if err := r.Refs.AppendReflog(headFile, object.ZeroHash, object.ZeroHash, "tm vcs init"); err != nil {
		return nil, err
	}
	return r, nil
}

// DefaultBranch trả về tên nhánh mặc định đã cấu hình.
func (r *Repo) DefaultBranch() string {
	return r.Config.GetString("tm.defaultbranch", "main")
}

// Identity trả về thông tin tác giả cho các thao tác ghi.
func (r *Repo) Identity() object.Identity {
	name, email := r.Config.ResolveIdentity()
	return object.Identity{Name: name, Email: email, When: time.Now()}
}

// AbsPath chuyển đường dẫn tương đối thành đường dẫn tuyệt đối trong repo.
func (r *Repo) AbsPath(p string) string {
	if filepath.IsAbs(p) {
		return filepath.Clean(p)
	}
	return filepath.Join(r.Root, filepath.FromSlash(p))
}

// RelPath chuyển đường dẫn tuyệt đối thành đường dẫn tương đối so với gốc repo,
// dùng dấu "/" và trả về lỗi nếu nằm ngoài repo.
func (r *Repo) RelPath(p string) (string, error) {
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(r.Root, abs)
	if err != nil {
		return "", err
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", worktree.ErrOutsideRepo{Path: p}
	}
	return filepath.ToSlash(rel), nil
}

// WorkPath trả về đường dẫn tuyệt đối của một đường dẫn trong repo.
func (r *Repo) WorkPath(rel string) string {
	return filepath.Join(r.Root, filepath.FromSlash(rel))
}

// ReadHeadRef đọc nội dung HEAD dưới dạng tên ref đầy đủ.
// Nếu HEAD đang tách rời (detached) thì trả về chuỗi rỗng.
func (r *Repo) ReadHeadRef() (string, error) {
	data, err := os.ReadFile(filepath.Join(r.GitDir, headFile))
	if err != nil {
		return "", err
	}
	s := strings.TrimSpace(string(data))
	if strings.HasPrefix(s, "ref:") {
		return strings.TrimSpace(strings.TrimPrefix(s, "ref:")), nil
	}
	return "", nil
}

// SetHeadSymlink trỏ HEAD tới một ref (dùng khi chuyển nhánh).
func (r *Repo) SetHeadSymlink(ref string) error {
	return r.Refs.WriteSymbolic(headFile, ref)
}

// SetHeadDetached ghim HEAD trực tiếp tới một commit.
func (r *Repo) SetHeadDetached(h object.Hash) error {
	return os.WriteFile(filepath.Join(r.GitDir, headFile), []byte(h.String()+"\n"), 0o644)
}

// Head trả về hash commit mà HEAD đang trỏ tới.
// Trả về ZeroHash nếu repo chưa có commit nào.
func (r *Repo) Head() (object.Hash, error) {
	h, err := r.Refs.Resolve(headFile)
	if err == storage.ErrRefNotFound {
		return object.ZeroHash, nil
	}
	return h, err
}

// ResolveHeadCommit trả về hash commit ở đầu HEAD, nếu có.
func (r *Repo) ResolveHeadCommit() (object.Hash, error) {
	return r.Head()
}

// CurrentBranch trả về tên nhánh hiện tại.
// Nếu HEAD tách rời thì trả về chuỗi rỗng.
func (r *Repo) CurrentBranch() (string, error) {
	ref, err := r.ReadHeadRef()
	if err != nil {
		return "", err
	}
	if ref == "" {
		return "", nil
	}
	return strings.TrimPrefix(ref, headsDir+"/"), nil
}

// Branches liệt kê tất cả nhánh cục bộ.
func (r *Repo) Branches() ([]string, error) {
	refs, err := r.Refs.List(headsDir + "/")
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(refs))
	for _, ref := range refs {
		out = append(out, strings.TrimPrefix(ref, headsDir+"/"))
	}
	return out, nil
}

// Tags liệt kê tất cả tag.
func (r *Repo) Tags() ([]string, error) {
	refs, err := r.Refs.List(tagsDir + "/")
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(refs))
	for _, ref := range refs {
		out = append(out, strings.TrimPrefix(ref, tagsDir+"/"))
	}
	return out, nil
}

// StateDir tạo và trả về thư mục trạng thái cho các thao tác nhiều bước
// (merge, rebase, cherry-pick đang dở dang).
func (r *Repo) StateDir(name string) (string, error) {
	p := filepath.Join(r.GitDir, name)
	if err := os.MkdirAll(p, 0o755); err != nil {
		return "", err
	}
	return p, nil
}

// StatePath trả về đường dẫn file trạng thái.
func (r *Repo) StatePath(name string) string {
	return filepath.Join(r.GitDir, name)
}

// HasState báo xem có file/thư mục trạng thái cho trước hay không.
func (r *Repo) HasState(name string) bool {
	_, err := os.Stat(r.StatePath(name))
	return err == nil
}

// ClearState xóa trạng thái của một thao tác.
func (r *Repo) ClearState(name string) error {
	return os.RemoveAll(r.StatePath(name))
}

// ReadState đọc file trạng thái (dòng đầu tiên là hash).
func (r *Repo) ReadState(name string) (string, error) {
	data, err := os.ReadFile(r.StatePath(name))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

// WriteState ghi file trạng thái.
func (r *Repo) WriteState(name, content string) error {
	if err := os.MkdirAll(filepath.Dir(r.StatePath(name)), 0o755); err != nil {
		return err
	}
	return os.WriteFile(r.StatePath(name), []byte(content), 0o644)
}

// SaveIndex lưu staging area xuống đĩa.
func (r *Repo) SaveIndex() error { return r.Index.Save() }
