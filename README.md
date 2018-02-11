# HTTP Cache

## Description

HTTP Cache is a caching reverse-proxy. It's a way to simulate a CDN like Cloudlare or Fastly on your development environment.

## Installation

```
go get -u github.com/panoplymedia/httpcache
```

## Example Usage

We would like to cache the output of our local server that is serving audio (mp3) files. Let's say that our local server runs on port `8080`. We will run the cache on port `3003` (any open port will do).

```
httpcache --port=3003 --origin=http://localhost:8080
```

Now we can point any other service that is using our audio service to port `3003`, our local caching service, instead of our actual service that is running at `8080`.

Now requests will be routed as follows:

- `http://localhost:3003/audio/first.mp3` cache miss -> `http://localhost:8080/audio/first.mp3`
- `http://localhost:3003/audio/first.mp3` cache hit -> cached file
- `http://localhost:3003/audio/second.mp3` cache miss -> `http://localhost:8080/audio/second.mp3`

## Help

```
httpcache --help
```

## Notes

The cache is transient, in that it will reset it's state after every run. So if it's running and you kill it, when you start it back up, it's cache will be empty.
