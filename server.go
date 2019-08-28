package polite

import (
	"context"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Server is like http.Server, but doesn't prematurely kill http2 connections
// on Shutdown. See below.
type Server struct {
	http.Server
	activeHandlersWg sync.WaitGroup
	shutdown         chan struct{}
}

// New wraps a http.Server, overloading the Shutdown and ServeTLS methods.
func New(hs http.Server) *Server {
	s := &Server{Server: hs, shutdown: make(chan struct{})}
	s.Handler = s.wrapHandler(s.Handler)
	return s
}

// Shutdown closes listeners and blocks until all of the server's handlers have
// returned, only then initialising shutdown of the underlying http server.
//
// Workaround for https://github.com/golang/go/issues/29764.
func (s *Server) Shutdown(ctx context.Context) error {
	close(s.shutdown)
	s.activeHandlersWg.Wait()
	// TODO Why do we need this to avoid an unexpected EOF?
	// Hunch without any investigation: when the handler function returns we
	// haven't completely finished cleanly closing the stream, so if we don't
	// wait for a moment Server.Shutdown will still aggresively close it's
	// connections, which will result in an EOF. Is a millisecond always
	// enough?
	time.Sleep(time.Millisecond)
	return s.Server.Shutdown(ctx)
}

// ServeTLS keeps a reference to the listener so that it can be closed on
// shutdown but otherwise delegates to the underlying http server.
func (s *Server) ServeTLS(l net.Listener, certFile, keyFile string) error {
	go func() {
		<-s.shutdown
		l.Close()
	}()

	err := s.Server.ServeTLS(l, certFile, keyFile)

	select {
	case <-s.shutdown:
		if isClosedConnErr(err) {
			// We caused the close error by closing the listener early.
			return http.ErrServerClosed
		}
	default:
	}

	return err
}

// Serve is unsupported. Use ServeTLS.
func (s *Server) Serve(l net.Listener) error {
	panic("unimplemented")
}

// ListenAndServe is unsupported. Use ServeTLS.
func (s *Server) ListenAndServe() error {
	panic("unimplemented")
}

// ListenAndServeTLS is unsupported. Use ServeTLS.
func (s *Server) ListenAndServeTLS(certFile, keyFile string) error {
	panic("unimplemented")
}

func (s *Server) wrapHandler(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.activeHandlersWg.Add(1)
		defer s.activeHandlersWg.Done()
		handler.ServeHTTP(w, r)
	})
}

func isClosedConnErr(err error) bool {
	return strings.Contains(err.Error(), "use of closed network connection")
}
