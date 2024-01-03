FROM golang:1.21 AS builder
ENV CGO_ENABLED 0
WORKDIR /go/src/app
ADD . .
RUN go build -o /redirect

FROM scratch
COPY --from=builder /redirect /redirect
CMD ["/redirect"]
