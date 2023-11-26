package httpserver

import (
	"context"
	"net"
	"net/http"
	"net/netip"
	"time"
)

var defaultTimeout = time.Duration(30 * time.Second)
var defaultMaxHeaderBytes = 8192

type httpServerOptions struct {
	addr           string
	timeout        *time.Duration
	maxHeaderBytes *int
	baseContext    func(net.Listener) context.Context
	handler        http.Handler
}

type HTTPServerOption func(options *httpServerOptions) error

func NewHTTPServer(options ...HTTPServerOption) (*http.Server, error) {
	var err error

	opts := httpServerOptions{
		timeout:        &defaultTimeout,
		maxHeaderBytes: &defaultMaxHeaderBytes,
		baseContext: func(net.Listener) context.Context {
			return context.Background()
		},
	}

	for _, option := range options {
		if err = option(&opts); err != nil {
			return nil, err
		}
	}

	httpServer := &http.Server{
		Addr:                         opts.addr,
		Handler:                      nil,
		DisableGeneralOptionsHandler: false,
		TLSConfig:                    nil,
		ReadTimeout:                  *opts.timeout,
		ReadHeaderTimeout:            *opts.timeout,
		WriteTimeout:                 *opts.timeout,
		IdleTimeout:                  *opts.timeout,
		MaxHeaderBytes:               *opts.maxHeaderBytes,
		ErrorLog:                     nil,
		BaseContext:                  opts.baseContext,
	}

	return httpServer, nil
}

func WithTimeout(t time.Duration) HTTPServerOption {
	return func(options *httpServerOptions) error {
		*options.timeout = t

		return nil
	}
}

func WithAddress(j *netip.AddrPort) HTTPServerOption {
	return func(options *httpServerOptions) error {
		if j != nil {
			options.addr = j.String()
		}

		return nil
	}
}

func WithMaxHeaderBytes(s int) HTTPServerOption {
	return func(options *httpServerOptions) error {
		if *options.maxHeaderBytes = s; s < 0 {
			*options.maxHeaderBytes = 0
		}

		return nil
	}
}

func WithBaseContext(c func(net.Listener) context.Context) HTTPServerOption {
	return func(options *httpServerOptions) error {
		options.baseContext = c

		return nil
	}
}

func WithHandler(h http.Handler) HTTPServerOption {
	return func(options *httpServerOptions) error {
		options.handler = h

		return nil
	}
}
