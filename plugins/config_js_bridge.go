package plugins

import (
	"github.com/robertkrimen/otto"
)

func (pm *PluginManager) InitConfigJSBridge() {
	pm.js.Set("GetConfig", func(call otto.FunctionCall) otto.Value {
		if len(call.ArgumentList) == 0 {
			if val, err := pm.js.ToValue(pm.cfg); err == nil {
				return val
			}
		}
		return otto.FalseValue()
	})
}
