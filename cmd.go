package main

import "os/exec"

func tesseractToText(imageFile string) (string, error) {
	var cmd *exec.Cmd
	cmd = exec.Command("tesseract", imageFile, "stdout", "-l", "eng", "--psm", "6")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}
