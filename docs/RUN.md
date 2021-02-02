# gitlab-api-client

The execution minimally requires certain parameters (url, token, certificate). Which can be incorporated in the configuration file ($HOME/.api-client.yaml) or passed in the execution.

Parameters by command line

* --api-url https://gitlab.localhost/api/v4/
* --private-token TOKEN_USER
* --trusted-certificates @PATH_CERTIFICATES

Parameter indicating configuration file

* --config [./docs/api-client.yaml](./api-client.yaml)

>NOTE: If the configuration file is not indicated, by default it looks to if it exists in the path $HOME/.api-client.yaml

## Run from existing binary in docs

```sh
cd ~/go/src/github.com/janusky/gitlab-api-client/docs

# Show help
./gitlab-api-client

# Help for a particular command
./gitlab-api-client list-projects -h

# Indicating the path of the configuration file
# Default location: $HOME/.api-client.yaml
./gitlab-api-client --config ./api-client.yaml

# Overriding properties of the api-client.yaml configuration file
./gitlab-api-client --api-url https://gitlab.localhost/api/v4/ \
  --private-token CAMBIAR_POR_TOKEN \
  --trusted-certificates=@$HOME/go/src/github.com/janusky/gitlab-api-client/docs/certificates.pem \
  --debug

# Use api-client.yaml configuration but change token
./gitlab-api-client --config ./api-client.yaml \
  --private-token CAMBIAR_POR_TOKEN \
  --debug
```

## Run from sources

```sh
cd ~/go/src/github.com/janusky/gitlab-api-client

# Config file and output in result
go run main.go --config ./docs/api-client.yaml > result

# Parameters
go run main.go --api-url https://gitlab.localhost/api/v4/ \
  --private-token CAMBIAR_POR_TOKEN \
  --trusted-certificates=@$HOME/go/src/github.com/janusky/gitlab-api-client/docs/certificates.pem \
  --debug
```
