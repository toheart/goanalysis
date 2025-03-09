package service

import (
	"github.com/google/wire"
	"github.com/toheart/goanalysis/internal/server/iface"
)

// ProviderSet is service providers.
var ProviderSet = wire.NewSet(NewAnalysisService, NewStaticAnalysisService, NewHttpServiceList)

func NewHttpServiceList(s *StaticAnalysisService, a *AnalysisService) []iface.InitGrpcHttp {
	return []iface.InitGrpcHttp{
		s,
		a,
	}
}
