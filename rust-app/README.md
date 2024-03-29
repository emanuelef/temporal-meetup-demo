Update to latest Rust version

```
rustup self update
rustup update stable
```
```
OTEL_EXPORTER_OTLP_ENDPOINT=https://api.honeycomb.io:443 OTEL_EXPORTER_OTLP_HEADERS="x-honeycomb-team=NMqf4vWfg4QkgHjeIBRj7d" OTEL_SERVICE_NAME=RustApp RUST_LOG="debug,h2=warn" cargo run
```