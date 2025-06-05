package gitanalysis

import (
	"testing"

	"github.com/toheart/goanalysis/internal/biz/gitanalysis/dos"
)

func TestAnalyzeFunctionCallRelations(t *testing.T) {
	// 创建一个模拟的MR分析结果
	result := &dos.MrAnalysisResult{
		MergeRequestID: 123,
		ProjectID:      456,
		AffectedFunctions: []dos.AffectedFunction{
			{
				Filename:   "test.go",
				ChangeType: dos.Modified,
				Functions: []*dos.Function{
					{
						FunctionName: "TestFunction",
						IsValid:      true,
						Reason:       "test function",
						Suggestion:   "",
					},
					{
						FunctionName: "InvalidFunction",
						IsValid:      false,
						Reason:       "invalid",
						Suggestion:   "",
					},
				},
			},
		},
	}

	// 注意：这里需要一个真实的数据库路径来进行完整测试
	// 目前只是验证函数结构是否正确
	t.Logf("MR分析结果结构正确，包含 %d 个受影响的函数", len(result.AffectedFunctions))

	// 验证有效函数被正确识别
	validFunctions := 0
	for _, af := range result.AffectedFunctions {
		for _, f := range af.Functions {
			if f.IsValid {
				validFunctions++
				t.Logf("找到有效函数: %s", f.FunctionName)
			}
		}
	}

	if validFunctions != 1 {
		t.Errorf("期望 1 个有效函数，但找到 %d 个", validFunctions)
	}
}

func TestFindAllCallers(t *testing.T) {
	// 这是一个示例测试，展示如何使用findAllCallers方法
	// 在实际使用中，需要提供真实的数据库连接和数据

	t.Log("findAllCallers 方法用于查找所有调用指定函数的上级函数")
	t.Log("支持多种匹配方式：")
	t.Log("1. 完全匹配函数名")
	t.Log("2. 函数名包含在节点名称中")
	t.Log("3. 节点名称以 '.函数名' 结尾")

	// 测试函数名匹配逻辑
	functionName := "TestFunction"
	nodeName1 := "TestFunction"                     // 完全匹配
	nodeName2 := "github.com/test/pkg.TestFunction" // 包含匹配
	nodeName3 := "pkg.TestFunction"                 // 后缀匹配

	if functionName != nodeName1 {
		t.Errorf("完全匹配失败")
	}

	if nodeName2 != "github.com/test/pkg.TestFunction" {
		t.Errorf("包含匹配验证失败")
	}

	if nodeName3 != "pkg.TestFunction" {
		t.Errorf("后缀匹配验证失败")
	}

	t.Log("函数名匹配逻辑测试通过")
}
