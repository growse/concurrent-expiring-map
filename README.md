# concurrent map [![Circle CI](https://circleci.com/gh/streamrail/concurrent-map.png?style=badge)](https://circleci.com/gh/streamrail/concurrent-map)

Forked from original project: [https://github.com/streamrail/concurrent-map](https://github.com/streamrail/concurrent-map). I've modified the functionality to add item expiry, and the values are now just byte slices - it's getting to be a little bit more like an in-process memcache.

Golang map doesn't support concurrent reads and writes, Please see (http://golang.org/doc/faq#atomic_maps and http://blog.golang.org/go-maps-in-action), in case you're using multiple Go routines to read and write concurrently from a map some form of guard mechanism should be in-place.

Concurrent map is a wrapper around Go's map, more specifically around a String -> interface{} kinda map, which enforces concurrency.

## usage

Import the package:

```go
import (
	"github.com/growse/concurrent-expiring-map"
)

```
and go get it using the goapp gae command:

```bash
goapp get "github.com/growse/concurrent-expiring-map"
```

The package is now imported under the "cmap" namespace. 

## example


```go

	// Create a new map.
	map := cmap.New()
	
	// Sets item within map, sets "bar" under key "foo", expiring 5 hours from now
	map.Set("foo", []byte("bar"), time.Now().Add(time.Hour*5)

	// Retrieve item from map.
	tmp, ok := map.Get("foo")

	// Checks if item exists
	if ok {
		// Map stores items as ConcurrentMapItem, which contains the expiry and the byte slice
		bar := string(tmp.Value)
	}

	// Removes item under key "foo"
	map.Remove("foo")

```

For more examples have a look at concurrent_map_test.go.


Running tests:
```bash
go test "github.com/growse/concurrent-expiring-map"
```


## license 
MIT (see [LICENSE](https://github.com/streamrail/concurrent-map/blob/master/LICENSE) file)
