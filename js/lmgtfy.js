RegisterCommand("lmgtfy", function(){
	var args = this.event.message.split(" "),
		source = this.event.args[0],
		cmd = args.shift(),
		nick = this.event.nick,
		url = "http://lmgtfy.com/?q=" + encodeURIComponent(args.join(" "));
	
	if(args.length > 0) {
		IRC.Privmsg(source, nick+": Let me google that for you - " + url);	
	} else {
		IRC.Privmsg(source, nick+": usage is - !lmgtfy [query]");
	}
}, "returns a 'let me google that for you' search url for the given query");