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

func (t *Tesseract) Detect(imageFile string, psm int, oem int) (string, error) {
	dir := os.Getenv("TESSDATA_DIR")
	if dir == "" {
		dir = "/Users/konchoo/Downloads"
	}
	prefix := "%s stdout -l eng --psm %d --tessdata-dir %s --oem %d -c tessedit_char_whitelist=0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	cmdStr := fmt.Sprintf(prefix, imageFile, psm, dir, oem)
	text, err := runThenGet(cmdStr)
	if err != nil {
		return "", err
	}
	text = strings.ReplaceAll(text, "\n", "")
	text = strings.TrimSpace(text)
	if "00" == text {
		text = "8"
	} else if ("o" == text || "O" == text) && oem == 0 {
		cmdStr = fmt.Sprintf(prefix, imageFile, 13, dir, 1)
		newText, err := runThenGet(cmdStr)
		if err != nil {
			return "", err
		}
		newText = strings.ReplaceAll(newText, "\n", "")
		if "0" == newText {
			text = newText
		}
	} else if "J" == text && oem == 0 {
		cmdStr = fmt.Sprintf(prefix, imageFile, 13, dir, 1)
		newText, err := runThenGet(cmdStr)
		if err != nil {
			return "", err
		}
		newText = strings.ReplaceAll(newText, "\n", "")
		if "j" == newText {
			text = newText
		}
	}
	return text, nil
}
func runThenGet(cmdStr string) (string, error) {
	var cmd *exec.Cmd
	cmd = exec.Command("tesseract", strings.Split(cmdStr, " ")...)
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
	text, err := t.Detect(img, psm, 1)
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
