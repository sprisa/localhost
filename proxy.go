package main

import (
	"context"
	"crypto/tls"
	_ "embed"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/samber/lo"
	"github.com/sprisa/localhost/util"
)

//go:embed ssl/cert.pem
var cert []byte

//go:embed ssl/key.pem
var certKey []byte

func StartProxyService(
	ctx context.Context,
	tlsCert tls.Certificate,
	addrIp net.IP,
	listenPort int,
	hostPort int,
	availableSubdomains []string,
) error {
	log := util.Log.With().Int("targetPort", hostPort).Logger()
	handler := http.NewServeMux()

	target := fmt.Sprintf("http://127.0.0.1:%d", hostPort)
	url, err := url.Parse(target)
	if err != nil {
		log.Err(err).Msg("error building url")
		return err
	}
	proxy := httputil.NewSingleHostReverseProxy(url)

	// Proxy request handler
	handler.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		res.Header().Add("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		proxy.ServeHTTP(res, req)
	})

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
		util.Log.Info().Msg("Shutting down proxy server.")
		closeError = server.Close()
	}()

	util.Log.Info().Msgf(
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
			util.Log.Info().Msg("Proxy server closed successfully.")
		default:
			return err
		}
	}
	return closeError
}
