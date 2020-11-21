package tls

import (
	"crypto/tls"
	"io/ioutil"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/fabiocicerchia/go-proxy-cache/config"
	"golang.org/x/crypto/acme/autocert"
)

// CertificatePair - Pair of certificate and key.
type CertificatePair struct {
	Cert string
	Key  string
}

// ServerOverrides - Overrides the http.Server configuration for TLS.
func ServerOverrides(
	server *http.Server,
	certManager *autocert.Manager,
	certPair *CertificatePair,
) {
	tlsConfig, err := Config(*&certPair.Cert, *&certPair.Key)
	if err != nil {
		log.Fatal(err)
		return
	}
	server.TLSConfig = tlsConfig

	if config.Config.Server.TLS.Auto {
		server.TLSConfig = certManager.TLSConfig()
	}
}

// Config - Returns a TLS configuration.
func Config(certFile string, keyFile string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	tlsConfig := &tls.Config{
		// Causes servers to use Go's default ciphersuite preferences,
		// which are tuned to avoid attacks. Does nothing on clients.
		PreferServerCipherSuites: true,
		// Only use curves which have assembly implementations
		CurvePreferences: []tls.CurveID{
			tls.CurveP256,
			tls.X25519, // Go 1.8 only
		},
		// TODO: Configurable
		MinVersion: tls.VersionTLS12,
		// TODO: Configurable
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305, // Go 1.8 only
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,   // Go 1.8 only
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,

			// Best disabled, as they don't provide Forward Secrecy,
			// but might be necessary for some clients
			// tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			// tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
		},
		Certificates: []tls.Certificate{cert},
	}

	return tlsConfig, nil
}

// InitCertManager - Initialise the Certification Manager for auto generation.
func InitCertManager(host string, email string) *autocert.Manager {
	cacheDir, err := ioutil.TempDir("", "cache_dir")
	if err != nil {
		log.Fatal(err)
		return nil
	}

	certManager := &autocert.Manager{
		Cache:      autocert.DirCache(cacheDir),
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(host),
		Email:      email,
	}

	return certManager
}
