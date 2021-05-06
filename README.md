# Vulcan Results Service

## Requirements
- go 1.13.x

## Build & Install
```
GO111MODULE=on go get github.com/adevinta/vulcan-results/cmd/vulcan-result
```

## Build, regenerate code from Goa DSL & install in a Docker container
```
docker run -ti golang:1.13-alpine /bin/sh
apk add git
# Cloning out of the GOPATH
cd /tmp
git clone git://github.com/adevinta/vulcan-results.git
cd vulcan-results
go mod download
go install github.com/goadesign/goa/goagen
sh clean.sh
sh build.sh
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
