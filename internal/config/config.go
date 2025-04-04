package config

import (
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"path/filepath"
	"time"
)

type SFTPConfig struct {
	Host            string        `yaml:"host"`
	Port            int           `yaml:"port"`
	Username        string        `yaml:"username"`
	Password        string        `yaml:"password"` // 或使用私钥
	PrivateKey      string        `yaml:"private_key"`
	WatchDir        string        `yaml:"watch_dir"`
	Interval        time.Duration `yaml:"interval"`
	FilePattern     string        `yaml:"file_pattern"` // 监控文件模式
	LogFile         string        `yaml:"logFile"`      //日志文件路径
	Manner          int8          `yaml:"manner"`
	IncrementalFull int8          `yaml:"incrementalFull"`
	WhetherDelete   bool          `yaml:"whetherDelete"`
	SupportCatalogs bool          `yaml:"SupportCatalogs"`
	LocalPaths      string        `yaml:"LocalPaths"`
	RemotePaths     string        `yaml:"RemotePaths"`
	TimeInterval    time.Duration `yaml:"timeInterval"`
}

func Config() (string, string) {
	// 获取可执行文件的完整路径
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	// 获取可执行文件所在的目录
	exPath := filepath.Dir(ex)
	relativePath := "config.yaml"
	return exPath, relativePath
}

type DefaultMailConfigProvider struct{}

func (d *DefaultMailConfigProvider) GetConfig() SFTPConfig {
	FilePath := filepath.Join(Config())
	YamlFile, err := os.Open(FilePath)
	if err != nil {
		log.Printf("配置文件打开失败: %v\n", err)
		return SFTPConfig{}
	}
	defer YamlFile.Close()

	// 解析yaml文件内容到MailConfig
	var YamlContent SFTPConfig
	err = yaml.NewDecoder(YamlFile).Decode(&YamlContent)
	if err != nil {
		log.Printf("配置文件解析失败: %v\n", err)
		return SFTPConfig{}
	}
	// 这里可以根据实际情况初始化 MailConfig
	return YamlContent
}
