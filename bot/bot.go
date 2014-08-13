package bot

import (
	"log"
	"net"
	"os"
	"time"

	"github.com/thoj/go-ircevent"
	"github.com/zenithar/aktarus/config"
	"github.com/zenithar/aktarus/debug"
	"github.com/zenithar/aktarus/plugins"
	"github.com/zenithar/aktarus/state"
)

type Bot struct {
	conn    *irc.Connection
	cfg     *config.Settings
	pm      *plugins.PluginManager
	state   *state.StateTracker
	Quitted chan bool
}

func (bot *Bot) Quit() {
	bot.conn.Quit()
	bot.Quitted <- true
}

func (bot *Bot) Connect() error {
	bot.InitCallbacks()
	// Connect
	if bot.cfg.Irc.Port != "" {
		return bot.conn.Connect(net.JoinHostPort(bot.cfg.Irc.Host, bot.cfg.Irc.Port))
	}
	return bot.conn.Connect(bot.cfg.Irc.Host)
}

func (bot *Bot) InitCallbacks() error {
	// Drop nick handling standardd callbacks (we will be using our own)
	bot.conn.ClearCallback("433")

	// Setup state tracker
	bot.state.InitStateCallbacks()

	// Handle built-in commands
	bot.conn.AddCallback("PRIVMSG", bot.RunBuiltinCommands)

	// Handle built-in callbacks
	bot.conn.AddCallback("433", bot.ReclaimNick)   // Reclaim stolen nicks
	bot.conn.AddCallback("JOIN", bot.AutoVoice)    // Autovoice people
	bot.conn.AddCallback("001", bot.SetBotState)   // Setup bot state
	bot.conn.AddCallback("477", bot.JoinChannels)  // Try to re-join channels
	bot.conn.AddCallback("001", bot.JoinChannels)  // Try to join channels on connect
	bot.conn.AddCallback("KICK", bot.JoinChannels) // Rejoin on kick
	bot.conn.AddCallback("PING", bot.JoinChannels) // Periodically try and rejoin if not already joined
	bot.conn.AddCallback("PONG", bot.JoinChannels) // Periodically try and rejoin if not already joined

	// Setup plugin callbacks
	bot.pm.InitPluginCallbacks()

	return nil
}

func New(cfg *config.Settings) (*Bot, error) {
	// Set up Irc Client
	client := irc.IRC(cfg.Irc.Nick, cfg.Irc.Username)

	if cfg.Irc.Version != "" {
		client.Version = cfg.Irc.Version
	}

	if cfg.Irc.Debug || cfg.Debug {
		client.Debug = true
	}
	if cfg.Irc.Timeout > 0 {
		// Set client timeout to configured amount in seconds
		client.Timeout = time.Duration(cfg.Irc.Timeout) * time.Second
	}
	if cfg.Irc.KeepAlive > 0 {
		// Set client keepalive duration to configured amount in seconds
		client.KeepAlive = time.Duration(cfg.Irc.KeepAlive) * time.Second
	}
	if cfg.Irc.PingFreq > 0 {
		// Set client pingfreq duration to configured amount in seconds
		client.PingFreq = time.Duration(cfg.Irc.PingFreq) * time.Second
	}
	if cfg.Irc.Debug || cfg.Debug {
		client.VerboseCallbackHandler = true
	}

	// Optionally, enable SSL
	client.UseTLS = cfg.Irc.Ssl

	// Setup IRC logger
	client.Log = log.New(os.Stdout, "[irc] ", log.LstdFlags)

	// Make bot instance
	bot := &Bot{
		cfg:     cfg,
		conn:    client,
		Quitted: make(chan bool, 1),
	}

	// Setup state tracker
	bot.state = state.New(cfg, client)

	// Give debug a window into the state handler
	debug.SetState(bot.state)

	// Setup plugin manager
	bot.pm = plugins.New(cfg, client, bot.state)

	// Boot up the plugin js environment
	bot.pm.InitJS()

	return bot, nil
}
