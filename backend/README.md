
# SMLCLoudPlatForm

## Environment Variable

### MongoDB
| Name        | Description            | Value |
|-------------|------------------------|-------|
| MODE / BC_ENV / APP_ENV / ENVIRONMENT | Runtime environment: `development/dev`, `uat`, `production/pro` | `development` |
| MONGODB_DEV_URI | MongoDB DEV connection URI (`mongodb://` or `mongodb+srv://`) | secret |
| MONGODB_DEV_DB | MongoDB DEV database name | '' |
| MONGODB_UAT_URI | MongoDB UAT connection URI (`mongodb://` or `mongodb+srv://`) | secret |
| MONGODB_UAT_DB | MongoDB UAT database name | '' |
| MONGODB_PRO_URI / MONGODB_PRODUCTION_URI | MongoDB PRO connection URI (`mongodb://` or `mongodb+srv://`) | secret |
| MONGODB_PRO_DB / MONGODB_PRODUCTION_DB | MongoDB PRO database name | '' |
| MONGODB_URI | Legacy/fallback MongoDB URI for DEV only | '' |
| MONGODB_DB | Legacy/fallback MongoDB DB for DEV only | '' |

MongoDB data must be separated by environment. DEV, UAT, and PRO must use different MongoDB locations or databases and different credentials. The location can be MongoDB Atlas or a private MongoDB deployment. Do not commit real MongoDB URI, password, token, or API key into this repository.

MongoDB rollout policy: start with fresh empty DEV/UAT/PRO databases. Do not migrate, import, upload, or copy old MongoDB data unless a separate migration task is explicitly approved with source, target, backup, and rollback plan.


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
    holdingcode String,
    status String,
    error_message String,
    progress Int32,
    createdat DateTime,
    updatedat DateTime,
    completed_at Nullable(DateTime)
) ENGINE = MergeTree()
ORDER BY (holdingcode, task_id, createdat);
