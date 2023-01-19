package image

import (
	mathUtil "Projet/utils"
	"image"
	"math"
)

func Filter(img *image.RGBA, i1, i2 uint, red, green, blue, alpha float64) { //Use percentages
	for i := i1; i < i2; i++ {
		switch i % 4 {
		case 0:
			if red >= 0 {
				img.Pix[i] = uint8(mathUtil.Clamp(int(math.Round(float64(img.Pix[i])*red)), 0, 255))
			}
		case 1:
			if green >= 0 {
				img.Pix[i] = uint8(mathUtil.Clamp(int(math.Round(float64(img.Pix[i])*green)), 0, 255))
			}
		case 2:
			if blue >= 0 {
				img.Pix[i] = uint8(mathUtil.Clamp(int(math.Round(float64(img.Pix[i])*blue)), 0, 255))
			}
		case 3:
			if alpha >= 0 {
				img.Pix[i] = uint8(mathUtil.Clamp(int(math.Round(float64(img.Pix[i])*alpha)), 0, 255))
			}
		}
	}
}
