package x

import (
	"errors"
	"fmt"
	"github.com/kanechoo/pdd-screenshot-link/tesseract"
	"gocv.io/x/gocv"
	"image"
	"log"
	"math"
	"os"
	"sort"
	"strings"
	"time"
)

var (
	//所有大小写字母和数字
	letters = strings.Split("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789", "")
	//大写字母
	upperLetters = strings.Split("ABCDEFGHIJKLMNOPQRSTUVWXYZ", "")
	//小写字母
	lowerLetters = strings.Split("abcdefghijklmnopqrstuvwxyz", "")
	//数字
	numbers = strings.Split("0123456789", "")
)

type QuoteFragment struct {
	//fragment提取到字母和数字应该存放的文件夹
	Dir string
	//Mat
	Mat *gocv.Mat
}
type LetterFragment struct {
	Mat  *gocv.Mat
	File string
}
type XProcessor struct {
	//原始截图文件路径
	OriginImg string
	//提取后需要分析的片段图片，图片和名称将和文件夹名称一致
	fragment []*QuoteFragment
	//文件夹
	dirs []string
	//文件夹和文件夹的mat对应关系的映射
	dirMatMap map[string][]*gocv.Mat
	//tesseract
	tesseract *tesseract.Tesseract
	//切割图片存放文件夹
	baseDir string
	//fragment文件夹
	fragmentDir string
}

func NewXProcessor(originImg string) *XProcessor {
	p := XProcessor{
		OriginImg: originImg,
		tesseract: tesseract.New(),
		dirs:      make([]string, 0),
		dirMatMap: make(map[string][]*gocv.Mat),
		fragment:  make([]*QuoteFragment, 0),
		baseDir:   fmt.Sprintf("images/%d", time.Now().UnixMilli()),
	}
	p.fragmentDir = fmt.Sprintf("%s/fragment", p.baseDir)
	return &p
}
func (p *XProcessor) Detect(letterFragment *LetterFragment) (string, error) {
	text, err := p.tesseract.Detect(letterFragment.File, 13)
	if err != nil {
		return "", err
	}
	text = strings.ReplaceAll(text, "\n", "")
	//特殊情况替换
	text = replaceUnCorrect(text)
	//是否空字符串
	if "" == text {
		//用psm 6再试一次
		text, err := p.tesseract.Detect(letterFragment.File, 6)
		if err != nil {
			return "", err
		}
		if text == "" {
			panic("识别失败")
		} else {
			s := strings.Split(text, "")[0]
			if isNumber(s) {
				return s, nil
			} else {
				return checkCase(s, letterFragment), nil
			}
		}
	}
	//首先需要判断是否是单个字母或者数字
	textArray := strings.Split(text, "")
	if allSameIgnoreCase(textArray) && len(textArray) > 1 {
		textArray = textArray[:1]
		text = textArray[0]
	}
	//如果都是相同的字母取第一个
	if len(textArray) > 1 {
		//用psm 6再试一次
		text, err := p.tesseract.Detect(letterFragment.File, 6)
		if err != nil {
			return "", err
		}
		if text == "" {
			panic("识别失败")
		} else {
			s := strings.Split(text, "")[0]
			if isNumber(s) {
				return s, nil
			} else {
				return checkCase(s, letterFragment), nil
			}
		}
	} else {
		if isNumber(text) {
			return text, nil
		} else {
			return checkCase(text, letterFragment), nil
		}
	}
}

