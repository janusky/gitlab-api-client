# gitlab-api-client

Aplicación con funcionalidades para interactuar con la API existente en GitLab Community Edition.

- <https://docs.gitlab.com/ce/api/>

## Uso

La ejecución cuenta con ayuda general y por comando.

```sh
# Crear ejecutable
# -ldflags "-X \"main.version=dev\" -X \"main.date=$(date '+%Y%m%d%H%M%S')\" -X \"main.commit=$(git rev-parse HEAD)\""
go build -ldflags "-X 'main.version=dev' -X 'main.date=$(date '+%Y%m%d%H%M')' -X 'main.commit=head'"

# Por defecto muestra la ayuda
./gitlab-api-client

# Ayuda de un comando en particular (./gitlab-api-client comando -h)
./gitlab-api-client list-projects -h
```

## Más información

Dentro de la carpeta [/docs](./docs) se encuentran archivos adicionales útiles para trabajar con está aplicación.

* [Ejecutar/Lanzar/Run](./docs/RUN.es.md)

**Referencias**

- <https://docs.gitlab.com/ee/api/README.html>
- <https://github.com/xanzy/go-gitlab>

## TODO

Completar pruebas en comandos

* [add-member](commands/addMember.go)
* [deploy-key](commands/deployKey.go)

Implementar otras funcionalidades de la API

* https://docs.gitlab.com/ee/api/api_resources.html
