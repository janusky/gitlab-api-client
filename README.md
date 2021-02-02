# gitlab-api-client

Application with functionalities to interact with the existing API in GitLab Community Edition.

- <https://docs.gitlab.com/ce/api/>

## Use

The execution has general and command help.

```sh
# Create executable
# -ldflags "-X \"main.version=dev\" -X \"main.date=$(date '+%Y%m%d%H%M%S')\" -X \"main.commit=$(git rev-parse HEAD)\""
go build -ldflags "-X 'main.version=dev' -X 'main.date=$(date '+%Y%m%d%H%M')' -X 'main.commit=head'"

# Show help
./gitlab-api-client

# Help for a particular command (./gitlab-api-client command -h)
./gitlab-api-client list-projects -h
```

## More information

Inside the folder [/docs](./docs) there are additional useful files to work with this application.

* [Run](./docs/RUN.md)

**Referencias**

- <https://docs.gitlab.com/ee/api/README.html>
- <https://github.com/xanzy/go-gitlab>

## TODO

End tests in commands

* [add-member](commands/addMember.go)
* [deploy-key](commands/deployKey.go)

Implement other API features

* https://docs.gitlab.com/ee/api/api_resources.html
