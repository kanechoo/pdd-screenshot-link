package main

import (
	"fmt"
	"github.com/kanechoo/pdd-screenshot-link/tesseract"
	"gocv.io/x/gocv"
	"image"
	"log"
	"math"
	"sort"
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
	test()
}
func test() {
	tessert := tesseract.New()
	o := gocv.IMRead("test8.jpg", gocv.IMReadColor)
	l := gocv.NewMat()
	gocv.CvtColor(o, &l, gocv.ColorBGRToGray)
	gocv.Threshold(l, &l, 0, 255, gocv.ThresholdOtsu)
	gocv.Resize(l, &l, image.Point{}, 1, 1, gocv.InterpolationLinear)
	r := gocv.NewMat()
	gocv.BitwiseNot(l, &r)
	f := gocv.NewMat()
	//25 10
	rectRange := gocv.GetStructuringElement(gocv.MorphRect, image.Point{X: 30, Y: 10})
	gocv.MorphologyEx(r, &f, gocv.MorphClose, rectRange)
	contours := gocv.FindContours(f, gocv.RetrievalTree, gocv.ChainApproxSimple)
	//gocv.DrawContours(&o, contours, -1, color.RGBA{
	//	R: 255,
	//	G: 0,
	//	B: 0,
	//	A: 1,
	//}, 0)
	for i := 0; i < contours.Size(); i++ {
		area := r.Region(gocv.BoundingRect(contours.At(i)))
		//gocv.IMWrite(fmt.Sprintf("%s/o_%d.jpg", folder, i), area)
		if area.Cols() > 905 && area.Cols() < 940 && tessert.DetectContains(&area, 6, "APP") {
			//裁切掉代码前缀部分
			area = area.Region(image.Rect(75, 0, area.Cols(), area.Rows()))
			area2 := gocv.NewMat()
			//2 8
			element := gocv.GetStructuringElement(gocv.MorphRect, image.Point{X: 2, Y: 8})
			gocv.MorphologyEx(area, &area2, gocv.MorphClose, element)
			gocv.IMWrite(fmt.Sprintf("%s/a_%d.jpg", "images", i), area2)
			contours2 := gocv.FindContours(area2, gocv.RetrievalExternal, gocv.ChainApproxSimple)
			//gocv.Erode(area, &area, gocv.GetStructuringElement(gocv.MorphRect, image.Point{X: 2, Y: 2}))
			mapData := make(map[int]*gocv.Mat)
			mapKeys := make([]int, 0)
			for j := 0; j < contours2.Size(); j++ {
				rect := gocv.BoundingRect(contours2.At(j))
				rect.Min.X = int(math.Max(0, float64(rect.Min.X-2)))
				region := area.Region(rect)
				count := gocv.ContourArea(contours2.At(j))
				if j == 30 {
					log.Printf("areaCount: %v", count)
				}
				left := gocv.NewMatWithSize(region.Rows(), 5, gocv.MatTypeCV8UC1)
				left.SetTo(gocv.Scalar{Val1: 0})
				right := gocv.NewMatWithSize(region.Rows(), 5, gocv.MatTypeCV8UC1)
				right.SetTo(gocv.Scalar{Val1: 0})
				gocv.Hconcat(left, region, &region)
				gocv.Hconcat(region, right, &region)
				top := gocv.NewMatWithSize(5, region.Cols(), gocv.MatTypeCV8UC1)
				top.SetTo(gocv.Scalar{Val1: 0})
				bottom := gocv.NewMatWithSize(5, region.Cols(), gocv.MatTypeCV8UC1)
				bottom.SetTo(gocv.Scalar{Val1: 0})
				gocv.Vconcat(top, region, &region)
				gocv.Vconcat(region, bottom, &region)
				//30 50
				if region.Rows() > 30 && region.Rows() < 55 {
					mapData[rect.Min.X] = &region
					mapKeys = append(mapKeys, rect.Min.X)
				}
			}
			//排序
			sort.Ints(mapKeys)
			for index, key := range mapKeys {
				gocv.IMWrite(fmt.Sprintf("%s/%d.jpg", "images", index), *mapData[key])
			}
			//os.Exit(0)
		}
	}
	gocv.IMWrite("test6.jpg", r)
}
