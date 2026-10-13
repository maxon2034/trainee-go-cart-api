FROM golang:1.26-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o cart_api /build/cmd/

FROM alpine

WORKDIR /app

COPY --from=builder /build/config ./config
COPY --from=builder /build/cart_api .

EXPOSE 8080

CMD ["./cart_api"]


