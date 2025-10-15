package dos

import "context"

// Provider 定义语言模型服务需要实现的接口
// 中文注释：用于解耦具体 LLM 实现，避免在业务代码中硬编码第三方 SDK
type Provider interface {
	// AnalyzeFile 基于 diff 内容分析函数变更
	AnalyzeFile(ctx context.Context, fileContent string) (LLMResponseList, error)
	// AnalyzeMR 基于 MR 标题与上下文生成审查与建议
	AnalyzeMR(ctx context.Context, title, content string) (*LLMReview, error)
	// GenerateTestCases 根据 MR 标题与分析上下文生成结构化测试用例
	GenerateTestCases(ctx context.Context, mergeRequestTitle string, analysisContext string) ([]TestCase, error)
}
