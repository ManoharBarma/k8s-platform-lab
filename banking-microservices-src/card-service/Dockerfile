FROM golang:1.25-alpine AS builder

ENV CGO_ENABLED=0 GOOS=linux
WORKDIR /src

COPY pkg/ ./pkg/
COPY proto/ ./proto/
COPY card-service/ ./card-service/

WORKDIR /src/card-service

RUN go mod download
RUN go build -trimpath -ldflags="-s -w" -o /out/app ./cmd

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /out/app /app
EXPOSE 8080 9090
ENTRYPOINT ["/app"]
