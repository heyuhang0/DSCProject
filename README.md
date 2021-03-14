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
## CLI: Command Line Interface
### Getting Started

**Start the server**

```go run cmd/server/main.go```

#### GET
**Description: For Client to send a GET request to Server**
```
Input Argument:
1. Address of Server (type string) i.e localhost:50051
2. Key (type string)

Expected Response: 
GET SUCCESSFUL, {key, value}

Command:
go run internal/command/main.go get <ADDRESS> <KEY>

Example:
go run internal/command/main.go get localhost:50051 hello
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

Command:
go run internal/command/main.go put <ADDRESS> <KEY> <VALUE>

Example:
go run internal/command/main.go put localhost:50051 test_key test_value
```