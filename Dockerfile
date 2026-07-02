FROM node:22-bookworm AS web-builder

WORKDIR /app/internal/site

COPY internal/site/package.json internal/site/package-lock.json ./
RUN npm ci

COPY internal/site ./
RUN npm run build

FROM golang:1.26.3-bookworm AS go-builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
COPY --from=web-builder /app/internal/site/dist ./internal/site/dist

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/beszel ./internal/cmd/hub \
	&& CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/beszel-agent ./internal/cmd/agent

FROM debian:bookworm-slim

RUN apt-get update \
	&& apt-get install -y --no-install-recommends ca-certificates \
	&& rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=go-builder /out/beszel /usr/local/bin/beszel
COPY --from=go-builder /out/beszel-agent /usr/local/bin/beszel-agent

EXPOSE 8090

ENV APP_URL=http://localhost:8090

ENTRYPOINT ["/usr/local/bin/beszel"]
CMD ["serve", "--http", "0.0.0.0:8090", "--dir", "/beszel_data"]
