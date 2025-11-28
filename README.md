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
