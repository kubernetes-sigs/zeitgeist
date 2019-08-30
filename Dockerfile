FROM golang:1.12-alpine as builder

RUN apk --update add git upx

WORKDIR /go/src/github.com/pluies/zeitgeist
ENV GO111MODULE=on
ADD go.mod .
ADD go.sum .
RUN go mod download

ADD . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o zeitgeist .

# Make things even smaller!
RUN upx zeitgeist

FROM scratch
# Copy trusted CAs for TLS
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
# And our executable!
COPY --from=builder /go/src/github.com/pluies/zeitgeist/zeitgeist /
# Run as non-root
USER 1001
CMD ["/zeitgeist"]
