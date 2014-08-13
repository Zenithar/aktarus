package state

func (st *StateTracker) associate(nick, channel string) *ChannelPrivileges {
	channelObj := st.channels[channel]
	nickObj := st.nicks[nick]

	// Haven't seen this channel before
	if channelObj == nil {
		// Create a channel object
		channelObj = &Channel{
			Name:  channel,
			Nicks: make(map[string]*ChannelPrivileges),
		}

		// Put it in the channels map
		st.channels[channel] = channelObj

		// Get some initial info about it
		st.conn.Mode(channelObj.Name)
		st.conn.Who(channelObj.Name)
	}

	// Not seen this nick before
	if nickObj == nil {
		// Create a nick object
		nickObj = &Nick{
			Nick:     nick,
			Channels: make(map[string]*ChannelPrivileges),
		}

		// Put it in the nicks map
		st.nicks[nick] = nickObj

		// Get some inital info about it
		st.conn.Who(nickObj.Nick)
	}

	privs := new(ChannelPrivileges)
	nickObj.Channels[channel] = privs
	channelObj.Nicks[nick] = privs
	return privs
}

func (st *StateTracker) disassociate(nick, channel string) {
	if channelObj, ok := st.channels[channel]; ok {
		delete(channelObj.Nicks, nick)
	}
	if nickObj, ok := st.nicks[nick]; ok {
		delete(nickObj.Channels, channel)
	}
}

func (st *StateTracker) changeNick(oldNick, newNick string) {
	if nick, ok := st.nicks[oldNick]; ok {
		nick.Nick = newNick
		st.nicks[newNick] = nick
		delete(st.nicks, oldNick)
	}
	for _, channel := range st.channels {
		if privs, ok := channel.Nicks[oldNick]; ok {
			channel.Nicks[newNick] = privs
			delete(channel.Nicks, oldNick)
		}
	}
}

func (st *StateTracker) deleteNick(nick string) {
	if _, ok := st.nicks[nick]; ok {
		delete(st.nicks, nick)
	}
	for _, channel := range st.channels {
		if _, ok := channel.Nicks[nick]; ok {
			delete(channel.Nicks, nick)
		}
	}
}

func (st *StateTracker) setTopic(channel, topic string) {
	channelObj := st.channels[channel]

	// Haven't seen this channel before
	if channelObj == nil {
		// Create a channel object
		channelObj = &Channel{
			Name:  channel,
			Nicks: make(map[string]*ChannelPrivileges),
		}

		// Put it in the channels map
		st.channels[channel] = channelObj

		// Get some initial info about it
		st.conn.Mode(channelObj.Name)
		st.conn.Who(channelObj.Name)
	}

	channelObj.Topic = topic
}
