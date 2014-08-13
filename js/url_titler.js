RegisterCallback("PRIVMSG", "Url Titler", function() {
	var url = UTILS.ExtractURL(this.event.message),
		target = this.event.args[0],
		nick = this.event.nick,
		me = IRC.GetNick();

	if(nick != me && url) {
		var title = UTILS.ExtractTitle(url);

		if(title) {
			// Don't bother for images
			if(title.indexOf("[image/gif]") > -1 || title.indexOf("[image/jpeg]") > -1 || title.indexOf("[image/png]") > -1) {
				return
			}

			IRC.Privmsg(target, "[Link] " + title)
		}
	}
});