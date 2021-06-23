# Spanner Backup Service

Go service for implementing scheduled Cloud Spanner backups.

## Usage

To use the service, set up some type of scheduler to send a POST request to the service on a set schedule. The request 
must contain the resource name of the database, in the [AIP][resource-names] format, and the time-to-live of the backup. 
The time-to-live must be in the Go [Duration][go-duration] string form.

[resource-names]: https://google.aip.dev/122
[go-duration]: https://golang.org/pkg/time/#Duration

An example setup in Google Cloud Platform using Cloud Scheduler and Cloud Run:

![GCP Service Setup][setup]

[setup]:./docs/spanner-backups-setup.png
An example request would be:
```json
{
  "database":"projects/my-project/instances/my-instance/databases/my-databas",
  "ttl":"336h"
}
```

## License
[MIT](https://choosealicense.com/licenses/mit/)
