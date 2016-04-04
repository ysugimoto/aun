## aun

Simple WebSocket Implementation over TCP/TLS.

### Installation

Get the package:

```
$ go get github.com/ysugimoto/aun
```

### Usage

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
ws.onmessage = (msg) {
    console.log(msg);
};
ws.send("Hello, aun!");
```

### License

MIT License.

### Author

Yoshiaki Sugimoto

