#!/bin/bash

# Copyright 2019 Adevinta

set -e

# Ensure the current version of gogen is installed

go install github.com/goadesign/goa/goagen


# Autogenerate content
# goagen bootstrap -d github.com/adevinta/vulcan-results/design

goagen app -d github.com/adevinta/vulcan-results/design
goagen client -d github.com/adevinta/vulcan-results/design
goagen swagger -d github.com/adevinta/vulcan-results/design


# Compile and install
go install ./...
