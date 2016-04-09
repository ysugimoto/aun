## aun

Simple WebSocket Implementation over TCP/TLS.

### Installation

Get the package:

```
$ go get github.com/ysugimoto/aun
```

### Basic Usage

#### TCP

Import this package and start server:

```
package main

import (
    "github.com/ysugimoto/aun"
)

func main() {
    // Listen 0.0.0.0:10021
    server := aun.NewServer("0.0.0.0", 10021)

    // Listen and server with 1024 buffer size messaging
    server.Listen(1024)
}
```

And connect from Browser or some clients:

```
let ws = new WebSocket("ws://localhost:10021");
ws.onmessage = (msg) {
    console.log(msg);
};
ws.send("Hello, aun!");
```

#### TLS

Import this package and start server with TLS configuration:

```
package main

import (
    "github.com/ysugimoto/aun"
    "crypto/tls"
)

func main() {
    // Listen 0.0.0.0:10021
    server := aun.NewServer("0.0.0.0", 10021)

    // TLS configuration
    cer, err := tls.LoadX509KeyPair("/path/to/server.pem", "/path/to/server.key")
    if err != nil {
        log.Println(err)
        return
    }
    config := &tls.Config{Certificates: []tls.Certificate{cer}}

    // Listen and server with 1024 buffer size messaging
    server.ListenTLS(1024, config)
}
```

And connect from Browser or some clients:

```
let ws = new WebSocket("wss://localhost:10021");
ws.onmessage = (msg) => {
    console.log(msg);
};
ws.send("Hello, aun!");
```

### CLI command

Get the command package:

```
$ go get github.com/ysugimoto/aun/cmd/aun
```

Run with some options:

| option | description                         | default   |
|--------|-------------------------------------|-----------|
|    -h  | Listen host                         | 127.0.0.1 |
|    -p  | Listen port                         | 12345     |
| --tls  | Using TLS                           | false     |
| --key  | key file path (with `--tls` option) | -         |
| --pem  | pem file path (with `--tls` option) | -         |

for example:

```
$ aun -h 0.0.0.0 -p 9999
```

will start server `0.0.0.0:9999`.

### License

MIT License.

### Author

Yoshiaki Sugimoto

