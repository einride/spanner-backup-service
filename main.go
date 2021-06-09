package main

import (
	"context"
	"log"
	"net/http"
	"time"

	database "cloud.google.com/go/spanner/admin/database/apiv1"
	"github.com/blendle/zapdriver"
	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap"
	adminpb "google.golang.org/genproto/googleapis/spanner/admin/database/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	config      Config
	adminClient *database.DatabaseAdminClient
	logger      *zap.Logger
}

// Config represents the specific configuration of this service.
type Config struct {
	Project   string
	Instance  string
	Database  string
	Frequency string
}

func main() {
	ctx := context.Background()
	var config Config
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
	server := &Server{
		config:      config,
		adminClient: adminClient,
		logger:      logger,
	}
	http.HandleFunc("/", server.handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func (s *Server) handler(w http.ResponseWriter, r *http.Request) {
	var expireTime time.Time
	// Match retention to the backup frequency
	switch s.config.Frequency {
	case "daily":
		expireTime = time.Now().AddDate(0, 0, 14)
	case "weekly":
		expireTime = time.Now().AddDate(0, 2, 0)
	case "monthly":
		expireTime = time.Now().AddDate(1, 0, 0)
	default:
		s.logger.Error(
			"Unsupported backup frequency",
			zap.String("frequency", s.config.Frequency),
			zap.String("allowed", "daily, weekly, monthly"),
		)
		http.Error(w, "Unsupported backup frequency. Allowed: daily, weekly, monthly", http.StatusBadRequest)
		return
	}

	req := adminpb.CreateBackupRequest{
		Parent:   "projects/" + s.config.Project + "/instances/" + s.config.Instance,
		BackupId: "auto-backup-" + s.config.Database + "-" + time.Now().Format("2006-01-02"),
		Backup: &adminpb.Backup{
			Database:    "projects/" + s.config.Project + "/instances/" + s.config.Instance + "/databases/" + s.config.Database,
			ExpireTime:  timestamppb.New(expireTime),
			VersionTime: timestamppb.Now(),
		},
	}
	if _, err := s.adminClient.CreateBackup(r.Context(), &req); err != nil {
		s.logger.Error(
			"Error when creating backup",
			zap.String("request", req.String()),
			zap.Error(err),
		)
		http.Error(w, "Error when creating backup", 400)
		return
	}
	s.logger.Info("Started backing up database...")
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
