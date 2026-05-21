package main

import (
	"SamaraAI/common/aihelper"
	"SamaraAI/common/mysql"
	"SamaraAI/common/rabbitmq"
	"SamaraAI/common/redis"
	appconfig "SamaraAI/config"
	"SamaraAI/dao/message"
	"SamaraAI/internal/config"
	"SamaraAI/internal/handler"
	"SamaraAI/internal/svc"
	"flag"
	"fmt"
	"log"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/samara.yaml", "the config file")

func readDataFromDB() error {
	manager := aihelper.GetGlobalManager()
	msgs, err := message.GetAllMessages()
	if err != nil {
		return err
	}
	for i := range msgs {
		m := &msgs[i]
		modelType := "1"
		cfg := make(map[string]interface{})
		helper, err := manager.GetOrCreateAIHelper(m.UserName, m.SessionID, modelType, cfg)
		if err != nil {
			log.Printf("[readDataFromDB] failed to create helper for user=%s session=%s: %v", m.UserName, m.SessionID, err)
			continue
		}
		log.Println("readDataFromDB init: ", helper.SessionID)
		helper.AddMessage(m.Content, m.UserName, m.IsUser, false)
	}
	log.Println("AIHelperManager init success")
	return nil
}

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	bizConf := appconfig.GetConfig()
	if bizConf.MainConfig.Host != "" {
		c.Host = bizConf.MainConfig.Host
	}
	if bizConf.MainConfig.Port != 0 {
		c.Port = bizConf.MainConfig.Port
	}

	if err := mysql.InitMysql(); err != nil {
		log.Println("InitMysql error:", err)
		return
	}

	if err := readDataFromDB(); err != nil {
		log.Println("readDataFromDB error:", err)
		return
	}

	redis.Init()
	log.Println("redis init success")
	rabbitmq.InitRabbitMQ()
	log.Println("rabbitmq init success")

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)
	handler.RegisterCustomHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
