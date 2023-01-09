package image

import (
	"image"
)

func Filter(img *image.NRGBA, red, green, blue, alpha int) { //Use percentages
	for i, _ := range img.Pix {
		switch i % 4 {
		case 0:
			if red >= 0 {
				img.Pix[i] = uint8(red)
			}
		case 1:
			if green >= 0 {
				img.Pix[i] = uint8(green)
			}
		case 2:
			if blue >= 0 {
				img.Pix[i] = uint8(blue)
			}
		case 3:
			if alpha >= 0 {
				img.Pix[i] = uint8(alpha)
			}
		}
	}
}

func RedFilter(img *image.NRGBA) {
	for i, _ := range img.Pix {
		if i%4 != 3 && i%4 != 0 {
			img.Pix[i] = 0
		}
	}
}

func BlueFilter(img *image.NRGBA) {
	for i, _ := range img.Pix {
		if i%4 != 3 && i%4 != 0 {
			img.Pix[i] = 0
		}
	}
}

func GreenFilter(img *image.NRGBA) {
	for i, _ := range img.Pix {
		if i%4 != 3 && i%4 != 0 {
			img.Pix[i] = 0
		}
	}
}
