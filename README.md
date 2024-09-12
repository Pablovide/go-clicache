## Go CLI Http cache proxy 

Run the module using
```bash
go run ./main.go -port <PORT> -origin <HTTP_SERVER>
```
This proxies the origin address, storing in memory cache the HTTP GET requests' responses.

Cache keys lifespan is 5 min.

Every 10 min or killing the process, the cache will be cleaned.