package spec

import (
	"fmt"
	"strings"
)
type Path   []string
type Paths []Path

func NewPath(path string) Path {
	return strings.Split(strings.TrimPrefix(path, "."), ".")
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

func (ps Paths) embeddedSymbol() (string, error) {
	if len(ps) == 0 {
		return "", fmt.Errorf("No paths found")
	}
	if len(ps[0]) == 0 {
		return "", fmt.Errorf("empty path")
	}
	return ps[0][len(ps[0])-1], nil
}

func (ps Paths) embeddedArraySymbol() (string, error) {
	if len(ps) == 0 {
		return "", fmt.Errorf("No paths found")
	}
	if len(ps[0]) <= 1 {
		return "", fmt.Errorf("path not long enough")
	}
	return ps[0][len(ps[0])-2] + "Item", nil
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
