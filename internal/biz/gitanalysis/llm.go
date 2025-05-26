package gitanalysis

import (
	"context"
	"encoding/json"
	"fmt"

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
	openai.ResponseFormatJSON = formatOutput()
	return &LLMAnalyzer{
		llm: llm,
	}, nil
}

func (a *LLMAnalyzer) AnalyzeFile(ctx context.Context, fileContent string) (dos.LLMResponseList, error) {

	prompt := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, `作为一个golang专业开发人员,
		请根据git diff中的内容中获取所有被改动的函数,注意改动的函数必须是在函数体内部进行有效修改;
		无效修改包含:
		1. 添加注释;
		2. 添加日志;
		3. 语句中添加空格
		4. 其余操作可通过上下文自行判断是否有效;
		其他要求：
		* 对有效函数进行分析，如果改动存在不合理或者代码不够清晰的情况，给出建议;
		* 如果不存在有效函数，则返回空列表;
		* 快速返回结果,不要继续输出空字符串;
		* 输出函数名要求: 如果为结构体方法，按照golang语言规范输出; 如果为普通函数，需要输出函数名;
		`),
		llms.TextParts(llms.ChatMessageTypeHuman, fileContent),
	}
	completion, err := a.llm.GenerateContent(ctx, prompt, llms.WithJSONMode())
	if err != nil {
		return nil, err
	}
	responseList := dos.LLMResponseList{}
	err = json.Unmarshal([]byte(completion.Choices[0].Content), &responseList)
	if err != nil {
		return nil, err
	}
	return responseList, nil
}

func (a *LLMAnalyzer) AnalyzeMR(ctx context.Context, title, content string) (string, error) {
	prompt := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, `作为一个golang专业开发人员,
		根据merge Request信息以及函数改动分析，给出审查意见以及测试用例建议。
		`),
		llms.TextParts(llms.ChatMessageTypeHuman, fmt.Sprintf(`merge Request主题: %s, 描述信息: %s`, title, content)),
	}
	completion, err := a.llm.GenerateContent(ctx, prompt)
	if err != nil {
		return "", err
	}
	return completion.Choices[0].Content, nil
}

// 修改为格式化输出
func formatOutput() *openai.ResponseFormat {
	return &openai.ResponseFormat{
		Type: "json_schema",
		JSONSchema: &openai.ResponseFormatJSONSchema{
			Name: "object",
			Schema: &openai.ResponseFormatJSONSchemaProperty{
				Type: "array",
				Items: &openai.ResponseFormatJSONSchemaProperty{
					Type: "object",
					Properties: map[string]*openai.ResponseFormatJSONSchemaProperty{
						"function_name": {
							Type:        "string",
							Description: "The name of the function",
						},
						"is_valid": {
							Type:        "boolean",
							Description: "Whether the change is valid",
						},
						"reason": {
							Type:        "string",
							Description: "reason of the change",
						},
						"suggestion": {
							Type:        "string",
							Description: "suggestion of the change",
						},
					},
					AdditionalProperties: false,
					Required:             []string{"function_name", "is_valid", "reason", "suggestion"},
				},
			},
			Strict: true,
		},
	}
}
