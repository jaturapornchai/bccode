
# SMLCLoudPlatForm

## Environment Variable

### MongoDB
| Name        | Description            | Value |
|-------------|------------------------|-------|
| MONGODB_URI | Mongodb Connection URI | ''    |
| MONGODB_DB  | Mongodb Database Name  | ''    |


### Redis

| Name                 | Description          | Value |
|----------------------|----------------------|-------|
| REDIS_CACHE_URI      | Redis connection uri | ''    |
| REDIS_CACHE_PASSWORD | Redis Password       | ''    |


## For wsl(ubuntu) Please Read

Install Kafkalib , Gcc
```

sudo apt-get install build-essential

```


## gen swagger

```
swag init -g swaggergen.go -o ./api/swagger/

```

### generate authorization pem key

```
openssl genrsa -out private.key 4096
```

```
openssl rsa -in private.key -pubout -out public.key
```

### Run Swagger 
```
swag init
go run main.go
```

```
http://localhost:1323/swagger/index.html
http://localhost:1323/swagger/doc.json

```

### Build Docker Command
```
docker build -t inventoryservice -f cmd/inventoryservice/Dockerfile .
docker build -t smlsoft/cloudauthentication -f ./cmd/authenticationservice/Dockerfile .
```

Get Mock Package
```
go install github.com/vektra/mockery/v2@latest
```

### Run Swagger With Make
```
make runswagger
```



## Github Registry 
## https://ghcr.io

```

```


# FOR M1 Run Please Read

```
brew install openssl
brew install librdkafka
brew install pkg-config
export PKG_CONFIG_PATH="/opt/homebrew/opt/openssl@3/lib/pkgconfig"
go build --tags dynamic main.go

```


## FOR M1 Build Docker and push
```
docker buildx create --use
docker buildx build --platform linux/amd64 --push -t <tag_to_push> .
```

## M1 Cannot Build Install
`https://www.baifachuan.com/posts/4862a3b1.html`

error
```
linux_syscall.c:67:13: error: implicit declaration of function 'setresgid' is invalid in C99 [-Werror,-Wimplicit-function-declaration]
linux_syscall.c:67:13: note: did you mean 'setregid'?
/Library/Developer/CommandLineTools/SDKs/MacOSX.sdk/usr/include/unistd.h:593:6: note: 'setregid' declared here
linux_syscall.c:73:13: error: implicit declaration of function 'setresuid' is invalid in C99 [-Werror,-Wimplicit-function-declaration]
linux_syscall.c:73:13: note: did you mean 'setreuid'?
/Library/Developer/CommandLineTools/SDKs/MacOSX.sdk/usr/include/unistd.h:595:6: note: 'setreuid' declared here
```

fix by 
```
brew install FiloSottile/musl-cross/musl-cross

```

and build with 
```
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 CC=x86_64-linux-musl-gcc  CXX=x86_64-linux-musl-g++  go build  -o go-app -tags musl main.go
```



CREATE TABLE task_status (
    task_id String,
    shop_id String,
    status String,
    error_message String,
    progress Int32,
    created_at DateTime,
    updated_at DateTime,
    completed_at Nullable(DateTime)
) ENGINE = MergeTree()
ORDER BY (shop_id, task_id, created_at);