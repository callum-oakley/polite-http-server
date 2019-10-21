# Polite HTTP server

This package exports a `Server` type which is a drop in replacement for
[`http.Server`][], but which will wait for long running HTTP/2 connections to
close before shutting down. This is a work around for [Go issue #29764][].

`Serve`, `ListenAndServe`, `ListenAndServeTLS` are not supported – please use
`ServeTLS`.

See [godoc][].

[`http.Server`]: https://golang.org/pkg/net/http/#Server
[Go issue #29764]: https://github.com/golang/go/issues/29764
[godoc]: https://godoc.org/github.com/callum-oakley/polite-http-server
