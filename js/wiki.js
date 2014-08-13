RegisterCommand("wiki", function(){
	var args = this.event.message.split(" "),
		source = this.event.args[0],
		cmd = args.shift(),
		nick = this.event.nick,
		url = "http://www.wikipedia.org/wiki/" + encodeURIComponent(args.join("_"));

	if(args.length > 0) {
		IRC.Privmsg(source, nick+": Wikipedia - " + url);	
	} else {
		IRC.Privmsg(source, nick+": usage is - !wiki [query]");
	}
}, "returns an wikipedia page url for the given query");
