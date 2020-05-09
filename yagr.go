package yagr

import (
	"context"
	"errors"
	"net/http"
	"regexp"
	"strings"
)

// handlerFunc is a convinience type around http.HandlerFunc
type handlerFunc func(http.ResponseWriter, *http.Request)

// Router is the main interface defining the core
// functionality along with useful metadata retreival functions
type Router interface {
	getRoot() map[string]*node
	Insert(fullpath string, method string, hf handlerFunc) (*router, error)
	Search(fullpath string, method string) (*node, error)
}

type router struct {
	root map[string]*node
}

// node represents a single path part of an entire route, it contains
// inofrmation regarding each step in the route chain and holds the registered
// route part that will be checked against the requests
type node struct {
	path    string
	params  []string
	nodes   map[string]*node
	methods map[string]handlerFunc
}

// get Params is a helper method to parse query parameter against
// the path structure registered
func (n *node) getParams(fullpath string) map[string]string {
	ps := splitFullpath(fullpath)
	paramsToValue := make(map[string]string)
	for i, v := range ps {
		if i%2 != 0 {
			k := n.params[i>>1]
			paramsToValue[k] = v
		}
	}
	return paramsToValue
}

// Option is a functional options type for the router
type Option func(ro *router)

// NewNode is a factory for the node struct
func NewNode(path string) *node {
	n := node{
		path:    path,
		params:  make([]string, 0),
		nodes:   make(map[string]*node),
		methods: make(map[string]handlerFunc, 10),
	}
	return &n
}

// NewRouter is a factory for the router struct
func NewRouter(opts ...Option) *router {
	ro := router{
		root: make(map[string]*node),
	}
	ro.root["/"] = NewNode("/")
	for _, opt := range opts {
		opt(&ro)
	}

	return &ro
}

// getRoot is a convinience method to retrieve the router
// root node
func (ro *router) getRoot() *node {
	return ro.root["/"]
}

const (
	reQueryParam = `\{([a-zA-Z]+)\:(int|string|float|bool)\}`
)

// Insert registers a new node in the router tree starting from the root.
// If same route is registered twice, the first one will be overwritten
func (ro *router) Insert(fullpath string, method string, hf handlerFunc) (*router, error) {
	re := regexp.MustCompile(reQueryParam)
	n := ro.getRoot()
	params := []string{}

	ps := splitFullpath(fullpath)

	// if root
	if ps[0] == "/" {
		n.methods[method] = hf
	}

	for _, p := range ps {

		// test if query param, than parse it accoringly
		// format is {name:<T>} where T is int|string|float|bool
		if re.MatchString(p) {
			kv := re.FindStringSubmatch(p)
			params = append(params, kv[1])
			continue
		}

		// test if path segment, in case it is a new one
		// create a new node on the route
		if _, exist := n.nodes[p]; !exist {
			n.nodes[p] = NewNode(p)
		}

		n = n.nodes[p]
	}
	n.methods[method] = hf
	n.params = params
	return ro, nil
}

// Search Traverse the route tree according to received path and checks
// against registered routes. In case a corresponding route was found and matches
// the path, query params and methods, the function will return the registered
// http handler function
func (ro *router) Search(fullpath string) (*node, error) {
	n := ro.getRoot()
	ps := splitFullpath(fullpath)

	if ps[0] == "/" {
		return n, nil
	}

	exist := false
	for i := 0; i < len(ps); i++ {
		n, exist = n.nodes[ps[i]]
		if !exist {
			return nil, errors.New("path is not registered")
		}

		// if parameter exist for the path, validate
		if len(n.params) != 0 {
			i++
		}
	}
	return n, nil
}

func splitFullpath(fullpath string) []string {
	ps := strings.Split(fullpath, "/")[1:]
	if len(ps) == 1 {
		return []string{"/"}
	}
	return ps
}

// yagr is a wrapper aroud the router which creates
// convinience layer to manage routes
type yagr struct {
	router *router
}

// YAGR
type YAGR interface {
	Get(path string, hf func(http.ResponseWriter, *http.Request))
	Post(path string, hf func(http.ResponseWriter, *http.Request))
	Put(path string, hf func(http.ResponseWriter, *http.Request))
	Delete(path string, hf func(http.ResponseWriter, *http.Request))
	Patch(path string, hf func(http.ResponseWriter, *http.Request))
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}
type YAGROption func(*yagr)

// NewYAGR returns a YAGR struct reference
func NewYAGR(opts ...YAGROption) *yagr {
	yagr := yagr{
		router: NewRouter(),
	}

	for _, opt := range opts {
		opt(&yagr)
	}

	return &yagr
}

// Get registers a new GET handler on the given path
func (yagr *yagr) Get(path string, hf handlerFunc) {
	_, err := yagr.router.Insert(path, "GET", hf)
	if err != nil {
		panic(err)
	}
}

// Post registers a new POST handler on the given path
func (yagr *yagr) Post(path string, hf handlerFunc) {
	_, err := yagr.router.Insert(path, "POST", hf)
	if err != nil {
		panic(err)
	}
}

// Put registers a new PUT handler on the given path
func (yagr *yagr) Put(path string, hf handlerFunc) {
	_, err := yagr.router.Insert(path, "PUT", hf)
	if err != nil {
		panic(err)
	}
}

// Delete registers a new DELETE handler on the given path
func (yagr *yagr) Delete(path string, hf handlerFunc) {
	_, err := yagr.router.Insert(path, "DELETE", hf)
	if err != nil {
		panic(err)
	}
}

// Patch registers a new PATCH handler on the given path
func (yagr *yagr) Patch(path string, hf handlerFunc) {
	_, err := yagr.router.Insert(path, "PATCH", hf)
	if err != nil {
		panic(err)
	}
}

type key string

func (yagr *yagr) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	n, err := yagr.router.Search(r.URL.Path)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	h, exist := n.methods[r.Method]
	if !exist {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	ctx := context.WithValue(r.Context(), key("params"), n.getParams(r.URL.Path))
	r = r.WithContext(ctx)
	h(w, r)
}
