import asyncio
from fastapi import FastAPI
from fastapi.responses import JSONResponse
from fastapi.middleware.gzip import GZipMiddleware
from fastapi.middleware.cors import CORSMiddleware
import httpx
import pandas as pd
from prophet import Prophet
import math
import time
from cachetools import TTLCache
from opentelemetry.instrumentation.fastapi import FastAPIInstrumentor
from opentelemetry import trace
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import (
    BatchSpanProcessor,
)
from opentelemetry.instrumentation.httpx import HTTPXClientInstrumentor
from dotenv import load_dotenv

load_dotenv() 

# OTLPSpanExporter

provider = TracerProvider()
processor = BatchSpanProcessor(OTLPSpanExporter())
provider.add_span_processor(processor)

# Sets the global default tracer provider
trace.set_tracer_provider(provider)

# Creates a tracer from the global tracer provider
tracer = trace.get_tracer("my.tracer.name")

app = FastAPI()
app.add_middleware(GZipMiddleware, minimum_size=0)  # Compress all responses
FastAPIInstrumentor.instrument_app(app)

HTTPXClientInstrumentor().instrument()

# CORS (Cross-Origin Resource Sharing) middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],  # Update this with the list of allowed origins
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# In-memory cache with a TTL (Time To Live) of 10 days
cache = TTLCache(maxsize=1000, ttl=864000)


@app.get("/health")
async def root():
    current_span = trace.get_current_span()
    current_span.set_attribute("operation.value", 1)
    current_span.add_event("Gonna try it!")
    with tracer.start_as_current_span("span-name") as span:
        span.add_event("Gonna try it!")
        await asyncio.sleep(2)
    return {"message": "Ciao"}


@app.get("/predict")
async def predict(repo: str):
    current_span = trace.get_current_span()
    # Check if the data is already in the cache
    if repo in cache:
        cached_data = cache[repo]
        print("Returning cached data.")
        current_span.add_event("Cache it!")
        return JSONResponse(content=cached_data)

    api_url = f"http://143.47.226.125:8080/allStars?repo={repo}"
    async with httpx.AsyncClient() as client:
        api_response = await client.get(api_url)
    api_data = api_response.json()

    # Extracting stars data from the JSON
    stars_data = api_data["stars"]

    # Creating a DataFrame
    df = pd.DataFrame(stars_data, columns=["ds", "y", "y2"])
    current_span.add_event("Created Dataframe")

    # Converting the 'ds' column to datetime format
    df["ds"] = pd.to_datetime(df["ds"], format="%d-%m-%Y")

    print(df.tail())

    start_time = time.time()

    m = Prophet()
    m.fit(df)

    future = m.make_future_dataframe(periods=60, freq="D")
    # print(future.tail())

    forecast = m.predict(future)
    current_span.add_event("Computed forecast")

    forecast["ds"] = forecast["ds"].dt.strftime("%Y-%m-%d")

    # Extract relevant information from the forecast
    forecast_data = forecast[["ds", "yhat", "yhat_lower", "yhat_upper"]]
    forecast_trend = forecast[["ds", "trend"]]

    # Round up 'yhat' to the next integer
    forecast_data.loc[:, "yhat"] = forecast_data["yhat"].apply(math.ceil).astype(int)

    end_time = time.time()  # Record the end time
    elapsed_time = end_time - start_time  # Calculate the elapsed time
    print(f"Prediction took {elapsed_time:.2f} seconds")

    last_60_forecast = forecast_data.tail(61)
    # Combine the API data and forecast data
    result = {
        "forecast_data": last_60_forecast.to_dict(orient="records"),
        "forecast_trend": forecast_trend.to_dict(orient="records"),
    }

    # Cache the result
    cache[repo] = result

    return JSONResponse(content=result)


if __name__ == "__main__":
    import uvicorn

    uvicorn.run("main:app", host="0.0.0.0", port=8082)
