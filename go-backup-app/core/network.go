// core/network.go
package core

import "io"

// TODO: FEATURE - Network Backup

// Uploader 定义了网络上传器的接口
type Uploader interface {
	Upload(path string, data io.Reader) error
}

func GetUploaderFor(destinationUrl string) (Uploader, error) {
	// TODO: 根据 URL scheme (s3://, ftp://, ...) 返回不同的上传器实现
	panic("Network backup not implemented")
}
