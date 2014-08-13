package bot

import (
	"fmt"
	"github.com/thoj/go-ircevent"
	"github.com/zenithar/aktarus/debug"
	"github.com/zenithar/aktarus/utils"
	"sort"
	"strings"
	"time"
)

func (bot *Bot) RunBuiltinCommands(event *irc.Event) {
	args := strings.Split(strings.TrimSpace(event.Message()), " ")
	command := args[0]

	// Bin the command from the arg list
	if len(args) > 1 {
		args = args[1:]
	} else {
		args = []string{}
	}

	// Get the current channel privileges for the nick sending this command
	privs, ok := bot.state.GetPrivs(event.Arguments[0], event.Nick)

	// Commands must be run by known users
	switch {
	case command == "!rejoin" && event.Arguments[0] == bot.conn.GetNick():
		bot.conn.Join(bot.cfg.Irc.NormalChannel)
		bot.conn.Join(bot.cfg.Irc.StaffChannel)
	case ok && command == "!reload":
		if privs.Owner || privs.Admin || privs.Op {
			bot.pm.InitJS()
			utils.IRCAction(bot.conn, event.Arguments[0], "has reloaded its plugins")
		} else {
			utils.IRCAction(bot.conn, event.Arguments[0], fmt.Sprintf("slaps %s's hands away from the op only controls", event.Nick))
		}
	case command == "!ping":
		bot.conn.Privmsg(event.Arguments[0], fmt.Sprintf("%s: PONG!", event.Nick))
	case ok && command == "!quit":
		if privs.Owner || privs.Admin || privs.Op {
			bot.Quit()
		} else {
			utils.IRCAction(bot.conn, event.Arguments[0], fmt.Sprintf("slaps %s's hands away from the op only controls", event.Nick))
		}
	case ok && command == "!voice":
		if privs.Owner || privs.Admin || privs.Op {
			bot.VoiceAll(event)
		} else {
			utils.IRCAction(bot.conn, event.Arguments[0], fmt.Sprintf("slaps %s's hands away from the op only controls", event.Nick))
		}
	case command == "!help":
		if len(args) == 0 {
			bot.ShowCommandList(event.Arguments[0], event.Nick)
		} else {
			bot.ShowCommandHelp(event.Arguments[0], event.Nick, args[0])
		}
	case ok && command == "!import" && len(args) >= 2 && bot.cfg.Irc.StaffChannel == event.Arguments[0]:
		if privs.Owner {
			overwrite := false
			if len(args) == 3 {
				overwrite = args[2] == "overwrite"
			}
			if err := bot.pm.ImportPlugin(event.Nick, args[0], args[1], overwrite); err != nil {
				bot.conn.Privmsg(bot.cfg.Irc.StaffChannel, fmt.Sprintf("ALERT: %s tried to use !import with %s and got this error: %s", event.Nick, args[0], err.Error()))
			} else {
				bot.conn.Privmsg(bot.cfg.Irc.StaffChannel, fmt.Sprintf("ALERT: %s successfully used !import with %s", event.Nick, args[0]))
			}
		} else {
			utils.IRCAction(bot.conn, event.Arguments[0], fmt.Sprintf("slaps %s's hands away from the op only controls", event.Nick))
		}
	case ok && command == "!debug" && bot.cfg.Irc.StaffChannel == event.Arguments[0]:
		if privs.Owner || privs.Admin || privs.Op {
			switch {
			case len(args) > 0 && args[0] == "on":
				port, alreadyRunning := debug.StartDebugServer()
				if alreadyRunning {
					bot.conn.Privmsg(bot.cfg.Irc.StaffChannel, fmt.Sprintf("%s: Debug server running on port %s", event.Nick, port))
				} else {
					bot.conn.Privmsg(bot.cfg.Irc.StaffChannel, fmt.Sprintf("%s: Debug server started on port %s", event.Nick, port))
				}
				bot.cfg.Debug = true
			case len(args) > 0 && args[0] == "off":
				debug.StopDebugServer()
				bot.conn.Privmsg(bot.cfg.Irc.StaffChannel, fmt.Sprintf("%s: Debug server stopped", event.Nick))
				bot.cfg.Debug = false
			case len(args) > 0 && args[0] == "status":
				status := debug.DebugServerStatus()
				bot.conn.Privmsg(bot.cfg.Irc.StaffChannel, fmt.Sprintf("%s: Debug server is %s", event.Nick, status))
			default:
				bot.conn.Privmsg(bot.cfg.Irc.StaffChannel, fmt.Sprintf("%s: usage - !debug [on|off]", event.Nick))
			}
			bot.conn.VerboseCallbackHandler = bot.cfg.Debug
		} else {
			utils.IRCAction(bot.conn, event.Arguments[0], fmt.Sprintf("slaps %s's hands away from the op only controls", event.Nick))
		}
	}
}

// Handler for reclaiming a stolen nick
func (bot *Bot) ReclaimNick(event *irc.Event) {
	if thief := bot.state.GetNick(bot.cfg.Irc.Nick); thief != nil {
		// Recover nick from thieves
		bot.conn.Privmsg("NickServ", fmt.Sprintf("RECOVER %s %s", bot.cfg.Irc.Nick, bot.cfg.Irc.NickPass))
		time.Sleep(time.Second)
		bot.conn.Privmsg("NickServ", fmt.Sprintf("RELEASE %s %s", bot.cfg.Irc.Nick, bot.cfg.Irc.NickPass))
	}
	bot.SetBotState(event)
}

