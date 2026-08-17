FROM golang:1.26-bookworm AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN mkdir -p /app/bin && \
    for dir in cmd/*; do \
        name=$(basename "$dir"); \
        go build -o "/app/bin/$name" "./$dir"; \
    done


FROM debian:bookworm-slim AS runner

COPY --from=builder /app/bin/ /usr/local/bin/

ENTRYPOINT ["server"]
