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

# FROM alpine
# RUN apk --no-cache add ca-certificates
# ^^ I don't believe ca-certificates are needed if not making outgoing conns?
FROM scratch
COPY --from=builder /home/tinygeoip/tinygeoip /tinygeoip
ENTRYPOINT [ "./tinygeoip" ]
