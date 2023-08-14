package main

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
)

type Object map[interface{}]interface{}

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
	return reflect.DeepEqual(o, other)
	// for k, v := range o {
	// 	if k == "description" {
	// 		continue
	// 	}
	// 	childObj, ok := v.(Object)
	// 	if ok {
	// 		childOtherObj, ok := other[k].(Object)
	// 		if !ok {
	// 			return false
	// 		}
	// 		if !childObj.isEqual(childOtherObj) {
	// 			return false
	// 		}
	// 	}else{
	// 		if reflect.DeepEqual(v, other[k]) {
	// 			continue
	// 		}
	// 	}
	// }
	// return true
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
