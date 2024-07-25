FROM docker.io/golang:alpine3.20 AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY cmd cmd
RUN CGO_ENABLED=0 GOOS=linux go build cmd/report.go

FROM docker.io/alpine:3.20
RUN apk add --update typst font-noto
WORKDIR /app
COPY --from=build /app/report .
COPY template.typ .
# Preload Typst packages 
RUN typst compile template.typ || true
CMD ["./report"]