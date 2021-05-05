package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/alessiosavi/GoGPUtils/helper"
	"github.com/alessiosavi/go-pyrecognizer-fe/datastructure"
	"github.com/alessiosavi/go-pyrecognizer-fe/utils/common"
	httputils "github.com/alessiosavi/go-pyrecognizer-fe/utils/http"
	redisutils "github.com/alessiosavi/go-pyrecognizer-fe/utils/redis"
	"github.com/dgrijalva/jwt-go"
	"github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/expvarhandler"
	"io/ioutil"
	"strings"
	"time"
)

func main() {
	Formatter := new(log.JSONFormatter)
	Formatter.TimestampFormat = "Jan _2 15:04:05.000000000"
	Formatter.PrettyPrint = false
	log.SetFormatter(Formatter)
	log.SetLevel(log.TraceLevel)
	log.SetReportCaller(true)
	log.Debug("HELLO WORLD!")

	log.Debug("Reading configuration file ...")
	data, err := ioutil.ReadFile("conf/test.json")
	if err != nil {
		panic(err)
	}
	var conf datastructure.Configuration
	if err = json.Unmarshal(data, &conf); err != nil {
		panic(err)
	}

	log.Debug("Connecting to DB ...")
	dbClient, err := redisutils.Connect("", "", 0)
	if err != nil {
		panic(err)
	}
	log.Debug("Binding HTTP ...")
	handleRequests(conf, dbClient)
}

// handleRequests Is delegated to map (BIND) the API methods to the HTTP URL
// It use a gzip handler that is usefull for reduce bandwitch usage while interacting with the middleware function
func handleRequests(cfg datastructure.Configuration, redisClient *redis.Client) {
	m := func(ctx *fasthttp.RequestCtx) {
		httputils.SecureRequest(ctx, cfg.SSL.Enabled)
		ctx.Response.Header.Set("go-pyrecognizer-fe", "$v0.0.1")

		// Avoid to print stats for the expvar handler
		if strings.Compare(string(ctx.Path()), "/stats") != 0 {
			log.Info(common.DebugRequest(ctx))
		}
		switch string(ctx.Path()) {
		case "/stats":
			expvarhandler.ExpvarHandler(ctx)
		case "/register":
			json.NewEncoder(ctx).Encode(RegisterUser(ctx, redisClient))
		case "/login":
			json.NewEncoder(ctx).Encode(LoginUser(ctx, redisClient))
		case "/remove":
			json.NewEncoder(ctx).Encode(RemoveUser(ctx, redisClient))
		case "/predict":
			break
		case "/train":
			break
		default:
			json.NewEncoder(ctx).Encode(datastructure.Response{Status: false, Description: "Url does not exists!", ErrorCode: "URL_NOT_EXIST", Data: nil})
			ctx.Response.SetStatusCode(404)
		}
	}
	// ==== GZIP HANDLER ====
	// The gzipHandler will serve a compress request only if the client request it with headers (Content-Type: gzip, deflate)
	gzipHandler := fasthttp.CompressHandlerLevel(m, fasthttp.CompressBestSpeed) // Compress data before sending (if requested by the client)
	log.Info("Binding services to @[", cfg.Host, ":", cfg.Port)

	// ==== SSL HANDLER + GZIP if requested ====
	if cfg.SSL.Enabled {
		log.Debug("SSL is enabled!")
		httputils.ListAndServerSSL(cfg.Host, cfg.SSL.Path, cfg.SSL.Cert, cfg.SSL.Key, cfg.Port, gzipHandler)
	}
	// ==== Simple GZIP HANDLER ====
	httputils.ListAndServerGZIP(cfg.Host, cfg.Port, gzipHandler)
}

