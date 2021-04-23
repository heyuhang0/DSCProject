# 50.041 Course Project

## Setup

```shell
# Clone project
git clone https://github.com/heyuhang0/DSCProject.git
cd DSCProject

# Install dependencies
go mod vendor
```

## Run Instructions

### Seed server

Seed servers are known to clients and all machines in the cluster.
At least one seed server is required to launch the cluster, which is for
node discovery and fault detection in addition to data storage.
With the default cluster configuration `./configs/default_config.json`,
3 seed servers are configured.

Please run the following commands in separated terminals to launch seed servers:

```shell
go run ./cmd/server -config ./configs/default_config.json -seed -index 1
go run ./cmd/server -config ./configs/default_config.json -seed -index 2
go run ./cmd/server -config ./configs/default_config.json -seed -index 3
```

After launching, you can view their dashboards with address http://127.0.0.1:8001/,
http://127.0.0.1:8002/, http://127.0.0.1:8003/.

![Screenshot of the dashboard page](https://user-images.githubusercontent.com/10456378/115833776-0779fb00-a447-11eb-8854-eb6d08ad620e.jpeg)

### Normal server

Normal servers are optional data nodes that can be added on demand. With the
default configuration, at least one normal server is needed to fulfill the
4 replicas requirements.

Use the following commands to launch normal servers in separated terminals:

```shell
go run ./cmd/server -config ./configs/default_config.json -index 4 -id 5004 -internalAddr localhost:5004 -externalAddr localhost:6004
go run ./cmd/server -config ./configs/default_config.json -index 5 -id 5005 -internalAddr localhost:5005 -externalAddr localhost:6005
...
```

View their dashboards with address http://127.0.0.1:8004/, http://127.0.0.1:8005/.

### CLI client

The CLI client is an interactive tool for debugging and testing the database.
You can launch it with the following command:

```shell
go run ./cmd/client -config ./configs/default_client.json
```

Following commands are provided in the CLI client:
```
# put <key> <value>
$ put hello world
OK

# get <key>
$ get hello
hello: world

# benchmark <num_requests> <num_goroutines>
$ benchmark 1000 16
PUT
  - RPS:           2919.1957849148225 req/s
  - Avg Latency:   5.135980099999996 ms
  - 90% Latency:   11.0326 ms
  - 99% Latency:   13.9993 ms
  - 99.9% Latency: 14.0022 ms
Get
  - RPS:           2985.0790822075855 req/s
  - Avg Latency:   5.0290164000000015 ms
  - 90% Latency:   11.9985 ms
  - 99% Latency:   14.0007 ms
  - 99.9% Latency: 15.50005 ms
```

For debugging, to specify a coordinator, use following commands:
```
# put-node <address> <key> <value>
$ put-node localhost:6001 hello world
=== PUT Request is called! ===
2021/04/23 15:10:52 PUT SUCCESSFUL: {hello: world}

# put-node <address> <key>
$ get-node localhost:6001 hello
=== GET Request is called! ===
GET SUCCESSFUL: hello world

# rps <address> <key> <num_requests>
$ rps localhost:6001 hello 1000
Number of Requests Per Second: 1179.5094208007195

# latencytime <address> <key> <num_requests> <percentile>
$ latencytime localhost:6001 hello 1000 99
99 th percentile latency: 0.0011612
```

### Sample use case: COVID-19 tracing application

*The sample tracing application is for the demonstration
purpose of this school project only. It is not intended
to simulate any real-world application.*

To launch the safe entry server, run
```shell
go run ./cmd/safeentry -address :8080 -config ./configs/default_client.json
```

Then you can visit the sample tracing application at http://localhost:8080/

![screenshot of the use case](https://user-images.githubusercontent.com/10456378/115834902-4a889e00-a448-11eb-8704-6f6f27ca84e2.png)