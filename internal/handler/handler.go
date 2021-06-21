package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	database "cloud.google.com/go/spanner/admin/database/apiv1"
	"github.com/googleapis/gax-go/v2"
	"go.einride.tech/aip/resourcename"
	"go.uber.org/zap"
	adminpb "google.golang.org/genproto/googleapis/spanner/admin/database/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type adminClient interface {
	CreateBackup(ctx context.Context, req *adminpb.CreateBackupRequest, opts ...gax.CallOption) (*database.CreateBackupOperation, error)
}

type Server struct {
	AdminClient adminClient
	Logger      *zap.Logger
}

// Config represents the specific configuration of this service.
type Config struct {
	Database string
	TTL      string
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var config Config
	err := json.NewDecoder(r.Body).Decode(&config)
	if err != nil {
		s.Logger.Error(
			"Bad Request",
		)
		return
	}
	ttl, err := time.ParseDuration(config.TTL)
	if err != nil {
		http.Error(w, "Error when parsing Time-To-Live", http.StatusBadRequest)
		return
	}
	var project, instance, database string
	if err := resourcename.Sscan(config.Database, "projects/{project}/instances/{instance}/databases/{database}", &project, &instance, &database); err != nil {
		http.Error(w, "Invalid database name.", http.StatusBadRequest)
		return
	}

	if _, err := s.AdminClient.CreateBackup(r.Context(), &adminpb.CreateBackupRequest{
		Parent:   "projects/" + project + "/instances/" + instance,
		BackupId: "spanner-backup-service-" + database + "-" + time.Now().Format("2006-01-02-1504"),
		Backup: &adminpb.Backup{
			Database:    "projects/" + project + "/instances/" + instance + "/databases/" + database,
			ExpireTime:  timestamppb.New(time.Now().Add(ttl)),
			VersionTime: timestamppb.Now(),
		},
	}); err != nil {
		s.Logger.Error(
			"Error when creating backup",
			zap.Error(err),
		)
		http.Error(w, "Error when creating backup", 400)
		return
	}
	s.Logger.Info("Started backing up database...")
}