func RegisterUser(ctx *fasthttp.RequestCtx, client *redis.Client) datastructure.Response {
	username, password := httputils.GetCredentials(ctx)
	if !httputils.ValidateCredentials(username, password) {
		return datastructure.Response{
			Status:      false,
			ErrorCode:   "ERROR_RETRIEVING_CREDENTIALS",
			Description: fmt.Sprintf("credentials not valid: [%s:%s] ", username, password),
			Data:        nil,
		}
	}
	result, err := client.Exists(context.TODO(), username).Result()
	if err != nil {
		return datastructure.Response{
			Status:      false,
			ErrorCode:   "ERROR_RETRIEVING_USER",
			Description: "unable to retrieve username: " + username,
			Data:        err,
		}
	}

	if result != 0 {
		return datastructure.Response{
			Status:      false,
			ErrorCode:   "USER_ALREADY_EXISTS",
			Description: fmt.Sprintf("user [%s] already exists", username),
			Data:        err,
		}
	}

	person := datastructure.Person{Username: username, Password: password}
	if err = client.Set(context.TODO(), username, helper.Marshal(person), -1).Err(); err != nil {
		return datastructure.Response{
			Status:      false,
			ErrorCode:   "UNABLE_REGISTER_USER",
			Description: "unable to register user: " + username,
			Data:        err.Error(),
		}
	}

	return LoginUser(ctx, client)
}