// Automatically give voice to users in channels in which the bot has Op
func (bot *Bot) AutoVoice(event *irc.Event) {
	if bot.cfg.Irc.AutoVoice {
		time.Sleep(time.Second) // Wait a second before we bother

		privs, ok := bot.state.GetPrivs(event.Arguments[0], bot.conn.GetNick())

		// Do we current have the rights for voicing people?
		if ok && (privs.Owner || privs.Admin || privs.Op || privs.HalfOp) {
			privs, ok = bot.state.GetPrivs(event.Arguments[0], event.Nick)

			// No need to autovoice
			if ok && (privs.Owner || privs.Admin || privs.Op || privs.HalfOp || privs.Voice) {
				return
			}

			// Set mode
			bot.conn.Mode(event.Arguments[0], fmt.Sprintf("+v %s", event.Nick))
		} else {
			// We can't grant voice yet
			bot.conn.Log.Printf("I don't have the privileges to grant voice in channel %s", event.Arguments[0])
		}
	}
}

// Give voice to all users that don't have it yet
func (bot *Bot) VoiceAll(event *irc.Event) {
	privs, ok := bot.state.GetPrivs(event.Arguments[0], bot.conn.GetNick())

	// Do we current have the rights for voicing people?
	if ok && (privs.Owner || privs.Admin || privs.Op || privs.HalfOp) {
		if channel := bot.state.GetChannel(event.Arguments[0]); channel != nil {
			for nick, privs := range channel.Nicks {
				// No need to autovoice
				if privs != nil && (privs.Owner || privs.Admin || privs.Op || privs.HalfOp || privs.Voice) {
					continue
				}

				// Set mode
				bot.conn.Mode(event.Arguments[0], fmt.Sprintf("+v %s", nick))
			}
		}
	} else {
		// We can't grant voice yet
		bot.conn.Log.Printf("I don't have the privileges to grant voice in channel %s", event.Arguments[0])
	}
}

// Function for setting up the botstate
func (bot *Bot) SetBotState(event *irc.Event) {
	if ghost := bot.state.GetNick(bot.cfg.Irc.Nick); ghost != nil && ghost != bot.state.Me() {
		// GHOST the old nick
		bot.conn.Privmsg("NickServ", fmt.Sprintf("GHOST %s %s", bot.cfg.Irc.Nick, bot.cfg.Irc.NickPass))
	}

	// Set up the nick
	bot.conn.Nick(bot.cfg.Irc.Nick)

	// Identify as the nick owner
	bot.conn.Privmsg("NickServ", fmt.Sprintf("IDENTIFY %s", bot.cfg.Irc.NickPass))

	// Tell IRC I'm a bot
	bot.conn.Mode(bot.cfg.Irc.Nick, "+B")
}

// Function for re-joining channels
func (bot *Bot) JoinChannels(event *irc.Event) {
	time.Sleep(time.Second) // Wait a second before we bother

	// Was I kicked?
	kickedFromNormal := event.Code == "KICK" && event.Arguments[1] == bot.state.Me().Nick && event.Arguments[0] == bot.cfg.Irc.NormalChannel
	kickedFromStaff := event.Code == "KICK" && event.Arguments[1] == bot.state.Me().Nick && event.Arguments[0] == bot.cfg.Irc.StaffChannel

	if _, ok := bot.state.GetPrivs(bot.cfg.Irc.NormalChannel, bot.state.Me().Nick); !ok || kickedFromNormal {
		bot.conn.Join(bot.cfg.Irc.NormalChannel)
	}

	if _, ok := bot.state.GetPrivs(bot.cfg.Irc.StaffChannel, bot.state.Me().Nick); !ok || kickedFromStaff {
		bot.conn.Join(bot.cfg.Irc.StaffChannel)
	}
}

// Print out the commands available
func (bot *Bot) ShowCommandList(source, nick string) {
	var commands []string = make([]string, 0)
	commands = append(commands, "!reload", "!ping", "!quit", "!help", "!import", "!debug", "!voice", "!rejoin")

	for cmd, _ := range bot.pm.CommandHelp() {
		commands = append(commands, cmd)
	}
	sort.Strings(commands)
	bot.conn.Privmsg(source, fmt.Sprintf("%s: available commands are: %s", nick, strings.Join(commands, ", ")))
}

// Print out the commands available
func (bot *Bot) ShowCommandHelp(source, nick, cmd string) {
	message := "unknown command, run `!help` to see what commands are available."
	switch cmd {
	case "!ping":
		message = fmt.Sprintf("makes `%s` reply with PONG!", bot.state.Me().Nick)
	case "!reload":
		message = "reloads the plugins"
	case "!quit":
		message = fmt.Sprintf("makes `%s` quit IRC", bot.state.Me().Nick)
	case "!help":
		message = "shows this message, smart ass"
	case "!voice":
		message = "grants voice to everyone in the channel who doesn't already have it"
	case "!import":
		message = "call like !import [url] [name] [overwrite], imports the plugin at [url] into [name].js and loads it into the bot. Will not overwrite unless [overwrite] is set to 'overwrite'"
	case "!debug":
		message = "starts/stops debugging server for inspecting the bot"
	case "!rejoin":
		message = "makes the bot rejoin it's standard channels. Only works via PM."
	default:
		if help, ok := bot.pm.CommandHelp()[cmd]; ok {
			message = help
		}
	}

	bot.conn.Privmsg(source, fmt.Sprintf("%s: %s - %s", nick, cmd, message))
}
