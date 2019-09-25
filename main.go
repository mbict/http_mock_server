package main

import (
	"flag"
	"github.com/radovskyb/watcher"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"os"
	"path"
	"regexp"
	"time"
)

type Config struct {
	Routes []Route
}

type Route struct {
	Path     string            `yaml:"path"`
	Method   string            `yaml:"method"`
	Payload  string            `yaml:"payload"`
	Query    map[string]string `yaml:"query"`
	Headers  map[string]string `yaml:"headers"`
	Response Response          `yaml:"response"`
}

type Response struct {
	Code    int               `yaml:"code"`
	Body    string            `yaml:"body"`
	Headers map[string]string `yaml:"headers"`
}

type RouteMatcher struct {
	Method   string
	Path     *regexp.Regexp
	Payload  *regexp.Regexp
	Query    map[string]*regexp.Regexp
	Headers  map[string]*regexp.Regexp
	Response ResponseRequest
}

type ResponseRequest struct {
	Code      int
	Body      []byte
	MediaType string
	Headers   map[string]string
}

func (m RouteMatcher) Match(request *http.Request) bool {
	if m.Method != request.Method {
		return false
	}

	if !m.Path.MatchString(request.URL.Path) {
		return false
	}

	for param, queryMatcher := range m.Query {
		value := request.URL.Query().Get(param)
		if !queryMatcher.MatchString(value) {
			return false
		}
	}

	for headerKey, headerMatcher := range m.Headers {
		value := request.Header.Get(headerKey)
		if !headerMatcher.MatchString(value) {
			return false
		}
	}

	body, _ := ioutil.ReadAll(request.Body)
	if !m.Payload.Match(body) {
		return false
	}

	return true
}

func main() {

	var addr = flag.String("http.addr", ":8080", "port to listen")
	var configRoutes = flag.String("config.routes", "routes.yaml", "routes definition file")
	var configDataDir = flag.String("config.data", "./data", "data dir when the config file is")

	flag.Parse()

	config := readConfig(path.Join(*configDataDir, *configRoutes))

	matchers := generateMatchers(config)

	//watch for changes in filesystem to reload mocking server
	w := watcher.New()
	w.SetMaxEvents(1)

	go func() {
		for {
			select {
			case <-w.Event:
				log.Println("file change detected, reloading config")
				matchers = generateMatchers(readConfig(path.Join(*configDataDir, *configRoutes)))
			case err := <-w.Error:
				log.Fatalln(err)
			case <-w.Closed:
				return
			}
		}
	}()

	if err := w.AddRecursive(*configDataDir); err != nil {
		log.Fatalln(err)
	}

	for path, f := range w.WatchedFiles() {
		log.Printf("watched files %s: %s\n", path, f.Name())
	}

	// Start the watching process
	go func() {
		if err := w.Start(time.Millisecond * 1000); err != nil {
			log.Fatalln(err)
		}
	}()

	http.HandleFunc("/", func(response http.ResponseWriter, request *http.Request) {
		for _, route := range matchers {
			if route.Match(request) {
				for headerKey, value := range route.Response.Headers {
					response.Header().Add(headerKey, value)
				}
				response.WriteHeader(route.Response.Code)
				response.Write(route.Response.Body)
				log.Printf("handled [%s] %s", request.Method, request.RequestURI)
				return
			}
		}
		log.Printf("ignored [%s] %s", request.Method, request.RequestURI)

		http.NotFound(response, request)
	})

	log.Println("server started on", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

func generateMatchers(config Config) []RouteMatcher {
	matchers := []RouteMatcher{}
	for _, route := range config.Routes {
		queryMatchers := map[string]*regexp.Regexp{}
		for param, value := range route.Query {
			if len(value) == 0 {
				value = ".*"
			}
			queryMatchers[param] = regexp.MustCompile("^" + value + "$")
		}

		headerMatchers := map[string]*regexp.Regexp{}
		for key, value := range route.Headers {
			if len(value) == 0 {
				value = ".*"
			}
			headerMatchers[key] = regexp.MustCompile("^" + value + "$")
		}

		if len(route.Path) == 0 {
			route.Path = ".*"
		}

		if len(route.Payload) == 0 {
			route.Payload = ".*"
		}

		//init empty map
		if route.Response.Headers == nil {
			route.Response.Headers = make(map[string]string)
		}

		//load in file as response if file is available
		var body = []byte(route.Response.Body)
		if _, err := os.Stat("./data/" + route.Response.Body); err == nil {
			mimeType := mime.TypeByExtension(path.Ext(route.Response.Body))
			if mimeType != "" && route.Response.Headers["Content-Type"] == "" {
				route.Response.Headers["Content-Type"] = mimeType
			}
			body, _ = ioutil.ReadFile("./data/" + route.Response.Body)
		}

		matchers = append(matchers, RouteMatcher{
			Method:  route.Method,
			Path:    regexp.MustCompile("^" + route.Path + "$"),
			Payload: regexp.MustCompile(route.Payload),
			Query:   queryMatchers,
			Headers: headerMatchers,
			Response: ResponseRequest{
				Code:    route.Response.Code,
				Body:    body,
				Headers: route.Response.Headers,
			},
		})
	}
	return matchers
}

func readConfig(filename string) Config {
	configFile, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	var config Config
	if err = yaml.Unmarshal(configFile, &config); err != nil {
		panic(err)
	}
	return config
}
