package biz

import (
	"github.com/google/wire"
	"github.com/toheart/goanalysis/internal/biz/analysis"
	"github.com/toheart/goanalysis/internal/biz/gitlab"
	"github.com/toheart/goanalysis/internal/biz/staticanalysis"
)

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(analysis.NewAnalysisBiz, staticanalysis.NewStaticAnalysisBiz, gitlab.NewGitLabBiz)
