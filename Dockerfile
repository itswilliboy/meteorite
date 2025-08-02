FROM golang:1.24.5-alpine AS builder

WORKDIR /workspace

COPY go.mod go.sum ./
RUN go mod download

COPY ./ ./

RUN CGO_ENABLED=0 GOOS=linux go build -installsuffix cgo -ldflags '-extldflags "-static"' -o server .

FROM oven/bun:1.2.19-alpine AS bun

WORKDIR /frontend

COPY frontend/package.json frontend/bun.lockb ./

RUN bun install --no-cache
COPY frontend/ ./
RUN bun run build

FROM scratch
WORKDIR /app

COPY migrations migrations
COPY --from=builder /workspace/server ./server

COPY --from=bun /frontend/dist/index.html ./
COPY --from=bun /frontend/dist/assets ./assets
COPY --from=bun /frontend/public ./public

EXPOSE 3000

CMD ["./server"]
