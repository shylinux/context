<canvas id="heart" width="400" height="400"></canvas>

## 简介

- 文档: <https://developer.mozilla.org/en-US/docs/Learn>
- 文档: <https://developer.mozilla.org/en-US/docs/Web/API/Canvas_API/Tutorial/>

### miniCAD在线绘图
<button class="cmd_shape heart"></button>
<button class="cmd_shape cycle"></button>
<button class="cmd_shape rect"></button>
<button class="cmd_shape line"></button>
<button class="cmd_shape text"></button>
<br/>
<button class="cmd_stroke stroke"></button>
<button class="cmd_stroke fill"></button>
<br/>
<button class="cmd_color black"></button>
<button class="cmd_color red"></button>
<button class="cmd_color yellow"></button>
<button class="cmd_color green"></button>
<button class="cmd_color purple"></button>
<button class="cmd_color blue"></button>
<button class="cmd_color white"></button>
<br/>
<canvas id="draw" width="400" height="400"
onmousemove="draw_move(event)"
onmouseup="draw_point(event)"
onclick="draw_point(event)"
ontouchstart="draw_point(event)"
ontouchend="draw_point(event)"
></canvas>
<br/>
<select class="config shape"></select>
<select class="config stroke"></select>
<select class="config color"></select>
<label class="config scale"></label>
<label class="config begin"></label>
<label class="config point"></label>
<br/>
<button class="control big"></button>
<button class="control small"></button>
<button class="control hide"></button>
<button class="control draw"></button>
<button class="control play"></button>
<input class="config interval"></label>
<br/>
<button class="control delete"></button>
<button class="control clear"></button>
<button class="control export"></button>
<button class="control import"></button>
<input class="config json"></label>
<br/>
<div id="fuck"></div>
<table id="draw_history"></table>

### canvas绘图
<canvas id="demo0" class="demo" width="120" height="120"></canvas>
```
<canvas id="canvas"></canvas>
<script>
  var canvas = document.getElementById("canvas");
  var ctx = canvas.getContext("2d");
  ctx.fillStyle = "green";
  ctx.fillRect(10,10,100,100);
</script>
```

### 画矩形

```
fillRect(x, y, width, height)
strokeRect(x, y, width, height)
clearRect(x, y, width, height)
```

### 画路径
<canvas id="demo2" class="demo" width="120" height="120"></canvas>
```
<canvas id="demo2" width="120" height="120"></canvas>
<script>
var ctx2 = document.getElementById("demo2").getContext("2d");
ctx2.beginPath();
ctx2.moveTo(60,10);
ctx2.lineTo(10,110);
ctx2.lineTo(110,110);
ctx2.fill();
</script>
```

```
beginPath()
moveTo(x, y)
LineTo(x, y)
closePath()
stroke()
fill()

arc(x, y, radius, startAngle, endAngle, anticlockwise)
arcTo(x1, y1, x2, y2, radius)
quadraticCurveTo(cp1x, cp1y, x, y)
bezierCurveTo(cp1x, cp1y, cp2x, cp2y, x, y)
new Path2D()
```
### 设置样式
<canvas id="demo3" class="demo" width="120" height="120"></canvas>
```
fillStyle = "red"
fillStyle = "#FF0000"
fillStyle = "rgb(255,0,0)"
fillStyle = "rgb(255,0,0,1)"
strokeStyle =

var img = new Image();
img.src = "img.png";
img.onLoad = function() {}
createPattern(img, style)

createLinearGradient(x1, y1, x2, y2)
createRadialGradient(x1, y1, r1, x2, y2, r2)
	addColorStop(position, color)

shadowOffsetX
shadowOffsetY
shadowBlur
shadowColor

lineWidth
lineCap
lineJoin
```
### 输出文字
```
font
textAlign
textBaseline
direction
measureText()
fillText(text, x, y[, maxWidth])
strokeText(text, x, y[, maxWidth])
```

### 坐标变换
```
save()
restore()
translate(x,y)
rotate(angle)
scale(x, y)
transform(a,b,c,d,e,f)
setTransform(a,b,c,d,e,f)
resetTransform()
```

<link rel="stylesheet" href="/wiki/html5.css" type="text/css"></link>
<script src="/wiki/html5.js"></script>
