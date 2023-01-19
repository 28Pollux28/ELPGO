package image

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
)

func clonePix(b []uint8) []byte {
	c := make([]uint8, len(b))
	copy(c, b)
	return c
}

func SaveImage(filePath string, img *image.RGBA) {
	imgFile, err := os.Create(filePath)
	defer imgFile.Close()
	if err != nil {
		fmt.Println("Error happened when creating file "+filePath+" :", err)
		return
	}
	err2 := png.Encode(imgFile, img.SubImage(img.Rect))
	if err2 != nil {
		fmt.Println("Error happened when encoding image :", err)
		return
	}
}

func LoadImage(filePath string) *image.RGBA {
	imgFile, err := os.Open(filePath)
	defer imgFile.Close()
	if err != nil {
		fmt.Println("Error happened when opening file "+filePath+" :", err)
		return nil
	}
	img, _, err2 := image.Decode(imgFile)
	//png.Encode(os.Stdout, img)
	//convert := color.NRGBAModel.Convert(img.At(0, 0))
	//get the nrgba of the img at the 0,0
	//fmt.Println(convert, convert[1], convert[2], convert[3])

	//fmt.Println(r>>8, g>>8)
	if err2 != nil {
		fmt.Println("Error happened when parsing image :", err)
		return nil
	}
	return ConvertImageToModel(img, color.RGBAModel).(*image.RGBA)
}

func CloneImage(src image.Image) draw.Image {
	switch s := src.(type) {
	case *image.Alpha:
		clone := *s
		clone.Pix = clonePix(s.Pix)
		return &clone
	case *image.Alpha16:
		clone := *s
		clone.Pix = clonePix(s.Pix)
		return &clone
	case *image.Gray:
		clone := *s
		clone.Pix = clonePix(s.Pix)
		return &clone
	case *image.Gray16:
		clone := *s
		clone.Pix = clonePix(s.Pix)
		return &clone
	case *image.NRGBA:
		clone := *s
		clone.Pix = clonePix(s.Pix)
		return &clone
	case *image.NRGBA64:
		clone := *s
		clone.Pix = clonePix(s.Pix)
		return &clone
	case *image.RGBA:
		clone := *s
		clone.Pix = clonePix(s.Pix)
		return &clone
	case *image.RGBA64:
		clone := *s
		clone.Pix = clonePix(s.Pix)
		return &clone
	}
	return nil
}

func ConvertImageToModel(src image.Image, model color.Model) draw.Image {
	switch model {
	case color.AlphaModel:
		dst := image.NewAlpha(src.Bounds())
		draw.Draw(dst, src.Bounds(), src, src.Bounds().Min, draw.Src)
		return dst
	case color.Alpha16Model:
		dst := image.NewAlpha16(src.Bounds())
		draw.Draw(dst, src.Bounds(), src, src.Bounds().Min, draw.Src)
		return dst
	case color.GrayModel:
		dst := image.NewGray(src.Bounds())
		draw.Draw(dst, src.Bounds(), src, src.Bounds().Min, draw.Src)
		return dst
	case color.Gray16Model:
		dst := image.NewGray16(src.Bounds())
		draw.Draw(dst, src.Bounds(), src, src.Bounds().Min, draw.Src)
		return dst
	case color.NRGBAModel:
		dst := image.NewNRGBA(src.Bounds())
		draw.Draw(dst, src.Bounds(), src, src.Bounds().Min, draw.Src)
		return dst
	case color.NRGBA64Model:
		dst := image.NewNRGBA64(src.Bounds())
		draw.Draw(dst, src.Bounds(), src, src.Bounds().Min, draw.Src)
		return dst
	case color.RGBAModel:
		dst := image.NewRGBA(src.Bounds())
		draw.Draw(dst, src.Bounds(), src, src.Bounds().Min, draw.Src)
		return dst
	case color.RGBA64Model:
		dst := image.NewRGBA64(src.Bounds())
		draw.Draw(dst, src.Bounds(), src, src.Bounds().Min, draw.Src)
		return dst
	default:
		dst := image.NewRGBA(src.Bounds())
		draw.Draw(dst, src.Bounds(), src, src.Bounds().Min, draw.Src)
		return dst
	}

}
