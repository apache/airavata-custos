FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /custos-signer .

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=build /custos-signer /usr/local/bin/custos-signer
# migrate runs with a path relative to the workdir
COPY --from=build /src/migrations /app/migrations
EXPOSE 8084
USER 65534:65534
ENTRYPOINT ["custos-signer"]
CMD ["serve", "--config", "/app/config.yaml", "--auto-migrate"]
