### TinyGithub

<div style="text-align:center;">
  <img src="resource/logo.png" alt="Logo" width="256" height="256">
</div>

A tiny git server used for learning git principle.

### Build

```shell
make
docker-compose --env-file=./docker-compose.env up
```

### How to Dev it

#### golang git server
- [tinygithub](https://github.com/adlternative/tinygithub)

1. write config.json
2. go run main.go server

#### vue frontend server
- [tinygithub-frontend](https://github.com/adlternative/tinygithub-frontend)

1. npm run serve

![resource/example.png](resource/example.png)