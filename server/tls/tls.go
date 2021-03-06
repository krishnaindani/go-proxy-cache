package tls

//                                                                         __
// .-----.-----.______.-----.----.-----.--.--.--.--.______.----.---.-.----|  |--.-----.
// |  _  |  _  |______|  _  |   _|  _  |_   _|  |  |______|  __|  _  |  __|     |  -__|
// |___  |_____|      |   __|__| |_____|__.__|___  |      |____|___._|____|__|__|_____|
// |_____|            |__|                   |_____|
//
// Copyright (c) 2020 Fabio Cicerchia. https://fabiocicerchia.it. MIT License
// Repo: https://github.com/fabiocicerchia/go-proxy-cache

import (
	crypto_tls "crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/fabiocicerchia/go-proxy-cache/config"
	"golang.org/x/crypto/acme/autocert"
)

var httpsDomains []string
var certificates map[string]*crypto_tls.Certificate
var tlsConfig *crypto_tls.Config

// ServerOverrides - Overrides the http.Server configuration for TLS.
func ServerOverrides(domain string, server http.Server, domainConfig config.Server) http.Server {
	newServer := server
	var err error

	tlsConfig, err = Config(domain, domainConfig.TLS.CertFile, domainConfig.TLS.KeyFile)
	if err != nil {
		log.Fatal(err)
		return newServer
	}
	newServer.TLSConfig = tlsConfig
	// TODO: check this: server.TLSNextProto = make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),

	if domainConfig.TLS.Auto {
		certManager := InitCertManager(domainConfig.Upstream.Host, domainConfig.TLS.Email)

		newServer.TLSConfig = certManager.TLSConfig()
	}

	return newServer
}

// Config - Returns a TLS configuration.
func Config(domain string, certFile string, keyFile string) (*crypto_tls.Config, error) {
	cert, err := crypto_tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	// G402 (CWE-295): TLS MinVersion too low. (Confidence: HIGH, Severity: HIGH)
	// It can be ignored as it is customisable, but the default is TLSv1.2.
	tlsConfig := &crypto_tls.Config{
		// Causes servers to use Go's default ciphersuite preferences,
		// which are tuned to avoid attacks. Does nothing on clients.
		PreferServerCipherSuites: true,
		CurvePreferences:         config.Config.Server.TLS.Override.CurvePreferences,
		MinVersion:               config.Config.Server.TLS.Override.MinVersion,
		MaxVersion:               config.Config.Server.TLS.Override.MaxVersion,
		CipherSuites:             config.Config.Server.TLS.Override.CipherSuites,
		GetCertificate:           returnCert,
	} // #nosec

	if len(certificates) == 0 {
		certificates = make(map[string]*crypto_tls.Certificate)
	}
	certificates[domain] = &cert

	// If GetCertificate is nil or returns nil, then the certificate is
	// retrieved from NameToCertificate. If NameToCertificate is nil, the
	// best element of Certificates will be used.
	// Ref: https://golang.org/pkg/crypto/tls/#Config.GetCertificate
	for _, c := range certificates {
		tlsConfig.Certificates = append(tlsConfig.Certificates, *c)
	}

	return tlsConfig, nil
}

func returnCert(helloInfo *crypto_tls.ClientHelloInfo) (*crypto_tls.Certificate, error) {
	log.Debugf("HelloInfo: %v\n", helloInfo)
	if val, ok := certificates[helloInfo.ServerName]; ok {
		return val, nil
	}
	return nil, fmt.Errorf("missing certificate for %s", helloInfo.ServerName)
}

// InitCertManager - Initialise the Certification Manager for auto generation.
func InitCertManager(host string, email string) *autocert.Manager {
	cacheDir, err := ioutil.TempDir("", "cache_dir")
	if err != nil {
		log.Fatal(err)
		return nil
	}

	httpsDomains = append(httpsDomains, host)

	certManager := &autocert.Manager{
		Cache:      autocert.DirCache(cacheDir),
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(httpsDomains...),
		Email:      email,
	}

	return certManager
}
