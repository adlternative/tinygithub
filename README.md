### TinyGithub

A tiny git server used for learning git principle.

### How to Use

1. compile
```shell
go build
```
2. edit config.json, you can see example in [config-example.json](config-example.json).
3. put your git repositories to "storage" which specified in config file.
```shell
ls /home/adl/test/tinygithub/repositories/
adlternative

ls /home/adl/test/tinygithub/repositories/adlternative 
CodeGPT.git  git.git
```
4. run tinygithub server
```shell
./tinygithub server
```
5. clone git repo via http protocol
```shell
git clone http://127.0.0.1:8080/adlternative/git.git
```

6. push git repo via http protocol
```shell
#edit something...
git push origin main
```