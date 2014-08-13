package utils

import (
	"github.com/robertkrimen/otto"
	"reflect"
)

func SliceToJavascriptArray(js *otto.Otto, slice interface{}) otto.Value {
	if reflect.TypeOf(slice).Kind() != reflect.Slice {
		panic("You must pass in a slice to utils.SliceToJavascriptArray")
	}
	val := reflect.ValueOf(slice)
	arr, _ := js.Object("([])")
	for i := 0; i < val.Len(); i++ {
		arr.Call("push", val.Index(i))
	}
	return arr.Value()
}
