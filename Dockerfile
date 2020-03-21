# STEP 1: Build executable binary
FROM golang:1.13 AS builder

WORKDIR /go/src/kindccm
COPY . .
RUN go get -d -v ./...
RUN GOOS=linux CGO_ENABLED=0 go build -o /go/bin/kindccm

# STEP 2: Build small image
FROM scratch

COPY --from=builder /go/bin/kindccm /bin/kindccm

CMD ["/bin/kindccm"]
