package biz

import (
	"github.com/google/wire"
	"github.com/toheart/goanalysis/internal/biz/analysis"
)

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(analysis.NewAnalysisBiz)
