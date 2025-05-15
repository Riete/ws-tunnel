# ws-tunnel

## Build
```
make build-linux-amd64
make build-linux-arm64
# or
make build-linux
```

## Global Flag
```
Global Flags:
      --log-level string   log level, one of debug, info, warn, error (default "info")
      --token string       auth token (default "ws-tunnel-token")
```
token is used for authentication  
if log-level is debug, log will also output to stdout

## Run  Server
Server accepts connections from Proxy and Client
```
Usage:
  ws-tunnel server [flags]

Flags:
  -h, --help                 help for server
  -l, --listen-port string   listen port (default "37452")
```

## Run Client
Client connects to Server with unique client id to establish ws tunnel
```
Usage:
  ws-tunnel client [flags]

Flags:
  -i, --client-id string     client id (default "test-client")
  -h, --help                 help for client
  -s, --server-addr string   server addr (default "127.0.0.1:37452")
      --wss                  if use wss
```

## Run Proxy
Proxy connects to Server to establish ws tunnel and run a local socks5 proxy server  
The traffic of the proxy server is sent through [client-id] tunnel  
Run multiple Proxies to handle custom traffic
```
Usage:
  ws-tunnel proxy [flags]

Flags:
  -i, --client-id string           client id (default "test-client")
  -h, --help                       help for proxy
  -b, --proxy-bind-ip string       proxy bind ip  (default "127.0.0.1")
  -p, --proxy-listen-port string   proxy listen port (default "2222")
  -s, --server-addr string         server addr (default "127.0.0.1:37452")
      --wss                        if use wss
```