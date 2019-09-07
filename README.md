# wcprof

A wall-clock based simple function profiler.

`wcprof` parses your go source then install timer to all functions (except function literals).

```
+--------------------------+-------+-------------+------------+------------+------------+
|           Name           | count |   sum(ms)   |  max(ms)   |  min(ms)   |  avg(ms)   |
+--------------------------+-------+-------------+------------+------------+------------+
| main/cpuIntensiveHandler |    33 | 3732.827265 | 242.233016 |  89.352146 | 113.115977 |
|     main/blockingHandler |     3 | 1515.822732 | 507.925529 | 502.406221 | 505.274244 |
+--------------------------+-------+-------------+------------+------------+------------+
```

## Installation

#### wcprof command

```
$ go get github.com/ocadaruma/wcprof/cmd/wcprof
```

#### wcprof library

Add `github.com/ocadaruma/wcprof` to your go project.

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

#### 3. Run application

Run your application and collect stats.

You can disable timer by environmental variable of calling `wcprof.Off()`.

```
WCPROF_OFF=1 go run your-app
```

or

```go
func main() {
	wcprof.Off()
	...
	// your codes
}
```
