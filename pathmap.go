package main

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-openapi/spec"
)

//PathMap contains the list of all paths defined in the Swagger files for each
//service
type PathMap struct {
	Services map[string]*PathItem `json:"services"`
}

//PathItem represents a single element from a path defined in a Swagger file
type PathItem struct {
	Key        string               `json:"key"`
	PathItems  map[string]*PathItem `json:"pathItems"`
	Verbs      map[string]*Verb     `json:"verbs"`
	Documented bool                 `json:"documented"`
}

//ParameterisedItemKey is the string put into the PathItem maps when the path item
//is parameterised so that map will always match even if the parameters are defined
//with different names in the swagger files
const ParameterisedItemKey = "{*}"

//Verb represents a verb applied to a path and contains the list of defined accept
//header, content-type header, and response code combinations that are documented
//as supported in the Swagger file
type Verb struct {
	Name            string                     `json:"name"`
	Produces        []string                   `json:"produces"`
	Consumes        []string                   `json:"consumes"`
	Responses       map[string]*Response       `json:"responses"`
	QueryParameters map[string]*QueryParameter `json:"queryParams"`
	Documented      bool                       `json:"documented"`
}

//Response represents a possible response code for the documented http request
type Response struct {
	Response   string `json:"code"`
	Covered    int    `json:"covered"`
	Documented bool   `json:"documented"`
}

//QueryParameter represents a possible query parameter for the documented http request
type QueryParameter struct {
	Key        string `json:"key"`
	Covered    int    `json:"covered"`
	Documented bool   `json:"documented"`
}

//NewPathMap constructs a PathMap and returns it
func NewPathMap() *PathMap {
	return &PathMap{
		Services: map[string]*PathItem{},
	}
}

//IsParameter returns true if this PathItem is a route parameter
func (p *PathItem) IsParameter() bool {
	r := regexp.MustCompile(`^\{[^\}]*\}$`)
	return r.MatchString(p.Key)
}

//MapKey returns a string that can be used to add the PathItem into a Map
//Note if the item is a path parameter this method returns "{*}" so that
//path parameters that are specified differently, but exist in the same
//route path position will still match
func (p *PathItem) MapKey() string {
	if p.IsParameter() {
		return ParameterisedItemKey
	}
	return p.Key
}

//ParameterisedChild returns true if the PathItems map of this PathItem
//contains a parameterised path item. This means that the child element
//in the path is parameterised e.g. /thisPathItem/{childItem}
func (p *PathItem) ParameterisedChild() bool {
	_, exists := p.PathItems[ParameterisedItemKey]
	return exists
}

//NewPathItem returns a new instance of PathItem for the given path element
func NewPathItem(pathelement string, documented bool) *PathItem {
	return &PathItem{
		Key:        pathelement,
		PathItems:  map[string]*PathItem{},
		Documented: documented,
	}
}

//NewVerb creates a new instance of Verb with the specified name
func NewVerb(verb string, documented bool, produces, consumes []string) *Verb {
	return &Verb{
		Name:            verb,
		Consumes:        consumes,
		Produces:        produces,
		Responses:       map[string]*Response{},
		QueryParameters: map[string]*QueryParameter{},
		Documented:      documented,
	}
}

//ReadSwagger reads all of the Swagger files specified in the passed configuration
func (pm *PathMap) ReadSwagger(c Config) error {
	for _, srv := range c.Services {
		_, exists := pm.Services[srv.RoutePath]
		if exists {
			return fmt.Errorf("Swagger for service '%s' already read. Check you haven't included the same file more than once", srv.RoutePath)
		}
		sr, err := NewSwaggerReader(srv.Swagger)
		if err != nil {
			return err
		}
		swgr, err := sr.GetSwaggerContent()
		if err != nil {
			return err
		}
		err = pm.MapSwaggerPaths(srv.RoutePath, swgr)
		if err != nil {
			return err
		}
	}
	return nil
}

//MapSwaggerPaths adds all of the paths from the passed Swagger definition
func (pm *PathMap) MapSwaggerPaths(route string, swgr *spec.Swagger) error {
	pi := NewPathItem(route, true)
	pm.Services[route] = pi
	for path, spi := range swgr.Paths.Paths {
		//Paths in Swagger should always begin with '/' so discard the first empty string
		lpi := pm.MapElementPath(pi, strings.Split(path, "/"), 1, true)
		err := pm.AddVerbToPathItem(lpi, spi)
		if err != nil {
			return fmt.Errorf("Error adding path '%s': %s", path, err.Error())
		}
	}
	return nil
}

