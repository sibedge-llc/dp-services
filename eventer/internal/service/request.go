package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

type RequestDataParser interface {
	Parse(out interface{}) error
}

type FormRequestDataParser struct {
	urlValues url.Values
}

func NewFormRequestDataParser(r *http.Request) (*FormRequestDataParser, error) {
	r.PostFormValue("") // Trigger init both form and post form data
	return &FormRequestDataParser{r.Form}, nil
}

func getOmittedSchemaFields(out interface{}) map[string]interface{} {
	if schema, ok := out.(map[string]interface{}); ok {
		return schema
	}
	v := reflect.ValueOf(out)
	if v.Kind() == reflect.Ptr && v.Elem().Kind() == reflect.Struct {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil
	}
	t := reflect.TypeOf(out)
	if t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct {
		t = t.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil
	}
	visibleFields := reflect.VisibleFields(t)
	schema := make(map[string]interface{}, len(visibleFields))
	for _, visibleField := range visibleFields {
		tag, ok := visibleField.Tag.Lookup("json")
		if !ok {
			continue
		}
		tagParts := strings.Split(tag, ",")
		if tagParts[len(tagParts)-1] != "omitempty" {
			continue
		}
		val := v.FieldByIndex(visibleField.Index).Interface()
		if tagParts[0] != "" {
			schema[tagParts[0]] = val
		} else {
			schema[visibleField.Name] = val
		}
	}
	return schema
}

func getRequestUrlValue(field string, requestVal string, val interface{}) (interface{}, error) {
	switch val.(type) {
	case string:
		return requestVal, nil
	case float64:
		f, err := strconv.ParseFloat(requestVal, 64)
		if err != nil {
			return nil, fmt.Errorf("field %s expects number but got %s: %w", field, requestVal, err)
		}
		return f, nil
	case int:
		f, err := strconv.ParseInt(requestVal, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("field %s expects number but got %s: %w", field, requestVal, err)
		}
		return int(f), nil
	case int64:
		f, err := strconv.ParseInt(requestVal, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("field %s expects number but got %s: %w", field, requestVal, err)
		}
		return f, nil
	case uint64:
		f, err := strconv.ParseUint(requestVal, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("field %s expects number but got %s: %w", field, requestVal, err)
		}
		return f, nil
	case bool:
		b, err := strconv.ParseBool(requestVal)
		if err != nil {
			return nil, fmt.Errorf("field %s expects boolean but got %s: %w", field, requestVal, err)
		}
		return b, nil
	default:
		return nil, fmt.Errorf("unsupported field type: %s %T", field, val)
	}
}

func (r *FormRequestDataParser) Parse(out interface{}) error {
	data, err := json.Marshal(out)
	if err != nil {
		return err
	}

	var schema map[string]interface{}
	err = json.Unmarshal(data, &schema)
	if err != nil {
		return err
	}

	vals := make(map[string]interface{}, len(schema))

	for field, val := range schema {
		if !r.urlValues.Has(field) {
			vals[field] = val
			continue
		}
		requestVal := r.urlValues.Get(field)
		v, err := getRequestUrlValue(field, requestVal, val)
		if err != nil {
			return err
		}
		vals[field] = v
	}

	omittedSchemaFields := getOmittedSchemaFields(out)
	for field, val := range omittedSchemaFields {
		if !r.urlValues.Has(field) {
			vals[field] = val
			continue
		}
		requestVal := r.urlValues.Get(field)
		v, err := getRequestUrlValue(field, requestVal, val)
		if err != nil {
			return err
		}
		vals[field] = v
	}

	data, err = json.Marshal(vals)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, out)
}

type JsonRequestDataParser struct {
	data []byte
}

func NewJsonRequestDataParser(r *http.Request) (*JsonRequestDataParser, error) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	// Make body be readable again for the futher processing
	r.Body = ioutil.NopCloser(bytes.NewReader(data))
	return &JsonRequestDataParser{data}, nil
}

func (r *JsonRequestDataParser) Parse(out interface{}) error {
	if json.Unmarshal(r.data, &json.RawMessage{}) != nil {
		return nil
	}
	return json.NewDecoder(bytes.NewReader(r.data)).Decode(&out)
}

func ParseRequest(r *http.Request, out interface{}) error {
	var err error
	var requestDataParser RequestDataParser
	switch r.Method {
	case http.MethodPost, http.MethodPut:
		mimeType, _, err := mime.ParseMediaType(r.Header.Get("content-type"))
		if err != nil {
			return err
		}
		switch mimeType {
		case "application/json":
			requestDataParser, err = NewJsonRequestDataParser(r)
		case "application/x-www-form-urlencoded", "multipart/form-data":
			requestDataParser, err = NewFormRequestDataParser(r)
		default:
			return fmt.Errorf("invalid content type: %s", mimeType)
		}
	case http.MethodGet, http.MethodDelete:
		requestDataParser, err = NewFormRequestDataParser(r)
	default:
		return fmt.Errorf("unexpected method type: %s", r.Method)
	}
	if err != nil {
		return err
	}

	err = requestDataParser.Parse(out)
	if err != nil {
		return err
	}
	return nil
}
