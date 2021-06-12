# Igor

Igor is a dialect of Go language with the support for
[structured concurrency](https://vorpus.org/blog/notes-on-structured-concurrency-or-go-statement-considered-harmful/).

*WARNING: The project is under development.*

## How does it differ from the standard go?

`go` statement is prohibited - very much like `goto` in most modern languages.

Instead, launch the goroutines in nurseries using `run` function:

```go
n := make(nursery)
run(n, foo())
err := n.Wait()
```

Note that Igor uses a calling convention that differs from the standard Go. Call the functions
from native packages with `gocc` specifier:

```go
fmt.Println.gocc("Hi there!")
```

# Nurseries in detail

Nursery holds a set of goroutines. The goroutine function MUST must have an `error` return type.
If goroutine exits with `nil` it is silently removed from the nursery. If goroutine exits with
an error, all the other goroutines in the nursery are immediately canceled. The error is stored
to be later returned to the owner of the nursery.

The owner, in addition to launching goroutines, can use the following functions on the nursery:

`Close()` immediately cancels all the goroutines in the nursery. There's no return value.

`Wait() error` waits until all the goroutines in the nursery are finished. If one of them fails,
the error is returned from the `Wait` function. If no goroutine has failed, `nil` is returned.

`Err() error` checks whether the nursery failed, but doesn't block. In case of previous failure,
it returns the error. If the nursery is in running state it returns `nil`. This function can be used
when periodic checking of the health of the nursery is desired.

## How does it work?

`igor` tool is a transpiler that takes files with extension `.igor` and produces standard Go files
with extension `.go`.

To install the transpiler:

```bash
go install github.com/sustrik/igor/igor
```

Run it like this:

```bash
igor
```

The transpiler compiles all the Igor files in the current directory and in all of its
subdirectories, recursively. Alternatively, starting directory can be specified on the command line:

```bash
igor ./go/src/github.com/foo/bar
```

The compiled Igor programs are linked with the `lib/igor` package which contains the language
runtime.