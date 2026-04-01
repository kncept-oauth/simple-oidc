package httpdispatcher

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/kncept-oauth/simple-oidc/service/webcontent"
)

type TemplateDispatcher struct {
	mu                        sync.Mutex
	devModeLiveFilesystemBase *string
	tmpl                      *template.Template
}

func NewTemplateDispatcher(devModeLiveFilesystemBase *string) *TemplateDispatcher {
	return &TemplateDispatcher{
		devModeLiveFilesystemBase: devModeLiveFilesystemBase,
	}
}

func (obj *TemplateDispatcher) RespondWithTemplate(filename string, statusCode int, res http.ResponseWriter, data any) {
	t := obj.templates()
	res.Header().Add("Content-Type", "text/html")
	// TODO: configurable option - execute template before writing to stream
	res.WriteHeader(statusCode)
	err := t.ExecuteTemplate(res, filename, data)
	// err := t.Execute(res, data)
	if err != nil {
		fmt.Printf("%v", err)
	}
}

func (obj *TemplateDispatcher) templates() *template.Template {
	obj.mu.Lock()
	defer obj.mu.Unlock()
	if obj.tmpl != nil {
		return obj.tmpl
	}

	templ := template.New("_")

	f := map[string]any{
		"Wrap": func(keyPairs ...any) any {
			keyPairsLen := len(keyPairs)
			if keyPairsLen%2 != 0 {
				panic("Must supply a full set of key pairs")
			}
			m := map[any]any{}
			i := 0
			for i < keyPairsLen {
				m[keyPairs[i]] = keyPairs[i+1]
				i = i + 2
			}
			return m
		},
		"Coalesce": func(str ...string) string {
			for _, s := range str {
				if s != "" {
					return s
				}
			}
			return ""
		},

		"time": func(args ...any) string {
			if len(args) == 0 {
				return ""
			}
			if t, ok := args[0].(time.Time); ok {
				if len(args) == 1 {
					return t.String()
				}
				if args[1] == "since" {
					duration := time.Since(t)

					seconds := int(duration.Seconds())
					if seconds < 60 {
						return fmt.Sprintf("%d seconds ago", seconds)
					}

					minutes := int(duration.Minutes())
					if minutes < 60 {
						return fmt.Sprintf("%d minutes ago", minutes)
					}

					hours := int(duration.Hours())
					if hours < 24 {
						return fmt.Sprintf("%d hours ago", hours)
					}

					days := int(hours / 24)
					if days < 30 {
						return fmt.Sprintf("%d days ago", days)
					}

					return fmt.Sprintf("%d months ago", int(days/30))

				}
				panic("unknown use of time function")
			} else {
				return ""
			}

		},
	}
	templ = templ.Funcs(f)
	if obj.devModeLiveFilesystemBase != nil {
		foundFiles := make([]string, 0, 0)
		snippets, err := os.ReadDir(fmt.Sprintf("%s/webcontent", *obj.devModeLiveFilesystemBase))
		for _, snippet := range snippets {
			if strings.HasSuffix(snippet.Name(), ".html") {
				foundFiles = append(foundFiles, fmt.Sprintf("%s/webcontent/%s", *obj.devModeLiveFilesystemBase, snippet.Name()))
			}
		}
		snippets, err = os.ReadDir(fmt.Sprintf("%s/webcontent/snippet", *obj.devModeLiveFilesystemBase))
		for _, snippet := range snippets {
			if strings.HasSuffix(snippet.Name(), ".snippet") {
				foundFiles = append(foundFiles, fmt.Sprintf("%s/webcontent/snippet/%s", *obj.devModeLiveFilesystemBase, snippet.Name()))
			}
		}

		templ, err := templ.ParseFiles(foundFiles...)
		// templ, err := templ.ParseGlob("*.html", "snippet/*.snippet")
		if err != nil {
			panic(err)
		}
		// do not cache (!!), force a re-read every time
		return templ
	} else {
		templ, err := templ.ParseFS(webcontent.Fs, "*.html", "snippet/*.snippet")
		if err != nil {
			panic(err)
		}
		obj.tmpl = templ
	}

	return templ
}
