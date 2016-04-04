## CloudyKit Router

CloudyKit Router was developed to be a very fast router on matching and retrieving of parameters.

### Main characteristics

 1. Use prefix tree with special nodes to match dynamic sentences and catch all.
 2. No allocations are all, matching and retrieve parameters don't allocates.
 3. Nodes has precedence when matching, text node then single sentence node then wildcard node

### Benchmarks

I benchmark CloudyKit router against "github.com/julienschmidt/httprouter" and the results are pretty good, the benchmark consist of test the routes listem below per each iteration.

```go
	/users
	/users/:userId
	/users/:userId/subscriptions
	/users/:userId/subscriptions/:subscription
	/assets/*file
```

Benchmark source: https://github.com/CloudyKit/benchmarks/router

####### Results
```text
Go 1.6
BenchmarkCloudyKitRouter-4       2000000               618 ns/op               0 B/op          0 allocs/op
BenchmarkHttprouterRouter-4      1000000              1104 ns/op             224 B/op          4 allocs/op

Go tip
BenchmarkCloudyKitRouter-4       3000000               492 ns/op               0 B/op          0 allocs/op
BenchmarkHttprouterRouter-4      2000000              1006 ns/op             224 B/op          4 allocs/op
```


***

### Precedence example

On the example below the router will test the routes in the following order, /users/list then /users/:userId then /users/*page.
```go
	router.AddRoute("GET","/users/:userId",...)
	router.AddRoute("GET","/users/*page",...)
	router.AddRoute("GET","/users/list",...)
```
