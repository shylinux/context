var canvas = document.getElementById("heart");
var ctx = canvas.getContext('2d');

var main_angle = 30;
function refreshHeart() {
	ctx.clearRect(0,0,400,400)
	drawHeart(ctx,200,200,60,main_angle);
	main_angle += 10;
	for (var i = 0; i < 10; i++) {
		var x = Math.random() * 400;
		var y = Math.random() * 400;
		var scale = Math.random() * 20+10;
		var angle = Math.random() * 360;
		drawHeart(ctx,x,y,scale,angle);
	}
	setTimeout(refreshHeart, 200);
}
setTimeout(refreshHeart, 200);

function drawHeart(ctx,x,y,scale,angle, style, stroke) {//{{{
	ctx.save();
	ctx.translate(x,y);
	ctx.rotate(angle/180*Math.PI);
	ctx.scale(scale, scale);
	heartPath(ctx);
	ctx.shadowColor = "gray";
	ctx.shadowOffsetX = 5;
	ctx.shadowOffsetY = 5;
	ctx.shadowBlur = 5;
	if (stroke == "stroke") {
		ctx.strokeStyle = style||"red";
		ctx.stroke();
	} else {
		ctx.fillStyle = style||"red";
		ctx.fill();
	}
	ctx.restore();
}
//}}}
function heartPath(ctx) {//{{{
	ctx.beginPath();
	ctx.arc(-1,0,1,Math.PI,0,false);
	ctx.arc(1,0,1,Math.PI,0,false);
	ctx.bezierCurveTo(1.9, 1.2, 0.6, 1.6, 0, 3.0);
	ctx.bezierCurveTo( -0.6, 1.6,-1.9, 1.2,-2,0);
	ctx.closePath();
}
//}}}

var ctx0 = document.getElementById("demo0").getContext("2d");
ctx0.fillStyle = "green";
ctx0.fillRect(10,10,100,100);

var ctx2 = document.getElementById("demo2").getContext("2d");
ctx2.beginPath();
ctx2.moveTo(60,10);
ctx2.lineTo(10,110);
ctx2.lineTo(110,110);
ctx2.fill();


function draw3() {
	for (var i = 0; i < 120; i+=20) {
		for (var j = 0; j < 120; j+=20) {
			r = Math.random()*255;
			g = Math.random()*255;
			b = Math.random()*255;
			ctx3.fillStyle = "rgb("+r+","+g+","+b+")";
			ctx3.fillRect(i, j, 20, 20);
		}
	}
}
var demo3 = document.getElementById("demo3");
var ctx3 = demo3.getContext("2d");
demo3.onclick = draw3;
draw3()


