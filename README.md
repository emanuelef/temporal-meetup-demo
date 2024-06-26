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

<img width="1444" alt="Screenshot 2024-04-21 at 18 57 02" src="https://github.com/emanuelef/temporal-meetup-demo/assets/48717/fe3e8670-7953-48a9-a130-00d617184298">

<img width="692" alt="Screenshot 2024-04-21 at 19 04 18" src="https://github.com/emanuelef/temporal-meetup-demo/assets/48717/6911983c-eb7a-4618-a299-86cdab83f215">

Bubble Up analysis

<img width="1405" alt="Screenshot 2024-04-21 at 19 07 14" src="https://github.com/emanuelef/temporal-meetup-demo/assets/48717/845df360-cabf-4eec-8576-508018209c0c">



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
