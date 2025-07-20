FROM golang:1.24.5-alpine AS builder

WORKDIR /workspace

COPY go.mod go.sum ./
RUN go mod download

COPY ./ ./

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o server .

RUN apk add --no-cache curl


FROM scratch

COPY migrations migrations
COPY --from=builder /workspace/server /server

EXPOSE 3000

CMD ["/server"]
