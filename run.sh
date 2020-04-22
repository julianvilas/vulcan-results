#!/bin/sh

export PORT=${PORT:-8080}
export DEBUG=${DEBUG:-false}
export PATH_STYLE=${PATH_STYLE:-false}

# Apply env variables
cat config.toml | envsubst > run.toml

./vulcan-results run.toml
