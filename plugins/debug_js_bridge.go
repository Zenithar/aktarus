package plugins

import (
	"github.com/robertkrimen/otto"
	"github.com/zenithar/aktarus/debug"
)

type pmDebugJSBridge struct {
	StackTrace func(call otto.FunctionCall) otto.Value
}

func (pm *PluginManager) InitDebugJSBridge() {
	bridge := &pmDebugJSBridge{
		StackTrace: func(call otto.FunctionCall) otto.Value {
			if len(call.ArgumentList) == 0 {
				debug.PrintStack()
				return otto.TrueValue()
			}
			return otto.FalseValue()
		},
	}
	pm.js.Set("DEBUG", bridge)
}
