# Shortener

URL shortener

## Running

Go must be installed, `cd` into directory, then:

```sh
> go get
> go run main.go
```

## Usage

### Shortening URL

```sh
> curl localhost:8080 -X POST -d '{"url": "https://example.com"}'

{"success":true,"data":{"hash":"iuxUhiM"}}
```

### Following shortened URL

```sh
> curl localhost:8080/iuxUhiM

<a href="https://example.com">Moved Permanently</a>.
```

### Getting URL stats

```sh
> curl localhost:8080/iuxUhiM/stats

{"success":true,"data":{"day":1,"week":1,"total":1}}
```

## Motivation

- Used BoltDB as a storage, this allows to save time compared to traditional SQL databases (no network roundtrip, less marshalling/unmarshalling overhead). It's also ACID compliant;

- Getting URL and updating it's stats are performed in two different transactions (getting URL is read-only and non-blocking, stats update is performed after redirect is served, in background);

- Short URL generated is random base-62 number. There is reasonably low probablitity that new number already exists and if it does, shortener tries 2 more times and logs error when fails (so this issue can be handled when collisions start being relevant);

- Stats processing is performed in a single loop over all data points in a non-blocking read-only transaction;

- Some of the decisions are related to 4-hour development time: I've used integration testing (would be better to isolate tests a bit more by mocking), decided to hardcode configuration. Also, adding a bit more logging and monitoring of current state would be a good idea.

  

