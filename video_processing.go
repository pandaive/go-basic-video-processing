package main

import (
	"image"
	"image/jpeg"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"sync"
)

func processVideo(timestamp string) {
	var path = "frames/" + timestamp + "/"

	//splitting video into images
	cmd := "ffmpeg -i " + path + "/file -f image2 " + path + "frames_%03d.jpg"
	_, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		log.Print(err)
	}

	//removing video
	err = os.Remove(path + "file")
	if err != nil {
		log.Print(err)
	}

	//creating channel and workers to handle concurrent processing of frames
	//I could put it in outside global function so the pool is for all requests together, but it would probably take me another evening
	//(I'll do it anyway because I'm curious if it works like this and if it speeds things up (and won't blow my CPU out on many requests))
	frameTasks := make(chan string, 64)
	nbWorkers := 10

	var wg sync.WaitGroup

	for i := 0; i < nbWorkers; i++ {
		wg.Add(1)
		go func() {
			for ft := range frameTasks {
				processFrame(ft)
			}
			wg.Done()
		}()
	}

	//read all frames and send them to processing workers
	files, _ := ioutil.ReadDir(path)
	for _, f := range files {
		frameTasks <- path + f.Name()
	}
	close(frameTasks)

	wg.Wait()

	//after processing of frames completed, put the video back
	//I see it produces bad quality, but I guess it's not that important for now - of course fixable, if I dig in
	cmd = "ffmpeg -i " + path + "frames_%03d.jpg -vcodec libx264 " + timestamp + ".mp4"
	_, err = exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		log.Print(err)
	}

	//removing whole directory
	err = os.RemoveAll(path)
	if err != nil {
		log.Print(err)
	}

	//let notification "system" know, that it's been processed
	notification <- timestamp

}

func processFrame(filename string) {
	imgfile, err := os.Open(filename)
	if err != nil {
		log.Print(err)
	}

	src, _, _ := image.Decode(imgfile)
	dimensions := src.Bounds()
	width, height := dimensions.Dx(), dimensions.Dy()

	// Create a new grayscale image
	gray := image.NewGray(image.Rectangle{image.Point{0, 0}, image.Point{width, height}})
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			gray.Set(x, y, src.At(x, y))
		}
	}

	imgfile.Close()
	outfile, _ := os.Create(filename)
	jpeg.Encode(outfile, gray, &jpeg.Options{80})
	outfile.Close()
}
