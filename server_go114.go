//go:build go1.14
// +build go1.14

package loong

import (
	"crypto/tls"
	"strings"
)

func CipherSuites() []*tls.CipherSuite {
	return tls.CipherSuites()
}

func SetCipherSuites(cfg *tls.Config, values string) {
	SetCipherSuitesWithNames(cfg, strings.Split(values, ","))
}

func SetCipherSuitesWithNames(cfg *tls.Config, values []string) {
	cfg.CipherSuites = nil
	for _, name := range values {
		name = strings.TrimSpace(name)
		name = strings.ToUpper(name)

		for _, cipherSuite := range tls.CipherSuites() {
			if cipherSuite.Name == name {
				cfg.CipherSuites = append(cfg.CipherSuites, cipherSuite.ID)
				break
			}
		}
		for _, cipherSuite := range tls.InsecureCipherSuites() {
			if cipherSuite.Name == name {
				cfg.CipherSuites = append(cfg.CipherSuites, cipherSuite.ID)
				break
			}
		}
	}
}
