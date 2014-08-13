RegisterCommand("wb", function(){
	var args = this.event.message.split(" "),
		source = this.event.args[0],
		cmd = args.shift(),
		nick = this.event.nick,
		base = "http://wurstball.de/",
		url = base + encodeURIComponent(args.join("_")),
		minimum = 1,
		maximum = 200000,
		rnd = base + Math.round(Math.exp(Math.random()*Math.log(maximum-minimum+1)))+minimum;

	if(args.length > 0) {
		IRC.Privmsg(source, nick+": Wurstball.de - " + url + "/");	
	} else {
		IRC.Privmsg(source, nick+": Wurstball.de - " + rnd + "/");
	}
}, "returns an wurstball page url for the given query");
