FROM docker.io/library/node:20-alpine AS frontend-builder
WORKDIR /app/web/portal
COPY web/portal/package*.json ./
RUN npm ci
COPY web/portal/ ./
RUN npm run build

FROM docker.io/library/golang:1.23-alpine AS builder

RUN apk add --no-cache git make gcc musl-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /bin/aegis-gateway ./cmd/aegis-gateway

RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /bin/merchant-mcp ./cmd/merchant-mcp

RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /bin/dashboard ./cmd/dashboard

RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /bin/demo-buyer ./cmd/demo-buyer

RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /bin/redteam ./cmd/redteam

RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /bin/seed-catalog ./cmd/seed-catalog

RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /bin/migrate ./cmd/migrate

RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /bin/portal ./cmd/portal

FROM docker.io/library/alpine:3.20


RUN apk add --no-cache ca-certificates postgresql-client tzdata

RUN addgroup -g 1000 -S aegis && adduser -u 1000 -S aegis -G aegis

WORKDIR /app

COPY --from=builder /bin/aegis-gateway /usr/local/bin/aegis-gateway
COPY --from=builder /bin/merchant-mcp /usr/local/bin/merchant-mcp
COPY --from=builder /bin/dashboard /usr/local/bin/dashboard
COPY --from=builder /bin/demo-buyer /usr/local/bin/demo-buyer
COPY --from=builder /bin/redteam /usr/local/bin/redteam
COPY --from=builder /bin/seed-catalog /usr/local/bin/seed-catalog
COPY --from=builder /bin/migrate /usr/local/bin/migrate
COPY --from=builder /bin/portal /usr/local/bin/portal

COPY config.yaml /app/config.yaml
COPY --from=frontend-builder /app/web/portal/dist /app/web/portal/dist
COPY migrations/ /app/migrations/

COPY scripts/entrypoint.sh /app/entrypoint.sh

USER root
RUN chmod +x /app/entrypoint.sh && chown -R aegis:aegis /app
USER aegis

EXPOSE 8081 8082 8083 8084

ENTRYPOINT ["/app/entrypoint.sh"]