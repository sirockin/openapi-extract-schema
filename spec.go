package main

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type Spec struct{ Object }

func (s Spec) Transform() Spec {
	requests := s.FindPath(requestSearchPath)
	groupedRequests := GroupObjects(requests)
	responses := s.FindPath(responseSearchPath)
	groupedResponses := GroupObjects(responses)
	fmt.Printf("Found %d request schema in %d groups\n", len(requests), len(groupedRequests))
	fmt.Printf("Found %d response schema in %d groups\n", len(responses), len(groupedResponses))

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
	for {
		embeddedObjects := s.FindPath(embeddedObjectSearchPath)
		groupedEmbeddedObjects := GroupObjects(embeddedObjects)
		embeddedArrayObjects := s.FindPath(embeddedArrayObjectSearchPath)
		groupedEmbeddedArrayObjects := GroupObjects(embeddedArrayObjects)
		fmt.Printf("Found %d embedded objects in %d groups\n", len(embeddedObjects), len(groupedEmbeddedObjects))
		fmt.Printf("Found %d embedded array objects in %d groups\n", len(embeddedArrayObjects), len(groupedEmbeddedArrayObjects))
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

func (s Spec) FindPath(path string) []ObjectWithPath {
	return s.findPath(NewPath(strings.TrimPrefix(path, "$.")))
}

func (s Spec) findPath(path Path) []ObjectWithPath {
	return s.Object.findPath(path, nil)
}

func (s Spec) schemasNode() Object {
	return s.Object.getOrCreateChildObject("components").
		getOrCreateChildObject("schemas")
}

func (s Spec) addObjectSchema(obj Object, name string) {
	s.schemasNode()[name] = copyObject(obj)
}

func (s Spec) replaceWithRefs(paths []Path, name string) {
	for _, path := range paths {
		s.replaceWithRef(path, name)
	}
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
		sb.WriteString(toTitle(v))
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

func toTitle(in string) string {
	return cases.Title(language.English).String(in)
}
