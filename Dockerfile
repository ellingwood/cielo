# Build frontend
FROM node:22-alpine AS frontend
WORKDIR /app/web
COPY web/package*.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

# Build backend
FROM golang:1.25-alpine AS backend
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /cielo ./cmd/cielo

# Final image
FROM alpine:3.21
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=backend /cielo /app/cielo
COPY --from=frontend /app/web/dist /app/web/dist
EXPOSE 8080
CMD ["/app/cielo"]
