<!DOCTYPE html>
<head>
<style>
	th {
		cursor:pointer;
		background-color:lightgray;
	}
	.time {
		padding-right:20px;
	}
	.size {
		text-align:right;
		padding-right:20px;
	}
	.name {
		text-align:left;
	}
	code {
		font-size:16px;
	}
</style>
</head>
<body>
<form method="POST" action="/upload" enctype="multipart/form-data">
	<input type="text" name="path" value="{{index . "file" 0}}">
	<input type="file" name="file">
	<input type="submit">
</form>
<button onclick="order('max')">max</button>
<button onclick="order('min')">min</button>
<script>
	function getCookie(name) {
		pattern = /([^=]+)=([^;]+);?\s*/g;
	}

	function order(what) {
		document.cookie = "order="+what;
		location.reload()
	}
	function list(what) {
		document.cookie = "list="+what;
		location.reload()
	}
</script>

<table>
<colgroup>{{range .append}}<col class="{{.}}">{{end}}</colgroup>
<tr>{{range .append}}<th class="{{.}}" onclick="list('{{.}}')">{{.}}</th>{{end}}</tr>
{{$meta := .}} {{$first := index .append 0}}
{{range $i, $k := index . $first}}
<tr>
	{{range $key := index $meta "append"}}
		{{if eq $key "name"}}
			<td class="{{$key}}">
				<a href="/download?file={{index $meta "file" 0}}/{{index $meta $key $i}}"><code>{{index $meta $key $i}}</code></a>
			</td>
		{{else}}
			<td class="{{$key}}">
				<code>{{index $meta $key $i}}</code>
			</td>
		{{end}}
	{{end}}
</tr>
{{end}}
</table>
</body>
