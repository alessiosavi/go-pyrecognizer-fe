package datastructure

import "github.com/dgrijalva/jwt-go"

// Configuration configuration is the structure for handle the configuration data
type Configuration struct {
	Host    string `json:"host,omitempty"` // Hostname to bind the service
	Port    int    `json:"port,omitempty"` // Port to bind the service
	Version string `json:"version,omitempty"`
	SSL     struct {
		Path    string `json:"path,omitempty"`
		Cert    string `json:"cert,omitempty"`
		Key     string `json:"key,omitempty"`
		Enabled bool   `json:"enabled,omitempty"`
	} `json:"ssl"`
	Redis struct {
		Host  string `json:"host,omitempty"`
		Port  string `json:"port,omitempty"`
		Token struct {
			Expire int `json:"expire,omitempty"`
			DB     int `json:"db,omitempty"`
		} `json:"token"`
	} `json:"redis"`
}

// Response status Structure used for populate the json response for the RESTfull HTTP API
type Response struct {
	Status      bool        `json:"Status"`      // Status of response [true,false] OK, KO
	ErrorCode   string      `json:"ErrorCode"`   // Code linked to the error (KO)
	Description string      `json:"Description"` // Description linked to the error (KO)
	Data        interface{} `json:"Data"`        // Generic data to return in the response
}

// MiddlewareRequest middlewareRequest Structure used for manage the request among the user and the external service
type MiddlewareRequest struct {
	Username string   `json:"username,omitempty"`
	Token    string   `json:"token,omitempty"`
	Service  string   `json:"service,omitempty"`
	Method   string   `json:"method,omitempty"`
	Headers  []string `json:"headers,omitempty"`
	Data     string   `json:"data,omitempty"`
}

// Person structure of a customer for save it into the DB during registration phase.
type Person struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Mail     string `json:"mail,omitempty"`
	Enabled  bool   `json:"enabled,omitempty"`
}

// CustomClaims store the jwt claims
type CustomClaims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}
