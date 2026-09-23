
# SMLCLoudPlatForm

## Environment Variable

### Runtime
| Name        | Description            | Value |
|-------------|------------------------|-------|
| BC_ENV / APP_ENV / RUN_ENV / ENVIRONMENT / MODE | Runtime environment: `development/dev`, `uat`, `production/pro` | `development` |

### PostgreSQL (ฐานข้อมูลเดียวของระบบ — ไม่มี MongoDB/Kafka/Redis/ClickHouse)
| Name        | Description            | Value |
|-------------|------------------------|-------|
| POSTGRES_HOST | PostgreSQL host | secret |
| POSTGRES_PORT | PostgreSQL port | `5432` |
| POSTGRES_DB_NAME | Central database name | secret |
| POSTGRES_USERNAME | PostgreSQL user | secret |
| POSTGRES_PASSWORD | PostgreSQL password | secret |
| POSTGRES_SSL_MODE | `disable`, `require`, `verify-ca`, `verify-full` | `disable` |

### Object storage (MinIO / S3)
| Name        | Description            | Value |
|-------------|------------------------|-------|
| S3_ENDPOINT / S3_REGION / S3_BUCKET_NAME | Bucket location | secret |
| S3_ACCESS_KEY_ID / S3_SECRET_ACCESS_KEY | Credentials | secret |
| S3_FORCE_PATH_STYLE | `true` for MinIO | `true` |

Do not commit real passwords, tokens, or API keys into this repository.


## For wsl(ubuntu) Please Read

Install Gcc
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
CGO_ENABLED=0 go build main.go

```


## FOR M1 Build Docker and push
```
docker buildx create --use
docker buildx build --platform linux/amd64 --push -t <tag_to_push> .
```
