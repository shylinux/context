package wiki

import (
	"contexts/ctx"
	mis "github.com/shylinux/toolkits"
	"strings"
	"toolkit"
)

// 图形接口
type Chart interface {
	Init(*ctx.Message, ...string) Chart
	Draw(*ctx.Message, int, int) Chart

	GetWidth(...string) int
	GetHeight(...string) int
}

// 图形基类
type Block struct {
	Text       string
	FontColor  string
	FontFamily string
	BackGround string

	FontSize int
	LineSize int
	Padding  int
	Margin   int

	Width, Height int

	TextData string
	RectData string
}

func (b *Block) Init(m *ctx.Message, arg ...string) Chart {
	b.Text = kit.Select(b.Text, arg, 0)
	b.FontColor = kit.Select("white", kit.Select(b.FontColor, arg, 1))
	b.BackGround = kit.Select("red", kit.Select(b.BackGround, arg, 2))
	b.FontSize = kit.Int(kit.Select("24", kit.Select(kit.Format(b.FontSize), arg, 3)))
	b.LineSize = kit.Int(kit.Select("12", kit.Select(kit.Format(b.LineSize), arg, 4)))
	return b
}
func (b *Block) Draw(m *ctx.Message, x, y int) Chart {
	m.Echo(`<rect x="%d" y="%d" width="%d" height="%d" fill="%s" %v/>`,
		x+b.Margin/2, y+b.Margin/2, b.GetWidth(), b.GetHeight(), b.BackGround, b.RectData)
	m.Echo("\n")
	m.Echo(`<text x="%d" y="%d" font-size="%d" style="dominant-baseline:middle;text-anchor:middle;" fill="%s" %v>%v</text>`,
		x+b.GetWidths()/2, y+b.GetHeights()/2, b.FontSize, b.FontColor, b.TextData, b.Text)
	m.Echo("\n")
	return b
}
func (b *Block) Data(root interface{}) {
	mis.Table(mis.Value(root, "data"), 0, 100, func(key string, value string) {
		b.TextData += key + "='" + value + "' "
	})
	mis.Table(mis.Value(root, "rect"), 0, 100, func(key string, value string) {
		b.RectData += key + "='" + value + "' "
	})
	b.FontColor = kit.Select(b.FontColor, mis.Value(root, "fg"))
	b.BackGround = kit.Select(b.BackGround, mis.Value(root, "bg"))
}
func (b *Block) GetWidth(str ...string) int {
	if b.Width != 0 {
		return b.Width
	}
	return len(kit.Select(b.Text, str, 0))*b.FontSize*6/10 + b.Padding
}
func (b *Block) GetHeight(str ...string) int {
	if b.Height != 0 {
		return b.Height
	}
	return b.FontSize*b.LineSize/10 + b.Padding
}
func (b *Block) GetWidths(str ...string) int {
	return b.GetWidth(str...) + b.Margin
}
func (b *Block) GetHeights(str ...string) int {
	return b.GetHeight() + b.Margin
}

// 树
type Chain struct {
	data map[string]interface{}
	max  map[int]int
	Block
}

