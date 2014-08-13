package plugins

import (
	"github.com/robertkrimen/otto"
	"github.com/thoj/go-ircevent"
	"github.com/zenithar/aktarus/config"
	"github.com/zenithar/aktarus/utils"
	"log"
	"strings"
)

type pluginFunc struct {
	function func(otto.Value)
	help     string
}

type Plugin struct {
	commands  map[string]*pluginFunc
	callbacks map[string][]*pluginFunc
	log       *log.Logger
	js        *otto.Otto
	cfg       *config.Settings
}

func (p *Plugin) SetCommand(name string, command otto.Value, help string) {
	if _, ok := p.commands[name]; ok {
		if p.cfg.Irc.Debug || p.cfg.Debug {
			p.log.Printf("Warning: Command `%s` was already defined. Overriding...", name)
		}
	}
	wrappedCommand := func(env otto.Value) {
		_, err := command.Call(env)
		if err != nil {
			p.log.Printf("Command `%s` errored: %s", name, err)
		}
	}
	p.commands[name] = &pluginFunc{
		function: wrappedCommand,
		help:     help,
	}
}

func (p *Plugin) AddCallback(eventCode string, name string, callback otto.Value) {
	wrappedCallback := func(env otto.Value) {
		_, err := callback.Call(env)
		if err != nil {
			p.log.Printf("Callback `%s` (%#v) for event code `%s` errored: %s", name, callback, eventCode, err)
		}
	}
	p.callbacks[eventCode] = append(p.callbacks[eventCode], &pluginFunc{
		function: wrappedCallback,
		help:     name,
	})
}

func (p *Plugin) RunCallbacks(event *irc.Event) {
	if callbacks, ok := p.callbacks[event.Code]; ok {
		if p.cfg.Irc.Debug || p.cfg.Debug {
			p.log.Printf("%v (%v) >> %#v\n", event.Code, len(callbacks), event)
		}

		for _, callback := range callbacks {
			callback.function(p.jsEnv(event))
		}
	}

	// Handle wildcard callbacks
	if callbacks, ok := p.callbacks["*"]; ok {
		if p.cfg.Irc.Debug || p.cfg.Debug {
			p.log.Printf("Wildcard %v (%v) >> %#v\n", event.Code, len(callbacks), event)
		}

		for _, callback := range callbacks {
			callback.function(p.jsEnv(event))
		}
	}
}

func (p *Plugin) RunCommand(event *irc.Event) bool {
	if event.Message()[0] == '!' && len(event.Message()) > 1 {
		call := strings.SplitN(event.Message()[1:], " ", 2)
		command := call[0]

		var ok bool
		if _, ok = p.commands[command]; ok {
			if p.cfg.Irc.Debug || p.cfg.Debug {
				p.log.Printf("%v (!%v) >> %#v\n", event.Code, command, event)
			}

			cmd := p.commands[command].function
			cmd(p.jsEnv(event))
		}

		return ok
	}
	return false
}

func (p *Plugin) CommandHelp() map[string]string {
	var commands map[string]string = make(map[string]string, 0)
	for name, cmd := range p.commands {
		commands["!"+name] = cmd.help
	}
	return commands
}

func (p *Plugin) eventToValue(event *irc.Event) otto.Value {
	obj, _ := p.js.Object("({})")
	obj.Set("code", event.Code)
	obj.Set("raw", event.Raw)
	obj.Set("nick", event.Nick)
	obj.Set("host", event.Host)
	obj.Set("source", event.Source)
	obj.Set("user", event.User)
	obj.Set("args", utils.SliceToJavascriptArray(p.js, event.Arguments))
	obj.Set("message", event.Message())
	return obj.Value()
}

func (p *Plugin) jsEnv(event *irc.Event) otto.Value {
	obj, _ := p.js.Object("({})")
	obj.Set("event", p.eventToValue(event))
	obj.Set("log", func(call otto.FunctionCall) otto.Value {
		if len(call.ArgumentList) == 1 && call.ArgumentList[0].IsString() {
			p.log.Println(call.ArgumentList[0].String())
			return otto.TrueValue()
		} else {
			return otto.FalseValue()
		}
	})
	return obj.Value()
}
