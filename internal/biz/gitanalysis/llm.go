package gitanalysis

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tmc/langchaingo/httputil"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/toheart/goanalysis/internal/biz/gitanalysis/dos"
)

type LLMAnalyzer struct {
	llm *openai.LLM
}

func NewLLMAnalyzer(apiToken string, apiURL string, model string) (*LLMAnalyzer, error) {
	llm, err := openai.New(
		openai.WithToken(apiToken),
		openai.WithBaseURL(apiURL),
		openai.WithModel(model),
		openai.WithHTTPClient(httputil.DebugHTTPClient),
	)
	if err != nil {
		return nil, err
	}
	return &LLMAnalyzer{
		llm: llm,
	}, nil
}

func (a *LLMAnalyzer) AnalyzeFile(ctx context.Context, fileContent string) (dos.LLMResponseList, error) {
	// 通过提示词约束返回严格 JSON 数组，兼容 DeepSeek
	prompt := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, `你是资深 Golang 开发。基于给定的 git diff，严格识别“有效改动函数”。
只允许返回一个 JSON 数组，数组元素类型为：
[
  {
    "function_name": string,   // 函数名；若为方法，按 Go 规范带接收者
    "is_valid": boolean,       // 是否为有效改动（函数体有语义变化）
    "reason": string,          // 有效/无效的原因
    "suggestion": string       // 如存在可改进建议则给出，否则允许为空字符串
  }
]
判定规则：
- 无效改动：仅注释、仅日志、仅空格/格式化等；
- 其余由上下文判断；
输出要求：
- 必须仅输出上述 JSON 数组，禁止任何额外文本、Markdown 或解释；
- 若没有有效函数，返回空数组 []。`),
		llms.TextParts(llms.ChatMessageTypeHuman, fileContent),
	}
	completion, err := a.llm.GenerateContent(ctx, prompt)
	if err != nil {
		return nil, err
	}
	raw := completion.Choices[0].Content
	responseList := dos.LLMResponseList{}
	if err := json.Unmarshal([]byte(raw), &responseList); err != nil {
		// 兼容性处理：提取首尾方括号之间的内容
		if sanitized := sanitizeJSONArray(raw); sanitized != "" {
			if err2 := json.Unmarshal([]byte(sanitized), &responseList); err2 == nil {
				return responseList, nil
			}
		}
		return nil, err
	}
	return responseList, nil
}

func (a *LLMAnalyzer) AnalyzeMR(ctx context.Context, title, content string) (*dos.LLMReview, error) {
	// 说明：明确要求仅输出 JSON，避免额外文本，确保可解析
	prompt := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, `你是资深 Golang 评审专家。请基于以下 Merge Request（MR）信息与函数改动，输出严格的 JSON。
必须仅返回一个 JSON 对象，禁止添加任何多余文本、解释或 Markdown。
JSON 结构如下：
{
  "review": string,                          // 对本次改动的总体审查意见，中文
  "commit_message_score": number,           // 提交说明质量评分，0-100，越高越好
  "go_issues_score": number,                // Go 代码问题评分，0-100，越高越好
  "test_suggestions": [
    {
      "title": string,                      // 测试建议标题
      "reason": string,                     // 建议原因
      "priority": "P0"|"P1"|"P2",      // 优先级
      "related_functions": [string]         // 相关函数名列表
    }
  ]
}
评分口径：
- commit_message_score：考虑主题清晰度、动机/范围说明、可追踪性
- go_issues_score：考虑潜在并发/边界/错误处理/日志与监控规范等问题
`),
		llms.TextParts(llms.ChatMessageTypeHuman, fmt.Sprintf(`MR 标题: %s
上下文: %s
请仅输出 JSON。`, title, content)),
	}
	// 仅用 prompt 约束为 JSON，避免与 AnalyzeFile 的 JSON Schema 互相干扰
	completion, err := a.llm.GenerateContent(ctx, prompt)
	if err != nil {
		return nil, err
	}
	var review dos.LLMReview
	raw := completion.Choices[0].Content
	if err := json.Unmarshal([]byte(raw), &review); err != nil {
		// 兼容性处理：提取花括号对象
		if sanitized := sanitizeJSONObject(raw); sanitized != "" {
			if err2 := json.Unmarshal([]byte(sanitized), &review); err2 == nil {
				return &review, nil
			}
		}
		return nil, err
	}
	return &review, nil
}

// sanitizeJSONArray 从文本中提取第一个 JSON 数组
func sanitizeJSONArray(s string) string {
	start := strings.Index(s, "[")
	end := strings.LastIndex(s, "]")
	if start >= 0 && end > start {
		return s[start : end+1]
	}
	return ""
}

// sanitizeJSONObject 从文本中提取第一个 JSON 对象
func sanitizeJSONObject(s string) string {
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")
	if start >= 0 && end > start {
		return s[start : end+1]
	}
	return ""
}
