package main

import (
	"github.com/schollz/progressbar/v3"
	"time"
)

func main() {
	bar := progressbar.Default(65535)
	for i:=0;i<65535 ;i++  {
		bar.Add(1)
		time.Sleep(1*time.Second)
	}
}
