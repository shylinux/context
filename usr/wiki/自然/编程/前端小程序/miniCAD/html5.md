<canvas id="heart" width="400" height="400"></canvas>

## 简介

- 文档: <https://developer.mozilla.org/en-US/docs/Learn>
- 文档: <https://developer.mozilla.org/en-US/docs/Web/API/Canvas_API/Tutorial/>

### miniCAD在线绘图
<div class="cmd_shape bar"></div>
<div class="cmd_stroke bar"></div>
<div class="cmd_color bar"></div>
<div class="ctrl_show bar"></div>
<canvas id="draw" width="400" height="400"
onmousemove="draw_move(event)"
onmouseup="draw_point(event)"
></canvas>
<div class="ctrl_status bar"></div>
<div class="debug_info"></div>
<div class="ctrl_data bar"></div>
<table>
<thead>
<tr>
<th>color</th>
<th>stroke</th>
<th>shape</th>
<th>x1</th>
<th>y1</th>
<th>x2</th>
<th>y2</th>
<th>text</th>
</tr>
</thead>
<tbody id="draw_history"></tbody>
</table>


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
