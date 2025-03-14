package service

import (
	"github.com/google/wire"
	"github.com/toheart/goanalysis/internal/server/iface"
)

// ProviderSet is service providers.
var ProviderSet = wire.NewSet(NewAnalysisService, NewStaticAnalysisService, NewHttpServiceList, NewFileManagerService)

func NewHttpServiceList(staticAnalysisService *StaticAnalysisService, analysisService *AnalysisService, fileManagerService *FileManagerService) []iface.InitGrpcHttp {
	return []iface.InitGrpcHttp{
		staticAnalysisService,
		analysisService,
		fileManagerService,
	}
}
