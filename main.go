package main

import (
	"sftp-monitor/internal/config"
	"sftp-monitor/internal/monitor"
	"sftp-monitor/internal/utils"
)

func main() {
	Config := &config.DefaultMailConfigProvider{}
	logGer := utils.LogGer{}
	logGer.Init(Config.GetConfig().LogFile)
	// 创建 SFTP 对象
	//sftp := &sftpClient.SFTP{}
	//ok := sftp.Connect(Config)
	switch Config.GetConfig().Manner {
	case 1:
		monitor.LocalWatcher(
			Config.GetConfig().LocalPaths,
			Config.GetConfig().RemotePaths,
			Config.GetConfig().Username,
			Config.GetConfig().Password,
			Config.GetConfig().Host,
			Config.GetConfig().Port,
			Config.GetConfig().Manner,
			Config.GetConfig().SupportCatalogs,
			Config.GetConfig().WhetherDelete,
		)
	case 2:
		monitor.RemoteWatcher(
			Config.GetConfig().RemotePaths,
			Config.GetConfig().LocalPaths,
			Config.GetConfig().Username,
			Config.GetConfig().Password,
			Config.GetConfig().Host,
			Config.GetConfig().Port,
			Config.GetConfig().TimeInterval,
		)
	}
}
