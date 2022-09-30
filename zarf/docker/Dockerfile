FROM golang:1.18 as builder

ENV CGO_ENABLED 0
ARG BUILD_REF

WORKDIR /service

COPY . .

RUN go build -ldflags "-X main.build=${BUILD_REF}" -o sales app/services/sales/main.go

FROM alpine:3.14

ARG BUILD_REF

COPY --from=builder /service/sales /service/sales

WORKDIR /service

CMD ["./sales"]
