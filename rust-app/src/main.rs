use actix_web::{get, middleware::Logger, web, App, HttpResponse, HttpServer, Responder};
use serde::Serialize;
use std::env;
use std::time::Duration;
use dotenv::dotenv;
use opentelemetry::trace::TraceError;
use opentelemetry::global;
use opentelemetry_sdk::{propagation::TraceContextPropagator, resource::{
    OsResourceDetector, ProcessResourceDetector, ResourceDetector,
    EnvResourceDetector, TelemetryResourceDetector,
    SdkProvidedResourceDetector,
}, runtime, trace as sdktrace};
use opentelemetry_otlp::{self, WithExportConfig};

#[derive(Debug, Serialize)]
struct GreetResponse {
    message: String,
}

#[get("/health")]
async fn health() -> impl Responder {
    HttpResponse::Ok().body("OK")
}

#[get("/hello/{name}")]
async fn greet(name: web::Path<String>) -> impl Responder {
    log::warn!("<---- /hello, name: {}", name);
    let response = GreetResponse {
        message: format!("Hello, {}!", name),
    };

    HttpResponse::Ok().json(serde_json::json!(response))
}

/* fn init_tracer() {
    global::set_text_map_propagator(TraceContextPropagator::new());
    let provider = TracerProvider::builder()
        .with_simple_exporter(SpanExporter::default())
        .build();
    global::set_tracer_provider(provider);
} */


fn init_tracer() -> Result<sdktrace::Tracer, TraceError> {
    global::set_text_map_propagator(TraceContextPropagator::new());
    let os_resource = OsResourceDetector.detect(Duration::from_secs(0));
    let process_resource = ProcessResourceDetector.detect(Duration::from_secs(0));
    let sdk_resource = SdkProvidedResourceDetector.detect(Duration::from_secs(0));
    let env_resource = EnvResourceDetector::new().detect(Duration::from_secs(0));
    let telemetry_resource = TelemetryResourceDetector.detect(Duration::from_secs(0));
    opentelemetry_otlp::new_pipeline()
        .tracing()
        .with_exporter(
            opentelemetry_otlp::new_exporter()
                .tonic()
                .with_endpoint(format!(
                    "{}{}",
                    env::var("OTEL_EXPORTER_OTLP_ENDPOINT")
                        .unwrap_or_else(|_| "http://otelcol:4317".to_string()),
                    "/v1/traces"
                )), 
        )
        .with_trace_config(
            sdktrace::config()
                .with_resource(os_resource.merge(&process_resource).merge(&sdk_resource).merge(&env_resource).merge(&telemetry_resource)),
        )
        .install_batch(runtime::Tokio)
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    std::env::set_var("RUST_LOG", "actix_web=info");
    env_logger::init();

    dotenv().ok();

    log::warn!("Starting web server on 0.0.0.0:8080");

    // Start the Actix web server
    HttpServer::new(|| App::new().wrap(Logger::default()).service(greet))
        .bind(("0.0.0.0", 8080))?
        .run()
        .await
}
