package server

import (
	"context"

	"github.com/MaroIshiku/dyniku/internal/models"
	"github.com/qdm12/goservices/httpserver"
)

func New(ctx context.Context, address, rootURL string, db Database,
	logger Logger, runner UpdateForcer,
	configPath, dataDir string, buildInfo models.BuildInformation,
) (server *httpserver.Server, err error) {
	return httpserver.New(httpserver.Settings{
		Handler: newHandler(ctx, rootURL, db, runner, configPath, dataDir, buildInfo),
		Address: &address,
		Logger:  logger,
	})
}
