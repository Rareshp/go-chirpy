package main 

import (
  "fmt"
  "time"
  "net/http"
  "errors"
  "strings"
  "github.com/golang-jwt/jwt/v5"
)

func (cfg *apiConfig) jwtCreateToken(issuer string, expireInSeconds int, id int) (string, error) {
  // Create a new token object, specifying signing method and the claims

  // Calculate the expiration time
  expireDuration := time.Duration(expireInSeconds) * time.Second

  token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
    Issuer: issuer,
    IssuedAt: jwt.NewNumericDate(time.Now().UTC()),
    ExpiresAt: jwt.NewNumericDate(time.Now().Add(expireDuration).UTC()),
    Subject: fmt.Sprint(id),
  })

  // Sign and get the complete encoded token as a string using the secret
  tokenString, err := token.SignedString([]byte(cfg.jwtSecret))
  if err != nil {
    return "", err
  }

  return tokenString, nil
}

func (cfg *apiConfig) jwtCreateAccessToken(id int) (string, error) {
  // access tokens have 1 hour 
  token, err := cfg.jwtCreateToken("chirpy-access", 3600, id)
  if err != nil {
    return "", err
  }
  return token, nil
}
func (cfg *apiConfig) jwtCreateRefreshToken(id int) (string, error) {
  // refresh tokens have 60 days = 
  token, err := cfg.jwtCreateToken("chirpy-refresh", 60 * 24 * 3600, id)
  if err != nil {
    return "", err
  }
  return token, nil
}

func (cfg *apiConfig) jwtParseToken(tokenString string) (string, error) {
  type MyCustomClaims struct {
    Foo string `json:"foo"`
    jwt.RegisteredClaims
  }

  token, err := jwt.ParseWithClaims(tokenString, &MyCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
    return []byte("AllYourBase"), nil
  })
  if err != nil {
    return "", err
  } else if claims, ok := token.Claims.(*MyCustomClaims); ok {
    fmt.Println(claims.Foo, claims.RegisteredClaims.Issuer)
  } 

  return "", nil
}

func ValidateJWT(tokenString, tokenSecret string) (string, error) {
	claimsStruct := jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(
		tokenString,
		&claimsStruct,
		func(token *jwt.Token) (interface{}, error) { return []byte(tokenSecret), nil },
	)
	if err != nil {
		return "", err
	}

	userIDString, err := token.Claims.GetSubject()
	if err != nil {
		return "", err
	}

	return userIDString, nil
}

func GetIssuer(tokenString, tokenSecret string) (string, error) { 
	claimsStruct := jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(
		tokenString,
		&claimsStruct,
		func(token *jwt.Token) (interface{}, error) { return []byte(tokenSecret), nil },
	)
	if err != nil {
		return "", err
	}

	issuerString, err := token.Claims.GetIssuer()
	if err != nil {
		return "", err
	}

	return issuerString, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("auth header is empty")
	}
	splitAuth := strings.Split(authHeader, " ")
	if len(splitAuth) < 2 || splitAuth[0] != "Bearer" {
		return "", errors.New("malformed authorization header")
	}

	return splitAuth[1], nil
}
