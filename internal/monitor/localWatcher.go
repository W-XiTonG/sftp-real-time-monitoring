package monitor

import (
	"github.com/fsnotify/fsnotify"
	"log"
	"path/filepath"
	"sftp-monitor/internal/handler"
	"sftp-monitor/internal/sftpClient"
)

func LocalWatcher(LocalPaths, RemotePaths, Username, Password, Host string, Port int, Manner int8, SupportCatalogs, whetherDelete bool) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Printf("创建监控器失败：%v", err)
		log.Fatal(err)
	}
	defer watcher.Close()
	// 要监控的目录
	directoryToWatch := LocalPaths
	err = watcher.Add(directoryToWatch)
	if err != nil {
		log.Printf("监控目录添加失败：%v", err)
		log.Fatal(err)
	}
	SftpClient := sftpClient.SFTP{}
	log.Println("正在检查SFTP连接……")
	err = SftpClient.Connect(Username, Password, Host, Port)
	if err != nil {
		log.Println("远程SFTP连接失败，请检查配置！")
		log.Fatal(err)
	} else {
		SftpClient.Close()
	}
	log.Printf("正在监控本地目录: %s", directoryToWatch)
	// 启动一个 goroutine 来处理监控事件
	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				// 提取路径中的最后一个元素
				baseName := filepath.Base(event.Name)
				log.Printf("事件发生: %s，变化的文件/目录: %s", event.Op.String(), event.Name)
				if event.Has(fsnotify.Write) {
					log.Printf("文件 %s 被修改", event.Name)
				}
				if event.Has(fsnotify.Create) {
					log.Printf("文件/目录 %s 被创建", event.Name)
					if Manner == 1 {
						handler.LocalUpload(event.Name, RemotePaths, baseName, Username, Password, Host, Port, SupportCatalogs, whetherDelete)
					}
				}
				if event.Has(fsnotify.Remove) {
					log.Printf("文件/目录 %s 被删除", event.Name)
				}
				if event.Has(fsnotify.Rename) {
					log.Printf("文件/目录 %s 被重命名", event.Name)
				}
				if event.Has(fsnotify.Chmod) {
					log.Printf("文件/目录 %s 的权限被更改", event.Name)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Printf("监控出错: %s", err)
			}
		}
	}()
	// 保持程序运行
	<-done
}
