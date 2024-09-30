package main

import (
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/kanechoo/pdd-screenshot-link/x"
	"gocv.io/x/gocv"
	"io/fs"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"
)

func init() {
	_ = os.Mkdir("images", fs.ModePerm)
}
func main() {
	engine := gin.Default()
	engine.Use(cors.Default())
	engine.Use(func(context *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				context.JSON(http.StatusInternalServerError, gin.H{"data": err, "code": http.StatusInternalServerError})
				context.Abort()
			}
		}()
		context.Next()
	})
	engine.POST("/upload", func(c *gin.Context) {
		system := c.Query("system")
		file, _ := c.FormFile("file")
		f, err := file.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"data": err.Error(), "code": http.StatusInternalServerError})
			return
		}
		b := make([]byte, file.Size)
		_, err = f.Read(b)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "code": http.StatusInternalServerError})
			return
		}
		result, err := handle(&b, system)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"data": err.Error(), "code": http.StatusInternalServerError})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": result, "code": http.StatusOK})
	})
	err := engine.Run(":8188")
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
}
func handle(b *[]byte, system string) (*[]string, error) {
	imageFile := fmt.Sprintf("images/%d.jpg", time.Now().UnixMilli())
	err := os.WriteFile(imageFile, *b, fs.ModePerm)
	if nil != err {
		return nil, err
	}
	p := x.NewXProcessor(imageFile)
	defer p.TrashMat()
	fragment, err := p.GetQuoteFragment()
	if nil != err {
		return nil, err
	}
	links := make([]string, 0)
	for i := 0; i < len(fragment); i++ {
		letterFragments := p.GetLetterFragment(fragment[i])
		if len(letterFragments) < 13 {
			return nil, fmt.Errorf("字母或者数字图片数量不足")
		}
		var code string
		for j := 0; j < len(letterFragments); j++ {
			letter, err := p.Detect(letterFragments[j])
			if nil != err {
				return nil, err
			}
			if system == "ios" && ("I" == letter || "l" == letter) {
				letter = adaptIAndLLetter(letterFragments[j].Mat)
			}
			code += letter
		}
		links = append(links, code)
		log.Printf("Link : %s\n", code)
	}
	for i := 0; i < len(links); i++ {
		links[i] = fmt.Sprintf("%c:/%s%s%s", randomCharacter(), "\u21e5", links[i], "\u21e4")
	}
	//反转数组
	for i, j := 0, len(links)-1; i < j; i, j = i+1, j-1 {
		links[i], links[j] = links[j], links[i]
	}
	return &links, nil
}
func adaptIAndLLetter(mat *gocv.Mat) string {
	height := mat.Rows()
	width := mat.Cols()
	minRow := height
	maxRow := -1
	// 遍历每一列
	for col := 0; col < width; col++ {
		for row := 0; row < height; row++ {
			if mat.GetUCharAt(row, col) == 255 { // 检查是否是白色像素
				if row < minRow {
					minRow = row
				}
				if row > maxRow {
					maxRow = row
				}
			}
		}
	}
	if maxRow-minRow > 31 {
		return "l"
	} else {
		return "I"
	}
}
func randomCharacter() byte {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	return charset[rand.Intn(len(charset))]
}
