FROM golang:1.13-alpine as builder
RUN apk --no-cache add git

# work outside of $GOPATH so modules are the default
WORKDIR /home/tinygeoip

# do deps first for caching layers
COPY go.mod .
COPY go.sum .
RUN go mod download

# actual binary build
COPY *.go ./
COPY cmd cmd
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
	go build -ldflags "-s -w" ./cmd/tinygeoip

# actual image is just distroless static with the binary
# https://github.com/GoogleContainerTools/distroless/tree/master/base
FROM gcr.io/distroless/static
COPY --from=builder /home/tinygeoip/tinygeoip /tinygeoip
ENTRYPOINT [ "/tinygeoip" ]
