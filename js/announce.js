RegisterCommand("announce", function() {
	var args = this.event.message.split(" "),
		source = this.event.args[0],
		cmd = args.shift(),
		nick = this.event.nick,
		cfg = GetConfig(),
		privs = IRC.GetPrivs(source, nick);

	if(privs && (privs.Owner || privs.Admin || privs.Op || privs.HalfOp) && source == cfg.Irc.StaffChannel) {
		IRC.Privmsg(cfg.Irc.NormalChannel, "NOTICE: " + args.join(" "))
	} else {
		IRC.Action(source, "slaps "+nick+"'s hands away from the op only controls")
	}
}, "announces a message to the normal channel");