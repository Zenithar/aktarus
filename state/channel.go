package state

import (
	"strconv"
)

type ChannelModes struct {
	// MODE +p, +s, +t, +n, +m
	Private, Secret, ProtectedTopic, NoExternalMsg, Moderated bool

	// MODE +i, +O, +z
	InviteOnly, OperOnly, SSLOnly bool

	// MODE +r, +Z
	Registered, AllSSL bool

	// MODE +k
	Key string

	// MODE +l
	Limit int
}

type ChannelPrivileges struct {
	// MODE +q, +a, +o, +h, +v
	Owner, Admin, Op, HalfOp, Voice bool
}

type Channel struct {
	Name, Topic string
	Modes       ChannelModes
	Nicks       map[string]*ChannelPrivileges
}

func (channel *Channel) ParseModes(modes string, modeargs ...string) {
	var modeop bool // true => add mode, false => remove mode
	for i := 0; i < len(modes); i++ {
		switch m := modes[i]; m {
		case '+':
			modeop = true
		case '-':
			modeop = false
		case 'i':
			channel.Modes.InviteOnly = modeop
		case 'm':
			channel.Modes.Moderated = modeop
		case 'n':
			channel.Modes.NoExternalMsg = modeop
		case 'p':
			channel.Modes.Private = modeop
		case 'r':
			channel.Modes.Registered = modeop
		case 's':
			channel.Modes.Secret = modeop
		case 't':
			channel.Modes.ProtectedTopic = modeop
		case 'z':
			channel.Modes.SSLOnly = modeop
		case 'Z':
			channel.Modes.AllSSL = modeop
		case 'O':
			channel.Modes.OperOnly = modeop
		case 'k':
			if modeop && len(modeargs) != 0 {
				channel.Modes.Key, modeargs = modeargs[0], modeargs[1:]
			} else if !modeop {
				channel.Modes.Key = ""
			}
		case 'l':
			if modeop && len(modeargs) != 0 {
				channel.Modes.Limit, _ = strconv.Atoi(modeargs[0])
				modeargs = modeargs[1:]
			} else if !modeop {
				channel.Modes.Limit = 0
			}
		case 'q', 'a', 'o', 'h', 'v':
			if len(modeargs) != 0 {
				if privs, ok := channel.Nicks[modeargs[0]]; ok {
					switch m {
					case 'q':
						privs.Owner = modeop
					case 'a':
						privs.Admin = modeop
					case 'o':
						privs.Op = modeop
					case 'h':
						privs.HalfOp = modeop
					case 'v':
						privs.Voice = modeop
					}
					modeargs = modeargs[1:]
				}
			}
		}
	}
}

func (channel *Channel) HasNick(nick string) (hasNick bool) {
	_, hasNick = channel.Nicks[nick]
	return
}
