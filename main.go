package main

import (
	"SamaraAI/common/aihelper"
	"SamaraAI/common/mysql"
	"SamaraAI/common/rabbitmq"
	"SamaraAI/common/redis"
	"SamaraAI/dao/message"
	"SamaraAI/internal/config"
	"SamaraAI/internal/handler"
	"SamaraAI/internal/svc"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/samara.yaml", "the config file")

func warnDeprecatedConfig() {
	if _, err := os.Stat("config/config.toml"); err == nil {
		log.Println("WARN: config/config.toml 已废弃，请仅使用 etc/samara.yaml")
	}
}

func readDataFromDB() error {
	cfg := config.Get()
	manager := aihelper.GetGlobalManager()
	msgs, err := message.GetAllMessages()
	if err != nil {
		return err
	}
	helperCfg := map[string]interface{}{
		"apiKey": cfg.OpenAI.ApiKey,
	}
	for i := range msgs {
		m := &msgs[i]
		helperCfg["username"] = m.UserName
		helper, err := manager.GetOrCreateAIHelper(m.UserName, m.SessionID, "1", helperCfg)
		if err != nil {
			log.Printf("[readDataFromDB] failed to create helper for user=%s session=%s: %v", m.UserName, m.SessionID, err)
			continue
		}
		helper.AddMessage(m.Content, m.UserName, m.IsUser, false)
	}
	log.Println("AIHelperManager init success")
	return nil
}

func main() {
	flag.Parse()
	warnDeprecatedConfig()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	config.InitGlobal(&c)

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

	fmt.Printf("Starting %s at %s:%d...\n", c.AppName, c.Host, c.Port)
	server.Start()
}
