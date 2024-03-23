FROM golang:1.22.1-alpine as builder
ARG ERDK_PAT
WORKDIR /app
COPY src ./src
COPY go.mod .
COPY go.sum .
RUN apk update && apk upgrade && \
    apk add --no-cache bash git openssh
RUN go mod download
RUN go build -o advanced-services-management ./src/main.go

FROM alpine:latest AS runner
ENV GIN_MODE=release
WORKDIR /app
COPY --from=builder /app/advanced-services-management .
EXPOSE 8080
ENTRYPOINT ["./advanced-services-management"]
