package routing

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// Constants

const (
	MethodPost   = "POST"
	MethodGet    = "GET"
	MethodPut    = "PUT"
	MethodDelete = "DELETE"
)

// Types

type RouteSpec struct {
	Name string
	Path string
}

type Controller interface {
	SetBasePath(basePath string)
	HandlerForRoute(route string, method string) func(w http.ResponseWriter, r *http.Request)
	ExposedRoutes() []RouteSpec
}

type Router map[string]map[string]Controller

// Public Methods

// Register a controller to handle requests to the specified api and route
func (self Router) RegisterRoute(api string, route string, controller Controller) {
	if self[api] == nil {
		self[api] = make(map[string]Controller)
	}
	controller.SetBasePath(api + route)
	self[api][route] = controller
}

// Public Functions

func BaseURLFromRequest(r *http.Request) string {
	u := r.URL

	if u.Scheme == "" {
		u.Scheme = "http"
	}

	if u.Host == "" {
		u.Host = r.Host
	}

	if u.Host == "" {
		u.Host = "localhost:8080"
	}

	u.Path = ""

	return u.String()
}

// Private Methods

func (self Router) handleAPIRoot(writer http.ResponseWriter, request *http.Request, routes map[string]Controller, apiPath string) {
	exposedRouteList := make(map[string]string)

	for route, controller := range routes {
		for _, routeSpec := range controller.ExposedRoutes() {
			fullURL := BaseURLFromRequest(request) + apiPath + route + routeSpec.Path
			exposedRouteList[routeSpec.Name] = fullURL
		}
	}

	bytes, _ := json.Marshal(exposedRouteList)
	writer.Write(bytes)
}

func (self Router) HandleRequest(writer http.ResponseWriter, request *http.Request) {
	path := request.URL.Path

	fmt.Printf("Handling request from [%s] %s...", request.Method, path)
	writer.Header().Set("Content-Type", "application/JSON")

	for api, routes := range self {
		apiPath := "/" + api
		if path == apiPath || path == apiPath+"/" {
			self.handleAPIRoot(writer, request, routes, apiPath)
			fmt.Printf("handled\n")
			return
		} else if strings.HasPrefix(path, apiPath) {
			for route, controller := range routes {
				routePath := apiPath + route
				if strings.HasPrefix(path, routePath) {
					remainingRoute := path[len(routePath):]
					handler := controller.HandlerForRoute(remainingRoute, request.Method)
					if handler != nil {
						handler(writer, request)
						fmt.Printf("handled\n")
						return
					}
				}
			}
		}
	}

	fmt.Printf("unhandled\n")
	http.Error(writer, http.StatusText(http.StatusNotFound), http.StatusNotFound)
}
