package webserver

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
)

func loadPEMKey(fileName string) (key *rsa.PrivateKey) {
	bytes, err := ioutil.ReadFile(fileName)
	checkError(err)

	block, _ := pem.Decode(bytes)
	if block == nil {
		fmt.Println("Cannot decode key pem payload")
		os.Exit(1)
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	checkError(err)

	return privateKey
}

func savePEMKey(fileName string, key *rsa.PrivateKey) {
	outFile, err := os.Create(fileName)
	checkError(err)
	defer outFile.Close()

	var privateKey = &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}

	err = pem.Encode(outFile, privateKey)
	checkError(err)
}
