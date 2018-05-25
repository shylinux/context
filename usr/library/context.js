ctx = {
	Cookie: function(name, value) {
		if (value == undefined) {
			var pattern = new RegExp(name+"=([^;]*);?");
			var result = pattern.exec(document.cookie);
			if (result && result.length > 0) {
				return result[1];
			}
			return "";
		}

		document.cookie = name+"="+value;
		return this.Cookie(name);
	},
	Search: function(name, value) {
		if (value == undefined) {
			var pattern = new RegExp(name+"=([^&#]*)");
			var result = pattern.exec(location.search);
			if (result && result.length > 0) {
				return result[1];
			}
			return "";
		}

		var args = {};
		var search = location.search.split("?");
		if (search.length > 1) {
			var searchs = search[1].split("&");
			for (var i = 0; i < searchs.length; i++) {
				var keys = searchs[i].split("=");
				args[keys[0]] = decodeURIComponent(keys[1]);
			}
		}
		args[name] = value;

		var arg = [];
		for (var k in args) {
			arg.push(k+"="+encodeURIComponent(args[k]));
		}
		location.search = arg.join("&");
	},
	POST: function(url, form, cb) {
		var xhr = new XMLHttpRequest();
		xhr.onreadystatechange = function() {
			switch (xhr.readyState) {
				case 4:
					switch (xhr.status) {
						case 200:
							var msg = JSON.parse(xhr.responseText);
							console.log(msg)
							typeof cb == "function" && cb(msg)
					}
					break;
			}
		}

		xhr.open("POST", url);
		xhr.setRequestHeader("Content-Type", "application/x-www-form-urlencoded");

		if (!("dir" in form)) {
			form = form || {}
			form["dir"] = this.Search("dir")
		}

		var args = [];
		for (k in form) {
			args.push(k+"="+encodeURIComponent(form[k]));
		}

		var arg = args.join("&");
		console.log(url)
		console.log(arg)
		xhr.send(arg);
	},
}

