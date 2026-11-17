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
