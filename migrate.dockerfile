# Local: docker compose build migrator
# CI:    docker build --build-arg BASE_IMAGE=$CI_REGISTRY/$CI_PROJECT_PATH/alpine -f migrate.dockerfile .
ARG BASE_IMAGE=alpine:3.20
FROM ${BASE_IMAGE} AS base
RUN apk add --no-cache ca-certificates curl
RUN curl -sSf https://atlasgo.sh | sh

# Runtime env (set by compose / CI):
#   DATABASE_URL  e.g. postgres://user:pass@postgres:5432/book_hive?sslmode=disable
FROM base
WORKDIR /app
COPY migrations/ ./migrations/
COPY schema/ ./schema/
COPY atlas.hcl ./

ENTRYPOINT ["atlas", "migrate", "apply", "--env", "remote"]
