package monitor

import (
	"log"
	"path/filepath"
	"sftp-monitor/internal/sftpClient"
	"time"
)

type DownloadFile struct {
	Name string
	Size int64
}

func RemoteWatcher(remoteDir, localDir, Username, Password, Host string, Port int, TimeInterval time.Duration) {
	SftpClient := sftpClient.SFTP{}
	err := SftpClient.Connect(Username, Password, Host, Port)
	if err != nil {
		log.Printf("SftpClient.Connect err: %v", err)
	}
	defer SftpClient.Close()
	downloadedFiles := make(map[string]DownloadFile)

	for {
		Client := sftpClient.SFTPClient{}
		// 列出远程目录中的文件
		files, err := Client.ReadDir(remoteDir)
		if err != nil {
			log.Printf("读取远程目录时出错: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}
		for _, file := range files {
			if !file.IsDir() {
				remotePath := filepath.Join(remoteDir, file.Name())
				localPath := filepath.Join(localDir, file.Name())

				// 检查文件是否已下载或文件大小是否有变化
				if downloaded, exists := downloadedFiles[file.Name()]; !exists || downloaded.Size != file.Size() {
					// 下载新文件或文件大小有变化的文件
					err = SftpClient.Download(remotePath, localPath)
					if err != nil {
						log.Printf("下载文件 %s 时出错: %v", file.Name(), err)
					} else {
						log.Printf("成功下载文件: %s", file.Name())
						// 更新已下载文件的信息
						downloadedFiles[file.Name()] = DownloadFile{
							Name: file.Name(),
							Size: file.Size(),
						}
					}
				}
			}
		}
		// 检查间隔
		time.Sleep(TimeInterval)
	}
}
