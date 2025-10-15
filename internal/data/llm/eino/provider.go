package eino

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cloudwego/eino-ext/components/model/openai"
	einomodel "github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/toheart/goanalysis/internal/biz/gitanalysis/dos"
)

// 中文注释：EinoProvider 基于 CloudWeGo Eino 封装的 LLM Provider，实现 dos.Provider 契约
type EinoProvider struct {
	logger *log.Helper
	model  einomodel.BaseChatModel
	// 中文注释：默认请求选项，可覆盖模型名、超时时间等
	defaultOpts []einomodel.Option
}

// NewEinoProvider 创建基于 Eino(OpenAI) 的 Provider
// 中文注释：沿用现有 OpenAI 配置(apiKey/apiBase/model)，支持 Azure/OpenAI 兼容
func NewEinoProvider(apiKey, apiBase, model string) (*EinoProvider, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("missing openai api key")
	}

	// 中文注释：构建 OpenAI ChatModel 配置，设置 BaseURL 以兼容自定义网关
	cfg := &openai.ChatModelConfig{
		APIKey:  apiKey,
		BaseURL: apiBase,
		Model:   model,
		Timeout: 60 * time.Second,
	}

	cm, err := openai.NewChatModel(context.Background(), cfg)
	if err != nil {
		return nil, err
	}

	p := &EinoProvider{
		logger: log.NewHelper(log.With(log.GetLogger(), "module", "eino-provider")),
		model:  cm,
	}
	// 中文注释：默认覆盖模型名（如需），可按需追加更多 eino 选项
	if model != "" {
		p.defaultOpts = append(p.defaultOpts, einomodel.WithModel(model))
	}
	return p, nil
}

// NewEinoProviderWithModel 允许第三方直接注入任意 Eino ChatModel 实现
// 中文注释：通过该构造，外部可自定义模型路由/鉴权/代理等，不受此处限制
func NewEinoProviderWithModel(chat einomodel.BaseChatModel) (*EinoProvider, error) {
	if chat == nil {
		return nil, fmt.Errorf("nil chat model")
	}
	return &EinoProvider{
		logger: log.NewHelper(log.With(log.GetLogger(), "module", "eino-provider")),
		model:  chat,
	}, nil
}

// AnalyzeFile 基于 diff 内容产出严格 JSON 数组（函数级别变更）
// 中文注释：issues 字段用于承载对该函数的“问题清单”，用于 MR 的质量汇总：
// - title：问题标题（一句话概括）
// - severity：问题严重级别（P0 严重/P1 警告/P2 建议）
// - detail：问题细节与定位（必要时给出修复要点）
func (p *EinoProvider) AnalyzeFile(ctx context.Context, fileContent string) (dos.LLMResponseList, error) {
	// 中文注释：提示词约束严格 JSON 数组输出；不得包含额外文本
	sys := schema.SystemMessage(`你是资深 Golang 代码评审与缺陷检测专家。输入是一段 git unified diff（可能是新增/修改）。
请完成：
1) 标出本次 diff 触及到的“全部函数/方法”（即使判断无效也要返回该函数，以便透明记录）。
2) 对每个函数：
   - is_valid：是否为有效改动（函数体语义变化）。
   - reason：有效/无效的判断理由（简洁、可复查）。
   - suggestion：如有改进建议请给出（不必重复 trivial）。
   - issues：问题清单，仅当存在“实际问题”时输出，作为 MR 质量汇总的直接数据来源；若无问题则为空数组。请按如下规则产出：
       · 每个函数最多 3 条问题；超出时优先保留 P0→P1→P2，且合并重复语义问题。
       · 每条问题需给出明确标题与细节；避免泛化或与本次改动无关的噪声。
     问题定义与分级：
       - P0：严重问题（并发安全、资源泄漏、错误吞噬、明显越界/空指针、数据竞争、不可恢复异常处理不当、敏感信息泄露等）。
       - P1：中等问题（边界缺失、错误处理不足、上下文/超时缺失、日志等级不当、复杂度过高影响可维护性等）。
       - P2：建议（命名/注释/可读性、幂等性与健壮性建议等）。
3) 仅输出一个严格 JSON 数组，禁止额外文字/Markdown/代码块。数组元素结构：
[
  {
    "function_name": string,
    "is_valid": boolean,
    "reason": string,
    "suggestion": string,
    "issues": [ { "title": string, "severity": "P0"|"P1"|"P2", "detail": string } ]
  }
]
无效判定（示例）：仅注释/日志/格式化、仅重命名且不影响语义、仅 import 顺序调整、接口声明无实现、仅新增/删除空行。`)
	usr := schema.UserMessage(fileContent)

	msg, err := p.model.Generate(ctx, []*schema.Message{sys, usr}, p.defaultOpts...)
	if err != nil {
		return nil, err
	}
	raw := msg.Content

	var list dos.LLMResponseList
	if err := json.Unmarshal([]byte(raw), &list); err != nil {
		// 中文注释：兼容性清洗，提取首尾 []
		if sanitized := sanitizeJSONArray(raw); sanitized != "" {
			if err2 := json.Unmarshal([]byte(sanitized), &list); err2 == nil {
				return list, nil
			}
		}
		p.logger.Warnf("failed to unmarshal AnalyzeFile json: %v", err)
		return nil, err
	}
	return list, nil
}

