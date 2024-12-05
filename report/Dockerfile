FROM docker.io/golang:alpine3.21 AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY cmd cmd
RUN CGO_ENABLED=0 GOOS=linux go build cmd/report.go

FROM docker.io/alpine:3.21
RUN apk add --no-cache typst font-noto
WORKDIR /app
COPY --from=build /app/report .
COPY *.typ .
# Preload Typst packages 
RUN typst compile submission-report.typ || true
CMD ["./report"]