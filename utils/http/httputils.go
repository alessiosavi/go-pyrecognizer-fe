package http

import (
	"encoding/json"
	stringutils "github.com/alessiosavi/GoGPUtils/string"
	"path"
	"strconv"

	fileutils "github.com/alessiosavi/GoGPUtils/files"
	log "github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
)

func ListAndServerGZIP(host string, _port int, gzipHandler fasthttp.RequestHandler) {
	port := strconv.Itoa(_port)
	log.Info("ListAndServerGZIP | Trying estabilishing connection @[http://", host, ":", port)
	err := fasthttp.ListenAndServe(host+":"+port, gzipHandler)
	if err != nil {
		panic(err)
	}
}

func ListAndServerSSL(host, _path, pub, priv string, _port int, gzipHandler fasthttp.RequestHandler) {
	pub = path.Join(_path, pub)
	priv = path.Join(_path, priv)
	if fileutils.FileExists(pub) && fileutils.FileExists(priv) {
		port := strconv.Itoa(_port)
		log.Info("ListAndServerSSL | Trying estabilishing connection @[https://", host, ":", port)
		err := fasthttp.ListenAndServeTLS(host+":"+port, pub, priv, gzipHandler)
		if err != nil {
			panic(err)
		}
	}
	log.Error("ListAndServerSSL | Unable to find certificates: pub[" + pub + "] | priv[" + priv + "]")
}

// SecureRequest Enhance the security with additional sec header
func SecureRequest(ctx *fasthttp.RequestCtx, ssl bool) {
	ctx.Response.Header.Set("Feature-Policy", "geolocation 'none'; microphone 'none'; camera 'self'")
	ctx.Response.Header.Set("Referrer-Policy", "no-referrer")
	ctx.Response.Header.Set("x-frame-options", "SAMEORIGIN")
	ctx.Response.Header.Set("X-Content-Type-Options", "nosniff")
	ctx.Response.Header.Set("X-Permitted-Cross-Domain-Policies", "none")
	ctx.Response.Header.Set("X-XSS-Protection", "1; mode=block")
	ctx.Response.Header.Set("Access-Control-Allow-Origin", "*")
	if ssl {
		ctx.Response.Header.Set("Content-Security-Policy", "upgrade-insecure-requests")
		ctx.Response.Header.Set("Strict-Transport-Security", "max-age=60; includeSubDomains; preload")
		ctx.Response.Header.Set("expect-ct", "max-age=60, enforce")
	}
}

// GetCredentials is delegated to extract the username and the password from the request body
func GetCredentials(ctx *fasthttp.RequestCtx) (string, string) {
	var user, pass string
	user = string(ctx.FormValue("user"))
	pass = string(ctx.FormValue("pass"))
	if stringutils.IsBlank(user) && stringutils.IsBlank(pass) {
		type req struct {
			User string `json:"user,omitempty"`
			Pass string `json:"pass,omitempty"`
		}
		var r req
		err := json.Unmarshal(ctx.PostBody(), &r)
		if err != nil {
			return "", ""
		}
		user = r.User
		pass = r.Pass
	}
	return user, pass
}

// ValidateUsername execute few check on the username in input
func ValidateUsername(username string) bool {
	if stringutils.IsBlank(username) {
		log.Warn("Username is empty :/")
		return false
	}
	if len(username) < 4 || len(username) > 32 {
		log.Warn("Username len not valid")
		return false
	}

	log.Debug("Username [", username, "] VALIDATED!")
	return true
}

// PasswordValidation execute few check on the password in input
func PasswordValidation(password string) bool {
	if stringutils.IsBlank(password) {
		log.Warn("Password is empty :/")
		return false
	}
	if len(password) < 4 || len(password) > 32 {
		log.Warn("Password len not valid")
		return false
	}
	log.Info("Password [", password, "] VALIDATED!")
	return true
}

// ValidateCredentials is wrapper for the multiple method for validate the input parameters
func ValidateCredentials(user string, pass string) bool {
	if ValidateUsername(user) && PasswordValidation(pass) {
		return true
	}
	return false
}
