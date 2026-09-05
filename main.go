package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/medidew/medisite/internal/config"
	"github.com/medidew/medisite/internal/content"
	"github.com/medidew/medisite/internal/logging"
	"github.com/medidew/medisite/internal/web"
	"go.uber.org/zap"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("loading config: %v", err)
	}

	logger, err := logging.New(cfg.LogFile)
	if err != nil {
		log.Fatalf("setting up logging: %v", err)
	}
	defer logger.Sync()

	site, err := content.Load(cfg.ContentDir, logger)
	if err != nil {
		logger.Fatal("loading content", zap.Error(err))
	}
	logger.Info("content loaded", zap.Int("posts", len(site.Posts)), zap.Int("projects", len(site.Projects)))

	srv, err := web.New(site, cfg.TemplatesDir, cfg.StaticDir, logger)
	if err != nil {
		logger.Fatal("starting server", zap.Error(err))
	}

	logger.Info("listening", zap.String("addr", cfg.Addr))
	logger.Fatal("server stopped", zap.Error(http.ListenAndServe(cfg.Addr, srv)))
}
