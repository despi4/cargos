FROM golang:1.26.1-alpine AS build
ARG SERVICE
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN test -n "$SERVICE" && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/api "./${SERVICE}/cmd/api"

FROM alpine:3.23
RUN apk add --no-cache ca-certificates && addgroup -S cargos && adduser -S cargos -G cargos
COPY --from=build /out/api /usr/local/bin/api
USER cargos
EXPOSE 8080
HEALTHCHECK --interval=5s --timeout=3s --start-period=10s --retries=10 CMD wget -q -O /dev/null http://127.0.0.1:8080/readyz || exit 1
ENTRYPOINT ["api"]
