package main

import (
	"simple-chatroom/models"
	"simple-chatroom/router"
	"simple-chatroom/utils"
	"time"

	"github.com/spf13/viper"
)

func main() {
	utils.InitConfig()
	utils.InitMySQL()
	utils.InitRedis()
	// 初始化定时器
	utils.Timer(time.Duration(viper.GetInt("timeout.DelayHeartbeat"))*time.Second, time.Duration(viper.GetInt("timeout.HeartbeatHz"))*time.Second, models.CleanConnection, "")
	r := router.Router()
	r.Run(viper.GetString("port.server.port")) // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
