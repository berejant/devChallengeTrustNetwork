# DEV Challenge XIX: Backend Online Round - Trust network

Implementation of [Online Round Task Backend | DEV Challenge XIX](https://docs.google.com/document/d/1fJcMrI3MQEze8QWbnL3ltiIR9_W1ShdL7KoFXUFNjpE/)


# Stack
1. [Arangodb](https://www.arangodb.com/) - Powerful Graph database with excellent [Performance](https://www.arangodb.com/2018/02/nosql-performance-benchmark-2018-mongodb-postgresql-orientdb-neo4j-arangodb/)
2. [Golang](https://go.dev/) - high performance language for make API-layer between client and database with minimal delay
3. [Gin Web Framework](https://github.com/gin-gonic/gin) - help build API fast: router, request validation, building response.
3. [Postman](https://www.postman.com/) - Useful tool which provide UI for create API tests and next export tests as JSON to run inside [docker runner](https://hub.docker.com/r/postman/newman/)

## Run app
> docker compose up -d

## See it works:
> curl -i http://127.0.0.1:8080/healthcheck

> curl -i -X POST http://127.0.0.1:8080/api/people

Note: Port 8080, protocol http. 
no sense in ssl on local, also by common rule port 8080 is http (I assume if we need ssl specs should include for port 8443)

## Run tests
> docker compose run postman

Example results:
```
┌─────────────────────────┬─────────────────┬─────────────────┐
│                         │        executed │          failed │
├─────────────────────────┼─────────────────┼─────────────────┤
│              iterations │               2 │               0 │
├─────────────────────────┼─────────────────┼─────────────────┤
│                requests │             280 │               0 │
├─────────────────────────┼─────────────────┼─────────────────┤
│            test-scripts │             560 │               0 │
├─────────────────────────┼─────────────────┼─────────────────┤
│      prerequest-scripts │               0 │               0 │
├─────────────────────────┼─────────────────┼─────────────────┤
│              assertions │             344 │               0 │
├─────────────────────────┴─────────────────┴─────────────────┤
│ total run duration: 6.9s                                    │
├─────────────────────────────────────────────────────────────┤
│ total data received: 2.47kB (approx)                        │
├─────────────────────────────────────────────────────────────┤
│ average response time: 5ms [min: 3ms, max: 97ms, s.d.: 5ms] │
└─────────────────────────────────────────────────────────────┘
```

## Run load testing
> docker compose run siege

Note:
 - load testing should be executed after postman tests (for populate database with data)
 - you can see siege log at `siege/log/siege.log`
 - load testing inside docker on same machine is not so representative. Better to use two hosts: API-server and siege-client.
 - for having complex load testing (closest to real data) we need to prepare Graph with complex edges and big depth.

Siege result on my machine:
 - Docker resources: 1.4 GHz Quad-Core Intel Core i5, 4 (v)CPUs, RAM 8 GB, 2 GB Swap.
```
Transactions:                  92998 hits
Availability:                 100.00 %
Elapsed time:                  59.98 secs
Data transferred:               5.34 MB
Response time:                  0.14 secs
Transaction rate:            1550.48 trans/sec
Throughput:                     0.09 MB/sec
Concurrency:                  220.02
Successful transactions:       86737
Failed transactions:               0
Longest transaction:            2.94
Shortest transaction:           0.00
```

## Optional. See visualization of Graph using UI
[ArangoDB WebUI Graph diagram](http://127.0.0.1:8529/_db/trustNetwork/_admin/aardvark/index.html#graph/trustNetwork)

Note: use button in top right corner `Show full graph` or change `Graph configue settings` for see whole picture.

## Optional. How to view tests.
Import collection `tests/DevChallangeTrustNetwork.postman_collection.json` into [Postman app](https://web.postman.co/).
Did not give collection public link because it can be associated with my personal data.