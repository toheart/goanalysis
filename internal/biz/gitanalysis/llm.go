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
	// 通过提示词约束返回严格 JSON 数组，兼容多模型
	prompt := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, `你是资深 Golang 代码变更分析器。输入是一段 git unified diff。
你的任务：识别“本次 diff 触及的所有函数或方法”，并逐一标注其是否为有效改动。不得因为无效而省略条目。
仅允许返回一个 JSON 数组，禁止任何额外文本/解释/Markdown/代码块。数组元素结构：
[
  {
    "function_name": string,   // 函数名；方法必须使用“*Type.Method”或“Type.Method”格式；普通函数直接使用函数名
    "is_valid": boolean,       // 是否为有效改动（函数体语义有变化）
    "reason": string,          // 有效/无效的原因，简洁精准
    "suggestion": string,      // 如存在可改进建议则给出，否则为空字符串
    "issues": [                // 可选：函数级问题列表
      { "title": string, "severity": "P0"|"P1"|"P2", "detail": string }
    ]
  }
]
判定准则（关键且必须遵守）：
- 以下改动均判定为无效（is_valid=false）：仅注释、仅日志、仅空格/格式化、仅重命名且不影响语义、仅 import 顺序调整、接口方法声明（无实现）、仅新增/删除空行。
- 若是接口方法声明或纯声明性改动，请设置 is_valid=false，reason 说明为“接口声明变更”或等价表述。
- 若存在并发/边界/错误处理/超时/上下文/资源释放等风险，可在 issues 中给出条目，severity 必须为 P0|P1|P2。
输出要求（务必满足）：
- 必须覆盖所有在 diff 中出现“函数签名或函数体”变动的函数；即使判断为无效也要返回该函数，设置 is_valid=false。
- 仅当 diff 完全未触及任何函数时，才返回空数组 []。
- 严格仅输出 JSON 数组，不要包含任何额外文字。`),
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
仅允许返回一个 JSON 对象，禁止任何多余文本/解释/Markdown。结构：
{
  "review": string,                          // 对本次改动的总体审查意见，中文
  "commit_message_score": number,            // 0-100
  "go_issues_score": number,                 // 0-100
  "test_suggestions": [
    {
      "title": string,
      "reason": string,
      "priority": "P0"|"P1"|"P2",
      "related_functions": [string]
    }
  ]
}
若无建议，test_suggestions 返回 []。`),
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

// GenerateTestCases 基于有效函数与其上游调用链，生成结构化测试用例（结构化返回）
func (a *LLMAnalyzer) GenerateTestCases(ctx context.Context, title, content string) ([]dos.TestCase, error) {
	// 仅允许输出 JSON 数组，元素为 TestCase 结构
	prompt := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, `你是资深测试架构师。基于 MR 标题与函数改动/上游调用链，生成高质量测试用例列表。
仅允许输出 JSON 数组，禁止任何额外文本/解释/Markdown。元素结构：
[
  {
    "id": string,                 // 格式：TC-<module>-<序号>
    "module": string,
    "title": string,
    "prerequisites": [string],
    "steps": [string],            // 步骤原子化、可执行
    "expected": [string],         // 期望原子化、可验证
    "priority": "P0"|"P1"|"P2",
    "related_functions": [string] // 必须来自上下文函数集合
  }
]
要求：
- 同时覆盖 正常/异常/边界 场景；
- 直接面向函数级改动与其影响面；
- 无用例则返回 []；严格仅输出 JSON 数组。`),
		llms.TextParts(llms.ChatMessageTypeHuman, fmt.Sprintf("MR 标题: %s\n上下文: %s\n请仅输出 JSON 数组。", title, content)),
	}
	completion, err := a.llm.GenerateContent(ctx, prompt)
	if err != nil {
		return nil, err
	}
	raw := completion.Choices[0].Content
	var cases []dos.TestCase
	if err := json.Unmarshal([]byte(raw), &cases); err != nil {
		if sanitized := sanitizeJSONArray(raw); sanitized != "" {
			if err2 := json.Unmarshal([]byte(sanitized), &cases); err2 == nil {
				return cases, nil
			}
		}
		return nil, err
	}
	return cases, nil
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
