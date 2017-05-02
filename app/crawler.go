package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	_ "strconv"
)

func main() {
	file, _ := os.Open("conf.json")
	decoder := json.NewDecoder(file)
	crawler := TwitchCrawler{}
	err := decoder.Decode(&crawler)
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Println(crawler)
}

type TwitchCrawler struct {
	ApiKey  string
	ApiLink string
}

func (crawler *TwitchCrawler) getImageList() (images map[string]string, err error) {
	client := &http.Client{}
	url := crawler.ApiLink + `/chat/emoticons`
	log.Printf("Url : %v",url)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Client-ID", crawler.ApiKey)
	req.Header.Set("Accept", "application/vnd.twitchtv.v5+json")
	response, err := client.Do(req)
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Error while getting list %v", response.Status)
	}
	bodyBytes, _ := ioutil.ReadAll(response.Body)
	bodyString := string(bodyBytes)
	log.Printf("Response %v", bodyString)
	log.Printf("error :%v", err)
	images = make(map[string]string)
	return images, err
}
func (crawler *TwitchCrawler) downloadImageFromUrl(imageUrl string, imageName string, path string) error {
	log.Printf("Attempting to get image %s", imageUrl)
	response, err := http.Get(imageUrl)
	defer response.Body.Close()
	if err != nil {
		log.Print(err)
		return err
	}
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("Error while getting file %v", response.Status)
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
