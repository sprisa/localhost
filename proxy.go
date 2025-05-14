package main

import (
	"context"
	"crypto/tls"
	_ "embed"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/samber/lo"
)

//go:embed ssl/cert.pem
var cert []byte

//go:embed ssl/key.pem
var certKey []byte

func StartProxyService(
	ctx context.Context,
	tlsCert tls.Certificate,
	addrIp string,
	listenPort int,
	hostPort int,
	availableSubdomains []string,
) error {
	log := Log.With().Int("targetPort", hostPort).Logger()
	handler := http.NewServeMux()

	target := fmt.Sprintf("http://127.0.0.1:%d", hostPort)
	url, err := url.Parse(target)
	if err != nil {
		log.Err(err).Msg("error building url")
		return err
	}
	proxy := httputil.NewSingleHostReverseProxy(url)

	// Proxy request handler
	handler.HandleFunc("/", proxy.ServeHTTP)

	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", addrIp, listenPort),
		Handler: handler,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{tlsCert},
		},
	}
	var closeError error
	// Shutdown handler
	go func() {
		<-ctx.Done()
		Log.Info().Msg("Shutting down proxy server.")
		closeError = server.Close()
	}()

	Log.Info().Msgf(
		"localhost proxy up\n%s",
		strings.Join(
			lo.Map(availableSubdomains, func(subdomain string, _ int) string {
				return fmt.Sprintf("  - https://%s.svc.host:%d", subdomain, listenPort)
			}),
			"\n",
		),
	)

	err = server.ListenAndServeTLS("", "")
	if err != nil {
		switch err {
		case http.ErrServerClosed:
			Log.Info().Msg("Proxy server closed successfully.")
		default:
			return err
		}
	}
	return closeError
}
