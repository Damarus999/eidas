package main

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/asn1"
	"encoding/pem"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/creditkudos/eidas"
	"github.com/creditkudos/eidas/qcstatements"
)

var countryCode = flag.String("country-code", "GB", "ISO-3166-1 Alpha 2 country code")
var orgName = flag.String("organization-name", "Credit Kudos Limited", "Organization name")
var orgID = flag.String("organization-id", "PSDGB-FCA-795791", "Organization ID")
var commonName = flag.String("common-name", "001580000103UAbAAM", "Common Name")
var roles = flag.String("roles", qcstatements.RoleAccountInformation, "eIDAS roles; comma-separated list from [PSP_AS, PSP_PI, PSP_AI, PSP_IC]")
var qcType = flag.String("type", "QWAC", "Certificate type; one of QWAC, QSIGN or QSEAL")

var outCSR = flag.String("csr", "out.csr", "Output file for CSR")
var outKey = flag.String("key", "out.key", "Output file for private key")

func writeCSR(path string, data []byte) (err error) {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer func() {
		if err2 := f.Close(); err2 != nil {
			if err == nil {
				err = err2
			}
		}
	}()
	return pem.Encode(f, &pem.Block{
		Type:  "CERTIFICATE REQUEST",
		Bytes: data,
	})
}

func writeKey(path string, key *rsa.PrivateKey) (err error) {
	pkcs8, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer func() {
		if err2 := f.Close(); err2 != nil {
			if err == nil {
				err = err2
			}
		}
	}()
	return pem.Encode(f, &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: pkcs8,
	})
}

func typeFromFlag(in string) (asn1.ObjectIdentifier, error) {
	if in == "QWAC" {
		return qcstatements.QWACType, nil
	} else if in == "QSEAL" {
		return qcstatements.QSEALType, nil
	}
	return nil, fmt.Errorf("Unknown QC type: %s", in)
}

func main() {
	flag.Parse()

	t, err := typeFromFlag(*qcType)
	if err != nil {
		log.Fatal(err)
	}

	d, key, err := eidas.GenerateCSR(
		*countryCode, *orgName, *orgID, *commonName, strings.Split(*roles, ","), t)
	if err != nil {
		log.Fatalf(":-( %v", err)
	}
	if err := writeCSR(*outCSR, d); err != nil {
		log.Fatalf("Failed to write CSR to %s: %v", *outCSR, err)
	}
	if err := writeKey(*outKey, key); err != nil {
		log.Fatalf("Failed to write key to %s: %v", *outKey, err)
	}
}
