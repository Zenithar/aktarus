package state

// Returns a Nick object
func (st *StateTracker) GetNick(n string) (nick *Nick) {
	nick, _ = st.nicks[n]
	return
}

// Returns a Channel object
func (st *StateTracker) GetChannel(c string) (channel *Channel) {
	channel, _ = st.channels[c]
	return
}

// Returns Nick object for the bot
func (st *StateTracker) Me() *Nick {
	return st.GetNick(st.conn.GetNick())
}

// Return string slice of known nicks
func (st *StateTracker) Nicks() (nicks []string) {
	nicks = make([]string, 0)
	for nick, _ := range st.nicks {
		nicks = append(nicks, nick)
	}
	return
}

// Return string slice of known channels
func (st *StateTracker) Channels() (channels []string) {
	channels = make([]string, 0)
	for channel, _ := range st.channels {
		channels = append(channels, channel)
	}
	return
}

// Returns a ChannelPrivs object for the given nick.channel
func (st *StateTracker) GetPrivs(c, n string) (privs *ChannelPrivileges, ok bool) {
	var channel *Channel
	if channel, ok = st.channels[c]; ok {
		privs, ok = channel.Nicks[n]
	}
	return
}
