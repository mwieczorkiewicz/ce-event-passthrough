FROM golang:alpine AS builder
RUN apk update && apk add --no-cache git

WORKDIR $GOPATH/src/ce-event-passthrough
COPY . .

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags="-w -s" -o /go/bin/ce-event-passthrough ./main.go 
FROM scratch
COPY --from=builder /go/bin/ce-event-passthrough /bin/ce-event-passthrough
ENTRYPOINT ["/bin/ce-event-passthrough"]