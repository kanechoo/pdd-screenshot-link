package main

import (
	"github.com/gin-gonic/gin"
	"log"
)

func main() {
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
		// This is the line that will cause the error
		screenshotImage := newScreenshotImage(file.Filename, b)
		result, err := screenshotImage.Execute()
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
