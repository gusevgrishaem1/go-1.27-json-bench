FROM golang:1.27-bookworm AS build
ARG GOEXPERIMENT
WORKDIR /src
COPY go.mod main.go ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /app/server .
FROM debian:bookworm-slim
COPY --from=build /app/server /server
EXPOSE 8080
ENTRYPOINT ["/server"]
