RegisterCommand("hand", function() {
	var args = this.event.message.split(" "),
		target = this.event.args[0],
		cmd = args.shift(),
		nick = this.event.nick;

	if(args.length == 1) {
		IRC.Action(target, "hands " + args[0] + " to " + nick)
		return
	} else if(args.length >= 2 && args[0] == "me") {
		IRC.Action(target, "hands " + args[1] + " to " + nick)
		return
	} else if(args.length == 3 && args[1] == "to"){
		IRC.Action(target, "hands " + args[0] + " to " + args[2] + ", courtesy of " + nick)
		return
	} else if(args.length > 3) {
		var message = args.join(" ")
		var idx = message.indexOf(" to ");
		if(idx > -1) {
			IRC.Action(target, "hands " + message.substring(0, idx) + message.substring(idx, message.length) + ", courtesy of " + nick)
			return
		}
	}
	IRC.Privmsg(target, nick + ": usage - !hand [object] to [recipient] OR !hand me [object]")
}, "!hand [object] to [recipient] or !hand me [object] to have the bot hand an item to someone")