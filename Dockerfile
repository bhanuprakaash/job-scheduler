FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/server ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/job-cli ./cmd/cli

FROM gcr.io/distroless/static-debian12

WORKDIR /

COPY --from=builder  /app/server /server
COPY --from=builder /app/job-cli /job-cli

EXPOSE 8080 50051 9090

USER nonroot:nonroot

ENTRYPOINT [ "/server" ]