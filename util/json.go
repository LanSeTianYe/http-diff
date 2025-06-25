package util

import (
	"errors"
	"strings"

	"github.com/oliveagle/jsonpath"
)

func SetJsonFieldToNil(jsonData interface{}, filedName string) (interface{}, error) {
	if strings.HasSuffix(filedName, ".") {
		return nil, errors.New("filedName cannot end with a dot")
	}

	if !strings.Contains(filedName, ".") {
		path := "$"
		subField := filedName

		return setNil(jsonData, path, subField)
	} else {

		dotIndex := strings.LastIndex(filedName, ".")

		path := "$." + filedName[:dotIndex]
		subField := filedName[dotIndex+1:]

		return setNil(jsonData, path, subField)
	}
}

func SetJsonFieldValue(jsonData interface{}, filedName string, value interface{}) error {
	if strings.HasSuffix(filedName, ".") {
		return errors.New("filedName cannot end with a dot")
	}

	if !strings.Contains(filedName, ".") {
		path := "$"
		subField := filedName

		return setValue(jsonData, path, subField, value)
	}

	if strings.Contains(filedName, ".") {
		dotIndex := strings.LastIndex(filedName, ".")

		path := "$." + filedName[:dotIndex]
		subField := filedName[dotIndex+1:]

		return setValue(jsonData, path, subField, value)
	}

	return nil
}

func GetFieldValue(jsonData interface{}, filedName string) (interface{}, error) {
	if strings.HasSuffix(filedName, ".") {
		return nil, errors.New("filedName cannot end with a dot")
	}

	path := "$" + filedName

	lookup, err := jsonpath.JsonPathLookup(jsonData, path)
	if err != nil {
		return nil, errors.New("filed value not found in jsonData [" + path + "]")
	}

	return lookup, nil
}

func setNil(jsonData interface{}, path, subField string) (interface{}, error) {
	lookup, err := jsonpath.JsonPathLookup(jsonData, path)
	if err != nil {
		return nil, err
	}

	var result interface{}
	if m, ok := lookup.(map[string]interface{}); ok {
		result = m[subField]
		m[subField] = nil
	} else {
		return nil, errors.New("jsonData is not a map [" + path + "." + subField + "]")
	}

	return result, nil
}

func setValue(jsonData interface{}, path, subField string, value interface{}) error {
	lookup, err := jsonpath.JsonPathLookup(jsonData, path)
	if err != nil {
		return err
	}

	if m, ok := lookup.(map[string]interface{}); ok {
		m[subField] = value
	} else {
		return errors.New("jsonData is not a map [" + path + "." + subField + "]")
	}

	return nil
}
