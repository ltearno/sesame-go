package webserver

import (
	"application/assetsgen"
	"application/tools"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gbrlsnchs/jwt"
	"github.com/julienschmidt/httprouter"

	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
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

func handlerCerts(w http.ResponseWriter, r *http.Request, p httprouter.Params, server *WebServer) {
	response := ServerDescription{
		Keys: []KeyDescription{
			KeyDescription{
				Kid: server.kid,
				Kty: "RSA",
				Alg: "RSA256",
				Use: "sig",
				N:   server.n,
				E:   server.e,
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

func generateToken(crypto *jwt.RSASHA) string {
	now := time.Now()
	pl := JwtPayload{
		Payload: jwt.Payload{
			Issuer:         "gbrlsnchs",
			Subject:        "test",
			Audience:       jwt.Audience{"https://golang.org", "https://jwt.io"},
			ExpirationTime: jwt.NumericDate(now.Add(24 * 30 * 12 * time.Hour)),
			NotBefore:      jwt.NumericDate(now.Add(30 * time.Minute)),
			IssuedAt:       jwt.NumericDate(now),
			JWTID:          "foobar",
		},
		Foo: "foo",
		Bar: 1337,
	}

	token, err := jwt.Sign(pl, crypto)
	if err != nil {
		fmt.Println("cannot sign ! ", err.Error())
	}

	return string(token)
}

func handlerPostWebUILogin(w http.ResponseWriter, r *http.Request, p httprouter.Params, server *WebServer) {
	login := r.FormValue("login")
	password := r.FormValue("password")
	redirectURI := r.FormValue("redirect_uri")

	if login == "aaa" && password == "aaa" {
		redirectResponse(w, redirectURI+"#access_token="+generateToken(server.crypto))
	} else {
		redirectResponse(w, "index.html?message=error&redirect_uri="+redirectURI)
	}
}

func handlerGetWebUI(w http.ResponseWriter, r *http.Request, p httprouter.Params, server *WebServer) {
	relativePath := p.ByName("requested_resource")
	if strings.HasPrefix(relativePath, "/") {
		relativePath = relativePath[1:]
	}

	rawContentBytes, err := assetsgen.Asset("assets/webui/" + relativePath)
	if err != nil {
		errorResponse(w, 404, fmt.Sprintf("not found '%s'", relativePath))
		return
	}

	content := string(rawContentBytes)
	contentType := "application/octet-stream"

	if strings.HasSuffix(relativePath, ".md") {
		context := &PageContext{
			Name: "First context member",
		}

		contentType = "application/markdown"
		interpolated := interpolateTemplate(relativePath, content, context)
		if interpolated != nil {
			content = *interpolated
		}
	} else if strings.HasSuffix(relativePath, ".css") {
		contentType = "text/css"
	} else if strings.HasSuffix(relativePath, ".js") {
		contentType = "application/javascript"
	} else if strings.HasSuffix(relativePath, ".html") {
		contentType = "text/html"
	}

	httpResponse(w, 200, contentType, content)
}

// injects the WebServer context in http-router handler
func (server *WebServer) makeHandler(handler func(http.ResponseWriter, *http.Request, httprouter.Params, *WebServer)) func(http.ResponseWriter, *http.Request, httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		handler(w, r, p, server)
	}
}

func (server *WebServer) init(router *httprouter.Router) {
	router.GET("/certs", server.makeHandler(handlerCerts))
	router.POST("/ui/index.html", server.makeHandler(handlerPostWebUILogin))
	router.GET("/ui/*requested_resource", server.makeHandler(handlerGetWebUI))
	router.GET("/toto/:test", server.makeHandler(handlerDemo))
}

type WebServer struct {
	name       string
	privateKey *rsa.PrivateKey
	crypto     *jwt.RSASHA
	n          string
	e          string
	kid        string
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

	n := base64.RawURLEncoding.EncodeToString(privateKey.PublicKey.N.Bytes())
	e := encodeUint64ToString(uint64(privateKey.PublicKey.E))
	checksum := sha512.New()
	checksum.Write(privateKey.PublicKey.N.Bytes())
	checksum.Write(encodeUint64ToBytes(uint64(privateKey.PublicKey.E)))
	kid := base64.RawURLEncoding.EncodeToString(checksum.Sum(make([]byte, 0)))

	router := httprouter.New()
	if router == nil {
		fmt.Printf("Failed to instantiate the router, exit\n")
	}

	crypto := jwt.NewRS256(jwt.RSAPrivateKey(privateKey))

	server := &WebServer{
		name:       "sesame",
		privateKey: privateKey,
		crypto:     crypto,
		n:          n,
		e:          e,
		kid:        kid,
	}

	server.init(router)

	fmt.Printf("\n you can use your internet browser to go here : http://127.0.0.1:%d\n", port)

	log.Fatal(http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", port), router))
}
