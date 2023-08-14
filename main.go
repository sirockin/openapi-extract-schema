package main

import (
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v2"
)

const (
	requestSearchPath        = "$.paths.*.*.requestBody.content.*.schema"
	responseSearchPath       = "$.paths.*.*.responses.*.content.*.schema"
	embeddedObjectSearchPath = "$.components.schemas.*.properties.*.[?(@type=='object')]"
	embeddedArrayObjectSearchPath = "$.components.schemas.*.*.*.*.[?(@type=='object')]"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: openapi-extract-schema {input-file} {output-file}")
		return
	}

	inputFileName := os.Args[1]
	outputFileName := os.Args[2]

	inStream, err := os.Open(inputFileName)
	if err != nil {
		panic(err)
	}

	outStream, err := os.Create(outputFileName)
	if err != nil {
		panic(err)
	}

	spec := Spec{}
	err = yaml.NewDecoder(inStream).Decode(&spec.Object)
	if err != nil {
		panic(err)
	}

	outSpec := spec.Transform()
	err = yaml.NewEncoder(outStream).Encode(outSpec.Object)
	if err != nil {
		panic(err)
	}
}

func (s Spec) Transform() Spec {
	requests := s.FindPath(requestSearchPath)
	groupedRequests := GroupObjects(requests)
	responses := s.FindPath(responseSearchPath)
	groupedResponses := GroupObjects(responses)
	fmt.Printf("Found %d request schema in %d groups\n", len(requests), len(groupedRequests))
	fmt.Printf("Found %d response schema in %d groups\n", len(responses), len(groupedResponses))

	for _, val := range groupedRequests {
		symbol, err := val.paths.requestSymbol()
		if err != nil {
			panic(err)
		}
		symbol = s.uniqueSymbol(symbol)
		s.moveToSchemas(val, symbol)
	}
	for _, val := range groupedResponses {
		symbol, err := val.paths.responseSymbol()
		if err != nil {
			panic(err)
		}
		symbol = s.uniqueSymbol(symbol)
		s.moveToSchemas(val, symbol)
	}

	embeddedObjects := s.FindPath(embeddedObjectSearchPath)
	groupedEmbeddedObjects := GroupObjects(embeddedObjects)
	embeddedArrayObjects := s.FindPath(embeddedArrayObjectSearchPath)
	groupedEmbeddedArrayObjects := GroupObjects(embeddedArrayObjects)
	fmt.Printf("Found %d embedded objects in %d groups\n", len(embeddedObjects), len(groupedEmbeddedObjects))
	fmt.Printf("Found %d embedded array objects in %d groups\n", len(embeddedArrayObjects), len(groupedEmbeddedArrayObjects))


	return s
}

type ObjectWithPath struct {
	object Object
	path   Path
}

type Paths []Path

type ObjectWithPaths struct {
	object Object
	paths  Paths
}

func (ps Paths) responseSymbol() (string, error) {
	if len(ps) == 0 {
		return "", fmt.Errorf("No paths found")
	}
	parts := []string{}
	verb, err := ps.commonValueAtIndex(2)
	if err != nil {
		return "", err
	}
	if verb != "" {
		parts = append(parts, toTitle(verb))
	}
	endpoint, err := ps.commonValueAtIndex(1)
	if err != nil {
		return "", err
	}
	if endpoint != "" {
		parts = append(parts, sanitizeURLPath(endpoint))
	}
	statusCode, err := ps.commonStatusCode()
	if err != nil {
		return "", err
	}
	if statusCode != "" {
		parts = append(parts, statusCode)
	}

	if len(parts) != 3 {
		parts = append([]string{"Common"}, parts...)
	}
	parts = append(parts, "Response")

	return strings.Join(parts, ""), nil
}

func (ps Paths) requestSymbol() (string, error) {
	if len(ps) == 0 {
		return "", fmt.Errorf("No paths found")
	}
	parts := []string{}
	verb, err := ps.commonValueAtIndex(2)
	if err != nil {
		return "", err
	}
	if verb != "" {
		parts = append(parts, toTitle(verb))
	}
	endpoint, err := ps.commonValueAtIndex(1)
	if err != nil {
		return "", err
	}
	if endpoint != "" {
		parts = append(parts, sanitizeURLPath(endpoint))
	}
	if len(parts) != 2 {
		parts = append([]string{"Common"}, parts...)
	}
	parts = append(parts, "Request")

	return strings.Join(parts, ""), nil
}

func (ps Paths) commonValueAtIndex(idx int) (string, error) {
	if len(ps) == 0 {
		return "", fmt.Errorf("No paths found")
	}
	var ret string
	for i, path := range ps {
		if len(path) <= idx {
			return "", fmt.Errorf("Path too short")
		}
		newVal := path[idx]
		if i == 0 {
			ret = newVal
		} else {
			if ret != newVal {
				return "", nil
			}
		}
	}
	return ret, nil
}

func (ps Paths) commonStatusCode() (string, error) {
	idx := 4
	if len(ps) == 0 {
		return "", fmt.Errorf("No paths found")
	}
	var ret string
	for i, path := range ps {
		if len(path) <= idx {
			return "", fmt.Errorf("Path too short")
		}
		newVal := path[idx]
		if i == 0 {
			ret = newVal
		} else {
			if ret != newVal {
				ret = ret[:1] + "xx"
				if ret[:1] != newVal[:1] {
					return "", nil
				}
			}
		}
	}
	return ret, nil
}

func sanitizeURLPath(in string) string {
	in = strings.Trim(in, "/")
	in = strings.ReplaceAll(in, "-", "/")
	vals := strings.Split(in, "/")
	var sb strings.Builder
	for _, v := range vals {
		sb.WriteString(toTitle(v))
	}
	return sb.String()
}

