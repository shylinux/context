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
        // arg && (url += ((url.indexOf("?")>-1)? "&": "?") + arg)

        var xhr = new XMLHttpRequest();
        // xhr.open("POST", url);
        xhr.open("POST", url);
        xhr.setRequestHeader("Content-Type", "application/x-www-form-urlencoded")
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

            if (msg.download_file) {
                window.open(msg.download_file.join(""))
            } else if (msg.page_redirect) {
                location.href = msg.page_redirect.join("")
            } else if (msg.page_refresh) {
                location.reload()
            }
            typeof cb == "function" && cb(msg)
        }
        xhr.send(arg);
    },
}

context.isMobile = navigator.userAgent.indexOf("Mobile") > -1
context.scroll_by = window.innerHeight/2

function right(arg) {
    if (arg == "true") {
        return true
    }
    if (arg == "false") {
        return false
    }
    if (arg) {
        return true
    }
    return false
}
function format_date(arg) {
    var date = arg.getDate()
    if (date < 10) {
        date = "0"+date
    }
    var month = arg.getMonth()+1
    if (month < 10) {
        month = "0"+month
    }
    var hour = arg.getHours()
    if (hour < 10) {
        hour = "0"+hour
    }
    var minute = arg.getMinutes()
    if (minute < 10) {
        minute = "0"+minute
    }
    var second = arg.getSeconds()
    if (second < 10) {
        second = "0"+second
    }
    return arg.getFullYear()+"-"+month+"-"+date+" "+hour+":"+minute+":"+second
}

function modify_node(which, html) {
    var node = which
    if (typeof which == "string") {
        node = document.querySelector(which)
    }

    html && typeof html == "string" && (node.innerHTML = html)
    if (html && typeof html == "object") {
        for (var k in html) {
            if (typeof html[k] == "object") {
                for (var d in html[k]) {
                    node[k][d] = html[k][d]
                }
                continue
            }
            node[k] = html[k]
        }
    }
    return node
}
function create_node(element, html) {
    var node = document.createElement(element)
    return modify_node(node, html)
}

function insert_child(parent, element, html, position) {
    var elm = create_node(element, html)
    return parent.insertBefore(elm, position || parent.firstElementChild)
}
function append_child(parent, element, html) {
    var elm = create_node(element, html)
    parent.append(elm)
    return elm
}
function insert_before(self, element, html) {
    var elm = create_node(element, html)
    return self.parentElement.insertBefore(elm, self)
}
function insert_button(which, value, callback) {
    insert_before(which, "input", {
        "type": "button", "value": value, "onclick": callback,
    })
}

function sort_table(table, index, sort_asc) {
    var list = table.querySelectorAll("tr")
    var new_list = []

    var is_time = true
    var is_number = true
    for (var i = 1; i < list.length; i++) {
        var value = Date.parse(list[i].childNodes[index].innerText)
        if (!(value > 0)) {
            is_time = false
        }

        var value = parseInt(list[i].childNodes[index].innerText)
        if (!(value >= 0 || value <= 0)) {
            is_number = false
        }

        new_list.push(list[i])
    }

    var sort_order = ""
    if (is_time) {
        if (sort_asc) {
            method = function(a, b) {return Date.parse(a) > Date.parse(b)}
            sort_order = "time"
        } else {
            method = function(a, b) {return Date.parse(a) < Date.parse(b)}
            sort_order = "time_r"
        }
    } else if (is_number) {
        if (sort_asc) {
            method = function(a, b) {return parseInt(a) > parseInt(b)}
            sort_order = "int"
        } else {
            method = function(a, b) {return parseInt(a) < parseInt(b)}
            sort_order = "int_r"
        }
    } else {
        if (sort_asc) {
            method = function(a, b) {return a > b}
            sort_order = "str"
        } else {
            method = function(a, b) {return a < b}
            sort_order = "str_r"
        }
    }

    list = new_list
    new_list = []
    for (var i = 0; i < list.length; i++) {
        list[i].parentElement && list[i].parentElement.removeChild(list[i])
        for (var j = i+1; j < list.length; j++) {
            if (typeof method == "function" && method(list[i].childNodes[index].innerText, list[j].childNodes[index].innerText)) {
                var temp = list[i]
                list[i] = list[j]
                list[j] = temp
            }
        }
        new_list.push(list[i])
    }

    for (var i = 0; i < new_list.length; i++) {
        table.appendChild(new_list[i])
    }
    return sort_order
}
function add_sort(append, field, cb) {
    append.onclick = function(event) {
        var target = event.target
        var dataset = target.dataset
        var nodes = target.parentElement.childNodes
        for (var i = 0; i < nodes.length; i++) {
            if (nodes[i] == target) {
                if (target.tagName == "TH") {
                    dataset["sort_asc"] = (dataset["sort_asc"] == "1") ? 0: 1
                    sort_table(append, i, dataset["sort_asc"] == "1")
                } else if (target.tagName == "TD") {
                    var tr = target.parentElement.parentElement.querySelector("tr")
                    if (tr.childNodes[i].innerText.startsWith(field)) {
                        typeof cb == "function" && cb(event)
                    }
                }
            }
        }
    }
}
function scroll_page(event, page) {
    var body = document.querySelector("body")

    switch (event.key) {
        case "h":
            if (event.ctrlKey) {
                window.scrollBy(-page.scroll_x*10, 0)
            } else {
                window.scrollBy(-page.scroll_x, 0)
            }
            break
        case "H":
            window.scrollBy(-body.scrollWidth, 0)
            break
        case "l":
            if (event.ctrlKey) {
                window.scrollBy(page.scroll_x*10, 0)
            } else {
                window.scrollBy(page.scroll_x, 0)
            }
            break
        case "L":
            window.scrollBy(body.scrollWidth, 0)
            break
        case "j":
            if (event.ctrlKey) {
                window.scrollBy(0, page.scroll_y*10)
            } else {
                window.scrollBy(0, page.scroll_y)
            }
            break
        case "J":
            window.scrollBy(0, body.scrollHeight)
            break
        case "k":
            if (event.ctrlKey) {
                window.scrollBy(0, -page.scroll_y*10)
            } else {
                window.scrollBy(0, -page.scroll_y)
            }
            break
        case "K":
            window.scrollBy(0, -body.scrollHeight)
            break
    }
    return
}

