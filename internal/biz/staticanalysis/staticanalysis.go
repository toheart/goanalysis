package staticanalysis

import (
	"os"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/toheart/goanalysis/internal/biz/entity"
	"github.com/toheart/goanalysis/internal/conf"
	"github.com/toheart/goanalysis/internal/data"
)

// StaticAnalysisBiz 静态分析业务逻辑
type StaticAnalysisBiz struct {
	conf *conf.Biz
	data *data.Data
	log  *log.Helper
}

// NewStaticAnalysisBiz 创建静态分析业务逻辑实例
func NewStaticAnalysisBiz(conf *conf.Biz, data *data.Data, logger log.Logger) *StaticAnalysisBiz {
	return &StaticAnalysisBiz{conf: conf, data: data, log: log.NewHelper(logger)}
}

// VerifyProjectPath 验证项目路径是否存在
func (s *StaticAnalysisBiz) VerifyProjectPath(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// GetStaticDBPath 获取静态分析数据库路径
func (s *StaticAnalysisBiz) GetStaticDBPath() string {
	return s.conf.StaticDBpath
}

// GetHotFunctions 获取热点函数
func (s *StaticAnalysisBiz) GetHotFunctions(sortBy string) ([]entity.HotFunction, error) {
	// 这里实现获取热点函数的逻辑
	// 示例数据
	return []entity.HotFunction{
		{
			Name:      "main.main",
			Package:   "main",
			CallCount: 1,
			TotalTime: "10ms",
			AvgTime:   "10ms",
		},
		{
			Name:      "internal/biz.Process",
			Package:   "internal/biz",
			CallCount: 45,
			TotalTime: "500ms",
			AvgTime:   "11.1ms",
		},
		{
			Name:      "internal/data.Query",
			Package:   "internal/data",
			CallCount: 78,
			TotalTime: "800ms",
			AvgTime:   "10.3ms",
		},
	}, nil
}

// GetFunctionAnalysis 获取函数调用关系分析
func (s *StaticAnalysisBiz) GetFunctionAnalysis(functionName, queryType, path string) ([]entity.FunctionNode, error) {
	// 这里实现获取函数调用关系分析的逻辑
	// 示例数据
	return []entity.FunctionNode{
		{
			ID:        "1",
			Name:      functionName,
			Package:   "main",
			CallCount: 10,
			AvgTime:   "5ms",
			Children:  []entity.FunctionNode{},
		},
	}, nil
}

// GetFunctionCallGraph 获取函数调用关系图
func (s *StaticAnalysisBiz) GetFunctionCallGraph(functionName string, depth int, direction string) ([]entity.FunctionGraphNode, []entity.FunctionGraphEdge, error) {
	// 这里实现获取函数调用关系图的逻辑
	// 示例数据
	nodes := []entity.FunctionGraphNode{
		{
			ID:        "1",
			Name:      functionName,
			Package:   "main",
			CallCount: 1,
			AvgTime:   "10ms",
			NodeType:  "root",
		},
		{
			ID:        "2",
			Name:      "subFunc",
			Package:   "main",
			CallCount: 5,
			AvgTime:   "2ms",
			NodeType:  "callee",
		},
	}

	edges := []entity.FunctionGraphEdge{
		{
			Source:   "1",
			Target:   "2",
			Label:    "calls",
			EdgeType: "root_to_callee",
		},
	}

	return nodes, edges, nil
}
