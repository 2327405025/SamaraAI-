package aihelper

import (
	"fmt"
	"strings"

	"github.com/cloudwego/eino/schema"
	mcpapi "github.com/mark3labs/mcp-go/mcp"
)

func convertMCPToolsToSchema(tools []mcpapi.Tool) ([]*schema.ToolInfo, error) {
	out := make([]*schema.ToolInfo, 0, len(tools))
	for _, tool := range tools {
		info, err := convertMCPTool(tool)
		if err != nil {
			return nil, fmt.Errorf("convert tool %s: %w", tool.Name, err)
		}
		out = append(out, info)
	}
	return out, nil
}

func convertMCPTool(tool mcpapi.Tool) (*schema.ToolInfo, error) {
	info := &schema.ToolInfo{
		Name: tool.Name,
		Desc: tool.Description,
	}
	if len(tool.InputSchema.Properties) == 0 {
		return info, nil
	}

	params := make(map[string]*schema.ParameterInfo)
	requiredSet := make(map[string]struct{}, len(tool.InputSchema.Required))
	for _, name := range tool.InputSchema.Required {
		requiredSet[name] = struct{}{}
	}

	for name, rawProp := range tool.InputSchema.Properties {
		prop, ok := rawProp.(map[string]interface{})
		if !ok {
			continue
		}
		pi := &schema.ParameterInfo{Type: schema.String}
		if desc, ok := prop["description"].(string); ok {
			pi.Desc = desc
		}
		if typ, ok := prop["type"].(string); ok {
			pi.Type = jsonTypeToSchemaType(typ)
		}
		if _, ok := requiredSet[name]; ok {
			pi.Required = true
		}
		params[name] = pi
	}
	if len(params) > 0 {
		info.ParamsOneOf = schema.NewParamsOneOfByParams(params)
	}
	return info, nil
}

func jsonTypeToSchemaType(t string) schema.DataType {
	switch t {
	case "integer", "number":
		return schema.Number
	case "boolean":
		return schema.Boolean
	case "array":
		return schema.Array
	case "object":
		return schema.Object
	default:
		return schema.String
	}
}

func buildToolCatalog(tools []mcpapi.Tool) string {
	if len(tools) == 0 {
		return "（当前 MCP 服务未提供任何工具）"
	}
	var b strings.Builder
	for i, t := range tools {
		fmt.Fprintf(&b, "%d. %s", i+1, t.Name)
		if t.Description != "" {
			fmt.Fprintf(&b, "：%s", t.Description)
		}
		b.WriteString("\n")
	}
	return strings.TrimSpace(b.String())
}

func buildMCPSystemPrompt(catalog string) string {
	return fmt.Sprintf(`你是 SamaraAI 的 MCP 工具助手。你的职责是通过 MCP 工具获取外部/实时信息，再回答用户。

当前 MCP 服务提供的工具：
%s

核心规则：
1. 用户问题若需要外部数据、实时信息或工具能力，必须先调用合适工具，仅基于工具返回结果作答。
2. 结合完整对话理解意图（例如上文已提到城市，后续「好的」表示同意查询）。
3. 若没有合适工具，或工具调用失败，必须明确告知用户「当前无法通过工具获取该信息」，禁止用训练数据编造日期、天气、新闻等事实。
4. 工具结果返回后，用简洁中文总结。`, catalog)
}
