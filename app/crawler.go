package service

import (
	"log"
	"net/http"
	_ "strconv"
	"os"
	"io"
)

func downloadImageFromUrl(imageUrl string,imageName string, path string) error {
	//Stub
	log.Printf("Attempting to get image %s", imageUrl)
	response, err := http.Get(imageUrl)
	if err != nil {
		log.Print(err)
		return err
	}
	defer response.Body.Close()

	file, err := os.Create(path + `\` + imageName)
	if err != nil {
		log.Print(err)
		return err
	}
	_, err = io.Copy(file, response.Body)
	defer file.Close()
	if err != nil {
		log.Print(err)
		return err
	}
	return nil
}
