## CloudyKit Router

CloudyKit Router was developed to be a very fast router on matching and retrieving of parameters.

### Main characteristics

 1. Use prefix tree with special nodes to match dynamic sentences and wildcard.
 2. No allocations are all, matching and retrieve parameters don't allocates.
 3. Nodes has precedence when matching, text node then single sentence node then wildcard node

***

### Precedence example

On the example below the router will test the routes in the following order, /users/list then /users/:userId then /users/*page.
```go
	router.AddRoute("GET","/users/:userId",...)
	router.AddRoute("GET","/users/*page",...)
	router.AddRoute("GET","/users/list",...)
```