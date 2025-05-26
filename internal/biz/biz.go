package biz

import (
	"github.com/google/wire"
	"github.com/toheart/goanalysis/internal/biz/analysis"
	"github.com/toheart/goanalysis/internal/biz/chanMgr"
	"github.com/toheart/goanalysis/internal/biz/filemanager"
	"github.com/toheart/goanalysis/internal/biz/staticanalysis"
)

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(
	analysis.NewAnalysisBiz,
	staticanalysis.NewStaticAnalysisBiz,
	chanMgr.NewChannelManager,
	filemanager.NewFileBiz,
)
