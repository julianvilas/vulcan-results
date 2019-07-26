# Vulcan Persistence Service

## Installing dependencies

Install mercurial `hg` client if not already installed. For example in OSX if you are using homebrew:
```
brew install hg
```

And get the dependencies:

```
go get -u github.com/goadesign/goa/...
```

## Code generation

From the project directory (inside of `$GOPATH/src`) just run:

```
./build.sh
```

It must show a result like

```
app
app/contexts.go
app/controllers.go
app/hrefs.go
app/media_types.go
app/user_types.go
app/test
app/test/healthcheck_testing.go
main.go
healthcheck.go
tool/vulcan-results-cli
tool/vulcan-results-cli/main.go
tool/cli
tool/cli/commands.go
client
client/client.go
client/healthcheck.go
client/user_types.go
client/media_types.go
swagger
swagger/swagger.json
swagger/swagger.yaml
```

You can also delete the auto-generated content executing `./clean.sh`.
