package service

import (
	"log"
	"net/http"
	_ "strconv"
	"os"
	"io"
	"fmt"
)

func downloadImageFromUrl(imageUrl string,imageName string, path string) error {
	log.Printf("Attempting to get image %s", imageUrl)
	response, err := http.Get(imageUrl)
	defer response.Body.Close()
	if err != nil {
		log.Print(err)
		return err
	}
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("Error while getting file %v",response.Status)
	}
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
