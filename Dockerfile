FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY bpcl-portal-api/go.mod bpcl-portal-api/go.sum ./
RUN go mod download
COPY bpcl-portal-api/ .
RUN CGO_ENABLED=0 go build -o bin/api ./cmd/api

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/bin/api .
EXPOSE 8080
CMD ["./api"]
