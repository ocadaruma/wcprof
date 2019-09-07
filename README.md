# wcprof

A wall-clock based simple profiler.

`wcprof` parses your go source then install timer to all top-level functions.

## installation

```
$ go get github.com/ocadaruma/wcprof
```

## Usage

#### 1. Install timer

**Your sources will be modified by executing this.**

```
$ wcprof --path /path/to/your/repo [--backup]
```

To prevent installing timer, add `// wcprof: OFF` to the function like this:

```go
// wcprof: OFF
func someFunction() {
}
```

#### 2. Show Stats

```go
// to Stdout
wcprof.DefaultRegistry().Print()

// to http.ResponseWriter
func SomeHandler(w http.ResponseWriter, r *http.Request) {
	wcprof.DefaultRegistry().Write(w)
}
```
