package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/kanechoo/pdd-screenshot-link/x"
	"log"
	"math/rand"
	"os"
	"time"
)

func main() {
	//高度超过40才可能是大写字母
	engine := gin.Default()
	engine.POST("/upload", func(c *gin.Context) {
		file, _ := c.FormFile("file")
		f, err := file.Open()
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		b := make([]byte, file.Size)
		_, err = f.Read(b)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		result := handle(&b)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"data": result})
	})
	err := engine.Run(":8188")
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
}
func handle(b *[]byte) *[]string {
	imageFile := fmt.Sprintf("images/%d.jpg", time.Now().UnixMilli())
	err := os.WriteFile(imageFile, *b, 0666)
	if nil != err {
		log.Fatal(err)
	}
	p := x.NewXProcessor(imageFile)
	fragment, err := p.GetQuoteFragment()
	if nil != err {
		log.Fatal(err)
	}
	links := make([]string, 0)
	for i := 0; i < len(fragment); i++ {
		letterFragments := p.GetLetterFragment(fragment[i])
		var code string
		for j := 0; j < len(letterFragments); j++ {
			letter, err := p.Detect(letterFragments[j])
			if nil != err {
				log.Fatal(err)
			}
			code += letter
		}
		links = append(links, code)
		log.Printf("Link : %s\n", code)
	}
	for i := 0; i < len(links); i++ {
		links[i] = fmt.Sprintf("%c:/%s%s%s", randomCharacter(), "\u21e5", links[i], "\u21e4")
	}
	return &links
}
func randomCharacter() byte {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	return charset[rand.Intn(len(charset))]
}
