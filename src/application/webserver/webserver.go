package webserver

import (
	"application/tools"
	"encoding/base64"
	"encoding/binary"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gbrlsnchs/jwt"
	"github.com/julienschmidt/httprouter"

	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"crypto/x509"
)

type JwtPayload struct {
	jwt.Payload
	Foo string `json:"foo,omitempty"`
	Bar int    `json:"bar,omitempty"`
}

type KeyDescription struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	Alg string `json:"alg"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type ServerDescription struct {
	Keys []KeyDescription `json:"keys"`
}

func encodeUint64ToBytes(v uint64) []byte {
	data := make([]byte, 8)
	binary.BigEndian.PutUint64(data, v)
	i := 0
	for ; i < len(data); i++ {
		if data[i] != 0x0 {
			break
		}
	}

	return data[i:]
}

func encodeUint64ToString(v uint64) string {
	return base64.RawURLEncoding.EncodeToString(encodeUint64ToBytes(v))
}

func handlerCerts(w http.ResponseWriter, r *http.Request, p httprouter.Params, server *WebServer) {
	n := base64.RawURLEncoding.EncodeToString(server.privateKey.PublicKey.N.Bytes())
	e := encodeUint64ToString(uint64(server.privateKey.PublicKey.E))

	checksum := sha512.New()
	checksum.Write(server.privateKey.PublicKey.N.Bytes())
	checksum.Write(encodeUint64ToBytes(uint64(server.privateKey.PublicKey.E)))
	kid := base64.RawURLEncoding.EncodeToString(checksum.Sum(make([]byte, 0)))

	response := ServerDescription{
		Keys: []KeyDescription{
			KeyDescription{
				Kid: kid,
				Kty: "RSA",
				Alg: "RSA256",
				Use: "sig",
				N:   n,
				E:   e,
			},
		},
	}

	jsonResponse(w, 200, response)
}

func handlerDemo(w http.ResponseWriter, r *http.Request, p httprouter.Params, server *WebServer) {
	test := p.ByName("test")

	now := time.Now()
	pl := JwtPayload{
		Payload: jwt.Payload{
			Issuer:         "gbrlsnchs",
			Subject:        test,
			Audience:       jwt.Audience{"https://golang.org", "https://jwt.io"},
			ExpirationTime: jwt.NumericDate(now.Add(24 * 30 * 12 * time.Hour)),
			NotBefore:      jwt.NumericDate(now.Add(30 * time.Minute)),
			IssuedAt:       jwt.NumericDate(now),
			JWTID:          "foobar",
		},
		Foo: "foo",
		Bar: 1337,
	}

	token, err := jwt.Sign(pl, server.crypto)
	if err != nil {
		fmt.Println("cannot sign ! ", err.Error())
	}

	jsonResponse(w, 200, struct {
		Toto string
	}{
		Toto: string(token),
	})
}

// injects the WebServer context in http-router handler
func (server *WebServer) makeHandler(handler func(http.ResponseWriter, *http.Request, httprouter.Params, *WebServer)) func(http.ResponseWriter, *http.Request, httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		handler(w, r, p, server)
	}
}

func (server *WebServer) init(router *httprouter.Router) {
	router.GET("/certs", server.makeHandler(handlerCerts))
	router.GET("/toto/:test", server.makeHandler(handlerDemo))
}

type WebServer struct {
	name       string
	privateKey *rsa.PrivateKey
	crypto     *jwt.RSASHA
}

// Start runs a webserver hosting the application
func Start(port int) {
	fmt.Println("starting web server")

	if !tools.ExistsFile("private.pem") {
		fmt.Println("generating private key...")
		reader := rand.Reader
		bitSize := 2048

		privateKey, err := rsa.GenerateKey(reader, bitSize)
		checkError(err)

		savePEMKey("private.pem", privateKey)
	}

	fmt.Println("loading private key...")
	privateKey := loadPEMKey("private.pem")

	router := httprouter.New()
	if router == nil {
		fmt.Printf("Failed to instantiate the router, exit\n")
	}

	crypto := jwt.NewRS256(jwt.RSAPrivateKey(privateKey))

	server := &WebServer{
		name:       "sesame",
		privateKey: privateKey,
		crypto:     crypto,
	}

	server.init(router)

	fmt.Printf("\n you can use your internet browser to go here : http://127.0.0.1:%d\n", port)

	log.Fatal(http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", port), router))
}

func checkError(err error) {
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
		os.Exit(1)
	}
}

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
