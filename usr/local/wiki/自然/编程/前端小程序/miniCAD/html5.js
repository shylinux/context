var canvas = document.getElementById("heart");//{{{
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
//}}}

var draw_history = [{shape:"hide"}];
var his = document.getElementById("draw_history");
var draw = document.getElementById("draw").getContext("2d");

var current_ctx = {//{{{
	agent: {},
	big_scale: 1.25,
	small_scale: 0.8,
	font: '32px sans-serif',

	index_point: false,
	begin_point: null,
	end_point: null,
	last_point: null,
	last_move: 0,

	config: {
		shape: {value: "rect", list: [
			{text: "移动", value: "move"},
			{text: "隐藏", value: "hide"},
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
		scale: {text: "比例", value: 1},
		offsetX: {text: "X偏移", value: 0},
		offsetY: {text: "Y偏移", value: 0},
		point: {text: "坐标", value: "0,0"},
		interval: {text: "间隔", value: 100},
		json: {text: "数据", value: ""},
	},
	command: {
		cmd_shape: {
			move: {text: "移动", key: "m", conf:{"shape": "move"}},
			hide: {text: "隐藏", key: "h"},
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
		ctrl_status: {
			shape: {type: "config"},
			stroke: {type: "config"},
			color: {type: "config"},
			scale: {type: "cache"},
			point: {type: "cache"},
		},
		ctrl_show: {
			big: {text: "放大", key: "+"},
			small: {text: "缩小", key: "-"},
			play: {text: "播放", key: "a"},
			interval: {type: "config", width: 30},
		},
		ctrl_data: {
			delete: {text: "删除", key: "d"},
			clear: {text: "清空", key: "q"},
			export: {text: "导出"},
			import: {text: "导入"},
			json: {type: "config", width: 80},
		},
		"": {},
	},
}
//}}}
function init(configs, commands) {//{{{
	current_ctx.agent.isChrome = window.navigator.userAgent.indexOf("Chrome")>-1;
	current_ctx.agent.isMobile = window.navigator.userAgent.indexOf("Mobile")>-1;

	for (var group in commands) {
		var cs = document.getElementsByClassName(group+" bar");
		for (var i = 0; i < cs.length; i++) {
			for (var which in commands[group]) {
				var command = commands[group][which];

				var config = configs[which];
				if (command.type == "cache") {
					var label = cs[i].appendChild(document.createElement("label"));
					label.innerText = config.text+": ";

					var label = cs[i].appendChild(document.createElement("label"));
					label.innerText = config.value;
					label.className = "config "+which;
				} else if (command.type == "config") {
					if (config.list) {
						var select = cs[i].appendChild(document.createElement("select"));
						select.className = "config "+which;

						for (var j in config.list) {
							var item = config.list[j];
							var option = select.appendChild(document.createElement("option"));
							option.value = item.value;
							option.text = item.text;
							if (config.value == item.value) {
								select.selectedIndex = j
							}
						}

						(function() {
							var bar = group;
							var key = which;
							select.onchange = function(event) {
								current_ctx.config[key].value = event.target[event.target.selectedIndex].value;
							}
						})();
					} else {
						var label = cs[i].appendChild(document.createElement("label"));
						label.innerText = config.text+": ";

						var input = cs[i].appendChild(document.createElement("input"));
						input.style.width = command.width+"px";
						input.value = config.value;
						input.className = "config "+which;

						(function() {
							var key = which;
							input.onblur = function(event) {
								current_ctx.config[key].value = event.target.value;
							}
							input.onkeyup = function(event) {
								switch (event.key) {
								case "Enter":
									current_ctx.config[key].value = event.target.value;
									break
								case "Escape":
									event.target.value = current_ctx.config[key].value;
									break
								}
							}
						})();
					}
				} else {
					var cmd = cs[i].appendChild(document.createElement("button"));
					cmd.className = group+" "+which;

					if (command.key) {
						control_map[command.key] = [group, which];
						cmd.innerText = command.text+"("+command.key+")";
					} else {
						cmd.innerText = command.text
					}

					(function() {
						var key = which;
						var bar = group;
						cmd.onclick = function(event) {
							action(event, key, bar);
						}
					})();
				}
			}
		}
	}
}//}}}
function conf(group, which, value) {//{{{
	var config = current_ctx[group][which];
	if (value != undefined) {
		config.value = value;
		var cs = document.getElementsByClassName(group+" "+which);
		for (var i = 0; i < cs.length; i++) {
			if (cs[i].nodeName == "LABEL") {
				cs[i].innerText = value;
			} else {
				cs[i].value = value;
			}
		}
	}
	return config.value
}//}}}
function info() {//{{{
	var list = []
	for (var i = 0; i < arguments.length; i++) {
		if (typeof arguments[i] == "object") {
			list.push("{")
			for (var k in arguments[i]) {
				list.push(k+": "+arguments[i][k]+",")
			}
			list.push("}")
		} else {
			list.push(arguments[i])
		}
	}

	var debug_info = document.getElementsByClassName("debug_info");
	for (var i = 0; i < debug_info.length; i++) {
		var p = debug_info[i].appendChild(document.createElement("p"));
		p.appendChild(document.createTextNode(list.join(" ")));
		debug_info[i].scrollTop+=100;
	}
}
//}}}

var control_map = {//{{{
	Escape: ["", "escape"],

	action: {
		"escape": [function() {
			current_ctx.begin_point = null;
			current_ctx.end_point = null;
		}],
		"hide": [function() {
			var s = {shape: "hide", time:1}
			add_history(his, s);
			draws(draw, s);
		}],
		"big": [function(){
			draw.scale(current_ctx.big_scale, current_ctx.big_scale);
			var m = draw.getTransform();
			conf("config", "scale", m.a);
		}],
		"small": [function(){
			draw.scale(current_ctx.small_scale, current_ctx.small_scale);
			var m = draw.getTransform();
			conf("config", "scale", m.a);
		}],
		"play": [function() {
			draw.resetTransform();
			conf("config", "scale", 1)
			conf("config", "offsetX", 0)
			conf("config", "offsetY", 0)

			refresh(conf("config", "interval"), 0, "", function() {
				var m = draw.getTransform();
				conf("config", "scale", m.a);
				conf("config", "offsetX", m.e)
				conf("config", "offsetY", m.f)
			});
			return false
		}],
		"delete": [function() {
			if (draw_history.length > 1) {
				var tr = his.rows[his.rows.length-1];
				tr.parentElement.removeChild(tr)
				draw_history.pop();
			}
		}],
		"clear": [function() {
			if (confirm("clear all?")) {
				draw_history.length = 1;
				var th = his.rows[0];
				his.innerHTML = "";
				his.appendChild(th);
			}
		}],
		"export": [function() {
			conf("config", "json", JSON.stringify(draw_history));
			return false
		}],
		"import": [function() {
			var im = JSON.parse(conf("config", "json"));
			for (var i in im) {
				add_history(his, im[i]);
				draws(draw, im[i]);
			}
		}],
		"default": [function(event, which, group) {
			var cs = document.getElementsByClassName(group);
			for (var i = 0; i < cs.length; i++) {
				cs[i].style.backgroundColor = "white";
			}

			var cs = document.getElementsByClassName(group+" "+which);
			for (var i = 0; i < cs.length; i++) {
				cs[i].style.backgroundColor = "lightblue";
			}
		}],
	}
}
//}}}
function control(event) {//{{{
	if (event.type == "keyup" && control_map[event.key]) {
		action(event, control_map[event.key][1], control_map[event.key][0]);
	}
}
//}}}
function action(event, which, group) {//{{{
	var w = control_map.action[which]? which: "default";
	while (control_map.action[w]) {
		var command = current_ctx.command[group][which] || {};
		for (var k in command.conf) {
			conf("config", k, command.conf[k]);
		}
		for (var i in control_map.action[w]) {
			var next = control_map.action[w][i](event, which, group);
			w = next || w;
		}
		for (var k in command.cmd) {
			action(event, k, command.cmd[k])
		}
		next == undefined && refresh()
		w = next;
	}
}
//}}}

function trans(point) {//{{{
	return {
		x: point.x/conf("config", "scale")-conf("config","offsetX"),
		y: point.y/conf("config", "scale")-conf("config","offsetY"),
	}
}
//}}}
function draw_point(event) {//{{{
	var point = trans({
		x: event.type == "touchstart"? event.touches[0].clientX: event.offsetX,
		y: event.type == "touchstart"? event.touches[0].clientY: event.offsetY,
	});
	conf("config", "point", parseInt(point.x)+","+parseInt(point.y));

	if (!current_ctx.begin_point) {
		current_ctx.begin_point = point;
		info(event.type, "begin_point: ", current_ctx.begin_point)
		return
	}
	current_ctx.end_point = point;
	info(event.type, "end_point: ", current_ctx.end_point)

	var s = {
		shape: conf("config", "shape"),
		stroke: conf("config", "stroke"),
		color: conf("config", "color"),
		begin_point: current_ctx.begin_point,
		end_point: current_ctx.end_point,
		text: conf("config", "shape") == "text"? prompt("请入文字", ""): "",
	};

	add_history(his, s);
	draws(draw, s);
	refresh();

	current_ctx.begin_point = null;
	current_ctx.end_point = null;
}
//}}}
function draw_move(event) {//{{{
	var point = trans({x:event.offsetX, y:event.offsetY});
	conf("config", "point", parseInt(point.x)+","+parseInt(point.y));

	if (current_ctx.agent.isMobile) {
		return
	}

	var color = conf("config", "color");
	var stroke = conf("config", "stroke");
	var shape = conf("config", "shape");

	if (current_ctx.begin_point) {
		switch (conf("config", "shape")) {
		case "move":
			var m = draw.getTransform()
			draw.translate(point.x-current_ctx.begin_point.x,point.y-current_ctx.begin_point.y);
			refresh();
			draw.setTransform(m)
			break
		default:
			refresh();
			draws(draw, {
				shape: shape, stroke: stroke, color: color,
				begin_point: current_ctx.begin_point,
				end_point: point,
				text: "",
			});
		}
	}
}
//}}}

function add_history(his, s) {//{{{
	s.index = draw_history.length;
	switch (s.shape) {
		case "move":
			s.type = "image"
	}

	draw_history.push(s);
	if (s.begin_point) {
		var begin_x = s.begin_point.x;
		var begin_y = s.begin_point.y;
		var end_x = s.end_point.x;
		var end_y = s.end_point.y;
	}

	var tr = his.appendChild(document.createElement("tr"))
	var headers = ["shape", "stroke", "color", "x1", "y1", "x2", "y2", "text"]
	var fields = [s.shape, s.stroke, s.color,
		parseInt(begin_x), parseInt(begin_y), parseInt(end_x), parseInt(end_y), s.text]

	for (var i in fields) {
		var td = tr.appendChild(document.createElement("td"));
		switch (headers[i]) {
			case "color":
			case "stroke":
			case "shape":
				var select = td.appendChild(document.createElement("select"));
				var list = current_ctx.config[headers[i]].list;
				for (var j in list) {
					var option = select.appendChild(document.createElement("option"));
					option.value = list[j].value;
					option.text = list[j].text;
					if (option.value == fields[i]) {
						select.selectedIndex = j;
					}
				}

				(function() {
					var index = headers[i];
					select.onchange = function(event) {
						draw_history[tr.rowIndex][index] = event.target[event.target.selectedIndex].value;
						refresh();
					}
				})();
				break
			default:
				var input = td.appendChild(document.createElement("input"));
				input.value = fields[i];
				input.style.width="46px";

				(function() {
					var row = draw_history[tr.rowIndex]
					var col = headers[i]
					input.onblur = input.onkeyup = function(event) {
						if (event.key && event.key != "Enter") {
							return
						}
						var value = event.target.value;
						switch (col) {
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
						refresh()
					}
				})()
		}
	}

	his.scrollTop+=100;
	return s;
}
//}}}
function refresh(time, i, last, done) {//{{{
	i = i || 0
	if (!last || last > draw_history.length) {
		last = draw_history.length;
	}

	if (time) {
		if (i < last) {
			draws(draw, draw_history[i]);
			if (draw_history[i].type == "image") {
				refresh(0, 0, i);
			}

			setTimeout(function(){refresh(time, i, last, done)}, draw_history[i].time||time);
			i++
			return
		}

		typeof done == "function" && done();
		return
	}

	for (i = i || 0; i < last; i++) {
		if (draw_history[i].type != "image") {
			draws(draw, draw_history[i]);
		}
	}
}
//}}}
function draws(draw, h) {//{{{
	if (h.begin_point) {
		var begin_x = h.begin_point.x;
		var begin_y = h.begin_point.y;
		var end_x = h.end_point.x;
		var end_y = h.end_point.y;
	}

	switch (h.shape) {
	case "init":
	case "move":
		draw.translate(end_x-begin_x, end_y-begin_y);
		var m = draw.getTransform();
		conf("config", "offsetX", m.e)
		conf("config", "offsetY", m.f)
		return
	}

	draw.save();

	if (h.color) {
		if (h.stroke == "stroke") {
			draw.strokeStyle = h.color;
		} else {
			draw.fillStyle = h.color;
		}
	}

	switch (h.shape) {
		case "hide":
			draw.clearRect(-conf("config", "offsetX")/conf("config", "scale"), -conf("config", "offsetY")/conf("config", "scale"), 400/conf("config", "scale"), 400/conf("config", "scale"));
			break
		case 'heart':
			r = Math.sqrt(Math.pow(begin_x-end_x, 2)+Math.pow(begin_y-end_y,2));
			a = Math.atan((end_y-begin_y)/(end_x-begin_x))/Math.PI*180;
			drawHeart(draw, begin_x, begin_y, r, a, h.color, h.stroke)
			break
		case 'cycle':
			draw.beginPath();
			r = Math.sqrt(Math.pow(begin_x-end_x, 2)+Math.pow(begin_y-end_y,2));
			draw.arc(begin_x, begin_y, r, 0, 2*Math.PI)
			draw[h.stroke]()
			break
		case 'line':
			draw.beginPath();
			draw.moveTo(begin_x, begin_y);
			draw.lineTo(end_x, end_y);
			draw[h.stroke]()
			break
		case 'rect':
			draw[h.stroke+"Rect"](begin_x, begin_y, end_x-begin_x, end_y-begin_y);
			break
		case 'text':
			draw.font = current_ctx.font;
			draw[h.stroke+"Text"](h.text, begin_x, begin_y, end_x-begin_x);
	}

	draw.restore();
}
//}}}

init(current_ctx.config, current_ctx.command);
action(null, "heart", "cmd_shape");
action(null, "red", "cmd_color");

