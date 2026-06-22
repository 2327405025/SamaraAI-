package aihelper

import (
	"SamaraAI/internal/config"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
	"sync"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
)

const maxMCPAgentSteps = 5

// MCPModel 通过 MCP 协议动态发现工具，多轮调用后再回答。
type MCPModel struct {
	llm         model.ToolCallingChatModel
	toolLLM     model.ToolCallingChatModel
	mcpClient   *client.Client
	username    string
	mcpBaseURL  string
	toolCatalog string
	toolsMu     sync.Mutex
	toolsReady  bool
}

func NewMCPModel(ctx context.Context, username string) (*MCPModel, error) {
	conf := config.Get()
	modelName := conf.Mcp.ChatModelName
	if modelName == "" {
		modelName = conf.OpenAI.Model
	}
	if modelName == "" {
		modelName = conf.Rag.ChatModelName
	}
	baseURL := conf.Rag.BaseUrl
	if conf.OpenAI.BaseUrl != "" {
		baseURL = conf.OpenAI.BaseUrl
	}

	llm, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL: baseURL,
		Model:   modelName,
		APIKey:  conf.OpenAI.ApiKey,
	})
	if err != nil {
		return nil, fmt.Errorf("create mcp model failed: %v", err)
	}

	mcpBaseURL := conf.Mcp.BaseURL
	if mcpBaseURL == "" {
		mcpBaseURL = "http://localhost:8081/mcp"
	}

	m := &MCPModel{
		llm:        llm,
		mcpBaseURL: mcpBaseURL,
		username:   username,
	}
	if err := m.syncToolsFromServer(ctx); err != nil {
		log.Printf("MCP sync tools warning (will retry on request): %v", err)
	}
	return m, nil
}

func (m *MCPModel) syncToolsFromServer(ctx context.Context) error {
	m.toolsMu.Lock()
	defer m.toolsMu.Unlock()

	cli, err := m.getMCPClientLocked(ctx)
	if err != nil {
		return err
	}

	result, err := cli.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		return fmt.Errorf("list mcp tools: %w", err)
	}

	schemaTools, err := convertMCPToolsToSchema(result.Tools)
	if err != nil {
		return err
	}

	toolLLM, err := m.llm.WithTools(schemaTools)
	if err != nil {
		return fmt.Errorf("bind tools to llm: %w", err)
	}

	m.toolLLM = toolLLM
	m.toolCatalog = buildToolCatalog(result.Tools)
	m.toolsReady = len(schemaTools) > 0
	log.Printf("MCP synced %d tools: %s", len(result.Tools), toolNames(result.Tools))
	return nil
}

func toolNames(tools []mcp.Tool) string {
	names := make([]string, len(tools))
	for i, t := range tools {
		names[i] = t.Name
	}
	return strings.Join(names, ", ")
}

func (m *MCPModel) ensureTools(ctx context.Context) error {
	if m.toolsReady && m.toolLLM != nil {
		return nil
	}
	return m.syncToolsFromServer(ctx)
}

func (m *MCPModel) getMCPClient(ctx context.Context) (*client.Client, error) {
	m.toolsMu.Lock()
	defer m.toolsMu.Unlock()
	return m.getMCPClientLocked(ctx)
}

func (m *MCPModel) getMCPClientLocked(ctx context.Context) (*client.Client, error) {
	if m.mcpClient != nil {
		return m.mcpClient, nil
	}
	httpTransport, err := transport.NewStreamableHTTP(m.mcpBaseURL)
	if err != nil {
		return nil, fmt.Errorf("create mcp transport failed: %v", err)
	}
	m.mcpClient = client.NewClient(httpTransport)

	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcp.Implementation{
		Name:    "MCP-Go AIHelper Client",
		Version: "1.0.0",
	}
	initRequest.Params.Capabilities = mcp.ClientCapabilities{}

	if _, err := m.mcpClient.Initialize(ctx, initRequest); err != nil {
		return nil, fmt.Errorf("mcp client initialize failed: %v", err)
	}
	return m.mcpClient, nil
}

func (m *MCPModel) buildConversationMessages(messages []*schema.Message) []*schema.Message {
	out := make([]*schema.Message, 0, len(messages)+1)
	out = append(out, schema.SystemMessage(buildMCPSystemPrompt(m.toolCatalog)))
	out = append(out, messages...)
	return out
}

