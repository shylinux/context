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

var select_pan = document.getElementById("select_pan");
var his = document.getElementById("draw_history");
var show = document.getElementById("show");
var item = document.getElementById("draw");
var draw = item.getContext("2d");

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

var current_ctx = {//{{{
	hide: 0,
	scale: 1,
	index_point: false,
	shape: 'cycle',
	stroke: 'stroke',
	color: 'red',
	font: '32px sans-serif',
	big_scale: 1.25,
	small_scale: 0.8,
	begin_point: null,
	end_point: null,
	list: {
		style: ["red", "green", "yellow", "blue"],
		stroke: ["fill", "stroke"],
		shape: ["heart", "cycle", "rect", "line"],
	}
}
//}}}
function display(which, group) {//{{{
	var cs = document.getElementsByClassName(group);
	for (var i = 0; i < cs.length; i++) {
		cs[i].style.backgroundColor = "white";
	}

	var cs = document.getElementsByClassName(group+" "+which);
	for (var i = 0; i < cs.length; i++) {
		cs[i].style.backgroundColor = "lightblue";
	}
}
//}}}
function action(event, s) {//{{{
	console.log(event);
	switch (s) {
		case "escape":
			current_ctx.begin_point = null;
			current_ctx.end_point = null;
			refresh();
			break
		case "big":
			current_ctx.scale *= current_ctx.big_scale;
			draw.scale(current_ctx.big_scale, current_ctx.big_scale);
			refresh();
			break
		case "small":
			current_ctx.scale *= current_ctx.small_scale;
			draw.scale(current_ctx.small_scale, current_ctx.small_scale);
			refresh();
			break
		case "clear":
			if (confirm("clear all?")) {
				action(event, "clear");
				his.innerHTML = "";
				draw_history = [];
				refresh();
			}
			break
		case "hide":
			current_ctx.hide = draw_history.length;
			refresh();
			break
		case "delete":
			draw_history.pop();
			refresh();
			break
		case "draw":
			current_ctx.hide = 0;
			refresh();
			break

		case "fill":
			current_ctx.stroke = "fill";
			select_pan.selectedIndex=1;
			break
		case "stroke":
			current_ctx.stroke = "stroke";
			select_pan.selectedIndex=0;
			break

		case "heart":
			current_ctx.shape = s;
			current_ctx.stroke = "fill";
			display("e", "control");
			select_pan.selectedIndex=1;
			break
		case "cycle":
			current_ctx.shape = s;
			display("c", "control");
			break
		case "rect":
			current_ctx.shape = s;
			display("r", "control");
			break
		case "line":
			current_ctx.shape = s;
			current_ctx.stroke = "stroke";
			select_pan.selectedIndex=0;
			display("v", "control");
			break
		case "text":
			current_ctx.shape = s;
			current_ctx.stroke = "fill";
			select_pan.selectedIndex=0;
			display("t", "control");
			break

		case "play":
			current_ctx.hide = 0;
			refresh(500, 0);
			break
	}
}
//}}}
function select(event, config, val) {//{{{
	var target = event.target;
	var value = target[target.selectedIndex].value;
	switch (config) {
		case "color":
			current_ctx.color = value;
			break
		case "stroke":
			current_ctx.stroke = value;
			break
		case "shape":
			display("r", "control");
			switch (value) {
				case "heart":
					current_ctx.stroke = "fill";
					display("h", "control");
					break
				case "cycle":
					display("c", "control");
					break
				case "rect":
					display("r", "control");
					break
				case "line":
					current_ctx.stroke = "stroke";
					display("v", "control");
					break
			}
			current_ctx.shape = value;
			break
	}
}
//}}}

var draw_history = [];
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
			draw.clearRect(0, 0,400/current_ctx.scale, 400/current_ctx.scale);
		}

		if (current_ctx.hide <= i && i < draw_history.length) {
			var h = draw_history[i];
			draws(draw, h.style, h.stroke, h.shape, h.begin_point, h.end_point);
			i++;
			setTimeout(function(){
				refresh(time, i);
			}, time);
		}
		return
	}

	draw.clearRect(0, 0,400/current_ctx.scale, 400/current_ctx.scale);

	for (var i in draw_history) {
		if (i <current_ctx.hide) {
			continue
		}
		var h = draw_history[i];
		draws(draw, h.style, h.stroke, h.shape, h.begin_point, h.end_point, h.text);
	}
}
//}}}
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
function show_debug(log, clear) {//{{{
	var fuck = document.getElementById("fuck");
	if (clear) {
		fuck.innerHTML = "";
	}
	var div = fuck.appendChild(document.createElement("div"));
	div.appendChild(document.createTextNode(log));
}
//}}}

function trans(point) {//{{{
	point.x /= current_ctx.scale;
	point.y /= current_ctx.scale;
	return point;
}
//}}}
function draw_point(event) {//{{{
	console.log("point");
	console.log(event);

	var point = trans({x:event.offsetX, y:event.offsetY});

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
	if (current_ctx.shape == "text") {
		text = prompt("请入文字", "");
	}

	draws(draw, current_ctx.color, current_ctx.stroke, current_ctx.shape, current_ctx.begin_point, current_ctx.end_point, text);
	draw_history.push({style:current_ctx.color, stroke:current_ctx.stroke, shape:current_ctx.shape, begin_point:current_ctx.begin_point, end_point:current_ctx.end_point, text:text})

	var headers = ["style", "stroke", "shape", "x1", "y1", "x2", "y2", "text"]
	if (his.rows.length == 0) {
		var tr = his.insertRow(-1);
		for (var i in headers) {
			var th = tr.appendChild(document.createElement("th"));
			th.appendChild(document.createTextNode(headers[i]))
		}
	}

	var tr = his.insertRow(-1);
	var fields = [current_ctx.color, current_ctx.stroke, current_ctx.shape,
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
	show_debug("", true)
	show.innerText="坐标: "+parseInt(point.x)+","+parseInt(point.y);

	if (current_ctx.shape == "move") {
	if (current_ctx.index_point) {
		draw.beginPath();
		draw.arc(point.x, point.y, 5, 0, 2*Math.PI);
		draw.fill();
	}
	}

	if (current_ctx.begin_point) {
		refresh();
		draws(draw, current_ctx.color, current_ctx.stroke, current_ctx.shape, current_ctx.begin_point, point, "");
	}
	return false
}
//}}}

