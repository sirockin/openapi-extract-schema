package spec

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v2"
)

const (
	requestSearchPath             = "$.paths.*.*.requestBody.content.*.schema"
	responseSearchPath            = "$.paths.*.*.responses.*.content.*.schema"
	embeddedObjectSearchPath      = "$.components.schemas.*.properties.*.[?(@type=='object')]"
	embeddedArrayObjectSearchPath = "$.components.schemas.*.*.*.*.[?(@type=='object')]"
)

type Spec struct{ object }

func NewFromYaml(reader io.Reader) (*Spec, error) {
	ret := Spec{}
	err := yaml.NewDecoder(reader).Decode(&ret.object)
	if err != nil {
		return nil, err
	}
	return &ret, nil
}

func (s Spec) ToYaml(writer io.Writer) error {
	return yaml.NewEncoder(writer).Encode(&s.object)
}

func (s Spec) Transform() Spec {
	requests := removeRefs(s.findStringPath(requestSearchPath))
	groupedRequests := groupObjects(requests)
	responses := removeRefs(s.findStringPath(responseSearchPath))
	groupedResponses := groupObjects(responses)
	fmt.Printf("Found %d embedded request schema in %d groups\n", len(requests), len(groupedRequests))
	fmt.Printf("Found %d embedded response schema in %d groups\n", len(responses), len(groupedResponses))

	for _, val := range groupedRequests {
		symbol := s.findMatchingSchema(val.object)
		if symbol == "" {
			var err error
			symbol, err = val.paths.requestSymbol()
			if err != nil {
				panic(err)
			}

			symbol = s.uniqueSymbol(symbol)
			s.addObjectSchema(val.object, symbol)
		}
		s.replaceWithRefs(val.paths, symbol)
	}
	for _, val := range groupedResponses {
		symbol := s.findMatchingSchema(val.object)
		if symbol == "" {
			var err error
			symbol, err = val.paths.responseSymbol()
			if err != nil {
				panic(err)
			}
			symbol = s.uniqueSymbol(symbol)
			s.addObjectSchema(val.object, symbol)
		}
		s.replaceWithRefs(val.paths, symbol)
	}

	// We need to do this iteratively since there may be more than one level of embedded object
	fmt.Printf("Checking components.schemas for embedded schemas:\n")
	for i:=1;;i++{
		fmt.Printf("\tIteration %d:\n", i)
		embeddedObjects := s.findStringPath(embeddedObjectSearchPath)
		groupedEmbeddedObjects := groupObjects(embeddedObjects)
		embeddedArrayObjects := s.findStringPath(embeddedArrayObjectSearchPath)
		groupedEmbeddedArrayObjects := groupObjects(embeddedArrayObjects)
		fmt.Printf("\t\tFound %d embedded objects in %d groups\n", len(embeddedObjects), len(groupedEmbeddedObjects))
		fmt.Printf("\t\tFound %d embedded array objects in %d groups\n", len(embeddedArrayObjects), len(groupedEmbeddedArrayObjects))
		if len(embeddedObjects) == 0 && len(embeddedArrayObjects) == 0 {
			break
		}
		for _, val := range groupedEmbeddedObjects {
			symbol := s.findMatchingSchema(val.object)
			if symbol == "" {
				var err error
				symbol, err = val.paths.embeddedSymbol()
				if err != nil {
					panic(err)
				}

				symbol = s.uniqueSymbol(symbol)
				s.addObjectSchema(val.object, symbol)
			}
			s.replaceWithRefs(val.paths, symbol)
		}
		for _, val := range groupedEmbeddedArrayObjects {
			symbol := s.findMatchingSchema(val.object)
			if symbol == "" {
				var err error
				symbol, err = val.paths.embeddedArraySymbol()
				if err != nil {
					panic(err)
				}

				symbol = s.uniqueSymbol(symbol)
				s.addObjectSchema(val.object, symbol)
			}
			s.replaceWithRefs(val.paths, symbol)
		}

	}
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

func removeRefs(in []objectWithPath) []objectWithPath {
	return filter(in, func(o objectWithPath) bool { 
		_, remove := o.object["$ref"]
		return !remove
	})
}

func filter[T any](slice []T, f func(T) bool) []T {
    var n []T
    for _, e := range slice {
        if f(e) {
            n = append(n, e)
        }
    }
    return n
}

func (s Spec) findStringPath(path string) []objectWithPath {
	return s.findPath(newPath(strings.TrimPrefix(path, "$.")))
}

func (s Spec) findPath(path _path) []objectWithPath {
	return s.object.findPath(path, nil)
}

func (s Spec) schemasNode() object {
	return s.object.getOrCreateChildObject("components").
		getOrCreateChildObject("schemas")
}

func (s Spec) addObjectSchema(obj object, name string) {
	s.schemasNode()[name] = copyObject(obj)
}

func (s Spec) replaceWithRefs(paths []_path, name string) {
	for _, path := range paths {
		s.replaceWithRef(path, name)
	}
}

func (s Spec) replaceWithRef(path _path, name string) {
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

func (s Spec) findMatchingSchema(obj object) string {
	for name, schema := range s.schemasNode() {
		schemaObj, ok := schema.(object)
		if !ok {
			continue
		}
		if schemaObj.isEqual(obj) {
			return fmt.Sprintf("%v", name)
		}
	}
	return ""
}

func sanitizeURLPath(in string) string {
	in = strings.Trim(in, "/")
	in = strings.ReplaceAll(in, "-", "/")
	vals := strings.Split(in, "/")
	var sb strings.Builder
	for _, v := range vals {
		sb.WriteString(capitalizeFirst(v))
	}
	return sb.String()
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

func groupObjects(objects []objectWithPath) []objectWithPaths {
	ret := []objectWithPaths{}
	for _, obj := range objects {
		if idx := findMatchingObjectWithPaths(obj.object, ret); idx >= 0 {
			ret[idx].paths = append(ret[idx].paths, obj.path)
		} else {
			ret = append(ret, objectWithPaths{object: obj.object, paths: []_path{obj.path}})
		}
	}
	return ret
}

func findMatchingObjectWithPaths(obj object, list []objectWithPaths) int {
	for i, owp := range list {
		if owp.object.isEqual(obj) {
			return i
		}
	}
	return -1
}

func toTitle(in string) string {
	return cases.Title(language.English).String(in)
}

func capitalizeFirst(in string) string {
	// return cases.Title(language.English).String(in)
	return strings.ToUpper(in[:1]) + in[1:]
}
