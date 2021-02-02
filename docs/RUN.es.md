# gitlab-api-client

La ejecución requiere mínimamente ciertos parámetros (url, token, certificado). Los cuales pueden incorporarse en el archivo de configuración ($HOME/.api-client.yaml) o pasados en la ejecución.

Parámetros por linea de comando

* --api-url https://gitlab.localhost/api/v4/
* --private-token TOKEN_USER
* --trusted-certificates @PATH_CERTIFICATES

Parámetros indicando archivo de configuración

* --config [./docs/api-client.yaml](./api-client.yaml)

>NOTA: Si no se indica el archivo de configuración, por defecto busca si existe en el path $HOME/.api-client.yaml.

## Ejecutar desde binario existente en `docs`

```sh
cd ~/go/src/github.com/janusky/gitlab-api-client/docs

# Ejecución con help general
./gitlab-api-client

# Ejecución help de un comando
./gitlab-api-client list-projects -h

# Indicando el path del archivo de configuración
# Por defecto lo busca como: $HOME/.api-client.yaml
./gitlab-api-client --config ./api-client.yaml

# Reemplazando propiedades del archivo de configuración api-client.yaml
./gitlab-api-client --api-url https://gitlab.localhost/api/v4/ \
  --private-token CAMBIAR_POR_TOKEN \
  --trusted-certificates=@$HOME/go/src/github.com/janusky/gitlab-api-client/docs/certificates.pem \
  --debug

# Utilizar la configuración de api-client.yaml pero cambiar token
./gitlab-api-client --config ./api-client.yaml \
  --private-token CAMBIAR_POR_TOKEN \
  --debug
```

## Ejecutar desde los fuentes

```sh
cd ~/go/src/github.com/janusky/gitlab-api-client

# Archivo config y salida en result
go run main.go --config ./docs/api-client.yaml > result

# Parámetros
go run main.go --api-url https://gitlab.localhost/api/v4/ \
  --private-token CAMBIAR_POR_TOKEN \
  --trusted-certificates=@$HOME/go/src/github.com/janusky/gitlab-api-client/docs/certificates.pem \
  --debug
```
