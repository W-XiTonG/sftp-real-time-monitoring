package sftpClient

import (
	"fmt"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"io"
	"log"
	"os"
	"path/filepath"
)

type SFTP struct {
	client   *ssh.Client
	sftpConn *sftp.Client
}
type SFTPClient struct {
	client *sftp.Client
}

// 与SFTP服务端建立连接
func (s *SFTP) Connect(Username, Password, Host string, Port int) error {
	//配置客户端
	sshConfig := &ssh.ClientConfig{
		User: Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 注意：生产环境不要用此方式，应验证主机密钥
	}
	// 连接SFTP服务器
	addr := fmt.Sprintf("%s:%d", Host, Port)
	var err error
	s.client, err = ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		log.Printf("连接失败：%v", err)
		return err
	}
	// 打开SFTP会话
	s.sftpConn, err = sftp.NewClient(s.client)
	if err != nil {
		log.Printf("打开SFTP会话失败：%v", err)
		return err
	}
	return nil
}

// Upload 上传文件到 SFTP 服务器
func (s *SFTP) Upload(localPath, remotePath string) error {
	// 打开本地文件
	localFile, err := os.Open(localPath)
	if err != nil {
		log.Printf("打开本地文件失败：%v|%v", localFile, err)
		return err
	}
	defer localFile.Close()

	log.Printf("正在创建远程文件：%s", remotePath)

	// 创建远程文件
	remoteFile, err := s.sftpConn.Create(remotePath)
	if err != nil {
		log.Printf("创建远程文件失败：%v|%v", remoteFile, err)
		return err
	}
	defer remoteFile.Close()

	// 复制文件内容
	_, err = io.Copy(remoteFile, localFile)
	if err != nil {
		log.Printf("复制文件内容失败：%v|%v", remoteFile, err)
	}
	return err
}

// uploadDirectory 递归上传目录
func (s *SFTP) UploadDirectory(localDir, remoteDir string) error {
	err := filepath.Walk(localDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(localDir, path)
		if err != nil {
			return err
		}
		remotePath := filepath.Join(remoteDir, relPath)
		if info.IsDir() {
			// 创建远程目录
			err = s.sftpConn.MkdirAll(remotePath)
			if err != nil {
				log.Printf("创建远程目录失败：%v|%v", remotePath, err)
				return err
			}
		} else {
			// 上传文件
			return s.Upload(path, remotePath)
		}
		return nil
	})
	return err
}

// Download 从 SFTP 服务器下载文件
func (s *SFTP) Download(remotePath, localPath string) error {
	// 打开远程文件
	remoteFile, err := s.sftpConn.Open(remotePath)
	if err != nil {
		return err
	}
	defer remoteFile.Close()

	// 创建本地文件
	//remoteFileName := filepath.Base(remotePath)
	//fullLocalPath := filepath.Join(localPath, remoteFileName)
	localFile, err := os.Create(localPath)
	if err != nil {
		return err
	}
	defer localFile.Close()

	// 复制文件内容
	_, err = io.Copy(localFile, remoteFile)
	return err
}

// ReadDir 读取指定远程目录中的文件和子目录信息
func (s *SFTPClient) ReadDir(remoteDir string) ([]os.FileInfo, error) {
	if s.client == nil {
		return nil, fmt.Errorf("SFTP 客户端未初始化")
	}
	return s.client.ReadDir(remoteDir)
}

// Close 关闭 SFTP 连接
func (s *SFTP) Close() error {
	if s.sftpConn != nil {
		s.sftpConn.Close()
	}
	if s.client != nil {
		s.client.Close()
	}
	return nil
}
