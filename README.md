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
