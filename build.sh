#!/bin/bash

set -e

# Autogenerate content
goagen bootstrap -d github.com/adevinta/vulcan-results/design

# Compile and install
go install ./...
