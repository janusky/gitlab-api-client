package utils

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/apex/log"
	"github.com/pkg/errors"
)

func PrintCSV(dataPrint []string) error {
	out := csv.NewWriter(os.Stdout)
	err := out.Write(dataPrint)
	out.Flush()
	return err
}

func PrintJSON(dataPrint interface{}) error {
	s, err := json.Marshal(dataPrint)
	if err != nil {
		log.Warnf("marshalling data %+v: %v", dataPrint, err)
	}
	fmt.Println(string(s))
	return err
}

func PrintPlain(dataPrint string) {
	if IsDebugEnabled() {
		log.Debug(dataPrint)
	} else {
		fmt.Println(dataPrint)
	}
}

func ExitOnError(err error) {
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}

func NotImplementedError(function string) error {
	return errors.Errorf("%s: not implemented: PRs are welcome at https://github.com/janusky/gitlab-api-client", function)
}

// IsDebugEnabled returns true if current log is Level set Debug
func IsDebugEnabled() bool {
	if logger, ok := log.Log.(*log.Logger); ok {
		return logger.Level == log.DebugLevel
	}
	return false
}

// HTTPClient creates a new http client
func HTTPClient(trustedCertificates [][]byte) (*http.Client, error) {
	pool, err := x509.SystemCertPool()
	if err != nil {
		return nil, errors.Wrap(err, "creating system certificate pool")
	}
	for _, cert := range trustedCertificates {
		if !pool.AppendCertsFromPEM(cert) {
			return nil, errors.Errorf("no certificates was parsed from %q", cert)
		}
	}
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: pool,
			},
		},
	}, nil
}
