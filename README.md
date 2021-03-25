# 50.041 Course Project

- cmd
  - client
  - server

- internal

  Files for this project only.
  - server

- pkg

  Packages used in this project
----
## Interactive Shell
### Features
1. GET
2. PUT
3. Request Per Second
4. Latency Percentile

### Getting Started

**Build & Start up all the servers**
```go
go build -o server-main cmd/server/main.go
./server-main n (where n is the server number)
```
**Build & Start up client**
```go
go build -o simshell simshell.go
./simshell
```

#### GET
**Description: For Client to send a GET request to Server**
```
Input Argument:
1. Address of Server (type string) i.e localhost:6001
2. Key (type string)

Expected Response: 
GET SUCCESSFUL, {key, value}
```
```bash
command: get <address> <key>
example: get localhost:6001 hello
```

#### PUT
**Description: For Client to send a PUT request to Server**
```
Input Argument:
1. Address of Server (type string) i.e localhost:50051,
2. Key (type string)
3. Value (type string)

Expected Response: 
PUT SUCCESSFUL, {key, value}
```
```bash
command: put <address> <key> <value>
example: get localhost:6001 hello prof
```

#### Request Per Second
```bash
rps localhost:6001 hello 1000
```

#### Latency Percentile

Input Argument:
1. address
2. key
3. no_requests
4. percentile

```bash
latencytime localhost:6001 hello 1000 99
```