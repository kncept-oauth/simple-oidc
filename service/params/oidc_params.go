package params

import (
	"fmt"
	"net/url"
	"reflect"
	"strings"
)

// omitempty for optional params
// this needs to match the openapi authcode claims
type OidcAuthCodeFlowParams struct {
	ResponseType string `json:"response_type"`
	ClientId     string `json:"client_id"`
	Scope        string `json:"scope"`
	RedirectUri  string `json:"redirect_uri"`

	State string `json:"state,omitempty"`
	Nonce string `json:"nonce,omitempty"`
}

func jsonTag(f reflect.StructField) []string {
	tag := f.Tag
	jsonTag := tag.Get("json")
	if jsonTag == "" {
		return []string{""}
	}
	return strings.Split(jsonTag, ",")
}
func tagIgnore(tag []string) bool {
	return len(tag) == 0 || tag[0] == "" || tag[0] == "_"
}
func tagOmitEmpty(tag []string) bool {
	for _, t := range tag {
		if t == "omitempty" {
			return true
		}
	}
	return false
}

func (obj *OidcAuthCodeFlowParams) ToQueryParams() string {
	return toQueryParams(obj)
}

func (obj *OidcAuthCodeFlowParams) IsValid() bool {
	if obj.ResponseType == "" {
		return false
	}
	if obj.ClientId == "" {
		return false
	}
	if obj.Scope == "" {
		return false
	}
	if obj.RedirectUri == "" {
		return false
	}

	return true
}

func (obj *OidcAuthCodeFlowParams) Merge(other *OidcAuthCodeFlowParams) {
	obj.ResponseType = fallbackString(obj.ResponseType, other.ResponseType)
	obj.ClientId = fallbackString(obj.ClientId, other.ClientId)
	obj.Scope = fallbackString(obj.Scope, other.Scope)
	obj.RedirectUri = fallbackString(obj.RedirectUri, other.RedirectUri)
	obj.State = fallbackString(obj.State, other.State)
	obj.Nonce = fallbackString(obj.Nonce, other.Nonce)
}

func fallbackString(v1, v2 string) string {
	if v2 == "" {
		return v1
	}
	return v2
}

// any pointer
func toQueryParams(obj any) string {
	s := ""
	t := reflect.TypeOf(obj).Elem()
	v := reflect.ValueOf(obj).Elem()

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tag := jsonTag(f)
		if tagIgnore(tag) {
			continue
		}
		strVal := v.FieldByName(f.Name).String()
		if strVal == "" && tagOmitEmpty(tag) {
			continue
		}
		param := fmt.Sprintf("%v=%v", tag[0], url.QueryEscape(strVal))
		if s == "" {
			s = param
		} else {
			s = fmt.Sprintf("%v&%v", s, param)
		}

	}
	return s

}

func QueryParamsToMap(requestUrl *url.URL) map[string]string {
	res := make(map[string]string)
	queryValues := requestUrl.Query()
	for key, val := range queryValues {
		if len(val) >= 1 {
			res[key] = val[0]
		}
	}
	return res
}

// simple
func OidcParamsFromMap(structMap map[string]string) *OidcAuthCodeFlowParams {
	dst := &OidcAuthCodeFlowParams{}
	StructFromMap(structMap, dst)
	return dst
}
func OidcParamsFromQuery(queryString string) (*OidcAuthCodeFlowParams, error) {
	queryString = strings.TrimPrefix(queryString, "?")
	u, err := url.Parse(fmt.Sprintf("?%v", queryString))
	if err != nil {
		return nil, err
	}
	return OidcParamsFromMap(QueryParamsToMap(u)), nil
}

func StructFromMap(structMap map[string]string, dst any) {
	// TODO: assert pointer-ness (and handle)
	t := reflect.TypeOf(dst)
	v := reflect.ValueOf(dst)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
		v = v.Elem()
	}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tag := jsonTag(f)
		if tagIgnore(tag) {
			continue
		}

		structVal := structMap[tag[0]]
		if structVal != "" {
			v.FieldByName(f.Name).SetString(structVal)
		}

	}
}
