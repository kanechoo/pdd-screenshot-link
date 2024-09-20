package main

import (
	"errors"
	"fmt"
	"gocv.io/x/gocv"
	"image"
	"image/color"
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
	originalMat := gocv.IMRead(imageFile, gocv.IMReadColor)
	//crop image
	prefixMat := originalMat.Region(image.Rectangle{
		Min: image.Point{X: 0, Y: 0},
		Max: image.Point{X: 175, Y: originalMat.Rows()},
	})
	gocv.CvtColor(prefixMat, &prefixMat, gocv.ColorBGRToGray)
	//adjust prefix image threshold
	prefixThresholdMat := gocv.NewMat()
	gocv.Threshold(prefixMat, &prefixThresholdMat, 190, 255, gocv.ThresholdBinary)
	//adjust prefix image threshold
	PrefixThresholdMat := gocv.NewMat()
	gocv.Threshold(prefixThresholdMat, &PrefixThresholdMat, 90, 255, gocv.ThresholdToZero)
	//output prefix image
	prefixMatFile := fmt.Sprintf("%s/%d_prefix_.%s", folder, random, s.fileSuffixName)
	gocv.IMWrite(prefixMatFile, PrefixThresholdMat)
	//tesseractToText get image full text
	prefixMatFullText, err := tesseractToText(prefixMatFile)
	if err != nil {
		return &failResult, err
	}
	s.PrefixImageFullText = prefixMatFullText
	log.Printf("Prefix Image Full Text:\n%s", prefixMatFullText)
	//crop image
	suffixMat := originalMat.Region(image.Rectangle{
		Min: image.Point{X: 217, Y: 750},
		Max: image.Point{X: originalMat.Cols(), Y: originalMat.Rows() - 250},
	})
	//0 400
	gocv.Resize(suffixMat, &suffixMat, image.Point{X: suffixMat.Cols() * 3, Y: suffixMat.Rows() * 3}, 0, 0, gocv.InterpolationLinear)
	suffixMatV2 := gocv.NewMat()
	gocv.CvtColor(suffixMat, &suffixMatV2, gocv.ColorBGRToGray)
	//adjust suffix image threshold
	suffixThresholdMat := gocv.NewMat()
	//165
	gocv.Threshold(suffixMatV2, &suffixThresholdMat, 0, 255, gocv.ThresholdOtsu)
	//sx := gocv.NewMat()
	//gocv.Sobel(suffixThresholdMat, &sx, gocv.MatTypeCV32F, 1, 0, 1, 1, 0, gocv.BorderDefault)
	//sy := gocv.NewMat()
	//gocv.Sobel(suffixThresholdMat, &sy, gocv.MatTypeCV32F, 0, 1, 1, 1, 0, gocv.BorderDefault)
	//sz := gocv.NewMat()
	//gocv.Magnitude(sx, sy, &sz)
	//gocv.Normalize(sz, &sz, 0, 255, gocv.NormMinMax)
	//output suffix image
	suffixMatFile := fmt.Sprintf("%s/%d_suffix_.%s", folder, random, s.fileSuffixName)
	sa := gocv.NewMat()
	//sz.ConvertTo(&sa, gocv.MatTypeCV8UC1)
	gocv.BitwiseNot(suffixThresholdMat, &sa)
	contours := gocv.FindContours(sa, gocv.RetrievalTree, gocv.ChainApproxSimple)
	println(contours.Size())
	gocv.DrawContours(&suffixMat, contours, -1, color.RGBA{
		R: 255,
		G: 255,
		B: 0,
		A: 1,
	}, 0)
	for i := 0; i < contours.Size(); i++ {
		region := sa.Region(gocv.BoundingRect(contours.At(i)))
		if gocv.ContourArea(contours.At(i)) > 100 {
			gocv.IMWrite(fmt.Sprintf("%s/%d.jpg", folder, i), region)
		}
	}
	gocv.IMWrite(suffixMatFile, suffixThresholdMat)
	//tesseractToText get image full text
	suffixMatFullText, err := tesseractToText(suffixMatFile)
	if err != nil {
		return &failResult, err
	}
	s.SuffixImageFullText = suffixMatFullText
	log.Printf("Suffix Image Full Text:\n%s", suffixMatFullText)
	prefixes := getPrefixFromFullText(s.PrefixImageFullText)
	suffixes := getSuffixFromFullText(s.SuffixImageFullText)
	result := combine(prefixes, suffixes)
	//clean images
	//	cleanImages([]string{imageFile, prefixMatFile, suffixMatFile})
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
