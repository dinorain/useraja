### Useraja: Golang auth microservice example with gRPC and REST endpoints

#### What have been used:
* [GRPC](https://grpc.io/) - gRPC
* [Echo](https://github.com/labstack/echo) - Go web application framework
* [sqlx](https://github.com/jmoiron/sqlx) - Extensions to database/sql.
* [pgx](https://github.com/jackc/pgx) - PostgreSQL driver and toolkit for Go
* [JWT](https://github.com/golang-jwt/jwt) - A Go implementation of JSON Web Tokens.
* [viper](https://github.com/spf13/viper) - A Go configuration with fangs
* [go-redis](https://github.com/go-redis/redis) - Redis client for Golang
* [zap](https://github.com/uber-go/zap) - Logger
* [validator](https://github.com/go-playground/validator) - Go Struct and Field validation
* [migrate](https://github.com/golang-migrate/migrate) - Database migrations. CLI and Golang library.
* [testify](https://github.com/stretchr/testify) - Testing toolkit
* [gomock](https://github.com/golang/mock) - Mocking framework
* [CompileDaemon](https://github.com/githubnemo/CompileDaemon) - Compile daemon for Go
* [Docker](https://www.docker.com/) - Docker

#### Docker compose files:
    docker-compose.local.yml - run postgresql, redis, aws, prometheus, grafana containers
    docker-compose.dev.yml - run all in docker

### Docker development usage:
    make develop

### Local development usage:
    make local
    make run

### Swagger:

http://localhost:5001/swagger/

### Test (Admin Login):

```sh
curl -X POST                                                   \
    -d '{
        	"email": "admin@gmail.com",
        	"password": "admin"
        }' \
    http://172.104.58.183:5001/swagger/index.html#/Users/post_user_login
```

