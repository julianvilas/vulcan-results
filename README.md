# Vulcan Results Service

## Requirements
- go
- dep

## Build
```
go get -d -v github.com/adevinta/vulcan-results
cd $GOPATH/src/github.com/adevinta/vulcan-results
dep ensure -v
go get ./...
```

## Config file example:
**config-example.toml**
```
# File where result logs are stored.
# Leave empty (or remove) for STDOUT.
LogFile = ""
Port = 8888
Debug = true

[Storage]
Region = "eu-west-1"
BucketVulnerableReports = "my-vulnerable-reports-bucket"
BucketReports = "my-reports-bucket"
BucketLogs = "my-check-logs-bucket"
LinkBase = "http://example.com/v1"
```

## Run
```
$GOPATH/bin/vulcan-results /path/to/config-example.toml
```

# Docker execute

Those are the variables you have to use:

|Variable|Description|Sample|
|---|---|---|
|PORT|Listen http port|8080|
|DEBUG||true|
|AWS_REGION|aws region|eu-west-1|
|BUCKET_REPORTS|Bucket name to store reports|bucket-reports|
|BUCKET_LOGS|Buckent name to store logs|bucket-logs|
|LINK_BASE|URL used for TBD|http://results/v1|

```bash
docker build . -t vr

# Use the default config.toml customized with env variables.
docker run --env-file ./local.env vr

# Use custom config.toml
docker run -v `pwd`/custom.toml:/app/config.toml vr
```
