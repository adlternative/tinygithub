### TinyGithub

A tiny git server used for learning git principle.

### How to Use

1. compile
```shell
go build
```
2. edit config.json, you can see example in [config-example.json](config-example.json).
3. run tinygithub server
```shell
./tinygithub server
```
4. do what you want in the website, login, register, create repo and so on...

![resource/example.png](resource/example.png)

5. clone & push
```shell
git clone http://127.0.0.1:8080/adlternative/git.git
...
git push
```