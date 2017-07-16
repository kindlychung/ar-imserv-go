package main

import (
	"net/http"
	"fmt"
	"time"
	"crypto/md5"
	"io"
	"strconv"
	"os"
	"html/template"
	"log"
	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/tiff"
	"github.com/disintegration/imaging"
	"image/color"
	"image"
	"github.com/satori/go.uuid"
)

func uploadJPG(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method:", r.Method)
	if r.Method == "GET" {
		crutime := time.Now().Unix()
		h := md5.New()
		io.WriteString(h, strconv.FormatInt(crutime, 10))
		token := fmt.Sprintf("%x", h.Sum(nil))

		t, _ := template.ParseFiles("uploadJPG.gtpl")
		t.Execute(w, token)
	} else {
		r.ParseMultipartForm(32 << 20)
		fmt.Printf("%v", r)
		file, _, err := r.FormFile("uploadFile")
		if err != nil {
			fmt.Println(err)
			return
		}
		defer file.Close()
		var outputFileRoot string = uuid.NewV4().String();
		var thumbnailFileRoot string = outputFileRoot + "_thumbnail"
		var fileDir = "./file_store/"
		var outputFile = fileDir + outputFileRoot + ".jpg"
		var thumbnailFile = fileDir + thumbnailFileRoot + ".jpg"
		f, err := os.OpenFile(outputFile, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer f.Close()
		io.Copy(f, file)
		processImg(outputFile, thumbnailFile)
		w.Write([]byte("snapshots/" + outputFileRoot + ".jpg"))
	}
}

func getOrientationType(filename string) tiff.DataType {
	f, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	x, err := exif.Decode(f)
	if err != nil {
		log.Fatal(fmt.Sprintf(`Decoding: %v`, err))
	}
	orientation, err := x.Get(exif.Orientation)
	if err != nil {
		return 1
	}
	return orientation.Type
}

func correctOrientation(img *image.Image, orientation tiff.DataType) bool {
	switch orientation {
	case 2:
		log.Println("Horizontally flipped.")
		*img = imaging.FlipH(*img)
	case 3:
		log.Println("Rotated 180 degrees.")
		*img = imaging.Rotate(*img, 180, color.Transparent)
	case 4:
		log.Println("Vertically flipped.")
		*img = imaging.FlipV(*img)
	case 5:
		log.Println("Rotated 90 degrees clockwise, then horizontall flipped.")
		*img = imaging.Rotate(imaging.FlipH(*img), 90, color.Transparent)
	case 6:
		log.Println("Rotated 90 degrees counter-clockwise.")
		*img = imaging.Rotate(*img, -90, color.Transparent)
	case 7:
		log.Println("Rotated 90 degrees counter-clockwise, then horizontall flipped.")
		*img = imaging.Rotate(imaging.FlipH(*img), -90, color.Transparent)
	case 8:
		log.Println("Rotated 90 degrees clockwise.")
		*img = imaging.Rotate(*img, 90, color.Transparent)
	default:
		log.Println("Upright position, or orientation info not available.")
		return false
	}
	return true
}

func createThumbnail(img *image.Image, height int) {
	log.Println("Creating thumnail...")
	rect := (*img).Bounds()
	w := float64(rect.Max.X)
	h := float64(rect.Max.Y)
	width := int((w * float64(height)) / h)
	*img = imaging.Resize(*img, width, height, imaging.Lanczos)
}

func processImg(src string, thumbnail string) bool {
	img, err := imaging.Open(src)
	if err != nil {
		log.Fatalf("Open failed: %v", err)
	}
	orientation := getOrientationType(src)
	neededCorrection := correctOrientation(&img, orientation)
	// Save the resulting image using JPEG format.
	if neededCorrection {
		err = imaging.Save(img, src)
		if err != nil {
			log.Fatalf("Save failed: %v", err)
		}
	}
	createThumbnail(&img, 48)
	imaging.Save(img, thumbnail)
	return neededCorrection
}

func main() {
	http.HandleFunc("/uploadJPG", uploadJPG)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

