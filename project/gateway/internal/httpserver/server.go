package httpserver

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"
)

func StartServers(public, internal *http.Server, stop <-chan struct{}) {
	var wg sync.WaitGroup

	start := func(name string, srv *http.Server) {
		wg.Add(1)
		go func() {
			log.Printf("%s server started on %s", name, srv.Addr)
			if err := srv.ListenAndServe(); err != http.ErrServerClosed {
				log.Fatalf("%s stopped with error: %v", name, err)
			}
			wg.Done()
		}()
	}

	start("Public API", public)
	start("Internal API", internal)

	<-stop
	log.Println("Shutting down servers...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	public.Shutdown(ctx)
	internal.Shutdown(ctx)

	wg.Wait()
	log.Println("Servers stopped gracefully")
}
