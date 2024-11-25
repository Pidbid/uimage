package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()

	e.GET("/generate", func(c echo.Context) error {
		// 获取查询参数
		size := c.QueryParam("size")
		round := c.QueryParam("round")
		colorParam := c.QueryParam("color")

		// 设置默认值
		if size == "" {
			size = "200*200"
		}
		if round == "" {
			round = "0"
		}
		if colorParam == "" {
			colorParam = "grey"
		}

		// 解析 size 参数
		width, height := parseSize(size)

		// 解析 round 参数
		roundPercent, _ := strconv.Atoi(round)
		if roundPercent < 0 {
			roundPercent = 0
		} else if roundPercent > 100 {
			roundPercent = 100
		}

		// 解析颜色参数
		backgroundColor := parseColor(colorParam)

		// 创建图片
		img := image.NewRGBA(image.Rect(0, 0, width, height))
		draw.Draw(img, img.Bounds(), &image.Uniform{backgroundColor}, image.Point{}, draw.Src)

		// 添加圆角
		addRoundedCorners(img, roundPercent)

		// 添加文字
		text := fmt.Sprintf("%d X %d", width, height)
		addText(img, text, color.Black, width, height)

		// 编码为 PNG 并返回 Base64
		var buf bytes.Buffer
		png.Encode(&buf, img)
		base64Str := base64.StdEncoding.EncodeToString(buf.Bytes())

		return c.String(http.StatusOK, base64Str)
	})

	e.Start(":8080")
}

func parseSize(size string) (int, int) {
	parts := strings.Split(size, "*")
	if len(parts) != 2 {
		return 200, 200
	}
	width, _ := strconv.Atoi(parts[0])
	height, _ := strconv.Atoi(parts[1])
	return width, height
}

func parseColor(colorParam string) color.Color {
	colors := map[string]color.RGBA{
		"grey":  {128, 128, 128, 255},
		"blue":  {0, 0, 255, 255},
		"green": {0, 128, 0, 255},
		"white": {255, 255, 255, 255},
		"black": {0, 0, 0, 255},
	}
	if c, ok := colors[colorParam]; ok {
		return c
	}
	return colors["grey"]
}

func addRoundedCorners(img *image.RGBA, percent int) {
	width := img.Bounds().Dx()
	height := img.Bounds().Dy()
	radius := (width + height) / 4 * percent / 100

	mask := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(mask, mask.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)

	for y := 0; y < radius; y++ {
		for x := 0; x < radius; x++ {
			if (x-radius)*(x-radius)+(y-radius)*(y-radius) > radius*radius {
				mask.Set(x, y, color.Transparent)
				mask.Set(width-x-1, y, color.Transparent)
				mask.Set(x, height-y-1, color.Transparent)
				mask.Set(width-x-1, height-y-1, color.Transparent)
			}
		}
	}

	draw.DrawMask(img, img.Bounds(), mask, image.Point{}, nil, image.Point{}, draw.Over)
}

func addText(img *image.RGBA, text string, clr color.Color, width, height int) {
	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetFont(loadFont("Roboto-Black.ttf")) // 使用 Roboto-Black.ttf 字体
	c.SetClip(img.Bounds())
	c.SetDst(img)
	c.SetSrc(image.NewUniform(clr))

	// 根据图片大小调整字号
	fontSize := float64(height) / 10.0
	c.SetFontSize(fontSize)

	// 计算文字的宽度和高度
	textBounds, _ := c.DrawString(text, freetype.Pt(0, 0))
	textWidth := textBounds.X.Ceil()
	textHeight := textBounds.Y.Ceil()

	// 计算文字的居中位置
	pt := freetype.Pt((width-textWidth)/2, (height+textHeight)/2)

	// 绘制文字
	_, err := c.DrawString(text, pt)
	if err != nil {
		return
	}
}

func loadFont(path string) *truetype.Font {
	fontBytes, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	f, err := freetype.ParseFont(fontBytes)
	if err != nil {
		panic(err)
	}
	return f
}
