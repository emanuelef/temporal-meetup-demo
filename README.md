# temporal-meetup-demo

## Requirements 
- [Docker compose](https://docs.docker.com/compose/install/)
- [Temporal CLI](https://docs.temporal.io/cli#install)
- [Honeycomb.io account](https://ui.honeycomb.io/signup) It's free and no CC required

## Run this sample

Make sure port 8080 is available on your system

### Make

1. Create the .env file with the Honeycomb Configuration API key (one off, only needed once)
`make create-env`
2. Start the example using Docker images on ghcr
`make start`
3. Trigger a Temporal workflow
`make service`
4. Temporal UI
`http://localhost:8233/namespaces/default/workflows`
5. Honeycomb UI
`https://ui.honeycomb.io` Make sure to select TemporalMeetup dataset in the upper left corner
6. Stop
`make stop`

### Honeycomb.io 

The provion API call will generate something like:

<img width="1342" alt="Screenshot 2024-04-14 at 23 44 39" src="https://github.com/emanuelef/temporal-meetup-demo/assets/48717/1bd18950-581b-4a22-9cac-3bd116f32ca7">

<img width="681" alt="Screenshot 2024-04-15 at 00 06 14" src="https://github.com/emanuelef/temporal-meetup-demo/assets/48717/2de2018c-a490-477f-bb85-f17c50e40b62">


### Manual

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

Questions

- Task Queues

```
awslocal s3api list-buckets
```

```
docker compose -f docker-compose-ghcr.yml up
```

## Links

https://mermaid.js.org/syntax/sequenceDiagram.html