func LoginUser(ctx *fasthttp.RequestCtx, client *redis.Client) datastructure.Response {
	username, password := httputils.GetCredentials(ctx)
	if !httputils.ValidateCredentials(username, password) {
		return datastructure.Response{
			Status:      false,
			ErrorCode:   "ERROR_RETRIEVING_CREDENTIALS",
			Description: fmt.Sprintf("credentials not valid: [%s:%s] ", username, password),
			Data:        nil,
		}
	}
	result, err := client.Exists(context.TODO(), username).Result()
	if err != nil {
		return datastructure.Response{
			Status:      false,
			ErrorCode:   "ERROR_RETRIEVING_USER",
			Description: "unable to retrieve username: " + username,
			Data:        err,
		}
	}
	if result == 0 {
		return datastructure.Response{
			Status:      false,
			ErrorCode:   "USER_DOES_NOT_EXISTS",
			Description: fmt.Sprintf("user [%s] does not exists", username),
			Data:        err,
		}
	}

	res, err := client.Get(context.TODO(), username).Result()
	if err != nil {
		return datastructure.Response{
			Status:      false,
			ErrorCode:   "UNABLE_LOGIN_USER",
			Description: "unable to login user: " + username,
			Data:        err.Error(),
		}
	}
	var person datastructure.Person
	if err = json.Unmarshal([]byte(res), &person); err != nil {
		return datastructure.Response{
			Status:      false,
			ErrorCode:   "UNABLE_UNMARSHAL_USER",
			Description: "unable to unmarshal user: " + username,
			Data:        err.Error(),
		}
	}

	if password != person.Password {
		return datastructure.Response{
			Status:      false,
			ErrorCode:   "INVALID_PASSWORD",
			Description: fmt.Sprintf("password [%s] is not valid for user %s", password, username),
			Data:        nil,
		}
	}

	claims := datastructure.CustomClaims{
		Username: username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: int64((time.Hour * 24).Seconds()),
			Issuer:    "go-pyrecognizer-fe",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	//FIXME: go-pyrecognizer-fe have to be replaced with a strong password
	signedToken, err := token.SignedString([]byte("go-pyrecognizer-fe"))

	if err = client.Set(context.TODO(), fmt.Sprintf("%s_token", username), signedToken, time.Hour*24).Err(); err != nil {
		return datastructure.Response{
			Status:      false,
			ErrorCode:   "UNABLE_SET_TOKEN",
			Description: "unable to set token for user: " + username,
			Data:        err.Error(),
		}
	}
	ctx.Response.Header.Add("token", signedToken)
	return datastructure.Response{
		Status:      true,
		ErrorCode:   "",
		Description: "",
		Data:        fmt.Sprintf(`{"token":"%s","type":"Bearer", "expires":"%d"}`, signedToken, int64((time.Hour * 24).Seconds())),
	}
}

func RemoveUser(ctx *fasthttp.RequestCtx, client *redis.Client) datastructure.Response {
	username, password := httputils.GetCredentials(ctx)
	if !httputils.ValidateCredentials(username, password) {
		return datastructure.Response{
			Status:      false,
			ErrorCode:   "ERROR_RETRIEVING_CREDENTIALS",
			Description: fmt.Sprintf("credentials not valid: [%s:%s] ", username, password),
			Data:        nil,
		}
	}
	result, err := client.Exists(context.TODO(), username).Result()
	if err != nil {
		return datastructure.Response{
			Status:      false,
			ErrorCode:   "ERROR_RETRIEVING_USER",
			Description: "unable to retrieve username: " + username,
			Data:        err,
		}
	}
	if result == 0 {
		return datastructure.Response{
			Status:      false,
			ErrorCode:   "USER_DOES_NOT_EXISTS",
			Description: fmt.Sprintf("user [%s] does not exists", username),
			Data:        err,
		}
	}

	res, err := client.Get(context.TODO(), username).Result()
	if err != nil {
		return datastructure.Response{
			Status:      false,
			ErrorCode:   "UNABLE_LOGIN_USER",
			Description: "unable to login user: " + username,
			Data:        err.Error(),
		}
	}
	var person datastructure.Person
	if err = json.Unmarshal([]byte(res), &person); err != nil {
		return datastructure.Response{
			Status:      false,
			ErrorCode:   "UNABLE_UNMARSHAL_USER",
			Description: "unable to unmarshal user: " + username,
			Data:        err.Error(),
		}
	}

	if password != person.Password {
		return datastructure.Response{
			Status:      false,
			ErrorCode:   "INVALID_PASSWORD",
			Description: fmt.Sprintf("password [%s] is not valid for user %s", password, username),
			Data:        nil,
		}
	}

	if err = client.Del(context.TODO(), username).Err(); err != nil {
		return datastructure.Response{
			Status:      false,
			ErrorCode:   "UNABLE_REMOVE_USER",
			Description: "unable to delete the user from database",
			Data:        username + " not removed",
		}
	}

	if err = client.Del(context.TODO(), username+"_token").Err(); err != nil {
		return datastructure.Response{
			Status:      false,
			ErrorCode:   "UNABLE_REMOVE_USER_TOKEN",
			Description: "unable to delete token of the user from database",
			Data:        username + " not removed",
		}
	}

	return datastructure.Response{
		Status:      true,
		ErrorCode:   "",
		Description: "",
		Data:        username + " removed",
	}
}

func VerifyToken(ctx *fasthttp.RequestCtx, client *redis.Client) (bool, datastructure.Response) {
	username, _ := httputils.GetCredentials(ctx)
	if !httputils.ValidateUsername(username) {
		return false, datastructure.Response{
			Status:      false,
			ErrorCode:   "ERROR_RETRIEVING_CREDENTIALS",
			Description: fmt.Sprintf("credentials not valid: [%s:%s] ", username, "_"),
			Data:        nil,
		}
	}

	token, err := client.Get(context.TODO(), username+"_token").Result()
	if err != nil {
		return false, datastructure.Response{
			Status:      false,
			ErrorCode:   "UNABLE_RETRIEVE_TOKEN",
			Description: "unable to retrieve the token from the database",
			Data:        nil,
		}
	}

	peek := ctx.Request.Header.Peek("token")
	if token != string(peek) {
		return false, datastructure.Response{
			Status:      false,
			ErrorCode:   "TOKEN_MISMATCH",
			Description: "token not match",
			Data:        nil,
		}
	}
	return true, datastructure.Response{}
}

func EnableUser(ctx *fasthttp.RequestCtx, client *redis.Client) datastructure.Response {
	var resp datastructure.Response
	return resp
}

func DisableUser(ctx *fasthttp.RequestCtx, client *redis.Client) datastructure.Response {
	var resp datastructure.Response
	return resp
}
