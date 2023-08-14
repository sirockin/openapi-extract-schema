package spec

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
)

type object map[interface{}]interface{}

func (o object) findPath(findPath _path, parentPath _path) []objectWithPath {
	if len(findPath) == 0 {
		return []objectWithPath{{object: o, path: parentPath}}
	}
	// from arbitrary depth '..'
	if findPath[0] == "" {
		ret := o.findPath(findPath[1:], parentPath)
		for k, v := range o {
			obj, ok := v.(object)
			if ok {
				key := fmt.Sprintf("%v", k)
				ret = append(ret, obj.findPath(findPath, append(parentPath, key))...)
			}
		}
		return ret
	}
	if findPath[0] == "*" {
		ret := []objectWithPath{}
		for k, v := range o {
			obj, ok := v.(object)
			if ok {
				key := fmt.Sprintf("%v", k)
				ret = append(ret, obj.findPath(findPath[1:], append(parentPath, key))...)
			}
		}
		return ret
	}

	exp := regexp.MustCompile(`^\[\?\(@([[:alnum:]]+)=='([[:alnum:]]+)'\)\]`)
	result := exp.FindStringSubmatch(findPath[0])
	if result != nil {
		if o[result[1]] == result[2] {
			return []objectWithPath{{object: o, path: parentPath}}
		}
		return nil
	}
	v, ok := o[findPath[0]]
	if !ok {
		// try again with int
		i, err := strconv.Atoi(findPath[0])
		if err != nil {
			return nil
		}
		v, ok = o[i]
	}
	if ok {
		obj, ok := v.(object)
		if ok {
			return obj.findPath(findPath[1:], append(parentPath, findPath[0]))
		}
	}
	return nil
}

func (o object) getOrCreateChildObject(name string) object {
	r, ok := o[name]
	if !ok {
		ret := object{}
		o[name] = ret
		return ret
	}

	ret, ok := r.(object)
	if !ok {
		panic(fmt.Errorf("%s is not object", name))
	}
	return ret
}

func (o object) isEqual(other object) bool {
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

func copyObject(m object) object {
	cp := make(object)
	for k, v := range m {
		vm, ok := v.(object)
		if ok {
			cp[k] = copyObject(vm)
		} else {
			cp[k] = v
		}
	}
	return cp
}
