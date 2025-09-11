// core/network.go
package core

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/jlaffaye/ftp"
)

// Uploader 定义了网络上传器的接口
type Uploader interface {
	Upload(path string, data io.Reader) error
	UploadWithResume(path string, data io.ReaderAt, size int64) error
	GetRemoteSize(path string) (int64, error)
	Close() error
}

// UploadConfig 上传配置
type UploadConfig struct {
	MaxRetries    int
	RetryInterval time.Duration
	ChunkSize     int64
}

// DefaultUploadConfig 返回默认上传配置
func DefaultUploadConfig() *UploadConfig {
	return &UploadConfig{
		MaxRetries:    3,
		RetryInterval: time.Second * 5,
		ChunkSize:     1024 * 1024, // 1MB
	}
}

// FTPUploader FTP上传器实现
type FTPUploader struct {
	conn   *ftp.ServerConn
	config *UploadConfig
}

// NewFTPUploader 创建FTP上传器
func NewFTPUploader(ftpURL string, config *UploadConfig) (*FTPUploader, error) {
	if config == nil {
		config = DefaultUploadConfig()
	}

	parsedURL, err := url.Parse(ftpURL)
	if err != nil {
		return nil, fmt.Errorf("invalid FTP URL: %w", err)
	}

	host := parsedURL.Host
	if parsedURL.Port() == "" {
		if parsedURL.Scheme == "ftps" {
			host += ":990"
		} else {
			host += ":21"
		}
	}

	var conn *ftp.ServerConn
	if parsedURL.Scheme == "ftps" {
		// FTPS (FTP over TLS/SSL)
		conn, err = ftp.Dial(host, ftp.DialWithTLS(&tls.Config{
			InsecureSkipVerify: false,
		}))
	} else {
		// Regular FTP
		conn, err = ftp.Dial(host)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to FTP server: %w", err)
	}

	// 认证
	var username, password string
	if parsedURL.User != nil {
		username = parsedURL.User.Username()
		password, _ = parsedURL.User.Password()
	}
	if username == "" {
		username = "anonymous"
	}

	if err := conn.Login(username, password); err != nil {
		conn.Quit()
		return nil, fmt.Errorf("FTP login failed: %w", err)
	}

	return &FTPUploader{
		conn:   conn,
		config: config,
	}, nil
}

// Upload 实现基本上传功能
func (f *FTPUploader) Upload(path string, data io.Reader) error {
	// 确保目录存在
	if err := f.ensureDirectoryExists(filepath.Dir(path)); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt <= f.config.MaxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(f.config.RetryInterval)
		}

		if err := f.conn.Stor(path, data); err != nil {
			lastErr = err
			continue
		}

		return nil
	}

	return fmt.Errorf("upload failed after %d attempts: %w", f.config.MaxRetries+1, lastErr)
}

// UploadWithResume 实现断点续传上传
func (f *FTPUploader) UploadWithResume(path string, data io.ReaderAt, size int64) error {
	// 确保目录存在
	if err := f.ensureDirectoryExists(filepath.Dir(path)); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// 检查远程文件大小
	remoteSize, err := f.GetRemoteSize(path)
	if err != nil && !isFileNotFoundError(err) {
		return fmt.Errorf("failed to check remote file size: %w", err)
	}

	startPos := remoteSize
	if startPos >= size {
		// 文件已完全上传
		return nil
	}

	// 创建从指定位置开始的reader
	remainingData := &sectionReader{
		r:    data,
		base: startPos,
		off:  startPos,
		n:    size - startPos,
	}

	var lastErr error
	for attempt := 0; attempt <= f.config.MaxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(f.config.RetryInterval)
			// 重新检查远程文件大小
			remoteSize, err = f.GetRemoteSize(path)
			if err != nil && !isFileNotFoundError(err) {
				lastErr = err
				continue
			}
			startPos = remoteSize
			if startPos >= size {
				return nil
			}
			// 重新设置reader位置
			remainingData = &sectionReader{
				r:    data,
				base: startPos,
				off:  startPos,
				n:    size - startPos,
			}
		}

		// 使用APPE命令进行断点续传
		if startPos > 0 {
			err = f.conn.Append(path, remainingData)
		} else {
			err = f.conn.Stor(path, remainingData)
		}

		if err != nil {
			lastErr = err
			continue
		}

		return nil
	}

	return fmt.Errorf("resume upload failed after %d attempts: %w", f.config.MaxRetries+1, lastErr)
}

// GetRemoteSize 获取远程文件大小
func (f *FTPUploader) GetRemoteSize(path string) (int64, error) {
	entries, err := f.conn.List(filepath.Dir(path))
	if err != nil {
		return 0, fmt.Errorf("failed to list directory: %w", err)
	}

	filename := filepath.Base(path)
	for _, entry := range entries {
		if entry.Name == filename && entry.Type == ftp.EntryTypeFile {
			return int64(entry.Size), nil
		}
	}

	return 0, errors.New("file not found")
}

// ensureDirectoryExists 确保目录存在
func (f *FTPUploader) ensureDirectoryExists(dirPath string) error {
	if dirPath == "" || dirPath == "." || dirPath == "/" {
		return nil
	}

	// 递归创建父目录
	parent := filepath.Dir(dirPath)
	if parent != dirPath {
		if err := f.ensureDirectoryExists(parent); err != nil {
			return err
		}
	}

	// 尝试创建目录（如果已存在会返回错误，但我们忽略它）
	f.conn.MakeDir(dirPath)
	return nil
}

// Close 关闭连接
func (f *FTPUploader) Close() error {
	if f.conn != nil {
		return f.conn.Quit()
	}
	return nil
}

// sectionReader 实现从指定位置开始读取的Reader
type sectionReader struct {
	r    io.ReaderAt
	base int64
	off  int64
	n    int64
}

func (s *sectionReader) Read(p []byte) (n int, err error) {
	if s.n <= 0 {
		return 0, io.EOF
	}
	if int64(len(p)) > s.n {
		p = p[0:s.n]
	}
	n, err = s.r.ReadAt(p, s.off)
	s.off += int64(n)
	s.n -= int64(n)
	return
}

// GetUploaderFor 根据URL创建对应的上传器
func GetUploaderFor(destinationUrl string) (Uploader, error) {
	parsedURL, err := url.Parse(destinationUrl)
	if err != nil {
		return nil, fmt.Errorf("invalid destination URL: %w", err)
	}

	switch strings.ToLower(parsedURL.Scheme) {
	case "ftp", "ftps":
		return NewFTPUploader(destinationUrl, nil)
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", parsedURL.Scheme)
	}
}

// isFileNotFoundError 检查是否为文件未找到错误
func isFileNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "not found") ||
		strings.Contains(errStr, "No such file") ||
		strings.Contains(errStr, "file not found")
}
