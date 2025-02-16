package main

import (
	"context"
	"fmt"
	"github.com/aviseu/jobs-backoffice/internal/app/domain"
	"log/slog"
	ohttp "net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/aviseu/jobs-backoffice/internal/app/domain/channel"
	"github.com/aviseu/jobs-backoffice/internal/app/domain/imports"
	"github.com/aviseu/jobs-backoffice/internal/app/domain/job"
	"github.com/aviseu/jobs-backoffice/internal/app/gateway"
	"github.com/aviseu/jobs-backoffice/internal/app/http"
	"github.com/aviseu/jobs-backoffice/internal/app/storage"
	"github.com/aviseu/jobs-backoffice/internal/app/storage/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	"github.com/kelseyhightower/envconfig"
	_ "github.com/lib/pq"
)

type config struct {
	DB      storage.Config
	Import  http.Config `split_words:"true"`
	Gateway gateway.Config
	Job     struct {
		Workers int `default:"10"`
		Buffer  int `default:"10"`
	}
	Log struct {
		Level slog.Level `default:"info"`
	}
}

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{})))

	if err := run(context.Background()); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	// load environment variables
	slog.Info("loading environment variables...")
	var cfg config
	if err := envconfig.Process("", &cfg); err != nil {
		return fmt.Errorf("failed to load environment variables: %w", err)
	}

	// configure logging
	slog.Info("configuring logging...")
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.Log.Level}))
	slog.SetDefault(log)

	// setup database
	slog.Info("setting up database...")
	db, err := storage.SetupDatabase(cfg.DB)
	if err != nil {
		return fmt.Errorf("failed to setup database: %w", err)
	}
	defer func(db *sqlx.DB) {
		err := db.Close()
		if err != nil {
			slog.Error(fmt.Errorf("failed to close database connection: %w", err).Error())
		}
	}(db)

	// migrate db
	slog.Info("migrating database...")
	if err := storage.MigrateDB(db); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	// services
	slog.Info("setting up services...")
	chr := postgres.NewChannelRepository(db)
	chs := channel.NewService(chr)
	ir := postgres.NewImportRepository(db)
	is := imports.NewService(ir)
	jr := postgres.NewJobRepository(db)
	js := job.NewService(jr, cfg.Job.Buffer, cfg.Job.Workers)

	f := gateway.NewFactory(js, is, ohttp.DefaultClient, cfg.Gateway, log)
	ia := domain.NewImportAction(chs, is, f)

	// start server
	server := http.SetupServer(ctx, cfg.Import, http.ImportRootHandler(ia, log))
	serverErrors := make(chan error, 1)
	go func() {
		slog.Info("starting server...")
		serverErrors <- server.ListenAndServe()
	}()

	// shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)

	case <-done:
		slog.Info("shutting down server...")

		ctx, cancel := context.WithTimeout(ctx, cfg.Import.ShutdownTimeout)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			return fmt.Errorf("failed to shutdown server: %w", err)
		}
	}

	return nil
}
