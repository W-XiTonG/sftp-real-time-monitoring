package handler

import (
	"log"
	"os"
	"path/filepath"
	"sftp-monitor/internal/sftpClient"
)

// 本地文件上传到远程服务器
func LocalUpload(file, RemotePaths, baseName, Username, Password, Host string, Port int, SupportCatalogs, whetherDelete bool) {
	remoteName := filepath.Join(RemotePaths, baseName)
	SftpClient := sftpClient.SFTP{}
	err := SftpClient.Connect(Username, Password, Host, Port)
	if err != nil {
		log.Printf("SftpClient.Connect err: %v", err)
	}
	defer SftpClient.Close()
	// 判断是否删除本地文件
	checkRemove := func(file string) {
		if whetherDelete {
			err = os.Remove(file)
			if err != nil {
				log.Printf("remove file %s err: %v", file, err)
			}
			log.Printf("remove file %s success", file)
		}
	}

	checkFileDir := func() bool {
		info, err := os.Stat(file)
		if err != nil {
			log.Printf("获取信息失败: %v", err)
		}
		return info.IsDir()
	}
	if !SupportCatalogs {
		if checkFileDir() {
			log.Printf("SupportCatalogs为false,跳过目录:%s", file)
			return
		}
		// 上传文件
		err = SftpClient.Upload(file, remoteName)
		if err != nil {
			log.Printf("SftpClient.Upload err: %v", err)
			return
		}
		log.Printf("SftpClient.Upload success:%s", file)
		checkRemove(file)
	} else {
		// 上传目录
		err := SftpClient.UploadDirectory(file, remoteName)
		if err != nil {
			log.Printf("SftpClient.Upload err: %v", err)
			return
		}
		log.Printf("SftpClient.Upload success:%s", file)
		checkRemove(file)
	}
}
