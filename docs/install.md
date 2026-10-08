# Cài đặt

Chọn đúng tệp trong `out/` theo hệ điều hành của bạn, rồi chép vào một thư mục
nằm trong `PATH`. Từ đó gõ `tm` là chạy.

## macOS

Bản dựng cho Apple Silicon, kiểm tra bằng `uname -m` cho ra `arm64`.

```bash
install -m 755 out/td-devcli-mac-arm-0.1.0 ~/bin/tm

echo 'export PATH="$PATH:$HOME/bin"' >> ~/.zshrc
source ~/.zshrc
tm version
```

Muốn cài cho toàn hệ thống thay vì chỉ tài khoản của bạn:

```bash
sudo install -m 755 out/td-devcli-mac-arm-0.1.0 /usr/local/bin/tm
```

Nếu sao chép tệp bằng trình duyệt, macOS có thể chặn vì không có chữ ký:

```bash
xattr -d com.apple.quarantine ~/bin/tm
```

## Linux

Bản dựng cho máy 64 bit thông thường, kiểm tra bằng `uname -m` cho ra `x86_64`.

```bash
install -m 755 out/td-devcli-linux-0.1.0 ~/.local/bin/tm

echo 'export PATH="$PATH:$HOME/.local/bin"' >> ~/.bashrc
source ~/.bashrc
```

Hoặc cài cho toàn hệ thống:

```bash
sudo install -m 755 out/td-devcli-linux-0.1.0 /usr/local/bin/tm
```

Kiểm tra tệp đã quyền chạy chưa:

```bash
ls -l ~/bin/tm          # phần cuối phải là rwxr-xr-x
chmod +x ~/bin/tm       # nếu chưa có
```

## Windows

Mở PowerShell ở thư mục chứa thư mục `out/`:

```powershell
New-Item -ItemType Directory -Force "$env:LOCALAPPDATA\tm"
Copy-Item out\td-devcli-windows-0.1.0.exe "$env:LOCALAPPDATA\tm\tm.exe"

# thêm vào Path của người dùng, có hiệu lực vĩnh viễn
[Environment]::SetEnvironmentVariable(
  "Path",
  [Environment]::GetEnvironmentVariable("Path", "User") + ";$env:LOCALAPPDATA\tm",
  "User"
)
```

Mở một cửa sổ PowerShell mới rồi kiểm tra:

```powershell
tm version
```

Trường hợp đang sử dụng PowerShell trong vscode, chạy lệnh sau để cập nhật path:

```powershell
$env:Path = [System.Environment]::GetEnvironmentVariable("Path","Machine") + ";" + [System.Environment]::GetEnvironmentVariable("Path","User")
```

Windows không cần cấp quyền thực thi như trên macOS và Linux. Nếu PowerShell chặn
việc chạy tệp, xem [Execution Policy](https://learn.microsoft.com/powershell/module/microsoft.powershell.security/set-executionpolicy).

## Dùng tạm mà không cài

Không cần chép vào `PATH` cũng chạy được, gọi thẳng tệp thôi:

```bash
./out/td-devcli-mac-arm-0.1.0 vcs status
```

## Thay bằng bản mới

devcli không lưu gì cạnh tệp thực thi, nên chép tệp mới đè lên tệp cũ là xong.
Cấu hình nằm ở `~/.config/tm/config` và dữ liệu mỗi dự án nằm trong thư mục
`.tmx`, cả hai đều không bị ảnh hưởng.

```bash
install -m 755 out/td-devcli-linux-1.2.0 ~/.local/bin/tm
```

```powershell
Copy-Item -Force out\td-devcli-windows-1.2.0.exe "$env:LOCALAPPDATA\tm\tm.exe"
```

## Gỡ bỏ

```bash
rm ~/bin/tm                    # hoặc ~/.local/bin/tm, /usr/local/bin/tm
```

Tệp thực thi là toàn bộ phần mềm, xoá nó là xong. Nếu muốn xoá nốt cấu hình:

```bash
rm -rf ~/.config/tm
```

Dữ liệu kho mã nguồn trong thư mục `.tmx` của từng dự án không bị đụng tới,
vì đó là lịch sử công việc của bạn. Xoá thủ công khi không cần nữa:

```bash
find . -type d -name .tmx -prune -exec rm -rf {} +
```
