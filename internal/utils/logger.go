package utils

import (
	"io"
	"log"
	"os"
)

type LogGer struct {
	LogFile *os.File
}

// Init方法用于初始化日志记录器
func (l *LogGer) Init(LogFile string) {
	var err error
	l.LogFile, err = os.OpenFile(LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	// 尝试获取文件信息
	_, err = os.Stat(LogFile)
	if os.IsNotExist(err) {
		// 文件不存在，创建文件
		var file *os.File
		file, err = os.Create(LogFile)
		if err != nil {
			log.Printf("创建文件 %s 失败: %v", LogFile, err)
		}
		// 关闭文件
		file.Close()
		log.Printf("文件 %s 已创建。\n", LogFile)
		return
	} else if err != nil {
		// 其他错误
		log.Printf("检查文件 %s 时出错: %v", LogFile, err)
	}
	// 创建一个 MultiWriter，将日志同时输出到标准输出和日志文件
	multiWriter := io.MultiWriter(os.Stdout, l.LogFile)
	// 设置日志输出至multiWriter
	log.SetOutput(multiWriter)
	// 设置日志标识位，添加日期和时间
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}
