// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package commands

import (
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/toheart/goanalysis/internal/biz/analysis"
	"github.com/toheart/goanalysis/internal/biz/chanMgr"
	"github.com/toheart/goanalysis/internal/biz/filemanager"
	"github.com/toheart/goanalysis/internal/biz/staticanalysis"
	"github.com/toheart/goanalysis/internal/conf"
	"github.com/toheart/goanalysis/internal/data"
	"github.com/toheart/goanalysis/internal/data/sqlite"
	"github.com/toheart/goanalysis/internal/server"
	"github.com/toheart/goanalysis/internal/service"
)

// Injectors from wire.go:

// wireApp init kratos application.
func wireApp(confServer *conf.Server, biz *conf.Biz, confData *conf.Data, logger log.Logger) (*kratos.App, func(), error) {
	dataData := data.NewData(logger)
	channelManager := chanMgr.NewChannelManager()
	staticAnalysisBiz := staticanalysis.NewStaticAnalysisBiz(biz, dataData, channelManager, logger)
	fileRepo, err := sqlite.NewFileEntDB(confData)
	if err != nil {
		return nil, nil, err
	}
	fileBiz := filemanager.NewFileBiz(biz, logger, fileRepo)
	staticAnalysisService := service.NewStaticAnalysisService(staticAnalysisBiz, logger)
	analysisBiz := analysis.NewAnalysisBiz(biz, dataData, logger)
	analysisService := service.NewAnalysisService(analysisBiz, logger)
	fileManagerService := service.NewFileManagerService(fileBiz, logger)
	v := service.NewHttpServiceList(staticAnalysisService, analysisService, fileManagerService)
	httpServer := server.NewHTTPServer(confServer, logger, staticAnalysisBiz, fileBiz, v...)
	grpcServer := server.NewGRPCServer(confServer, logger, v...)
	v2 := server.NewServerList(httpServer, grpcServer)
	app := newApp(logger, v2)
	return app, func() {
	}, nil
}
