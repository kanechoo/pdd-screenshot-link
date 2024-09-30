package tesseract

import (
	"fmt"
	"gocv.io/x/gocv"
	"log"
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

func (t *Tesseract) Detect(imageFile string, commands string) (string, error) {
	cmdStr := getCommandLine(&imageFile, commands)
	text, err := runThenGet(cmdStr)
	if err != nil {
		return "", err
	}
	text = strings.ReplaceAll(text, "\n", "")
	text = strings.TrimSpace(text)
	if "00" == text || "oo" == text || "0o" == text || "o0" == text {
		text = "8"
	} else if ("o" == text || "O" == text) && strings.Contains(*cmdStr, "--oem 0") {
		cmdStrV2 := getCommandLine(&imageFile, "--psm 13 --oem 1")
		newText, err := runThenGet(cmdStrV2)
		if err != nil {
			return "", err
		}
		newText = strings.ReplaceAll(newText, "\n", "")
		if "0" == newText {
			text = newText
		}
	} else if "J" == text && strings.Contains(*cmdStr, "--oem 0") {
		cmdStrV2 := getCommandLine(&imageFile, "--psm 13 --oem 1")
		newText, err := runThenGet(cmdStrV2)
		if err != nil {
			return "", err
		}
		newText = strings.ReplaceAll(newText, "\n", "")
		if "j" == newText {
			text = newText
		}
	} else if "9" == text && strings.Contains(*cmdStr, "--oem 0") && strings.Contains(*cmdStr, "--psm 13") {
		cmdStrV2 := getCommandLine(&imageFile, "--psm 13 --oem 1")
		newText, err := runThenGet(cmdStrV2)
		if err != nil {
			return "", err
		}
		newText = strings.ReplaceAll(newText, "\n", "")
		if "g" == newText {
			text = newText
		}
	} else if "i" == text && strings.Contains(*cmdStr, "--oem 0") && strings.Contains(*cmdStr, "--psm 13") {
		cmdStrV2 := getCommandLine(&imageFile, "--psm 13 --oem 1")
		newText, err := runThenGet(cmdStrV2)
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
func getCommandLine(imageFile *string, commands string) *string {
	dir := os.Getenv("TESSDATA_DIR")
	if dir == "" {
		dir = "/Users/konchoo/Downloads"
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s stdout", *imageFile))
	commands = " " + commands
	sb.WriteString(commands)
	if !strings.Contains(commands, "-l") {
		sb.WriteString(" -l eng")
	}
	if !strings.Contains(commands, "--psm") {
		sb.WriteString(" --psm 5")
	}
	if !strings.Contains(commands, "--oem") {
		sb.WriteString(" --oem 0")
	}
	if !strings.Contains(commands, "--tessdata-dir") && !strings.Contains(sb.String(), "--oem 1") {
		sb.WriteString(fmt.Sprintf(" --tessdata-dir %s", dir))
	}
	if !strings.Contains(sb.String(), "-c tessedit_char_whitelist") && strings.Contains(sb.String(), "-l eng") {
		sb.WriteString(" -c tessedit_char_whitelist=0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz")
	}
	s := sb.String()
	return &s
}
func runThenGet(cmdStr *string) (string, error) {
	var cmd *exec.Cmd
	var cmdArray []string
	for _, s := range strings.Split(*cmdStr, " ") {
		if "" != s {
			cmdArray = append(cmdArray, s)
		}
	}
	cmd = exec.Command("tesseract", cmdArray...)
	out, err := cmd.Output()
	if err != nil {
		log.Printf("运行命令【%s】失败：%v", strings.Join(cmdArray, " "), err)
		return "", err
	}
	return string(out), nil
}
func (t *Tesseract) DetectFromMat(mat *gocv.Mat, remove bool, commands string) (string, error) {
	random := time.Now().UnixMilli()
	img := fmt.Sprintf("%s/%d.jpg", folder, random)
	gocv.IMWrite(img, *mat)
	text, err := t.Detect(img, commands)
	if remove {
		_ = os.Remove(img)
	}
	return text, err
}
