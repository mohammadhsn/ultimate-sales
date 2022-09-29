FROM golang:1.18 as builder

ENV CGO_ENABLED 0
ARG BUILD_REF

WORKDIR /service

COPY main.go go.mod go.sum ./

RUN go build -ldflags "-X main.build=${BUILD_REF}" -o sales-api

FROM alpine:3.14

ARG BUILD_REF

COPY --from=builder /service/sales-api /service/sales-api

WORKDIR /service

CMD ["./sales-api"]
