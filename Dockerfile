FROM --platform=$BUILDPLATFORM golang:1.26 AS build

WORKDIR /src

ENV GOMODCACHE=/go/pkg/mod
ENV GOCACHE=/go/.cache/go-build

COPY go.mod .
COPY go.sum .
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY config/ ./config
COPY utils/ ./utils
COPY report/ ./report
COPY run/ ./run
COPY cmd/ ./cmd
COPY VERSION .

ARG CGO_ENABLED=0
# Docker buildx sets these automatically from --platform; declare them to make them available as ENV
ARG TARGETOS=linux
ARG TARGETARCH=amd64
ARG MAIN_PKG=/src/cmd/main.go
ARG BIN_NAME=expirator
ARG VERSION_SYMBOL=github.com/goodieshq/expirator/utils.expiratorVersion

ENV CGO_ENABLED=${CGO_ENABLED} \
    GOOS=${TARGETOS} \
    GOARCH=${TARGETARCH}

# Read VERSION file and inject via -X
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/go/.cache/go-build \
    set -eu; \
    VERSION="$(tr -d '\r\n' < VERSION)"; \
    echo "Building ${BIN_NAME} version=${VERSION}"; \
    go build -trimpath \
      -ldflags="-s -w -X ${VERSION_SYMBOL}=${VERSION}" \
      -o /out/${BIN_NAME} \
      ${MAIN_PKG}

# Runtime
FROM gcr.io/distroless/static-debian12:nonroot AS runtime
WORKDIR /app
ARG BIN_NAME=expirator
COPY --from=build /out/${BIN_NAME} /app/${BIN_NAME}
USER nonroot:nonroot
ENTRYPOINT ["/app/expirator"]
CMD ["--format", "json"]