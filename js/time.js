RegisterCommand("time", function(){
	var source = this.event.args[0],
		nick = this.event.nick,
		date = new Date();
	IRC.Privmsg(source, nick+": the time is " + date);
}, "returns the current time for the bot");