func (b *Chain) Init(m *ctx.Message, arg ...string) Chart {
	// 解数据
	b.data = mis.Parse(nil, "", b.show(m, arg[0])...).(map[string]interface{})
	b.FontColor = kit.Select("white", arg, 1)
	b.BackGround = kit.Select("red", arg, 2)
	b.FontSize = kit.Int(kit.Select("24", arg, 3))
	b.LineSize = kit.Int(kit.Select("12", arg, 4))
	b.Padding = kit.Int(kit.Select("8", arg, 5))
	b.Margin = kit.Int(kit.Select("8", arg, 6))
	m.Log("info", "data %v", kit.Formats(b.data))

	// 计算尺寸
	b.max = map[int]int{}
	b.Height = b.size(m, b.data, 0, b.max) * b.GetHeights()
	width := 0
	for _, v := range b.max {
		width += b.GetWidths(strings.Repeat(" ", v))
	}
	b.Width = width
	m.Log("info", "data %v", kit.Formats(b.data))
	return b
}
func (b *Chain) Draw(m *ctx.Message, x, y int) Chart {
	return b.draw(m, b.data, 0, b.max, x, y, &Block{})
}
func (b *Chain) show(m *ctx.Message, str string) (res []string) {
	miss := []int{}
	list := mis.Split(str, "\n")
	for _, line := range list {
		// 计算缩进
		dep := 0
	loop:
		for _, v := range []rune(line) {
			switch v {
			case ' ':
				dep++
			case '\t':
				dep += 4
			default:
				break loop
			}
		}

		// 计算层次
		if len(miss) > 0 {
			if miss[len(miss)-1] > dep {
				for i := len(miss) - 1; i >= 0; i-- {
					if miss[i] < dep {
						break
					}
					res = append(res, "]", "}")
					miss = miss[:i]
				}
				miss = append(miss, dep)
			} else if miss[len(miss)-1] < dep {
				miss = append(miss, dep)
			} else {
				res = append(res, "]", "}")
			}
		} else {
			miss = append(miss, dep)
		}

		// 输出节点
		word := mis.Split(line)
		res = append(res, "{", "meta", "{", "text")
		res = append(res, word...)
		res = append(res, "}", "list", "[")
	}
	return
}
func (b *Chain) size(m *ctx.Message, root map[string]interface{}, depth int, width map[int]int) int {
	meta := root["meta"].(map[string]interface{})

	// 最大宽度
	text := kit.Format(meta["text"])
	if len(text) > width[depth] {
		width[depth] = len(text)
	}

	// 计算高度
	height := 0
	if list, ok := root["list"].([]interface{}); ok && len(list) > 0 {
		kit.Map(root["list"], "", func(index int, value map[string]interface{}) {
			height += b.size(m, value, depth+1, width)
		})
	} else {
		height = 1
	}

	meta["height"] = height
	return height
}
func (b *Chain) draw(m *ctx.Message, root map[string]interface{}, depth int, width map[int]int, x, y int, p *Block) Chart {
	meta := root["meta"].(map[string]interface{})
	b.Width, b.Height = 0, 0

	// 当前节点
	block := &Block{
		BackGround: kit.Select(b.BackGround, kit.Select(p.BackGround, meta["bg"])),
		FontColor:  kit.Select(b.FontColor, kit.Select(p.FontColor, meta["fg"])),
		FontSize:   b.FontSize,
		LineSize:   b.LineSize,
		Padding:    b.Padding,
		Margin:     b.Margin,
		Width:      b.GetWidth(strings.Repeat(" ", width[depth])),
	}

	block.Data(root["meta"])
	block.Init(m, kit.Format(meta["text"])).Draw(m, x, y+(kit.Int(meta["height"])-1)*b.GetHeights()/2)

	// 递归节点
	kit.Map(root["list"], "", func(index int, value map[string]interface{}) {
		b.draw(m, value, depth+1, width, x+b.GetWidths(strings.Repeat(" ", width[depth])), y, block)
		y += kit.Int(kit.Chain(value, "meta.height")) * b.GetHeights()
	})
	return b
}

// 表
type Table struct {
	data [][]string
	max  map[int]int
	Block
}

func (b *Table) Init(m *ctx.Message, arg ...string) Chart {
	// 解析数据
	b.max = map[int]int{}
	for _, v := range mis.Split(arg[0], "\n") {
		l := mis.Split(v)
		for i, v := range l {
			switch data := mis.Parse(nil, "", mis.Split(v)...).(type) {
			case map[string]interface{}:
				v = kit.Select("", data["text"])
			}
			if len(v) > b.max[i] {
				b.max[i] = len(v)
			}
		}
		b.data = append(b.data, l)
	}
	b.FontColor = kit.Select("white", arg, 1)
	b.BackGround = kit.Select("red", arg, 2)
	b.FontSize = kit.Int(kit.Select("24", arg, 3))
	b.LineSize = kit.Int(kit.Select("12", arg, 4))
	b.Padding = kit.Int(kit.Select("8", arg, 5))
	b.Margin = kit.Int(kit.Select("8", arg, 6))

	// 计算尺寸
	width := 0
	for _, v := range b.max {
		width += b.GetWidths(strings.Repeat(" ", v))
	}
	b.Width = width
	b.Height = len(b.data) * b.GetHeights()

	m.Log("info", "data %v", kit.Formats(b.data))
	return b
}
func (b *Table) Draw(m *ctx.Message, x, y int) Chart {
	b.Width, b.Height = 0, 0
	for n, line := range b.data {
		for i, text := range line {
			l := 0
			for j := 0; j < i; j++ {
				l += b.GetWidths(strings.Repeat(" ", b.max[i]))
			}
			block := &Block{
				BackGround: kit.Select(b.BackGround),
				FontColor:  kit.Select(b.FontColor),
				FontSize:   b.FontSize,
				LineSize:   b.LineSize,
				Padding:    b.Padding,
				Margin:     b.Margin,
				Width:      b.GetWidth(strings.Repeat(" ", b.max[i])),
			}

			switch data := mis.Parse(nil, "", mis.Split(text)...).(type) {
			case map[string]interface{}:
				text = kit.Select(text, data["text"])
				block.Data(data)
			}
			block.Init(m, text).Draw(m, x+l, y+n*b.GetHeights())
		}
	}
	return b
}
