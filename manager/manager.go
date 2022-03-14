package manager

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ermos/annotation/parser"
	"github.com/gofiber/fiber/v2"
	"net/http"
	"strconv"
	"strings"
)

type Handler struct{}

type Manager struct {
	HTTP struct {
		Method     string
		RequestURI string
	}
	User       map[string]interface{}
	Param      map[string]interface{}
	Query      map[string]string
	Payload    map[string]interface{}
	data       map[string]interface{}
	annotation parser.API
}

// New allows creating new manager instance
func New(a parser.API, c *fiber.Ctx) (m *Manager, status int, err error) {
	m = &Manager{
		annotation: a,
	}

	m.data = make(map[string]interface{})

	status, err = m.setRequest(c)

	return
}

func (m *Manager) Set(key string, data interface{}) {
	m.data[key] = data
}

func (m *Manager) Get(key string) interface{} {
	return m.data[key]
}

func (m *Manager) setRequest(c *fiber.Ctx) (status int, err error) {

	m.setQueryParams(c)

	err = m.setParams(c, m.annotation)
	if err != nil {
		return http.StatusBadRequest, err
	}

	if c.Method() == "POST" || c.Method() == "PUT" {
		ct := strings.Split(string(c.Request().Header.ContentType()), ";")

		switch strings.ToLower(ct[0]) {
		case "application/json":
			err = m.getPayloadJSON(c, m.annotation)
			if err != nil {
				return http.StatusBadRequest, err
			}
		default:
			return http.StatusBadRequest, errors.New(ct[0] + " is not supported by this API")
		}
	}

	return
}

// setParams allows getting parameters from url and convert it into the good type
func (m *Manager) setParams(c *fiber.Ctx, a parser.API) error {
	var err error
	result := make(map[string]interface{})

	for _, param := range a.Validate.Params {
		result[param.Key], err = convert(param.Type, c.Params(param.Key))
		if err != nil {
			return fmt.Errorf("%s's type is incorrect for this field", param.Key)
		}
	}

	m.Param = result

	return nil
}

// convert allows converting interface to the expected type
func convert(trueType string, value interface{}) (interface{}, error) {
	var valueString string

	switch value.(type) {
	case int:
		valueString = fmt.Sprintf("%d", value.(int))
	case bool:
		valueString = fmt.Sprintf("%t", value.(bool))
	case float64:
		if trueType != "int" {
			valueString = fmt.Sprintf("%2.f", value.(float64))
		} else {
			valueString = fmt.Sprintf("%0.f", value.(float64))
		}
	case string:
		valueString = value.(string)
	default:
		if trueType == "map" {
			marshal, err := json.Marshal(value)
			if err != nil {
				return nil, errors.New("can't parse map type")
			}

			valueString = string(marshal)
		} else {
			return nil, errors.New("type not found")
		}
	}

	switch strings.ToLower(trueType) {
	case "int":
		rInt, err := strconv.Atoi(valueString)
		if err != nil {
			return rInt, errors.New(`Impossible de convertir ` + valueString + ` en int`)
		}
		return rInt, nil
	case "float64":
		rFloat64, err := strconv.ParseFloat(valueString, 64)
		if err != nil {
			return rFloat64, errors.New(`Impossible de convertir ` + valueString + ` en float64`)
		}
		return rFloat64, nil
	case "bool":
		rBool, err := strconv.ParseBool(valueString)
		if err != nil {
			return rBool, errors.New(`Impossible de convertir ` + valueString + ` en bool`)
		}
		return rBool, nil
	case "string", "map":
		return valueString, nil
	default:
		return value, fmt.Errorf("%s's type is not supported", trueType)
	}
}

// getPayloadJSON allows parsing payload written with JSON
func (m *Manager) getPayloadJSON(c *fiber.Ctx, a parser.API) error {
	var value interface{}
	var err error
	var data map[string]interface{}
	result := make(map[string]interface{})

	if len(a.Validate.Payload) <= 0 {
		return nil
	}

	err = json.Unmarshal(c.Body(), &data)
	if err != nil {
		return err
	}

	for _, body := range a.Validate.Payload {
		if !body.Nullable && (data[body.Key] == "" || data[body.Key] == nil) {
			return fmt.Errorf("%s's key is required in payload", body.Key)
		}

		if data[body.Key] == "" || data[body.Key] == nil {
			result[body.Key] = nil
			continue
		}

		value, err = convert(body.Type, data[body.Key])
		if err != nil {
			return err
		}

		result[body.Key] = value
	}

	m.Payload = result

	return nil
}

// setQueryParams allows to get query parameters from URL
func (m *Manager) setQueryParams(c *fiber.Ctx) {
	list := make(map[string]string)

	split := strings.Split(c.OriginalURL(), "?")

	if len(split) < 2 {
		m.Query = list
		return
	}

	query := strings.Split(split[1], "&")
	for _, q := range query {
		split = strings.Split(q, "=")
		if len(split) == 1 {
			list[split[0]] = split[0]
		} else {
			list[split[0]] = split[1]
		}
	}

	m.Query = list
	return
}
