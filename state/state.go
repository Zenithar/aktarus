package state

import (
	"github.com/thoj/go-ircevent"
	"github.com/zenithar/aktarus/config"
	"sync"
)

type StateTracker struct {
	channels map[string]*Channel
	nicks    map[string]*Nick
	conn     *irc.Connection
	mutex    sync.Mutex
	cfg      *config.Settings
}

func New(cfg *config.Settings, conn *irc.Connection) *StateTracker {
	state := &StateTracker{
		channels: make(map[string]*Channel),
		nicks:    make(map[string]*Nick),
		conn:     conn,
		cfg:      cfg,
	}
	state.nicks[cfg.Irc.Nick] = &Nick{
		Nick:     cfg.Irc.Nick,
		Channels: make(map[string]*ChannelPrivileges),
	}
	return state
}
