# temporal-meetup-demo

### Steps to run this sample:
1) Run a [Temporal service](https://github.com/temporalio/samples-go/tree/main/#how-to-use).

One way could be just to use the Temporal CLI.  

```bash
temporal server start-dev
```

UI http://localhost:8233

2) Run the following command to start the worker
```bash
go run opentelemetry/worker/main.go
```
3) In another terminal, run the following command to run the workflow
```bash
go run opentelemetry/starter/main.go
```