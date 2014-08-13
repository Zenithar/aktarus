package plugins

import (
	"github.com/robertkrimen/otto"
	"github.com/zenithar/aktarus/state"
	"github.com/zenithar/aktarus/utils"
)

type pmIRCJSBridge struct {
	Nick, GetNick, SendRaw, Privmsg, Notice, Action,
	Part, Join, Who, Whois, Mode, Nicks, Channels,
	Topic, Away, Invite, Oper, GetPrivs, Redispatch func(call otto.FunctionCall) otto.Value
}

func stateNickToValue(nick *state.Nick) (val otto.Value) {
	return
}

func (pm *PluginManager) InitIRCJSBridge() {
	bridge := &pmIRCJSBridge{
		Nick: func(call otto.FunctionCall) otto.Value {
			if len(call.ArgumentList) == 1 && call.ArgumentList[0].IsString() {
				pm.conn.Nick(call.Argument(0).String())
				return otto.TrueValue()
			} else {
				return otto.FalseValue()
			}
		},
		GetNick: func(call otto.FunctionCall) otto.Value {
			val, err := otto.ToValue(pm.conn.GetNick())
			if err != nil {
				return otto.NullValue()
			}
			return val
		},
		SendRaw: func(call otto.FunctionCall) otto.Value {
			if len(call.ArgumentList) == 1 && call.ArgumentList[0].IsString() {
				pm.conn.SendRaw(call.Argument(0).String())
				return otto.TrueValue()
			} else {
				return otto.FalseValue()
			}
		},
		Privmsg: func(call otto.FunctionCall) otto.Value {
			if len(call.ArgumentList) == 2 && call.ArgumentList[0].IsString() && call.ArgumentList[1].IsString() {
				pm.conn.Privmsg(call.Argument(0).String(), call.Argument(1).String())
				return otto.TrueValue()
			} else {
				return otto.FalseValue()
			}
		},
		Notice: func(call otto.FunctionCall) otto.Value {
			if len(call.ArgumentList) == 2 && call.ArgumentList[0].IsString() && call.ArgumentList[1].IsString() {
				pm.conn.Notice(call.Argument(0).String(), call.Argument(1).String())
				return otto.TrueValue()
			} else {
				return otto.FalseValue()
			}
		},
		Part: func(call otto.FunctionCall) otto.Value {
			if len(call.ArgumentList) == 1 && call.ArgumentList[0].IsString() {
				pm.conn.Part(call.Argument(0).String())
				return otto.TrueValue()
			} else {
				return otto.FalseValue()
			}
		},
		Join: func(call otto.FunctionCall) otto.Value {
			if len(call.ArgumentList) == 1 && call.ArgumentList[0].IsString() {
				pm.conn.Join(call.Argument(0).String())
				return otto.TrueValue()
			} else {
				return otto.FalseValue()
			}
		},
		Who: func(call otto.FunctionCall) otto.Value {
			if len(call.ArgumentList) == 1 && call.ArgumentList[0].IsString() {
				pm.conn.Who(call.Argument(0).String())
				return otto.TrueValue()
			} else {
				return otto.FalseValue()
			}
		},
		Whois: func(call otto.FunctionCall) otto.Value {
			if len(call.ArgumentList) == 1 && call.ArgumentList[0].IsString() {
				pm.conn.Whois(call.Argument(0).String())
				return otto.TrueValue()
			} else {
				return otto.FalseValue()
			}
		},
		Mode: func(call otto.FunctionCall) otto.Value {
			if len(call.ArgumentList) == 1 && call.ArgumentList[0].IsString() {
				pm.conn.Mode(call.Argument(0).String())
				return otto.TrueValue()
			} else {
				if len(call.ArgumentList) > 1 && call.ArgumentList[0].IsString() {
					var args []string
					for _, arg := range call.ArgumentList[1:] {
						if !arg.IsString() {
							return otto.FalseValue()
						}
						args = append(args, arg.String())
					}
					pm.conn.Mode(call.Argument(0).String(), args...)
				}
				return otto.FalseValue()
			}
		},
		Nicks: func(call otto.FunctionCall) otto.Value {
			return utils.SliceToJavascriptArray(pm.js, pm.state.Nicks())
		},
		Channels: func(call otto.FunctionCall) otto.Value {
			return utils.SliceToJavascriptArray(pm.js, pm.state.Channels())
		},
		Action: func(call otto.FunctionCall) otto.Value {
			if len(call.ArgumentList) == 2 && call.ArgumentList[0].IsString() && call.ArgumentList[1].IsString() {
				utils.IRCAction(pm.conn, call.Argument(0).String(), call.Argument(1).String())
				return otto.TrueValue()
			} else {
				return otto.FalseValue()
			}
		},
		Topic: func(call otto.FunctionCall) otto.Value {
			switch {
			case len(call.ArgumentList) == 1 && call.ArgumentList[0].IsString():
				utils.IRCTopic(pm.conn, call.Argument(0).String())
				return otto.TrueValue()
			case len(call.ArgumentList) == 2 && call.ArgumentList[0].IsString() && call.ArgumentList[1].IsString():
				utils.IRCTopic(pm.conn, call.Argument(0).String(), call.Argument(1).String())
				return otto.TrueValue()
			default:
				return otto.FalseValue()
			}
		},
		Away: func(call otto.FunctionCall) otto.Value {
			switch {
			case len(call.ArgumentList) == 0:
				utils.IRCAway(pm.conn)
				return otto.TrueValue()
			case len(call.ArgumentList) == 1 && call.ArgumentList[0].IsString():
				utils.IRCAway(pm.conn, call.Argument(0).String())
				return otto.TrueValue()
			default:
				return otto.FalseValue()
			}
		},
		Oper: func(call otto.FunctionCall) otto.Value {
			if len(call.ArgumentList) == 2 && call.ArgumentList[0].IsString() && call.ArgumentList[1].IsString() {
				utils.IRCOper(pm.conn, call.Argument(0).String(), call.Argument(1).String())
				return otto.TrueValue()
			} else {
				return otto.FalseValue()
			}
		},
		Invite: func(call otto.FunctionCall) otto.Value {
			if len(call.ArgumentList) == 2 && call.ArgumentList[0].IsString() && call.ArgumentList[1].IsString() {
				utils.IRCInvite(pm.conn, call.Argument(0).String(), call.Argument(1).String())
				return otto.TrueValue()
			} else {
				return otto.FalseValue()
			}
		},
		GetPrivs: func(call otto.FunctionCall) otto.Value {
			if len(call.ArgumentList) == 2 && call.ArgumentList[0].IsString() && call.ArgumentList[1].IsString() {
				if privs, ok := pm.state.GetPrivs(call.Argument(0).String(), call.Argument(1).String()); ok {
					if val, err := pm.js.ToValue(privs); err == nil {
						return val
					}
				}
			}
			return otto.FalseValue()
		},
		Redispatch: func(call otto.FunctionCall) otto.Value {
			if len(call.ArgumentList) >= 7 {
				arguments := make([]string, 0)
				for _, arg := range call.ArgumentList {
					if !arg.IsString() {
						return otto.FalseValue()
					}
					arguments = append(arguments, arg.String())
				}
				utils.IRCRedispatch(
					pm.conn,
					arguments[0],
					arguments[1],
					arguments[2],
					arguments[3],
					arguments[4],
					arguments[5],
					arguments[6:]...,
				)
				return otto.TrueValue()
			}
			return otto.FalseValue()
		},
	}
	pm.js.Set("IRC", bridge)
}