var current_ctx = {//{{{
	big_scale: 1.25,
	small_scale: 0.8,
	font: '32px sans-serif',

	index_point: false,
	begin_point: null,
	end_point: null,

	config: {
		shape: {value: "rect", list: [
			{text: "心形", value: "heart"},
			{text: "圆形", value: "cycle"},
			{text: "矩形", value: "rect"},
			{text: "直线", value: "line"},
			{text: "文字", value: "text"},
		]},
		stroke: {value: "stroke", list: [
			{text: "画笔", value: "stroke"},
			{text: "画刷", value: "fill"},
		]},
		color: {value: "green", list: [
			{text: "黑色", value: "black"},
			{text: "红色", value: "red"},
			{text: "黄色", value: "yellow"},
			{text: "绿色", value: "green"},
			{text: "蓝色", value: "blue"},
			{text: "紫色", value: "purple"},
			{text: "白色", value: "white"},
		]},
		scale: {label: "比例", value: 1, width: 40},
		point: {label: "坐标", value: "0,0", width: 60},
		begin: {label: "隐藏", value: 0, width: 40},
		interval: {label: "间隔", value: 100, width: 40},
		json: {label: "数据", value: "", width: 100},
	},
	command: {
		cmd_shape: {
			heart: {text: "心形", key: "e",
				conf:{"shape": "heart", "stroke": "fill"},
				cmd:{"fill": "cmd_stroke"},
			},
			cycle: {text: "圆形", key: "c", conf:{"shape": "cycle"}},
			rect: {text: "矩形", key: "r", conf:{"shape": "rect"}},
			line: {text: "直线", key: "v",
				conf:{"shape": "line", "stroke": "stroke"},
				cmd:{"stroke": "cmd_stroke"},
			},
			text: {text: "文字", key: "t",
				conf:{"shape": "text", "stroke": "fill"},
				cmd:{"fill": "cmd_stroke"},
			},
		},
		cmd_stroke: {
			stroke: {text: "画笔", key: "s", conf:{"stroke": "stroke"}},
			fill: {text: "画刷", key: "f", conf:{"stroke": "fill"}},
		},
		cmd_color: {
			black: {text: "黑色", key: "", conf:{"color": "black"}},
			red: {text: "红色", key: "", conf:{"color": "red"}},
			yellow: {text: "黄色", key: "", conf:{"color": "yellow"}},
			green: {text: "绿色", key: "", conf:{"color": "green"}},
			purple: {text: "紫色", key: "", conf:{"color": "purple"}},
			blue: {text: "蓝色", key: "", conf:{"color": "blue"}},
			white: {text: "白色", key: "", conf:{"color": "white"}},
		},
		control: {
			big: {text: "放大", key: "b"},
			small: {text: "缩小", key: "m"},
			hide: {text: "隐藏"},
			draw: {text: "恢复"},
			play: {text: "播放", key: "a"},
			delete: {text: "删除", key: "d"},
			clear: {text: "清空", key: "q"},
			export: {text: "导出"},
			import: {text: "导入"},
		}
	},

	list: {
		style: ["red", "green", "yellow", "blue"],
		stroke: ["fill", "stroke"],
		shape: ["heart", "cycle", "rect", "line"],
	}
}
//}}}
function init(configs, commands) {//{{{
	for (var k in configs) {
		var config = configs[k];
		var cs = document.getElementsByClassName("config "+k);

		for (var i = 0; i < cs.length; i++) {
			if (config.list) {
				cs[i].innerHTML = "";
				for (var j in config.list) {
					var item = config.list[j];
					var option = cs[i].appendChild(document.createElement("option"));
					option.value = item.value
					option.text = item.text
					if (config.value == item.value) {
						cs[i].selectedIndex = j
					}
				}
				(function() {
					var key = k;
					cs[i].onchange = function(event) {
						conf('config', key, event);
					}
				})();
			} else {
				cs[i].value = config.value;
				(function() {
					var key = k;
					if (config.label) {
						var label = cs[i].parentElement.insertBefore(document.createElement("label"), cs[i]);
						label.appendChild(document.createTextNode(config.label+": "));
					}
					if (config.width) {
						cs[i].style.width = config.width+"px";
					}
					cs[i].onkeyup = cs[i].onblur = function(event) {
						conf('config', key, event);
					}
					conf('config', key, config.value);
				})();
			}
		}
	}

	for (var group in commands) {
		for (var which in commands[group]) {
			var command = commands[group][which];
			var cs = document.getElementsByClassName(group+" "+which);
			for (var i = 0; i < cs.length; i++) {
				if (command.key) {
					cs[i].innerText = command.text+"("+command.key+")";
				} else {
					cs[i].innerText = command.text
				}
				(function() {
					var key = which;
					var key1 = group;
					cs[i].onclick = function(event) {
						action(event, key, key1);
					}
				})();
			}
		}
	}
}//}}}
function conf(group, which, value) {//{{{
	if (value instanceof Event) {
		var event = value;
		switch (event.type) {
			case "change":
				current_ctx[group][which].value = event.target[event.target.selectedIndex].value;
				break
			case "keyup":
				switch (event.key) {
					case "Enter":
						current_ctx[group][which].value = event.target.value;
						break
					case "Escape":
						event.target.value = current_ctx[group][which].value;
						break
				}
				break
			case "blur":
				if (current_ctx[group][which].value == event.target.value) {
					break
				}
				if (confirm("save value?")) {
					current_ctx[group][which].value = event.target.value;
				} else {
					event.target.value = current_ctx[group][which].value;
				}
				break
		}
		console.log("conf");
		console.log(event);
		return
	}

	var config = current_ctx[group][which];
	if (value != undefined) {
		var cs = document.getElementsByClassName(group+" "+which);
		for (var i = 0; i < cs.length; i++) {
			config.value = value;
			if (cs[i].nodeName == "LABEL") {
				cs[i].innerText = value;
			} else {
				cs[i].value = value;
			}
		}
	}
	return config.value
}//}}}
init(current_ctx.config, current_ctx.command);

