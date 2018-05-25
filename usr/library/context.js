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
		var args = {};
		var search = location.search.split("?");
		if (search.length > 1) {
			var searchs = search[1].split("&");
			for (var i = 0; i < searchs.length; i++) {
				var keys = searchs[i].split("=");
				args[keys[0]] = decodeURIComponent(keys[1]);
			}
		}

		if (typeof name == "object") {
			for (var k in name) {
				if (name[k] != undefined) {
					args[k] = name[k];
				}
			}
		} else if (value == undefined) {
			return args[name];
		} else {
			args[name] = value;
		}

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
							var msg = JSON.parse(xhr.responseText||'{"result":[]}');
							msg && console.log(msg)
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
		if (!("module" in form)) {
			form = form || {}
			form["module"] = this.Search("module")
		}

		var args = [];
		for (k in form) {
			args.push(k+"="+encodeURIComponent(form[k]));
		}

		var arg = args.join("&");
		console.log("POST: "+url+"?"+arg);
		xhr.send(arg);
	},
	Cap: function(cap, cb) {
		this.POST("", {ccc:"cache", name:cap}, function(msg) {
			typeof cb == "function" && cb(msg.result.join(""));
		});
	},
	Conf: function(conf, value, cb) {
		if (typeof value == "function") {
			cb = value;
			value = undefined;
		}

		var args = {ccc:"config", name:conf};
		if (value != undefined) {
			args.value = value
		}

		this.POST("", args, function(msg) {
			typeof cb == "function" && cb(msg.result.join(""));
		});
	},
	Cmd: function(cmd, value, cb) {
		if (typeof value == "function") {
			cb = value;
			value = undefined;
		}

		var args = {ccc:"command", name:cmd};
		if (value != undefined) {
			args.value = value
		}

		this.POST("", args, cb);
	},
	Module: function(module, domain) {
		this.Search({module:module, domain:domain})
	},
}