//AddVerbToPathItem adds the information for the swagger endpoint to the PathItem map
func (pm *PathMap) AddVerbToPathItem(pi *PathItem, spi spec.PathItem) error {
	if spi.Get != nil {
		err := pm.CreateAndAddVerb(pi, "GET", spi.Get)
		if err != nil {
			return err
		}
	}
	if spi.Put != nil {
		err := pm.CreateAndAddVerb(pi, "PUT", spi.Put)
		if err != nil {
			return err
		}
	}
	if spi.Post != nil {
		err := pm.CreateAndAddVerb(pi, "POST", spi.Post)
		if err != nil {
			return err
		}
	}
	if spi.Delete != nil {
		err := pm.CreateAndAddVerb(pi, "DELETE", spi.Delete)
		if err != nil {
			return err
		}
	}
	if spi.Head != nil {
		err := pm.CreateAndAddVerb(pi, "HEAD", spi.Head)
		if err != nil {
			return err
		}
	}
	if spi.Options != nil {
		err := pm.CreateAndAddVerb(pi, "OPTIONS", spi.Options)
		if err != nil {
			return err
		}
	}
	if spi.Patch != nil {
		err := pm.CreateAndAddVerb(pi, "PATCH", spi.Patch)
		if err != nil {
			return err
		}
	}
	return nil
}

//CreateAndAddVerb ensures the passed verb is added to the passed PathItem
func (pm *PathMap) CreateAndAddVerb(pi *PathItem, verb string, op *spec.Operation) error {
	v, exists := pi.Verbs[verb]
	if !exists {
		v = NewVerb(verb, true, op.Produces, op.Consumes)
		pi.Verbs[verb] = v
		for code := range op.Responses.StatusCodeResponses {
			strcode := strconv.Itoa(code)
			v.Responses[strcode] = &Response{
				Response:   strcode,
				Documented: true,
			}
		}
		for _, param := range op.Parameters {
			if "query" == param.In {
				v.QueryParameters[param.Name] = &QueryParameter{
					Key:        param.Name,
					Documented: true,
				}
			}
		}
		return nil
	}
	return fmt.Errorf("Multiple definitions of the '%s' verb on the same path", verb)
}

//MapElementPath adds the path elements into the PathItem maps creating them as necessary
//and returns the leaf PathItem
func (pm *PathMap) MapElementPath(parent *PathItem, elements []string, index int, documented bool) *PathItem {
	item := NewPathItem(elements[index], documented)
	pi, exists := parent.PathItems[item.MapKey()]
	if !exists {
		if !documented && parent.ParameterisedChild() {
			//Match this element against the parameter
			pi = parent.PathItems[ParameterisedItemKey]
		} else {
			pi = item
			parent.PathItems[item.MapKey()] = item
		}
	}
	if index+1 == len(elements) {
		if pi.Verbs == nil {
			pi.Verbs = map[string]*Verb{}
		}
		return pi
	}
	return pm.MapElementPath(pi, elements, index+1, documented)
}

//JSON returns the content of this PathMap as a json object structure
func (pm *PathMap) JSON() string {
	c, _ := json.MarshalIndent(pm, "", "	")
	return string(c)
}

//CheckRequestLogEntry checks the passed request log entry against
//this PathMap. If necessary it will add PathItems if the URL is not
//documented. Verbs, Query Parameters, Response codes used in the log entry
//that are documented will have their coverage count incremented
func (pm *PathMap) CheckRequestLogEntry(le RequestLogEntry) {
	srv, exists := pm.Services[le.Service]
	if !exists {
		srv = NewPathItem(le.Service, false)
		pm.Services[le.Service] = srv
	}
	lpi := pm.MapElementPath(srv, le.PathElements, 0, false)
	v, exists := lpi.Verbs[le.Method]
	if !exists {
		v = NewVerb(le.Method, false, []string{}, []string{})
		lpi.Verbs[le.Method] = v
	}
	resp, exists := v.Responses[le.Response]
	if !exists {
		resp = &Response{
			Response:   le.Response,
			Documented: false,
		}
		v.Responses[le.Response] = resp
	}
	resp.Covered = resp.Covered + 1
	for qelem := range le.Query {
		qparam, exists := v.QueryParameters[qelem]
		if !exists {
			qparam = &QueryParameter{
				Key:        qelem,
				Documented: false,
			}
			v.QueryParameters[qelem] = qparam
		}
		qparam.Covered = qparam.Covered + 1
	}
}
