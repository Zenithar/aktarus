RegisterCommand("invite", function() {
	var args = this.event.message.split(" "),
		source = this.event.args[0],
		cmd = args.shift(),
		nick = this.event.nick,
		cfg = GetConfig(),
		privs = IRC.GetPrivs(source, nick);

	if(privs && (privs.Owner || privs.Admin || privs.Op) && source == cfg.Irc.StaffChannel) {
		IRC.Invite(args[0], cfg.Irc.StaffChannel)
	} else {
		IRC.Action(source, "slaps "+nick+"'s hands away from the op only controls")
	}
}, "invites a user to the staff channel");