# Step 1

Let's get the basic project configured. We'll keep the project structure pretty simple.

## Set up a Go project

Now create a directory for your project and navigate to it. We'll create a new Go project in this directory using the following CLI command:

`go mod init balloons`

That should create a new go.mod file that looks like this, more or less:

```text
module balloons

go 1.27.4
```

Your Go version might be different, but as long as it's 1.27.x it should be fine.

## Get your source on

Next up, let's create a `src` directory and throw a `main.go` in there.

I created a `main.go` with this content:

```go
package main

import "fmt"

func main() {
	fmt.Println("Ok, it runs.")
}
```

Now I can run `go run ./src` from the terminal and get:

```text
Ok, it runs.
```

## Get GoDom added.

We can look at the [GoDom Docs](https://github.com/ewaldhorn/godom) and see how to add it to our project:

Run `go get github.com/ewaldhorn/godom` and the dependency should be added. Now our `go.mod` file will look something like this:

```text
module balloons

go 1.27.4

require github.com/ewaldhorn/godom v1.0.6 // indirect
```

Run `go mod tidy` in the terminal to ensure it's all ready to rock.

## Now let's update our main.go

Now that we have the dependency, let's use it!

Update `main.go`:

```go
package main

import (
	"github.com/ewaldhorn/godom/dom"
	"github.com/ewaldhorn/godom/html"
)

func main() {
	// First step is to init the dom
	dom.Init()

	// now we can create elements
	// we won't need the actual container so we ignore the variable
	_ = html.Div().Class("container").Child(html.H1().Text("Hello from GoDom").Build()).AppendTo(dom.Body)

	// need to block this from exiting, otherwise wasm terminates our app
	select {}
}
```

What we want to do, is initialise the godom library, it grabs some handles to the document, body, head etc. This allows us to interact with the DOM from Go.  Then we go ahead and create a new `div` into which we place a `h1` header and append the container to the document body. Finally, because we don't want the app to exit (our later code will need to keep running), we do a `select {}` which forces Go to keep the app running.

## Compiling to WebAssembly

Now that we have the code, we need to compile it to Web Assembly. Go is really good at cross-compiling, it's one of my favourite features of the language. For our purposes, we want to target the `wasm` architecture, with the `js` operating system. This tells Go to target a browser environment. Let's create a `build.sh` file with this content:

```bash
#!/usr/bin/env bash

# let's create an output directory so we keep things tidy
mkdir -p dist

# have Go target wasm directly and output our binary in the dist directory
GOOS=js GOARCH=wasm go build -trimpath -ldflags="-s -w" -o ./dist/app.wasm ./src
```

First we set up an output directory called `dist` to keep things neat.
Then we task Go with building our `app.wasm` binary, using the appropriate flags, to the `dist` directory. Remember to give the `build.sh` script execution permissions with `chmod +x build.sh`.

Now you should be able to run `./build.sh` in your step 1 directory and it should output the binary to the `dist` directory.

## Yowsers that's a big file

Yeah, compiling Go to wasm can produce some chunky binaries. On my machine, just this basic Go app is 2 032 743 bytes! We can add some flags to our `build.sh` script to try to trim things just a bit. Change the build line to:

```bash
GOOS=js GOARCH=wasm go build -trimpath -ldflags="-s -w" -o ./dist/app.wasm ./src
```

Now run the build again. For my, the file size came down to 1 987 618 bytes. Still huge, but that's the price we pay for the powerful Go runtime. If you are interested in smaller binaries, check out [TinyGo](https://tinygo.org/), it is an interesting project that compiles smaller and tighter Go binaries. I even have a [TinyCanvas](https://github.com/ewaldhorn/tinycanvas) project with demos for using TinyGo to target wasm as well.

## How do we run this

Ok, it's compiled, but it is still not useable. After all, a .wasm file does not a website make. We need an `index.html` page and something that will boot the Go WebAssembly environment and run our wasm binary.

First, we need an `index.html` file. To keep things organised, I'm going to create a directory called `html` and place my HTML document(s) there.

For now, let's just get a basic `index.html` in place:

```html
<!DOCTYPE html>
<html>
  <head>
    <meta charset="utf-8" />
    <title>GoDom Tutorial Step 1</title>
  </head>
  <body>
      <h1>Just Checking</h1>
  </body>
</html>
```

## Finally running it

Now that we have an index file, we should serve it. I like `http-server`, but you can use whatever you want. I'll create a simple `run.sh` file that copies my index file into the `dist` directory and then serve that directory:

```bash
```

Remember to make the `run.sh` file executable with the `chmod +x run.sh` command. Now when you invoice `./run.sh`, you should see something like:
```bash
Starting up http-server, serving ./dist/

http-server version: 14.1.1

http-server settings: 
CORS: disabled
Cache: 3600 seconds
Connection Timeout: 120 seconds
Directory Listings: visible
AutoIndex: visible
Serve GZIP Files: false
Serve Brotli Files: false
Default File Extension: none

Available on:
  http://127.0.0.1:9000
  http://xxx.xxx.xx.xxx:9000
Hit CTRL-C to stop the server
```

And you should be able to open `localhost:9000` in your browser and see our index file! Of course, you use whatever local HTTP server works for you, but you will need a server of some kind, wasm can't be loaded by just opening the `index.html` directly from disk.
