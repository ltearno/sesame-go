package webserver

import (
	"application/assetsgen"
	"application/tools"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gbrlsnchs/jwt"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"

	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
)

type JwtPayload struct {
	jwt.Payload
	Uuid  string `json:"uuid"`
	Role  string `json:"role"`
	Roles string `json:"roles"`
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

func generateToken(server *WebServer, userId string, duration uint64) string {
	now := time.Now()
	pl := JwtPayload{
		Payload: jwt.Payload{
			Issuer:         server.config.IssuerUrl,
			Subject:        server.config.Company,
			Audience:       jwt.Audience{"IDP"},
			ExpirationTime: jwt.NumericDate(now.Add(time.Duration(duration) * time.Second)),
			IssuedAt:       jwt.NumericDate(now),
			JWTID:          uuid.New().String(),
		},

		Uuid:  userId,
		Role:  "{}",
		Roles: "{}",
	}

	token, err := jwt.Sign(pl, server.crypto, jwt.KeyID(server.kid))
	if err != nil {
		fmt.Println("cannot sign ! ", err.Error())
		return ""
	}

	fmt.Println(now, " : generated jwt token for ", userId, " jti ", pl.Payload.JWTID, " duration ", duration)

	return string(token)
}

func handlerCreateIdToken(w http.ResponseWriter, r *http.Request, p httprouter.Params, server *WebServer) {
	// be careful that the x-forwarded-user header is used as is
	// because it is assumed that the sesame executable is protected
	// by a gateway and runs in a safe environment.
	userID := r.Header.Get("x-forwarded-user")
	if _, ok := server.config.Users[userID]; ok {
		jsonResponse(w, 200, struct {
			Token string `json:"token"`
		}{Token: generateToken(server, userID, server.config.IdTokenDurationSecs)})
	} else {
		unauthorizedResponse(w)
	}
}

func handlerPostAuth(w http.ResponseWriter, r *http.Request, p httprouter.Params, server *WebServer) {
	// be careful that the x-forwarded-user header is used as is
	// because it is assumed that the sesame executable is protected
	// by a gateway and runs in a safe environment.
	userID := r.Header.Get("x-forwarded-user")
	if _, ok := server.config.Users[userID]; ok {
		jsonResponse(w, 200, struct {
			Token string `json:"token"`
		}{Token: generateToken(server, userID, server.config.TokenApiDurationSecs)})
	} else {
		unauthorizedResponse(w)
	}
}

func handlerPostWebUILogin(w http.ResponseWriter, r *http.Request, p httprouter.Params, server *WebServer) {
	login := r.FormValue("login")
	password := r.FormValue("password")
	redirectURI := r.FormValue("redirect_uri")

	if server.config.Users[login] == password {
		redirectResponse(w, redirectURI+"#access_token="+generateToken(server, login, server.config.TokenDurationSecs))
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
	router.POST("/create-id-token", server.makeHandler(handlerCreateIdToken))
	router.POST("/auth", server.makeHandler(handlerPostAuth))
}

type WebServer struct {
	config     *ConfigurationFile
	name       string
	privateKey *rsa.PrivateKey
	crypto     *jwt.RSASHA
	n          string
	e          string
	kid        string
}

// Start runs a webserver hosting the application
func Start(port int, workingDir string) {
	fmt.Println("starting web server")

	if !tools.ExistsFile(filepath.Join(workingDir, "private.pem")) {
		fmt.Println("generating private key...")
		reader := rand.Reader
		bitSize := 2048

		privateKey, err := rsa.GenerateKey(reader, bitSize)
		checkError(err)

		savePEMKey(filepath.Join(workingDir, "private.pem"), privateKey)
	}

	fmt.Println("loading private key...")
	privateKey := loadPEMKey(filepath.Join(workingDir, "private.pem"))

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
		config:     ReadConfiguration(filepath.Join(workingDir, "configuration.json")),
		name:       "sesame",
		privateKey: privateKey,
		crypto:     crypto,
		n:          n,
		e:          e,
		kid:        kid,
	}

	server.init(router)

	fmt.Printf("\n you can use your internet browser to go here : https://127.0.0.1:%d\n", port)

	log.Fatal(http.ListenAndServeTLS(fmt.Sprintf("0.0.0.0:%d", port), filepath.Join(workingDir, "tls.cert.pem"), filepath.Join(workingDir, "tls.key.pem"), router))
}
