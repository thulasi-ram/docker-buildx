ARG BUILDX_VERSION=0.25.0
ARG DOCKER_VERSION=28.3.0-dind
ARG GOLANG_VERSION=1.24

FROM --platform=$BUILDPLATFORM golang:${GOLANG_VERSION} AS build

COPY . /src
WORKDIR /src

ARG TARGETOS TARGETARCH INSECURE
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg \
    make build

FROM docker/buildx-bin:${BUILDX_VERSION} AS buildx-bin
FROM docker:${DOCKER_VERSION}

COPY Corefile /etc/coredns/Corefile
COPY --from=buildx-bin /buildx /usr/libexec/docker/cli-plugins/docker-buildx
COPY --from=build /src/plugin-docker-buildx /bin/plugin-docker-buildx

RUN apk --update --no-cache add coredns git


ENV DOCKER_HOST=unix:///var/run/docker.sock

ENTRYPOINT ["/usr/local/bin/dockerd-entrypoint.sh", "plugin-docker-buildx"]
