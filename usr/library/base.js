
function update(event, module, details, key) {
	if (event) {
		window[key+"timer"] = !window[key+"timer"];
	}
	if (!window[key+"timer"]) {
		return
	}
	console.log("update "+key)
	setTimeout(function() {
		action(event, module, details, key);
		update(null, module, details, key);
	}, refresh_time);
}

function input(event, module, details, key) {
	if (event.code == "Enter") {
		action(event, module, details, key);
	}
}
function action(event, module, details, key) {
	var input = document.getElementsByClassName(key+"_input");
	for (var i = 0; i < input.length; i++ ){
		if (input[i].value != "") {
			details.push(input[i].value)
		}
	}
	ctx.POST("", {module:module, details:details}, function(msg) {
		if (msg && msg.result) {
			var result = document.getElementsByClassName(key+"_result")[0];
			result.innerHTML = msg.result;
		}
		if (!msg || !msg.append || msg.append.length < 1) {
			return
		}

		var append = document.getElementsByClassName(key+"_append")[0];
		if (append.rows.length == msg[msg.append[0]].length+1) {
			return
		}
		append.innerHTML = '';
		var tr = append.insertRow(0);
		for (var i in msg.append) {
			var th = tr.appendChild(document.createElement("th"));
			th.appendChild(document.createTextNode(msg.append[i]));
		}

		for (var i = 0; i < msg[msg.append[0]].length; i++) {
			var tr = append.insertRow(1);
			for (var j in msg.append) {
				var td = tr.appendChild(document.createElement("td"));
				td.appendChild(document.createTextNode(msg[msg.append[j]][i]));
			}
		}
		var div = append.parent;

	})
	return false;
}