type (
	Object map[interface{}]interface{}
	Spec   struct{ Object }
	Path   []string
)

func NewPath(path string) Path {
	return strings.Split(strings.TrimPrefix(path, "."), ".")
}

func (s Spec) uniqueSymbol(symbol string) string {
	for s.symbolExists(symbol) {
		symbol = nextSymbol(symbol)
	}
	return symbol
}

func (s Spec) symbolExists(symbol string) bool {
	_, exists := s.schemasNode()[symbol]
	return exists
}

func nextSymbol(symbol string) string {
	sym := []rune(symbol)
	var numDigits int
	for numDigits = 0; unicode.IsDigit(sym[len(sym)-1-numDigits]); numDigits++ {
		fmt.Printf("numDigits: %d\n", numDigits)
	}
	if numDigits == 0 {
		return symbol + "2"
	}
	startIndex := len(sym) - (numDigits)
	suffix := sym[startIndex:]
	prefix := sym[:len(sym)-(numDigits)]
	lastNum, err := strconv.Atoi(string(suffix))
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%s%d", string(prefix), (lastNum + 1))
}

func GroupObjects(objects []ObjectWithPath) []ObjectWithPaths {
	ret := []ObjectWithPaths{}
	for _, obj := range objects {
		if idx := findMatchingObjectWithPaths(obj.object, ret); idx >= 0 {
			ret[idx].paths = append(ret[idx].paths, obj.path)
		} else {
			ret = append(ret, ObjectWithPaths{object: obj.object, paths: []Path{obj.path}})
		}
	}
	return ret
}

func findMatchingObjectWithPaths(object Object, list []ObjectWithPaths) int {
	for i, owp := range list {
		if owp.object.isEqual(object) {
			return i
		}
	}
	return -1
}

func (s Spec) FindPath(path string) []ObjectWithPath {
	return s.findPath(NewPath(strings.TrimPrefix(path, "$.")))
}

func (s Spec) findPath(path Path) []ObjectWithPath {
	return s.Object.findPath(path, nil)
}

func (o Object) findPath(path Path, parent Path) []ObjectWithPath {
	if len(path) == 0 {
		return []ObjectWithPath{{object: o, path: parent}}
	}
	// from arbitrary depth '..'
	if path[0] == "" {
		ret := o.findPath(path[1:], parent)
		for k, v := range o {
			obj, ok := v.(Object)
			if ok {
				key := fmt.Sprintf("%v", k)
				ret = append(ret, obj.findPath(path, append(parent, key))...)
			}
		}
		return ret
	}
	if path[0] == "*" {
		ret := []ObjectWithPath{}
		for k, v := range o {
			obj, ok := v.(Object)
			if ok {
				key := fmt.Sprintf("%v", k)
				ret = append(ret, obj.findPath(path[1:], append(parent, key))...)
			}
		}
		return ret
	}

	exp := regexp.MustCompile(`^\[\?\(@([[:alnum:]]+)=='([[:alnum:]]+)'\)\]`)
	result := exp.FindStringSubmatch(path[0])
	if result != nil {
		if o[result[1]] == result[2] {
			return []ObjectWithPath{{object: o, path: parent}}
		}
		return nil
	}
	v, ok := o[path[0]]
	if !ok {
		// try again with int
		i, err := strconv.Atoi(path[0])
		if err != nil {
			return nil
		}
		v, ok = o[i]
	}
	if ok {
		obj, ok := v.(Object)
		if ok {
			return obj.findPath(path[1:], append(parent, path[0]))
		}
	}
	return nil
}

func (s Spec) schemasNode() Object {
	return s.Object.getOrCreateChildObject("components").
		getOrCreateChildObject("schemas")
}

func (o Object) getOrCreateChildObject(name string) Object {
	r, ok := o[name]
	if !ok {
		ret := Object{}
		o[name] = ret
		return ret
	}

	ret, ok := r.(Object)
	if !ok {
		panic(fmt.Errorf("%s is not object", name))
	}
	return ret
}

func (o Object) isEqual(other Object) bool {
	// TODO: perhaps exclude stuff like 'description' from comparison
	return reflect.DeepEqual(o, other)
}

func (s Spec) moveToSchemas(objPath ObjectWithPaths, name string) {
	s.addToSchemas(objPath.object, name)
	for _, path := range objPath.paths {
		s.replaceWithRef(path, name)
	}
}

func (s Spec) addToSchemas(obj Object, name string){
	s.schemasNode()[name] = copyObject(obj)
}

func (s Spec) replaceWithRef(path Path, name string) {
	found := s.findPath(path)
	if len(found) != 1 {
		panic(fmt.Errorf("expected to find 1 object, found %d", len(found)))
	}
	obj := found[0].object
	// Remove all existing keys
	for k := range obj {
		delete(obj, k)
	}
	obj["$ref"] = fmt.Sprintf("#/components/schemas/%s", name)
}

func (s Spec) findMatchingSchema(obj Object) string {
	for name, schema := range s.schemasNode() {
		schemaObj, ok := schema.(Object)
		if !ok{
			continue
		}
		if schemaObj.isEqual(obj) {
			return fmt.Sprintf("%v", name)
		}
	}
	return ""
}

func copyObject(m Object) Object {
	cp := make(Object)
	for k, v := range m {
		vm, ok := v.(Object)
		if ok {
			cp[k] = copyObject(vm)
		} else {
			cp[k] = v
		}
	}
	return cp
}

func toTitle(in string) string {
	return cases.Title(language.English).String(in)
}
