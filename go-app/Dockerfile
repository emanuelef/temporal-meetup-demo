FROM golang:1.22.3-alpine AS builder
WORKDIR /usr/src/app/
COPY . .
WORKDIR /usr/src/app/api_service
RUN go build -o /go/bin/

FROM alpine:latest AS runner
ENV GIN_MODE=release
WORKDIR /usr/src/app/
COPY --from=builder /go/bin/ .
EXPOSE 8080
ENTRYPOINT [ "./api_service" ]