var control_map = {//{{{
	s: "stroke",
	f: "fill",

	e: "heart",
	c: "cycle",
	r: "rect",
	v: "line",
	t: "text",

	d: "delete",

	b: "big",
	m: "small",
	a: "play",

	Escape: "escape",
}
//}}}
function control(event) {//{{{
	if (event.type == "keyup") {
		action(event, control_map[event.key]);
	}
}
//}}}
function action(event, which, group) {//{{{
	console.log(event);
	switch (which) {
		case "big":
			conf("config", "scale", (conf("config", "scale")*current_ctx.big_scale).toFixed(3));
			draw.scale(current_ctx.big_scale, current_ctx.big_scale);
			refresh();
			break
		case "small":
			conf("config", "scale", (conf("config", "scale")*current_ctx.small_scale).toFixed(3));
			draw.scale(current_ctx.small_scale, current_ctx.small_scale);
			refresh();
			break
		case "escape":
			current_ctx.begin_point = null;
			current_ctx.end_point = null;
			refresh();
			break
		case "clear":
			if (confirm("clear all?")) {
				his.innerHTML = "";
				draw_history = [];
				refresh();
			}
			break
		case "hide":
			conf("config", "begin", draw_history.length);
			refresh();
			break
		case "delete":
			draw_history.pop();
			refresh();
			break
		case "draw":
			conf("config", "begin", 0);
			refresh();
			break
		case "play":
			conf("config", "begin", 0);
			refresh(conf("config", "interval"), 0);
			break
		case "export":
			conf("config", "json", JSON.stringify(draw_history));
			break
		case "import":
			draw_history = JSON.parse(conf("config", "json"));
			refresh();
			break
		default:
			if (!which || !group) {
				break
			}
			var cs = document.getElementsByClassName(group);
			for (var i = 0; i < cs.length; i++) {
				cs[i].style.backgroundColor = "white";
			}

			var cs = document.getElementsByClassName(group+" "+which);
			for (var i = 0; i < cs.length; i++) {
				cs[i].style.backgroundColor = "lightblue";
			}

			var command = current_ctx.command[group][which];
			for (var k in command.conf) {
				conf("config", k, command.conf[k]);
			}
			for (var k in command.cmd) {
				action(event, k, command.cmd[k])
			}
	}
}
//}}}
action(null, "heart", "cmd_shape");
action(null, "red", "cmd_color");

function trans(point) {//{{{
	point.x /= conf("config", "scale");
	point.y /= conf("config", "scale");
	return point;
}
//}}}
function show_debug(log, clear) {//{{{
	var fuck = document.getElementById("fuck");
	// if (clear) {
	// 	fuck.innerHTML = "";
	// }
	var div = fuck.appendChild(document.createElement("div"));
	div.appendChild(document.createTextNode(log));
}
//}}}
function draw_point(event) {//{{{
	console.log("point");
	console.log(event);
	show_debug(event.type)
	var point = trans({x:event.offsetX, y:event.offsetY});
	if (event.type == "touchstart") {
		var point = trans({x:event.touches[0].clientX, y:event.touches[0].clientY});
	}

	if (current_ctx.index_point) {
		draw.beginPath();
		draw.arc(point.x, point.y, 5, 0, 2*Math.PI);
		draw.fill();
	}

	if (!current_ctx.begin_point) {
		current_ctx.begin_point = point;
		return
	}
	current_ctx.end_point = point;
	console.log(current_ctx.end_point);
	var text = "";
	if (conf("config", "shape") == "text") {
		text = prompt("请入文字", "");
	}

	draws(draw, conf("config", "color"), conf("config", "stroke"), conf("config", "shape"), current_ctx.begin_point, current_ctx.end_point, text);
	draw_history.push({style:conf("config", "color"), stroke:conf("config", "stroke"), shape:conf("config", "shape"), begin_point:current_ctx.begin_point, end_point:current_ctx.end_point, text:text})

	var headers = ["style", "stroke", "shape", "x1", "y1", "x2", "y2", "text"]
	if (his.rows.length == 0) {
		var tr = his.insertRow(-1);
		for (var i in headers) {
			var th = tr.appendChild(document.createElement("th"));
			th.appendChild(document.createTextNode(headers[i]))
		}
	}

	var tr = his.insertRow(-1);
	var fields = [conf("config", "color"), conf("config", "stroke"), conf("config", "shape"),
		parseInt(current_ctx.begin_point.x), parseInt(current_ctx.begin_point.y),
		parseInt(current_ctx.end_point.x), parseInt(current_ctx.end_point.y), text]

	for (var i in fields) {
		var td = tr.appendChild(document.createElement("td"));
		switch (headers[i]) {
			case "shape":
			case "style":
			case "stroke":
				var select = td.appendChild(document.createElement("select"));
				(function() {
					var index = [headers[i]];
					select.onchange = function(event) {
						draw_history[tr.rowIndex-1][index] = event.target[event.target.selectedIndex].text;
						refresh();
					}
				})();
				var shapes = current_ctx.list[headers[i]];
				for (var j in shapes) {
					var option = select.appendChild(document.createElement("option"));
					option.text = shapes[j];
					if (option.text == fields[i]) {
						select.selectedIndex = j;
					}
				}
				break
			default:
				var input = td.appendChild(document.createElement("input"));
				input.value = fields[i];
				input.style.width="40px";
				input.dataset.row = tr.rowIndex-1
				input.dataset.col = headers[i]
				input.onkeyup = modify
		}

	}

	current_ctx.begin_point = null;
	current_ctx.end_point = null;
}
//}}}
function draw_move(event) {//{{{
	var point = trans({x:event.offsetX, y:event.offsetY});
	conf("config", "point", parseInt(point.x)+","+parseInt(point.y));

	if (current_ctx.begin_point) {
		refresh();
		draws(draw, conf("config", "color"), conf("config", "stroke"), conf("config", "shape"), current_ctx.begin_point, point, "");
	}
}
//}}}

