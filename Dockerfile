FROM golang:1.24.4-bookworm AS builder

ENV APP_PORT=3080
ENV METRICS_PORT=3088

WORKDIR /build
COPY src/* ./
COPY .git ./
RUN go mod tidy
RUN APP_VERSION=$(git describe --tags || echo '0.0.0') && APP_BUILD_DATE=$(date +%Y%m%d.%H%M) &&\
    CGO_ENABLED=0 GOOS=linux \
    go build -o ledlighter -ldflags "-X main.appVersion=$APP_VERSION -X main.buildDate=$APP_BUILD_DATE" .

FROM alpine:3.21.3
RUN addgroup -S web && adduser -S web -G web
USER web
WORKDIR /
COPY --from=builder /build/ledlighter /bin
HEALTHCHECK --interval=30s --timeout=3s \
    CMD curl -f http://localhost:${APP_PORT}/healthz || exit 1
EXPOSE $APP_PORT
EXPOSE $METRICS_PORT
USER root
ENTRYPOINT ["/bin/ledlighter"]
USER web
