package main

import (
	"context"
	"log"
	"net/http"

	database "cloud.google.com/go/spanner/admin/database/apiv1"
	"github.com/blendle/zapdriver"
	"github.com/einride/spanner-backup-service/internal/handler"
	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap"
)

func main() {
	ctx := context.Background()
	var config handler.Config
	if err := envconfig.Process("config", &config); err != nil {
		log.Panic(err)
	}
	logger, cleanup, err := initLogger()
	if err != nil {
		log.Panic(err)
	}
	defer cleanup()
	adminClient, cleanup, err := initAdminClient(ctx, logger)
	if err != nil {
		logger.Panic(
			"createBackup.NewDatabaseAdminClient",
			zap.Error(err),
		)
	}
	defer cleanup()
	server := &handler.Server{
		AdminClient: adminClient,
		Logger:      logger,
	}
	http.HandleFunc("/", server.ServeHTTP)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// initLogger initiates a new zap logger that conforms to the JSON structure for Cloud Logging.
func initLogger() (logger *zap.Logger, cleanup func(), err error) {
	zapConfig := zap.NewProductionConfig()
	zapConfig.EncoderConfig = zapdriver.NewProductionEncoderConfig()
	zapOptions := []zap.Option{
		zapdriver.WrapCore(
			zapdriver.ServiceName("spanner-auto-backup"),
			zapdriver.ReportAllErrors(true),
		)}
	logger, err = zapConfig.Build(zapOptions...)
	if err != nil {
		return nil, nil, err
	}
	logger = logger.WithOptions(zap.AddStacktrace(zap.ErrorLevel))
	logger.Info(
		"logger initialized",
	)
	cleanup = func() {
		logger.Info("closing logger, goodbye")
		_ = logger.Sync()
	}
	return logger, cleanup, nil
}

// initAdminClient initiates a new DatabaseAdminClient which can be used to perform operations on a Cloud Spanner Database.
func initAdminClient(ctx context.Context, logger *zap.Logger) (adminClient *database.DatabaseAdminClient, cleanup func(), err error) {
	adminClient, err = database.NewDatabaseAdminClient(ctx)
	if err != nil {
		return nil, nil, err
	}
	cleanup = func() {
		err := adminClient.Close()
		if err != nil {
			logger.Error(
				"Error when closing Admin Client",
				zap.Error(err),
			)
		}
	}
	return adminClient, cleanup, nil
}
