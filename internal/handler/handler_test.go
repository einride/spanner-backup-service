package handler

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	database "cloud.google.com/go/spanner/admin/database/apiv1"
	"github.com/googleapis/gax-go/v2"
	"go.uber.org/zap/zaptest"
	adminpb "google.golang.org/genproto/googleapis/spanner/admin/database/v1"
	"gotest.tools/v3/assert"
)

func TestHandler(t *testing.T) {
	logger := zaptest.NewLogger(t)
	t.Cleanup(func() {
		_ = logger.Sync()
	})
	cases := []struct {
		Name      string
		Body      []byte
		BackupErr error
		Expected  int
	}{
		{
			"Normal Request",
			[]byte(`{"database":"projects/test/instances/test/databases/test","ttl":"2h"}`),
			nil,
			200,
		},
		{
			"Invalid TTL",
			[]byte(`{"database":"projects/test/instances/test/databases/test","ttl":"2d"}`),
			nil,
			400,
		},
		{
			"Invalid Database Name",
			[]byte(`{"database":"projects//instances/test/databases/test","ttl":"2h"}`),
			fmt.Errorf("no such database exists"),
			400,
		},
		{
			"Invalid Database Name",
			[]byte(`{"database":"wrong/test/instances/test/databases/test","ttl":"2h"}`),
			fmt.Errorf("no such database exists"),
			400,
		},
		{
			"Invalid JSON",
			[]byte(`{"wrong":"projects/test/instances/test/databases/test","ttl":"2h"}`),
			fmt.Errorf("no such database exists"),
			400,
		},
	}
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			server := Server{
				Logger: logger,
				AdminClient: &adminClientMock{
					err: tc.BackupErr,
				},
			}
			//nolint: noctx
			req, err := http.NewRequest("POST", "/", bytes.NewBuffer(tc.Body))
			assert.NilError(t, err)
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(server.ServeHTTP)
			handler.ServeHTTP(rr, req)
			assert.Equal(t, rr.Code, tc.Expected)
		})
	}
}

type adminClientMock struct {
	createBackupRequest  *adminpb.CreateBackupRequest
	createBackupResponse *database.CreateBackupOperation
	err                  error
}

func (m *adminClientMock) CreateBackup(
	_ context.Context,
	req *adminpb.CreateBackupRequest,
	_ ...gax.CallOption,
) (*database.CreateBackupOperation, error) {
	m.createBackupRequest = req
	if m.err != nil {
		return nil, m.err
	}
	return m.createBackupResponse, nil
}
