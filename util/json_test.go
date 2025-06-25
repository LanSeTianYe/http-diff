package util

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
)

func TestSetJsonFieldToNil(t *testing.T) {
	var data1 interface{}
	var data2 interface{}

	err := json.Unmarshal([]byte(`{"name": "Alice", "age": {"s":"a"}}`), &data1)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal([]byte(`{"name": "Alice", "age": 31}`), &data2)
	if err != nil {
		panic(err)
	}

	diff1 := cmp.Diff(data1, data2)
	assert.NotEqual(t, "", diff1)

	_, err = SetJsonFieldToNil(data1, "age")
	assert.Nil(t, err)

	_, err = SetJsonFieldToNil(data2, "age")
	assert.Nil(t, err)

	diff2 := cmp.Diff(data1, data2)
	assert.Equal(t, "", diff2)
}

func TestSetJsonFieldToNilMutilLevel(t *testing.T) {
	var data1 interface{}
	var data2 interface{}

	err := json.Unmarshal([]byte(`{"name": "Alice", "address": {"city":"beijing"}}`), &data1)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal([]byte(`{"name": "Alice", "address": {"city":"luoyang"}}`), &data2)
	if err != nil {
		panic(err)
	}

	diff1 := cmp.Diff(data1, data2)
	assert.NotEqual(t, "", diff1)

	_, err = SetJsonFieldToNil(data1, "address.city")
	assert.Nil(t, err)

	_, err = SetJsonFieldToNil(data2, "address.city")
	assert.Nil(t, err)

	diff2 := cmp.Diff(data1, data2)
	assert.Equal(t, "", diff2)
}

func TestSetJsonFieldValue(t *testing.T) {
	var data1 interface{}
	var data2 interface{}

	err := json.Unmarshal([]byte(`{"name": "Alice", "age": {"s":"a"}}`), &data1)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal([]byte(`{"name": "Alice", "age": 31}`), &data2)
	if err != nil {
		panic(err)
	}

	diff1 := cmp.Diff(data1, data2)
	assert.NotEqual(t, "", diff1)

	err = SetJsonFieldValue(data1, "age", "10")
	assert.Nil(t, err)

	err = SetJsonFieldValue(data2, "age", "10")
	assert.Nil(t, err)

	diff2 := cmp.Diff(data1, data2)
	assert.Equal(t, "", diff2)
}

func TestSetJsonFieldValueMutilLevel(t *testing.T) {
	var data1 interface{}
	var data2 interface{}

	err := json.Unmarshal([]byte(`{"name": "Alice", "address": {"city":"beijing"}}`), &data1)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal([]byte(`{"name": "Alice", "address": {"city":"luoyang"}}`), &data2)
	if err != nil {
		panic(err)
	}

	diff1 := cmp.Diff(data1, data2)
	assert.NotEqual(t, "", diff1)

	err = SetJsonFieldValue(data1, "address.city", "shanghai")
	assert.Nil(t, err)

	err = SetJsonFieldValue(data2, "address.city", "shanghai")
	assert.Nil(t, err)

	diff2 := cmp.Diff(data1, data2)
	assert.Equal(t, "", diff2)
}

func TestGetFieldValue(t *testing.T) {
	var data1 interface{}
	var data2 interface{}

	err := json.Unmarshal([]byte(`{"name": "Alice", "address": {"city":"beijing"}}`), &data1)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal([]byte(`{"name": "Alice", "address": {"city":"luoyang"}}`), &data2)
	if err != nil {
		panic(err)
	}

	value1, err1 := GetFieldValue(data1, "address.city")
	assert.Nil(t, err1)
	assert.Equal(t, "beijing", value1)

	value2, err2 := GetFieldValue(data2, "address.city")
	assert.Nil(t, err2)
	assert.Equal(t, "luoyang", value2)
}
