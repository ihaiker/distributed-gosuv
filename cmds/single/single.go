package single

import (
	runtimeKit "github.com/ihaiker/gokit/runtime"
	"github.com/ihaiker/sudis/conf"
	"github.com/ihaiker/sudis/master"
	"github.com/ihaiker/sudis/server"
	"github.com/spf13/cobra"
	"time"
)

var Cmd = &cobra.Command{
	Use: "single", Short: "独立模式启动(默认命令)", Long: "独立模式启动，提供管理控制台服务和进程管理服务。",
	RunE: func(cmd *cobra.Command, args []string) error {
		if conf.Config.Server.Key == "" {
			conf.Config.Server.Key = "single"
		}
		listener := runtimeKit.NewListener()
		if err := master.StartAt(listener); err != nil {
			return err
		}
		if err := server.StartAt(listener); err != nil {
			return err
		}
		return listener.WaitTimeout(time.Second * 20)
	},
}
