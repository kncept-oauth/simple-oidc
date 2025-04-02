package dispatcher

import (
	"net/url"
	"testing"

	"github.com/kncept-oauth/simple-oidc/service/gen/api"
	"github.com/kncept-oauth/simple-oidc/service/params"
)

func TestDispatcherHandlerType(t *testing.T) {
	assertIsApiHandler(&oapiDispatcher{})
}
func assertIsApiHandler(receiver api.Handler) {}

func TestUrlParamParsing(t *testing.T) {
	u, err := url.Parse("?queryKey=queryVal1&k2=v2")
	if err != nil {
		t.Fatalf("url.Parse %v", err)
	}
	templateMap := params.QueryParamsToMap(u)
	if templateMap == nil {
		t.Fatalf("no templateMap returned")
	}
	if len(templateMap) != 2 {
		t.Fatalf("Expected length 2, got %v", len(templateMap))
	}
	tv, ok := templateMap["xxx"]
	if ok || tv != "" {
		t.Fatalf("should not have found result: xxx")
	}

	tv = templateMap["queryKey"]
	if tv != "queryVal1" {
		t.Fatalf("unexpected query value: %s", tv)
	}

}
