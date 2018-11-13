context = {
	Search: function(key, value) {
		var args = {};
		var search = location.search.split("?");
		if (search.length > 1) {
			var searchs = search[1].split("&");
			for (var i = 0; i < searchs.length; i++) {
				var keys = searchs[i].split("=");
				args[keys[0]] = decodeURIComponent(keys[1]);
			}
		}

        if (key == undefined) {
            return args
        } else if (typeof key == "object") {
			for (var k in key) {
				if (key[k] != undefined) {
					args[k] = key[k];
				}
			}
		} else if (value == undefined) {
			return args[key] || this.Cookie(key);
		} else {
			args[key] = value;
		}

		var arg = [];
		for (var k in args) {
			arg.push(k+"="+encodeURIComponent(args[k]));
		}
		location.search = arg.join("&");
	},
	Cookie: function(key, value) {
        if (key == undefined) {
            cs = {}
            cookies = document.cookie.split("; ")
            for (var i = 0; i < cookies.length; i++) {
                cookie = cookies[i].split("=")
                cs[cookie[0]] = cookie[1]
            }
            return cs
        }
        if (typeof key == "object") {
            for (var k in key) {
                document.cookie = k+"="+key[k];
            }
            return this.Cookie()
        }

		if (value == undefined) {
			var pattern = new RegExp(key+"=([^;]*);?");
			var result = pattern.exec(document.cookie);
			if (result && result.length > 0) {
				return result[1];
			}
			return "";
		}

		document.cookie = key+"="+value+";path=/";
		return this.Cookie(key);
	},
    Cache: function(key, cb, sync) {
        if (key == undefined) {
            return this.cache
        }
        if (this.cache && !sync) {
            typeof cb == "function" && cb(this.cache[key])
            return this.cache[key]
        }

        var that = this
        this.GET("", {"componet_group": "login", "componet_order": "userinfo"}, function(msg) {
            msg = msg[0]
            that.cache = {}
            for (var i = 0; i < msg.append.length; i++) {
                that.cache[msg.append[i]] = msg[msg.append[i]].join("")
            }
            typeof cb == "function" && cb(that.cache[key])
        })
	},
    GET: function(url, form, cb) {
        form = form || {}

        var args = [];
        for (var k in form) {
            if (form[k] instanceof Array) {
                for (i in form[k]) {
                    args.push(k+"="+encodeURIComponent(form[k][i]));
                }
            } else if (form[k] != undefined) {
                args.push(k+"="+encodeURIComponent(form[k]));
            }
        }

        var arg = args.join("&");
        arg && (url += ((url.indexOf("?")>-1)? "&": "?") + arg)
        console.log("GET: "+url);

        var xhr = new XMLHttpRequest();
        xhr.open("GET", url);
        xhr.setRequestHeader("Accept", "application/json")

        xhr.onreadystatechange = function() {
            if (xhr.readyState != 4) {
                return
            }
            if (xhr.status != 200) {
                return
            }

            try {
                var msg = JSON.parse(xhr.responseText||'{"result":[]}')
            } catch (e) {
                var msg = {"result": [xhr.responseText]}
            }

            console.log(msg)
            msg.result && console.log(msg.result.join(""))
            if (msg.download_file) {
                window.open(msg.download_file.join(""))
            } else if (msg.page_redirect) {
                location.href = msg.page_redirect.join("")
            } else if (msg.page_refresh) {
                location.reload()
            }
            typeof cb == "function" && cb(msg)
        }
        xhr.send();
    },
}

context.isMobile = navigator.userAgent.indexOf("Mobile") > -1

function insert_child(parent, element, html, position) {
    var elm = document.createElement(element)
    html && (elm.innerHTML = html)
    return parent.insertBefore(elm, position || parent.firstElementChild)
}
function append_child(parent, element, html) {
    var elm = document.createElement(element)
    html && typeof html == "string" && (elm.innerHTML = html)
    if (typeof html == "object") {
        for (var k in html) {
            if (typeof html[k] == "object") {
                for (var d in html[k]) {
                    elm[k][d] = html[k][d]
                }
                continue
            }

            elm[k] = html[k]
        }
    }
    parent.append(elm)
    return elm
}
function insert_before(self, element, html) {
    var elm = document.createElement(element)
    html && typeof html == "string" && (elm.innerHTML = html)
    if (typeof html == "object") {
        for (var k in html) {
            elm[k] = html[k]
        }
    }
    return self.parentElement.insertBefore(elm, self)
}

