package main

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v2"
)

const (
	requestSearchPath  = ".paths.*.*.requestBody.content.*.schema"
	responseSearchPath = ".paths.*.*.responses.*.content.*.schema"
)

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
	statusCode, err := ps.commonValueAtIndex(8)
	if err != nil {
		return "", err
	}
	if endpoint != "" {
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


func sanitizeURLPath(in string) string {
	in = strings.Trim(in, "/")
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

	// for _, val := range groupedRequests {
	// 	symbol := s.uniqueSymbol(val.requestSymbol())
	// 	s.moveToSchemas(val, symbol)
	// }
	// for _, val := range responses {
	// 	s.moveToSchemas(val, val.path.generateSchemaNameFromResponse())
	// }

	return s
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
		// TODO: perhaps exclude stuff like 'description' from comparison
		if reflect.DeepEqual(owp.object, object) {
			return i
		}
	}
	return -1
}

func (s Spec) FindPath(path string) []ObjectWithPath {
	path = strings.TrimPrefix(path, ".")
	return s.Object.findPath(NewPath(path), nil)
}

func (o Object) findPath(path Path, parent Path) []ObjectWithPath {
	if len(path) == 0 {
		return []ObjectWithPath{{object: o, path: parent}}
	}
	switch path[0] {
	case "*":
		ret := []ObjectWithPath{}
		for k, v := range o {
			obj, ok := v.(Object)
			if ok {
				key := fmt.Sprintf("%v", k)
				ret = append(ret, obj.findPath(path[1:], append(parent, key))...)
			}
		}
		return ret
	default:
		v, ok := o[path[0]]
		if ok {
			obj, ok := v.(Object)
			if ok {
				return obj.findPath(path[1:], append(parent, path[0]))
			}
		}
		return nil
	}
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

func (s Spec) moveToSchemas(objPath ObjectWithPath, name string) {
	s.schemasNode()[name] = objPath.object
}

func toTitle(in string) string {
	return cases.Title(language.English).String(in)
}
