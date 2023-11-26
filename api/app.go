package api

import (
	"context"
	"net/http"
	httpserver "swapi/api/server"
)

func Start(ctx context.Context, cancel context.CancelCauseFunc, options ...httpserver.HTTPServerOption) error {

	httpServer, err := httpserver.NewHTTPServer(options...)
	if err != nil {
		return err
	}

	retCh := make(chan error)

	go func() {
		retCh <- httpServer.ListenAndServe()
		close(retCh)
	}()

	select {
	case <-ctx.Done():

	case err := <-retCh:
		if err != http.ErrServerClosed {
			cancel(err)
		}
	}

	wctx, wcancel := context.WithTimeout(context.Background(), 30)
	defer wcancel()

	return httpServer.Shutdown(wctx)
}
