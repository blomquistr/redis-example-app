# Go Redis Example

A working Redis example for some internal training/testing

## Notes to Self - How a Golang Web Server Glues Together

- All applications in go start with a `main.go` file; this is the entrypoint to your app.

```go
// we have to be in package main in main.go - it's a convention that the compiler understand
package main

// here we're importing something. In this case, it's another package in our app, the server
// that we're going to run. We might import something else if we're running, say, a CLI or
// a local application that doesn't run a web server, but in this example app and refresher
// it's a web server, so we import our server package.
import (
 "github.com/blomquistr/go-redis-example/v2/internal/server"
)

// main is where we enter the application; the compiler will run this when you execute the
// compiled program. In our case, we want our server to do some work, so we have written
// a Run method in our server package, and inside main we're going to call the Run()
func main() {
 server.Run()
}
```
