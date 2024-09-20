package main

import (
	"gocv.io/x/gocv"
	"image"
)

func main() {
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
	test()
}
func test() {
	l := gocv.IMRead("test.jpg", gocv.IMReadGrayScale)
	r := gocv.IMRead("test4.jpg", gocv.IMReadGrayScale)
	f := gocv.NewMat()
	gocv.Hconcat(l, r, &f)
	gocv.Threshold(f, &f, 0, 255, gocv.ThresholdOtsu)
	gocv.Resize(f, &f, image.Point{X: int(float64(f.Cols()) * 2), Y: int(float64(f.Rows()) * 2)}, .0, 0, gocv.InterpolationArea)
	gocv.IMWrite("test6.jpg", f)
}
