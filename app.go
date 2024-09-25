package main

import (
	"github.com/kanechoo/pdd-screenshot-link/x"
	"log"
)

func main() {
	//高度超过40才可能是大写字母
	//engine := gin.Default()
	//engine.POST("/upload", func(c *gin.Context) {
	//	file, _ := c.FormFile("file")
	//	f, err := file.Open()
	//	if err != nil {
	//		c.JSON(500, gin.H{"error": err.Error()})
	//		return
	//	}
	//	b := make([]byte, file.Size)
	//	_, err = f.Read(b)
	//	if err != nil {
	//		c.JSON(500, gin.H{"error": err.Error()})
	//		return
	//	}
	//	// This is the line that will cause the error
	//	screenshotImage := newScreenshotImage(file.Filename, b)
	//	result, err := screenshotImage.Execute()
	//	if err != nil {
	//		c.JSON(500, gin.H{"error": err.Error()})
	//		return
	//	}
	//	c.JSON(200, gin.H{"data": result})
	//})
	//err := engine.Run(":8188")
	//if err != nil {
	//	log.Printf("Error: %v", err)
	//	return
	//}
	p := x.NewXProcessor("test8.jpg")
	fragment, err := p.GetQuoteFragment()
	if nil != err {
		log.Fatal(err)
	}
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
		log.Printf("检测到的link为: %s\n", code)
	}
}
