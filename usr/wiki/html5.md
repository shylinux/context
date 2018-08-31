<canvas id="heart" width="400" height="400"></canvas>

## 简介

- 文档: <https://developer.mozilla.org/en-US/docs/Learn>
- 文档: <https://developer.mozilla.org/en-US/docs/Web/API/Canvas_API/Tutorial/>

### miniCAD在线绘图
<label>颜色: </label>
<select onchange="select(event, 'color')">
	<option>red</option>
	<option>green</option>
	<option>yellow</option>
	<option>blue</option>
	<option>black</option>
	<option>white</option>
	<option>purple</option>
</select>
<label>画笔: </label>
<select id="select_pan" onchange="select(event, 'stroke')">
	<option>stroke</option>
	<option>fill</option>
</select>
<label>图形: </label>
<select id="select_pan" onchange="select(event, 'shape')">
	<option>heart</option>
	<option>cycle</option>
	<option>rect</option>
	<option>line</option>
</select>
<label id="show">坐标: 0,0</label>
<br/>
<button class="control e" onclick="action(event, 'heart')">画心(e)</button>
<button class="control c" onclick="action(event, 'cycle')">画圆\(c\)</button>
<button class="control r" onclick="action(event, 'rect')">矩形\(r\)</button>
<button class="control v" onclick="action(event, 'line')">直线(v)</button>
<button class="control t" onclick="action(event, 'text')">文字(t)</button>
<button class="control a" onclick="action(event, 'play')">播放\(a\)</button>
<br/>
<canvas id="draw" width="400" height="400"
onmousemove="draw_move(event)"
onmouseup="draw_point(event)"
></canvas>
<br/>
<button class="control" onclick="action(event, 'move')">追踪</button>
<button class="control b" onclick="action(event, 'big')">放大(b)</button>
<button class="control m" onclick="action(event, 'small')">缩小(m)</button>
<button class="control" onclick="action(event, 'hide')">隐藏</button>
<button class="control" onclick="action(event, 'draw')">恢复</button>
<button class="control d" onclick="action(event, 'delete')">删除\(d\)</button>
<button class="control" onclick="action(event, 'clear')">清空\(q\)</button>
<br/>
<div id="fuck">
</div>
<div style="clear:both">
</div>
<br/>
<div><table id="draw_history"></table></div>

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
