package mw

import (
	"context"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/vitego/config"
	"github.com/vitego/router/manager"
	"net/http"
	"strings"
)

type AuthenticateMiddleware struct{}

func (AuthenticateMiddleware) Run(ctx context.Context, m *manager.Manager, w http.ResponseWriter, r *http.Request) (status int, err error) {
	t, err := parseAuthorization(r)
	if err != nil {
		// @TODO on ajoute juste pas l'utilisateur
		return 0, nil
	}

	data, err := jwt.Parse(t, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(config.Get("auth.jwt.secretKey")), nil
	})
	if err != nil {
		// @TODO on ajoute juste pas l'utilisateur
		return 0, nil
	}

	claims, ok := data.Claims.(jwt.MapClaims)
	if !ok || !data.Valid {
		// @TODO on ajoute juste pas l'utilisateur
		return 0, nil
	}

	var stringRoles []string
	for _, userRole := range claims["roles"].([]interface{}) {
		stringRoles = append(stringRoles, userRole.(string))
	}

	m.User = make(map[string]interface{})
	m.User["token"] = t
	m.User["issuer"] = claims["issuer"].(string)
	m.User["id"] = int(claims["id"].(float64))
	m.User["roles"] = stringRoles
	m.User["exp"] = int(claims["exp"].(float64))

	return 0, nil
}

func parseAuthorization(r *http.Request) (jwt string, err error) {
	split := strings.Split(r.Header.Get("Authorization"), "Bearer ")
	if len(split) != 2 {
		return "", errors.New("jwt token malformed")
	}
	return split[1], nil
}
