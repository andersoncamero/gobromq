FROM golang:1.23-alpine AS builder


RUN apk --no-cache add ca-certificates git


WORKDIR /app


COPY go.mod ./
COPY go.sum ./
RUN go mod download


COPY . .


RUN CGO_ENABLED=0 GOOS=linux go build -o gobromq cmd/gobromq/main.go


FROM alpine:latest


RUN apk --no-cache add ca-certificates tzdata


RUN addgroup -S gobromq && adduser -S gobromq -G gobromq


WORKDIR /home/gobromq


COPY --from=builder /app/gobromq .


RUN mkdir -p /data && chown -R gobromq:gobromq /data /home/gobromq

USER gobromq


EXPOSE 1884


CMD ["./gobromq"]