func (m *MCPModel) GenerateResponse(ctx context.Context, messages []*schema.Message) (*schema.Message, error) {
	if len(messages) == 0 {
		return nil, fmt.Errorf("no messages provided")
	}
	if err := m.ensureTools(ctx); err != nil {
		return nil, err
	}
	if m.toolLLM == nil {
		return nil, fmt.Errorf("mcp server has no tools available")
	}

	working, direct, err := m.runToolRounds(ctx, m.buildConversationMessages(messages))
	if err != nil {
		return nil, err
	}
	if direct != nil {
		return direct, nil
	}
	resp, err := m.toolLLM.Generate(ctx, working)
	if err != nil {
		return nil, fmt.Errorf("mcp final generate failed: %v", err)
	}
	return resp, nil
}

func (m *MCPModel) StreamResponse(ctx context.Context, messages []*schema.Message, cb StreamCallback) (string, error) {
	if len(messages) == 0 {
		return "", fmt.Errorf("no messages provided")
	}
	if err := m.ensureTools(ctx); err != nil {
		return "", err
	}
	if m.toolLLM == nil {
		return "", fmt.Errorf("mcp server has no tools available")
	}

	working, direct, err := m.runToolRounds(ctx, m.buildConversationMessages(messages))
	if err != nil {
		return "", err
	}
	if direct != nil {
		if direct.Content != "" {
			cb(direct.Content)
		}
		return direct.Content, nil
	}

	stream, err := m.toolLLM.Stream(ctx, working)
	if err != nil {
		return "", fmt.Errorf("mcp final stream failed: %v", err)
	}
	defer stream.Close()

	var full strings.Builder
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("mcp stream recv failed: %v", err)
		}
		if msg != nil && msg.Content != "" {
			full.WriteString(msg.Content)
			cb(msg.Content)
		}
	}
	return full.String(), nil
}

// runToolRounds 执行多轮工具调用；若模型直接回答则返回 direct，否则返回待合成的上下文。
func (m *MCPModel) runToolRounds(ctx context.Context, msgs []*schema.Message) ([]*schema.Message, *schema.Message, error) {
	cli, err := m.getMCPClient(ctx)
	if err != nil {
		return nil, nil, err
	}

	working := append([]*schema.Message(nil), msgs...)
	for step := 0; step < maxMCPAgentSteps; step++ {
		resp, err := m.toolLLM.Generate(ctx, working)
		if err != nil {
			return nil, nil, fmt.Errorf("mcp agent step %d failed: %w", step, err)
		}
		if len(resp.ToolCalls) == 0 {
			return working, resp, nil
		}

		working = append(working, resp)
		for _, tc := range resp.ToolCalls {
			args, err := parseToolCallArgs(tc.Function.Arguments)
			if err != nil {
				return nil, nil, fmt.Errorf("parse tool args for %s: %w", tc.Function.Name, err)
			}
			result, err := m.callMCPTool(ctx, cli, tc.Function.Name, args)
			if err != nil {
				log.Printf("MCP tool %s failed: %v", tc.Function.Name, err)
				result = fmt.Sprintf("工具调用失败: %v", err)
			}
			working = append(working, schema.ToolMessage(result, tc.ID, schema.WithToolName(tc.Function.Name)))
		}
	}
	return working, nil, nil
}

func parseToolCallArgs(raw string) (map[string]interface{}, error) {
	if strings.TrimSpace(raw) == "" {
		return map[string]interface{}{}, nil
	}
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &args); err != nil {
		return nil, err
	}
	return args, nil
}

func (m *MCPModel) callMCPTool(ctx context.Context, cli *client.Client, toolName string, args map[string]interface{}) (string, error) {
	result, err := cli.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      toolName,
			Arguments: args,
		},
	})
	if err != nil {
		return "", err
	}
	var text string
	for _, content := range result.Content {
		if textContent, ok := content.(mcp.TextContent); ok {
			text += textContent.Text + "\n"
		}
	}
	return strings.TrimSpace(text), nil
}

func (m *MCPModel) GetModelType() string { return "3" }

func (m *MCPModel) Close() {
	if m.mcpClient != nil {
		m.mcpClient.Close()
	}
}