func replaceUnCorrect(text string) string {
	switch text {
	case "(0)":
		return "0"
	case "|":
		return "l"
	}
	return text
}
func allSameIgnoreCase(arrays []string) bool {
	if len(arrays) == 0 {
		return true // 空切片视为“相同”
	}
	first := strings.ToLower(arrays[0])
	for _, str := range arrays {
		if strings.ToLower(str) != first {
			return false
		}
	}
	return true
}
func checkCase(letter string, fragment *LetterFragment) string {
	if isLower(letter) {
		return letter
	} else {
		if fragment.Mat.Rows() <= 40 {
			return strings.ToLower(letter)
		}
		return letter
	}
}
func isUpper(letter string) bool {
	for _, upperLetter := range upperLetters {
		if upperLetter == letter {
			return true
		}
	}
	return false
}
func isLower(letter string) bool {
	for _, lowerLetter := range lowerLetters {
		if lowerLetter == letter {
			return true
		}
	}
	return false
}
func isNumber(letter string) bool {
	for _, number := range numbers {
		if number == letter {
			return true
		}
	}
	return false
}
func (p *XProcessor) GetLetterFragment(fragment *QuoteFragment) []*LetterFragment {
	//裁切掉代码前缀部分，相当于【f:/-lWExrk9cpkEUw复制并打开拼多多APP】的片段图片 大改裁切成 【lWExrk9cpkEUw复制并打开拼多多APP】
	subFragment := fragment.Mat.Region(image.Rect(75, 0, fragment.Mat.Cols(), fragment.Mat.Rows()))
	//模糊并查找字母和数字的轮廓
	contours := blurFindContours(subFragment, 2, 8)
	letterMatMap := make(map[int]*gocv.Mat)
	letterMatKeys := make([]int, 0)
	for index := 0; index < contours.Size(); index++ {
		rectangle := gocv.BoundingRect(contours.At(index))
		//向左偏移2个像素（这样可以获取更完整的字母和数字的图像）
		rectangle.Min.X = int(math.Max(0, float64(rectangle.Min.X-2)))
		//根据轮廓提取字母或者数字的图像
		letterMat := subFragment.Region(rectangle)
		//裁切后的图像左右上下各扩展5个像素（这样tesseract识别更准确一点）
		leftExpandMat := gocv.NewMatWithSize(letterMat.Rows(), 5, gocv.MatTypeCV8UC1)
		leftExpandMat.SetTo(gocv.Scalar{Val1: 0})
		rightExpandMat := gocv.NewMatWithSize(letterMat.Rows(), 5, gocv.MatTypeCV8UC1)
		rightExpandMat.SetTo(gocv.Scalar{Val1: 0})
		//水平方向拼接
		gocv.Hconcat(leftExpandMat, letterMat, &letterMat)
		gocv.Hconcat(letterMat, rightExpandMat, &letterMat)
		topExpandMat := gocv.NewMatWithSize(5, letterMat.Cols(), gocv.MatTypeCV8UC1)
		topExpandMat.SetTo(gocv.Scalar{Val1: 0})
		bottomExpandMat := gocv.NewMatWithSize(5, letterMat.Cols(), gocv.MatTypeCV8UC1)
		bottomExpandMat.SetTo(gocv.Scalar{Val1: 0})
		//垂直方向拼接
		gocv.Vconcat(topExpandMat, letterMat, &letterMat)
		gocv.Vconcat(letterMat, bottomExpandMat, &letterMat)
		//30 50 (字母和数字代码在提取和拼接后垂直方向长度在30到55之间)
		if letterMat.Rows() > 30 && letterMat.Rows() < 55 {
			//根据水平方向X坐标作为key，存储字母和数字的Mat
			letterMatMap[rectangle.Min.X] = &letterMat
			letterMatKeys = append(letterMatKeys, rectangle.Min.X)
		}
	}
	//根据X坐标排序，从小到大
	sort.Ints(letterMatKeys)
	result := make([]*LetterFragment, 0)
	//只需要提取前13个字母和数字
	for index, key := range letterMatKeys {
		if index > 12 {
			break
		}
		mat := letterMatMap[key]
		//输出到文件夹
		name := fmt.Sprintf("%s/%d.jpg", fragment.Dir, index)
		autoMkdirs(name)
		gocv.IMWrite(name, *mat)
		result = append(result, &LetterFragment{
			Mat:  mat,
			File: name,
		})
	}
	return result
}
func blurFindContours(origin gocv.Mat, X int, Y int) gocv.PointsVector {
	target := gocv.NewMat()
	//2 8
	gocv.MorphologyEx(origin, &target, gocv.MorphClose, gocv.GetStructuringElement(gocv.MorphRect, image.Point{X: X, Y: Y}))
	return gocv.FindContours(target, gocv.RetrievalExternal, gocv.ChainApproxSimple)
}
func (p *XProcessor) GetQuoteFragment() ([]*QuoteFragment, error) {
	if p.fragment != nil && len(p.fragment) > 0 {
		return p.fragment, nil
	}
	err := p.extractFragment()
	return p.fragment, err
}
func (p *XProcessor) extractFragment() error {
	if p.OriginImg == "" {
		return errors.New("截图文件路径为空")
	}
	//加载截图
	originMat := gocv.IMRead(p.OriginImg, gocv.IMReadColor)
	originGrayMat := gocv.NewMat()
	//转换为灰度图
	gocv.CvtColor(originMat, &originGrayMat, gocv.ColorBGRToGray)
	//使用Otsu算法进行二值化
	gocv.Threshold(originGrayMat, &originGrayMat, 0, 255, gocv.ThresholdOtsu)
	reverseMat := gocv.NewMat()
	//按位反转图片到黑底白字
	gocv.BitwiseNot(originGrayMat, &reverseMat)
	//从模糊结构后的Mat中提取轮廓
	fragmentContours := blurFindContours(reverseMat, 30, 10)
	//遍历轮廓
	for i := 0; i < fragmentContours.Size(); i++ {
		//根据轮廓裁切图片(要从反转的Mat中裁切，而不是从模糊后的Mat中裁切，模糊的Mat只是为了提取轮廓)
		fragmentMat := reverseMat.Region(gocv.BoundingRect(fragmentContours.At(i)))
		//905和940是根据实际情况调整的，根据实际情况调整（这个长度的且含有APP字符串是为了裁切出含有类似【f:/-lWExrk9cpkEUw复制并打开拼多多APP】的片段图片)
		if fragmentMat.Cols() > 905 && fragmentMat.Cols() < 940 && p.tesseract.DetectContains(&fragmentMat, 6, "APP") {
			//输出裁切后的图片到fragment文件夹
			name := fmt.Sprintf("%s/%d.jpg", p.fragmentDir, i)
			autoMkdirs(name)
			gocv.IMWrite(name, fragmentMat)
			p.fragment = append(p.fragment, &QuoteFragment{
				Dir: fmt.Sprintf("%s/%d", p.fragmentDir, i),
				Mat: &fragmentMat,
			})
		}
	}
	return nil
}
func autoMkdirs(file string) {
	if "" == file {
		return
	}
	lastIndex := strings.LastIndex(file, "/")
	if lastIndex != -1 {
		err := os.MkdirAll(file[:lastIndex], os.ModePerm)
		if err != nil {
			log.Printf("创建文件夹:【%s】失败：%v", file[:lastIndex], err)
		}
	}
}
