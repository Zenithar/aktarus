package state

type NickModes struct {
	Bot, Invisible, Oper, WallOps, HiddenHost, SSL bool
}

type Nick struct {
	Nick, User, Host, Name string
	Modes                  NickModes
	Channels               map[string]*ChannelPrivileges
}

func (n *Nick) InChannel(channel string) (isIn bool) {
	_, isIn = n.Channels[channel]
	return
}
