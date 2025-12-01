# Supanova Server

Go server for Supanova Radiation Protection Services learning platform

#### Prerequisite files:
- .env
- init.sql

#### Setup:
```
make dep
```

#### Run with docker:
```
docker-compose up -d
```

#### Run without docker:
```
docker-compose up -d postgres # run only postgres via docker
make run
```

#### Lint:
```
make lint
```

#### Tests:
- Ensure the following env vars are set if using docker with colima:
```
set:
export DOCKER_HOST=unix://${HOME}/.colima/default/docker.sock
export TESTCONTAINERS_RYUK_DISABLED=true
```
Then run:
```
make test
```

#### Generate db queries:
```
make sqlc
```

#### Create a db migration:
```
make migrate/create name=<migration_name>
```

#### Generate mocks:
```
make mocks
```

#### Get Firebase access token (for testing):
Get an ID token for API authentication:
```bash
go run cmd/access_token/main.go -api-key=<firebase_web_api_key> -email=<email> -password=<password>
```

Get verbose output (includes user ID, expiration, refresh token):
```bash
go run cmd/access_token/main.go -api-key=<firebase_web_api_key> -email=<email> -password=<password> -verbose
```

**Note:** The Firebase Web API Key can be found in Firebase Console → Project Settings → Web API Key
