FROM golang:1.18 as builder

ENV CGO_ENABLED 0
ARG BUILD_REF

WORKDIR /service
COPY . .

# Build the admin binary.
RUN go build -ldflags "-X main.build=${BUILD_REF}" -o admin app/tooling/admin/main.go

# Build the service binary
RUN go build -ldflags "-X main.build=${BUILD_REF}" -o sales app/services/sales/main.go

FROM alpine:3.14

COPY --from=builder /service/sales /service/sales
COPY --from=builder /service/admin /service/admin

CMD ["./sales"]
