package photoProcessing

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"strconv"
	"strings"

	"github.com/minio/minio-go/v7"
	"golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"

	"github.com/Kost0/L3/internal/minIO"
	"github.com/Kost0/L3/internal/repository"

	"github.com/wb-go/wbf/dbpg"
)

func ProcessPhoto(client *minio.Client, photo *repository.Photo, db *dbpg.DB) error {
	if photo.ResizeTo != "" {
		parts := strings.Split(photo.ResizeTo, "x")

		width, err := strconv.Atoi(parts[0])
		if err != nil {
			photo.Status = "failed"
			return err
		}
		height, err := strconv.Atoi(parts[1])
		if err != nil {
			photo.Status = "failed"
			return err
		}

		newPhoto, err := resizePhoto(client, photo, width, height)
		if err != nil {
			photo.Status = "failed"
			return err
		}

		err = repository.UpdatePhotoData(db, newPhoto)
		if err != nil {
			return err
		}
	} else if photo.WatermarkText != "" {
		newPhoto, err := AddWatermark(client, photo)
		if err != nil {
			photo.Status = "failed"
			return err
		}

		err = repository.UpdatePhotoData(db, newPhoto)
		if err != nil {
			return err
		}
	} else {
		newPhoto, err := GenThumbnail(client, photo)
		if err != nil {
			photo.Status = "failed"
			return err
		}

		err = repository.UpdatePhotoData(db, newPhoto)
		if err != nil {
			return err
		}
	}

	return nil
}

func resizePhoto(client *minio.Client, photo *repository.Photo, width, height int) (*repository.Photo, error) {
	bucketName := "images"

	src, err := minIO.GetPhoto(client, bucketName, photo.UUID.String())
	if err != nil {
		return nil, err
	}

	img, format, err := image.Decode(src)
	if err != nil {
		return nil, err
	}

	dst := image.NewRGBA(image.Rect(0, 0, width, height))

	draw.NearestNeighbor.Scale(dst, dst.Bounds(), img, img.Bounds(), draw.Src, nil)

	buf, err := encodeImage(dst, format)
	if err != nil {
		return nil, err
	}

	contentType := fmt.Sprintf("image/%s", format)

	err = minIO.UploadFileFromReader(client, bucketName, photo.UUID.String(), buf, int64(buf.Len()), contentType)
	if err != nil {
		return nil, err
	}

	photo.Status = "done"

	return photo, nil
}

func GenThumbnail(client *minio.Client, photo *repository.Photo) (*repository.Photo, error) {
	return resizePhoto(client, photo, 300, 300)
}

func AddWatermark(client *minio.Client, photo *repository.Photo) (*repository.Photo, error) {
	bucketName := "images"

	src, err := minIO.GetPhoto(client, bucketName, photo.UUID.String())
	if err != nil {
		return nil, err
	}

	img, format, err := image.Decode(src)
	if err != nil {
		return nil, err
	}

	textColor := color.RGBA{R: 255, G: 255, B: 255, A: 200}
	face := basicfont.Face7x13

	b := img.Bounds()
	dst := image.NewRGBA(b)

	draw.Draw(dst, b, img, image.Point{}, draw.Src)

	d := &font.Drawer{
		Dst:  dst,
		Src:  image.NewUniform(textColor),
		Face: face,
	}

	textWidth := font.MeasureString(face, photo.WatermarkText).Ceil()
	//textHeight := face.Metrics().Height.Ceil()
	ascent := face.Metrics().Ascent.Ceil()

	x := b.Dx() - textWidth - 20
	y := b.Dy() - 20

	if x < 0 {
		x = 10
	}
	if y < ascent+10 {
		y = ascent + 10
	}

	d.Dot = fixed.Point26_6{
		X: fixed.I(x),
		Y: fixed.I(y),
	}

	d.DrawString(photo.WatermarkText)

	buf, err := encodeImage(dst, format)
	if err != nil {
		return nil, err
	}

	contentType := fmt.Sprintf("image/%s", format)

	err = minIO.UploadFileFromReader(client, bucketName, photo.UUID.String(), buf, int64(buf.Len()), contentType)
	if err != nil {
		return nil, err
	}

	photo.Status = "done"

	return photo, nil
}

func encodeImage(img image.Image, format string) (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)
	var err error

	switch format {
	case "png":
		err = png.Encode(buf, img)
	case "jpeg":
		err = jpeg.Encode(buf, img, &jpeg.Options{Quality: 90})
	default:
		return nil, fmt.Errorf("Wrong format: %s", format)
	}

	if err != nil {
		return nil, err
	}

	return buf, nil
}
