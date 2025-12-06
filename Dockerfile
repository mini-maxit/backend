FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o backend.o cmd/app/main.go

FROM alpine/curl:8.14.1

WORKDIR /app

# Install Atlas binary
RUN curl -L https://release.ariga.io/atlas/atlas-linux-amd64-latest -o /usr/local/bin/atlas && \
    chmod +x /usr/local/bin/atlas

COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/atlas.hcl .
COPY --from=builder /app/backend.o .
COPY --from=builder /app/docs ./docs

CMD ["./backend.o"]
