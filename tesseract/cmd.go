package tesseract

import (
	"fmt"
	"gocv.io/x/gocv"
	"os"
	"os/exec"
	"strings"
	"time"
)

const folder = "images"

type Tesseract struct {
	psm int
}

func New() *Tesseract {
	return &Tesseract{
		psm: 13,
	}
}

func (t *Tesseract) setPsm(psm int) {
	t.psm = psm
}

func (t *Tesseract) Detect(imageFile string, psm int) (string, error) {
	var cmd *exec.Cmd
	cmd = exec.Command("tesseract", imageFile, "stdout", "-l", "eng", "--psm", fmt.Sprintf("%d", psm))
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}
func (t *Tesseract) DetectFromMat(mat *gocv.Mat, psm int, remove bool) (string, error) {
	random := time.Now().UnixMilli()
	img := fmt.Sprintf("%s/%d.jpg", folder, random)
	gocv.IMWrite(img, *mat)
	text, err := t.Detect(img, psm)
	if remove {
		cleanImage(img)
	}
	return text, err
}
func (t *Tesseract) DetectContains(mat *gocv.Mat, psm int, subtext string) bool {
	text, err := t.DetectFromMat(mat, psm, true)
	if err != nil {
		return false
	}
	return strings.Contains(text, subtext)
}
func cleanImage(img string) bool {
	err := os.Remove(img)
	if err != nil {
		return false
	}
	return true
}
