package x

import (
	"errors"
	"github.com/kanechoo/pdd-screenshot-link/tesseract"
	"gocv.io/x/gocv"
	"image"
)

type XProcessor struct {
	//原始截图文件路径
	OriginImg string
	//提取后需要分析的片段图片，图片和名称将和文件夹名称一致
	fragment []*gocv.Mat
	//文件夹
	dirs []string
	//文件夹和文件夹的mat对应关系的映射
	dirMatMap map[string][]*gocv.Mat
	//tesseract
	tesseract *tesseract.Tesseract
}

func newXProcessor(originImg string) *XProcessor {
	return &XProcessor{
		OriginImg: originImg,
		tesseract: tesseract.New(),
		dirs:      make([]string, 0),
		dirMatMap: make(map[string][]*gocv.Mat),
		fragment:  make([]*gocv.Mat, 0),
	}
}
func (p *XProcessor) GetFragment() ([]*gocv.Mat, error) {
	if p.fragment != nil || len(p.fragment) > 0 {
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
	blurMat := gocv.NewMat()
	blurStructuring := gocv.GetStructuringElement(gocv.MorphRect, image.Point{X: 30, Y: 10})
	//模糊结构，把分离的字体例如【i】之类的上面和下面链接到一起
	gocv.MorphologyEx(reverseMat, &blurMat, gocv.MorphClose, blurStructuring)
	//从模糊结构后的Mat中提取轮廓
	fragmentContours := gocv.FindContours(blurMat, gocv.RetrievalTree, gocv.ChainApproxSimple)
	//遍历轮廓
	for i := 0; i < fragmentContours.Size(); i++ {
		//根据轮廓裁切图片(要从反转的Mat中裁切，而不是从模糊后的Mat中裁切，模糊的Mat只是为了提取轮廓)
		fragmentMat := reverseMat.Region(gocv.BoundingRect(fragmentContours.At(i)))
		//905和940是根据实际情况调整的，根据实际情况调整（这个长度的且含有APP字符串是为了裁切出含有类似【f:/-lWExrk9cpkEUw复制并打开拼多多APP】的片段图片)
		if fragmentMat.Cols() > 905 && fragmentMat.Cols() < 940 && p.tesseract.DetectContains(&fragmentMat, 6, "APP") {
			p.fragment = append(p.fragment, &fragmentMat)
		}
	}
	return nil
}
