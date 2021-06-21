module github.com/einride/spanner-backup-service

go 1.16

require (
	cloud.google.com/go/spanner v1.20.0
	github.com/blendle/zapdriver v1.3.1
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/googleapis/gax-go/v2 v2.0.5
	github.com/kelseyhightower/envconfig v1.4.0
	go.einride.tech/aip v0.42.0
	go.uber.org/zap v1.17.0
	golang.org/x/net v0.0.0-20210525063256-abc453219eb5 // indirect
	golang.org/x/sys v0.0.0-20210608053332-aa57babbf139 // indirect
	google.golang.org/api v0.48.0 // indirect
	google.golang.org/genproto v0.0.0-20210607140030-00d4fb20b1ae
	google.golang.org/protobuf v1.26.0
	gotest.tools/v3 v3.0.3
)
