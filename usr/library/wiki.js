function jumpto(url) {
	if (url == "..") {
		var ps = locaiton.href.split();
	}
	location.href=url;
}

function keyup(event) {
	console.log(event);
	if (typeof window.control == "function") {
		control(event);
	}
	if (event.key == "z") {
		var input = document.getElementsByClassName("query_input")[0];
		if (!window.query_show) {
			window.query_show = true;
			var query = document.getElementsByClassName("query_menu")[0];
			var input = query.getElementsByTagName("input")[0];
			input.style.visibility = "visible";
			input.style.width = "80px";
			input.focus();
		}
	return true
	}
	return true
}

document.onkeyup = keyup;
function toggle(side) {
	if (side == "left") {
		window.left_list_hide = !window.left_list_hide;
		var list = document.getElementsByClassName("list")[0];
		var content = document.getElementsByClassName("content")[0];
		if (left_list_hide) {
			list.style.visibility = "hidden";
			list.style.width="0px";
			list.style.height="0px";
			list.style["min-width"]="0px";
			content.style.width="100%";
		} else {
			list.style.visibility = "visible";
			list.style.width="15%";
			list.style.height="100%";
			list.style["min-width"]="180px";
			content.style.width="85%";
		}
	}
}

function menu() {
	var max = 0;
	var min = 1000;
	var list = [];
	var hs = ["h2", "h3"];
	for (var i = 0; i < hs.length; i++) {
		var head = document.getElementsByTagName(hs[i]);
		for (var j = 0; j < head.length; j++) {
			head[j].id = "head"+head[j].offsetTop;
			head[j].onclick = function(event) {
			}
			list.push({"level": hs[i], "position": head[j].offsetTop, "title": head[j].innerText, "hash": head[j].id})
			if (head[j].offsetTop > max) {
				max = head[j].offsetTop;
			}
			if (head[j].offsetTop < min) {
				min = head[j].offsetTop;
			}
		}
	}
	max = max - min;

	var link = [];
	var a = document.getElementsByTagName("a");
	for (var i = 0; i < a.length; i++) {
		link.push({href: a[i].href, title: a[i].innerText});
	}

	for (var i = 0; i < list.length-1; i++) {
		for (var j = i+1; j < list.length; j++) {
			if (list[j].position < list[i].position) {
				var a = list[i];
				list[i] = list[j];
				list[j] = a;
			}
		}
	}

	var index2 = -1;
	var index3 = 0;
	for (var i = 0; i < list.length; i++) {
		if (list[i].level == "h2") {
			index2++;
			index3=0;
		} else {
			index3++;
			list[i].index3 = index3;
		}
		list[i].index2 = index2;
	}

	var m = document.getElementsByClassName("menu");
	for (var i = 0; i < m.length; i++) {
		for (var j = 0; j < list.length; j++) {
			var text = list[j].index2+"."
			if (list[j].level == "h3") {
				text += list[j].index3
			}
			text += " "
			text += list[j].title;

			var h = document.getElementById(list[j].hash)
			h.innerText = text

			var one = m[i].appendChild(document.createElement("div"));
			var a = one.appendChild(document.createElement("a"));
			a.href = "#"+list[j].hash;
			a.innerText = text+" ("+parseInt((list[j].position-min)/max*100)+"%)";

			one.className = list[j].level;
		}
	}

	var m = document.getElementsByClassName("link");
	for (var i = 0; i < m.length; i++) {
		var one = m[i].appendChild(document.createElement("div"));
		var a = one.appendChild(document.createTextNode("相关链接: "));

		for (var j = 0; j < link.length; j++) {
			var one = m[i].appendChild(document.createElement("div"));
			var a = one.appendChild(document.createTextNode(link[j].title+": "));
			var a = one.appendChild(document.createElement("a"));
			a.href = link[j].href
			a.innerText = a.href
		}
	}

	var m = document.getElementsByTagName("pre");
	for (var i = 0; i < m.length; i++) {
		var line = (m[i].clientHeight-10)/15
		// if (line < 3) {
		// 	continue
		// }
		console.log(m[i].clientHeight)
		var nu = m[i].parentElement.insertBefore(document.createElement("div"), m[i]);
		nu.className = "number1"

		for (var j = 1; j <= line; j++) {
			console.log(j)
			var li = nu.appendChild(document.createElement("div"));
			li.appendChild(document.createTextNode(""+j));
		}
	}
}

function query(event) {
	if (event) {
		if (event.code == "Enter") {
			jumpto("/wiki/?query="+encodeURIComponent(event.target.value));
		}
		console.log("what")
		return true
	}
	window.query_show = !window.query_show;
	var query = document.getElementsByClassName("query_menu")[0];
	var input = query.getElementsByTagName("input")[0];
	if (window.query_show) {
		input.style.visibility = "visible";
		input.style.width = "80px";

	} else {
		input.style.visibility = "hidden";
		input.style.width = "0px";
	}
}

var tags_list = {};
ctx.GET("/wiki/define.json", undefined, function(msg){
	tags_list = msg["define"];
})

function tags(event) {
	console.log(event);

    if (event.srcElement.tagName == "CODE") {
        var tag = document.getSelection().toString();
        console.log(tag);
        if (tag && tag.length > 0 && tags_list[tag]) {
			var position = tags_list[tag].position;
			if (position.length == 1) {
				jumpto("/wiki/src/"+position[0].file+"#hash_"+position[0].line);
			} else {
				jumpto("/wiki/?query="+encodeURIComponent(tag));
			}
        }
    }
}

document.onmouseup = tags;
window.onload = function() {
	toggle();
	if (location.href.endsWith(".md")) {
		menu();
	}
}
