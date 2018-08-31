ctx = {
	Cookie: function(key, value) {//{{{
		if (value == undefined) {
			var pattern = new RegExp(key+"=([^;]*);?");
			var result = pattern.exec(document.cookie);
			if (result && result.length > 0) {
				return result[1];
			}
			return "";
		}

		document.cookie = key+"="+value;
		return this.Cookie(key);
	},//}}}
	Search: function(key, value) {//{{{
		var args = {};
		var search = location.search.split("?");
		if (search.length > 1) {
			var searchs = search[1].split("&");
			for (var i = 0; i < searchs.length; i++) {
				var keys = searchs[i].split("=");
				args[keys[0]] = decodeURIComponent(keys[1]);
			}
		}

		if (typeof key == "object") {
			for (var k in key) {
				if (key[k] != undefined) {
					args[k] = key[k];
				}
			}
		} else if (value == undefined) {
			return args[key] || "";
		} else {
			args[key] = value;
		}

		var arg = [];
		for (var k in args) {
			arg.push(k+"="+encodeURIComponent(args[k]));
		}
		location.search = arg.join("&");
	},//}}}
	GET: function(url, form, cb) {//{{{
		var xhr = new XMLHttpRequest();
		xhr.onreadystatechange = function() {
			switch (xhr.readyState) {
				case 4:
					switch (xhr.status) {
						case 200:
							try {
								var msg = JSON.parse(xhr.responseText||'{"result":[]}');
							} catch (e) {
								msg = {"result": [xhr.responseText]}
							}

							msg && console.log(msg)
							msg.result && console.log(msg.result.join(""));
							typeof cb == "function" && cb(msg)
					}
					break;
			}
		}

		form = form || {}
		form["dir"] = form["dir"] || this.Search("dir") || undefined
		form["module"] = form["module"] || this.Search("module") || undefined
		form["domain"] = form["domain"] || this.Search("domain") || undefined

		var args = [];
		for (k in form) {
			if (form[k] instanceof Array) {
				for (i in form[k]) {
					args.push(k+"="+encodeURIComponent(form[k][i]));
				}
			} else if (form[k] != undefined) {
				args.push(k+"="+encodeURIComponent(form[k]));
			}
		}

		var arg = args.join("&");
        if (arg) {
            url += "?"+arg
        }

		xhr.open("GET", url);
		console.log("GET: "+url+"?"+arg);
		xhr.send();
	},//}}}
	POST: function(url, form, cb) {//{{{
		var xhr = new XMLHttpRequest();
		xhr.onreadystatechange = function() {
			switch (xhr.readyState) {
				case 4:
					switch (xhr.status) {
						case 200:
							try {
								var msg = JSON.parse(xhr.responseText||'{"result":[]}');
							} catch (e) {
								msg = {"result": [xhr.responseText]}
							}

							msg && console.log(msg)
							msg.result && console.log(msg.result.join(""));
							typeof cb == "function" && cb(msg)
					}
					break;
			}
		}

		xhr.open("POST", url);
		xhr.setRequestHeader("Content-Type", "application/x-www-form-urlencoded");


		form = form || {}
		form["dir"] = form["dir"] || this.Search("dir") || undefined
		form["module"] = form["module"] || this.Search("module") || undefined
		form["domain"] = form["domain"] || this.Search("domain") || undefined

		var args = [];
		for (k in form) {
			if (form[k] instanceof Array) {
				for (i in form[k]) {
					args.push(k+"="+encodeURIComponent(form[k][i]));
				}
			} else if (form[k] != undefined) {
				args.push(k+"="+encodeURIComponent(form[k]));
			}
		}

		var arg = args.join("&");
		console.log("POST: "+url+"?"+arg);
		xhr.send(arg);
	},//}}}
	Refresh: function() {//{{{
		location.assign(location.href);
	},//}}}

	Cap: function(cap, cb) {//{{{
		if (typeof cap == "function") {
			cb = cap;
			cap = undefined;
		}

		var args = {ccc:"cache"};
		if (cap != undefined) {
			args.name = cap;
		}

		this.POST("", args, function(msg) {
			var value = msg.result.join("");
			typeof cb == "function" && cb(value);
		});
	},//}}}
	Conf: function(name, value, cb) {//{{{
		if (typeof name == "function") {
			value = name;
			name = undefined;
		}
		if (typeof value == "function") {
			cb = value;
			value = undefined;
		}

		var args = {ccc:"config"};
		if (name != undefined) {
			args.name = name
		}
		if (value != undefined) {
			args.value = value
		}

		this.POST("", args, function(msg) {
			var value = msg.result.join("");
			typeof cb == "function" && cb(value);
		});
	},//}}}
	Cmd: function(cmd, value, cb) {//{{{
		if (typeof cmd == "function") {
			value = cmd;
			cmd = undefined;
		}
		if (typeof value == "function") {
			cb = value;
			value = undefined;
		}

		var args = {ccc:"command"};
		if (cmd != undefined) {
			args.name = cmd
		}
		if (value != undefined) {
			args.value = JSON.stringify(value)
		}

		this.POST("", args, cb);
	},//}}}
	Module: function(module, domain) {//{{{
		this.Search({module:module, domain:domain})
	},//}}}
};