// AnalyzeMR 基于 MR 标题+上下文，返回审查与测试建议（严格 JSON 对象）
func (p *EinoProvider) AnalyzeMR(ctx context.Context, title, content string) (*dos.LLMReview, error) {
	sys := schema.SystemMessage(`你是资深 Golang 评审专家。请基于“MR 标题 + 函数改动摘要”输出严格 JSON 对象（仅对象本身）。
字段语义与约束如下：
{
  "review": string,                       // 中文撰写：对本次改动的总体评审（<=300字，覆盖关键风险/影响面/改进建议，避免赘述与代码块）
  "commit_message_score": number,         // 整数 0-100：提交信息质量（范围/上下文/可追溯性/关联需求），越高越好
  "go_issues_score": number,              // 整数 0-100：Go 代码质量与最佳实践符合度（并发/错误处理/边界/可读性），越高越好
  "test_suggestions": [
    {
      "title": string,                    // 中文一句话用例建议标题（清晰可检）
      "reason": string,                   // 中文说明建议原因（关联风险/场景）
      "priority": "P0"|"P1"|"P2",       // P0 严重、P1 中、P2 建议
      "related_functions": [string]       // 仅允许使用输入改动集中出现的函数名；若无则为 []
    }
  ]
}
输出要求：
- 仅输出 JSON 对象；不得包含任何额外文字/注释/Markdown/代码块。
- 所有自然语言字段用中文撰写；数值字段为整数；不得编造不存在的函数名。
- 若无测试建议，请返回 "test_suggestions": []。`)
	usr := schema.UserMessage(fmt.Sprintf("MR 标题: %s\n上下文: %s\n请仅输出 JSON。", title, content))

	msg, err := p.model.Generate(ctx, []*schema.Message{sys, usr}, p.defaultOpts...)
	if err != nil {
		return nil, err
	}
	raw := msg.Content
	var review dos.LLMReview
	if err := json.Unmarshal([]byte(raw), &review); err != nil {
		if sanitized := sanitizeJSONObject(raw); sanitized != "" {
			if err2 := json.Unmarshal([]byte(sanitized), &review); err2 == nil {
				return &review, nil
			}
		}
		p.logger.Warnf("failed to unmarshal AnalyzeMR json: %v", err)
		return nil, err
	}
	return &review, nil
}

// GenerateTestCases 基于受影响函数与调用链生成结构化测试用例（严格 JSON 数组）
func (p *EinoProvider) GenerateTestCases(ctx context.Context, title, content string) ([]dos.TestCase, error) {
	sys := schema.SystemMessage(`你是资深测试架构师。请基于“MR 标题 + 函数改动/上游调用链摘要”输出严格 JSON 数组（仅数组本身）。
数组中每个测试用例对象字段语义如下：
[
  {
    "id": string,                         // 测试用例唯一标识，格式建议：TC-<module>-<序号>
    "module": string,                     // 受影响的模块/子系统名（简短可读）
    "title": string,                      // 中文用例标题（面向验证目标，清晰可检）
    "prerequisites": [string],            // 前置条件列表（环境/数据/配置），无则为 []
    "steps": [string],                    // 测试步骤，原子化可执行，按顺序排列
    "expected": [string],                 // 期望结果，客观可验证，按步骤对应
    "priority": "P0"|"P1"|"P2",        // P0 高优先，P1 中，P2 低
    "related_functions": [string]         // 关联的改动函数名，仅使用输入中出现的函数，未知则为 []
  }
]
生成要求：
- 覆盖 正常/异常/边界 场景；步骤与期望一一对应，避免模糊表达；自然语言字段用中文撰写。
- 仅输出 JSON 数组；不得包含任何额外文字/注释/Markdown/代码块。
- 若无合适用例，请返回空数组 []。`)
	usr := schema.UserMessage(fmt.Sprintf("MR 标题: %s\n上下文: %s\n请仅输出 JSON 数组。", title, content))

	msg, err := p.model.Generate(ctx, []*schema.Message{sys, usr}, p.defaultOpts...)
	if err != nil {
		return nil, err
	}
	raw := msg.Content
	var cases []dos.TestCase
	if err := json.Unmarshal([]byte(raw), &cases); err != nil {
		if sanitized := sanitizeJSONArray(raw); sanitized != "" {
			if err2 := json.Unmarshal([]byte(sanitized), &cases); err2 == nil {
				return cases, nil
			}
		}
		p.logger.Warnf("failed to unmarshal GenerateTestCases json: %v", err)
		return nil, err
	}
	return cases, nil
}

// 中文注释：从文本中提取首个 JSON 数组
func sanitizeJSONArray(s string) string {
	start := strings.Index(s, "[")
	end := strings.LastIndex(s, "]")
	if start >= 0 && end > start {
		return s[start : end+1]
	}
	return ""
}

// 中文注释：从文本中提取首个 JSON 对象
func sanitizeJSONObject(s string) string {
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")
	if start >= 0 && end > start {
		return s[start : end+1]
	}
	return ""
}
