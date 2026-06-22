package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	mcpclient "SamaraAI/common/mcp/client"
	mcpserver "SamaraAI/common/mcp/server"
)

func main() {
	mode := flag.String("mode", "", "运行模式: server 或 client")
	httpAddr := flag.String("http-addr", ":8081", "HTTP服务器地址")
	city := flag.String("city", "", "要查询天气的城市名称")
	flag.Parse()

	if *mode == "" {
		fmt.Println("Error: 您必须指定模式使用 --mode (server 或 client)")
		flag.Usage()
		os.Exit(1)
	}

	if *mode == "server" {
		fmt.Println("启动 MCP 服务器...")
		if err := mcpserver.StartServer(*httpAddr); err != nil {
			log.Fatalf("服务器错误: %v", err)
		}
	} else if *mode == "client" {
		if *city == "" {
			fmt.Println("Error: 您必须指定城市名称使用 --city")
			flag.Usage()
			os.Exit(1)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		httpURL := fmt.Sprintf("http://localhost%s/mcp", *httpAddr)
		if (*httpAddr)[0] != ':' {
			httpURL = fmt.Sprintf("http://%s/mcp", *httpAddr)
		}

		client, err := mcpclient.NewMCPClient(httpURL)
		if err != nil {
			log.Fatalf("创建客户端失败: %v", err)
		}
		defer client.Close()

		if _, err := client.Initialize(ctx); err != nil {
			log.Fatalf("初始化失败: %v", err)
		}

		if err := client.Ping(ctx); err != nil {
			log.Fatalf("健康检查失败: %v", err)
		}

		result, err := client.CallWeatherTool(ctx, *city)
		if err != nil {
			log.Fatalf("调用工具失败: %v", err)
		}

		fmt.Println("\n天气查询结果:")
		fmt.Println(client.GetToolResultText(result))
		fmt.Println("\n客户端初始化成功。正在关闭...")
	}
}
