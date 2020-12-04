FROM golang:1.14 AS builder
ENV CGO_ENABLED 0
WORKDIR /go/src/app
ADD . .
RUN go build -o /redirect

FROM scratch
COPY --from=builder /redirect /redirect