var draw_history = [];
var his = document.getElementById("draw_history");
function modify(event, row, col) {//{{{
	if (event.key != "Enter") {
		return
	}
	console.log("modify");
	console.log(event);
	console.log();
	var data = event.target.dataset;
	var row = draw_history[data.row];
	var value = event.target.value;
	switch (data.col) {
		case "x1":
			row.begin_point.x = value;
			break
		case "y1":
			row.begin_point.y = value;
			break
		case "x2":
			row.end_point.x = value;
			break
		case "y2":
			row.end_point.y = value;
			break
		default:
			row[data.col] = value;
	}

	refresh();
}
//}}}
function refresh(time, i) {//{{{
	if (time) {
		if (0 == i) {
			draw.clearRect(0, 0,400/conf("config", "scale"), 400/conf("config", "scale"));
		}

		if (conf("config", "begin") <= i && i < draw_history.length) {
			var h = draw_history[i];
			draws(draw, h.style, h.stroke, h.shape, h.begin_point, h.end_point);
			i++;
			setTimeout(function(){
				refresh(time, i);
			}, time);
		}
		return
	}

	draw.clearRect(0, 0,400/conf("config", "scale"), 400/conf("config", "scale"));

	for (var i in draw_history) {
		if (i < conf("config", "begin")) {
			continue
		}
		var h = draw_history[i];
		draws(draw, h.style, h.stroke, h.shape, h.begin_point, h.end_point, h.text);
	}
}
//}}}

var item = document.getElementById("draw");
var draw = item.getContext("2d");
function draws(draw, style, stroke, shape, begin_point, end_point, text) {//{{{
	draw.save();
	begin_x = begin_point.x;
	begin_y = begin_point.y;
	end_x = end_point.x;
	end_y = end_point.y;

	if (current_ctx.index_point) {
	draw.beginPath();
	draw.arc(begin_x, begin_y, 5, 0, 2*Math.PI)
	draw.fill()
	}

	if (current_ctx.index_point) {
	draw.beginPath();
	draw.arc(end_x, end_y, 5, 0, 2*Math.PI)
	draw.fill()
	}

	switch (shape) {
		case 'heart':
			r = Math.sqrt(Math.pow(begin_x-end_x, 2)+Math.pow(begin_y-end_y,2));
			a = Math.atan((end_y-begin_y)/(end_x-begin_x))/Math.PI*180;
			drawHeart(draw, begin_x, begin_y, r, a, style, stroke)
			break
		case 'cycle':
			draw.beginPath();
			r = Math.sqrt(Math.pow(begin_x-end_x, 2)+Math.pow(begin_y-end_y,2));
			draw.arc(begin_x, begin_y, r, 0, 2*Math.PI)
			if (stroke == "stroke") {
				if (style) {
					draw.strokeStyle = style;
				}
				draw.stroke()
			} else {
				if (style) {
					draw.fillStyle = style;
				}
				draw.fill()
			}
			break
		case 'line':
			draw.beginPath();
			draw.moveTo(begin_x, begin_y);
			draw.lineTo(end_x, end_y);
			if (stroke == "stroke") {
				if (style) {
					draw.strokeStyle = style;
				}
				draw.stroke()
			} else {
				if (style) {
					draw.fillStyle = style;
				}
				draw.fill()
			}
			break
		case 'rect':
			if (stroke == "stroke") {
				if (style) {
					draw.strokeStyle = style;
				}
				draw.strokeRect(begin_x, begin_y, end_x-begin_x, end_y-begin_y);
			} else {
				if (style) {
					draw.fillStyle = style;
				}
				draw.fillRect(begin_x, begin_y, end_x-begin_x, end_y-begin_y);
			}
			break
		case 'text':
			if (stroke == "stroke") {
				if (style) {
					draw.strokeStyle = style;
				}
				draw.font = current_ctx.font;
				draw.strokeText(text, begin_x, begin_y, end_x-begin_x);
			} else {
				if (style) {
					draw.fillStyle = style;
				}
				draw.font = current_ctx.font;
				draw.fillText(text, begin_x, begin_y, end_x-begin_x);
			}
	}

	draw.restore();
}
//}}}

