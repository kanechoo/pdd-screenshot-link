package main

import (
	"errors"
	"fmt"
	"gocv.io/x/gocv"
	"image"
	"log"
	"os"
	"regexp"
	"strings"
	"time"
)

const folder = "images"

type Prefix string
type Suffix string
type Result string
type ScreenshotImage struct {
	filename            string
	Image               []byte
	PrefixImage         string
	SuffixImage         string
	PrefixImageFullText string
	SuffixImageFullText string
	Result              []Result
	//
	fileSuffixName string
}

func newScreenshotImage(filename string, image []byte) *ScreenshotImage {
	stat, err := os.Stat(folder)
	if err != nil {
		if os.IsNotExist(err) {
			err := os.Mkdir(folder, 0755)
			if err != nil {
				panic(err)
			}
		}
	} else {
		if !stat.IsDir() {
			err := os.Mkdir(folder, 0755)
			if err != nil {
				panic(err)
			}
		}
	}
	return &ScreenshotImage{
		filename: filename,
		Image:    image,
	}
}

func (s *ScreenshotImage) Execute() (*[]Result, error) {
	failResult := make([]Result, 0)
	//Check filename is valid
	if s.filename == "" {
		return &failResult, errors.New("filename is empty")
	}
	strings.Split(s.filename, ".")
	filenameParts := strings.Split(s.filename, ".")
	if len(filenameParts) < 2 {
		return &failResult, errors.New("filename is invalid")
	}
	s.fileSuffixName = strings.Split(s.filename, ".")[len(filenameParts)-1]
	random := time.Now().UnixMilli()
	imageFile := fmt.Sprintf("%s/%d_origin_.%s", folder, random, s.fileSuffixName)
	err := os.WriteFile(imageFile, s.Image, 0644)
	if err != nil {
		return &failResult, err
	}
	//load original image
	originalImage := gocv.IMRead(imageFile, gocv.IMReadColor)
	//convert rgb color to gray
	grayImage := gocv.NewMat()
	gocv.CvtColor(originalImage, &grayImage, gocv.ColorBGRToGray)
	//crop image
	prefixImage := grayImage.Region(image.Rectangle{
		Min: image.Point{X: 0, Y: 0},
		Max: image.Point{X: 175, Y: grayImage.Rows()},
	})
	//adjust prefix image threshold
	prefixThresholdImage := gocv.NewMat()
	gocv.Threshold(prefixImage, &prefixThresholdImage, 190, 255, gocv.ThresholdBinary)
	//adjust prefix image threshold
	PrefixThresholdImage2 := gocv.NewMat()
	gocv.Threshold(prefixThresholdImage, &PrefixThresholdImage2, 90, 255, gocv.ThresholdToZero)
	//output prefix image
	prefixImageFile := fmt.Sprintf("%s/%d_prefix_.%s", folder, random, s.fileSuffixName)
	gocv.IMWrite(prefixImageFile, PrefixThresholdImage2)
	//tesseract get image full text
	prefixImageFullText, err := tesseract(prefixImageFile)
	if err != nil {
		return &failResult, err
	}
	s.PrefixImageFullText = prefixImageFullText
	log.Printf("Prefix Image Full Text:\n%s", prefixImageFullText)
	//crop image
	suffixImage := grayImage.Region(image.Rectangle{
		Min: image.Point{X: 214, Y: 0},
		Max: image.Point{X: grayImage.Cols(), Y: grayImage.Rows()},
	})
	//adjust suffix image threshold
	suffixThresholdImage := gocv.NewMat()
	gocv.Threshold(suffixImage, &suffixThresholdImage, 240, 255, gocv.ThresholdBinary)
	//adjust suffix image threshold
	suffixThresholdImage2 := gocv.NewMat()
	gocv.Threshold(suffixThresholdImage, &suffixThresholdImage2, 90, 255, gocv.ThresholdToZero)
	//output suffix image
	suffixImageFile := fmt.Sprintf("%s/%d_suffix_.%s", folder, random, s.fileSuffixName)
	gocv.IMWrite(suffixImageFile, suffixThresholdImage2)
	//tesseract get image full text
	suffixImageFullText, err := tesseract(suffixImageFile)
	if err != nil {
		return &failResult, err
	}
	s.SuffixImageFullText = suffixImageFullText
	log.Printf("Suffix Image Full Text:\n%s", suffixImageFullText)
	prefixes := getPrefixFromFullText(s.PrefixImageFullText)
	suffixes := getSuffixFromFullText(s.SuffixImageFullText)
	result := combine(prefixes, suffixes)
	//clean images
	//	cleanImages([]string{imageFile, prefixImageFile, suffixImageFile})
	return result, nil
}
func getPrefixFromFullText(prefixFullText string) *[]Prefix {
	regExp := regexp.MustCompile(`^[A-Za-z0-9]+:$`)
	result := make([]Prefix, 0)
	for _, lineText := range strings.Split(prefixFullText, "\n") {
		lineText = strings.TrimSpace(lineText)
		if regExp.MatchString(lineText) {
			result = append(result, Prefix(lineText))
		}
	}
	return &result
}
func getSuffixFromFullText(suffixFullText string) *[]Suffix {
	result := make([]Suffix, 0)
	for _, lineText := range strings.Split(suffixFullText, "\n") {
		lineText = strings.TrimSpace(lineText)
		if strings.Contains(lineText, "APP") {
			//cut first 13 characters
			code := lineText[:13]
			//add unicode 21e5 as first character and 21e4 as last character
			code = fmt.Sprintf("%c%s%c", 0x21e5, code, 0x21e4)
			result = append(result, Suffix(code))
		}
	}
	return &result
}
func combine(prefixes *[]Prefix, suffixes *[]Suffix) *[]Result {
	result := make([]Result, 0)
	for i := 0; i < len(*prefixes); i++ {
		if i < len(*suffixes) {
			result = append(result, Result(fmt.Sprintf("%s/%s 复制并打开拼多多APP", (*prefixes)[i], (*suffixes)[i])))
		}
	}
	return &result
}
func cleanImages(files []string) {
	for _, file := range files {
		err := os.Remove(file)
		if err != nil {
			log.Printf("Error: %v", err)
		}
	}
}
