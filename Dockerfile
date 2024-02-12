FROM --platform=$BUILDPLATFORM tonistiigi/xx AS xx

FROM --platform=$BUILDPLATFORM golang:1.22-alpine AS xbuild
WORKDIR /src
COPY --from=xx / /
ARG TARGETPLATFORM
ENV CGO_ENABLED=0
RUN --mount=target=. \
    --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg \
    xx-go build -trimpath -ldflags "-s -w" -o /dist/tinygeoip ./cmd/tinygeoip && \
    xx-verify --static /dist/tinygeoip

FROM cgr.dev/chainguard/static:latest
COPY --from=xbuild /dist/tinygeoip /app/tinygeoip
ENTRYPOINT ["/app/tinygeoip"]
