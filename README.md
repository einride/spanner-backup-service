# Spanner Backup Service

Go service for implementing scheduled Cloud Spanner backups.

## Usage
To use the service, simply set the desired project, instance, database and the frequency of the backups 
(daily, weekly, monthly) as environment variables with RESOURCE as a prefix. To schedule the backups, use a service like
Cloud Scheduler to send an empty POST request to the service on a set schedule. The service will then create a backup of the 
specified database with a retention to that matches the frequency of the backups. 

The current frequency to retention
time is:

 - Daily frequency, two weeks retention.
 - Weekly frequency, two months retention.
 - Monthly frequency, one year retention. 

## License
[MIT](https://choosealicense.com/licenses/mit/)
