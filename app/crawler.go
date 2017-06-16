package app

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"
)

type TwitchImageMeta struct {
	Regex  string
	Images map[string]interface{}
}
type doneFunc func()

func (crawler *TwitchCrawler) spawnWorkers(limit uint, proc doneFunc) (chan<- DownloadImageData, chan error) {
	c := make(chan DownloadImageData)
	res := make(chan error)

	for i := uint(0); i < limit; i++ {
		go func() {
			for {
				a, ok := <-c
				if !ok {
					proc()
					return
				}
				crawler.DownloadImageFromUrl(a, res)
			}
		}()
	}
	return c, res
}
func (crawler *TwitchCrawler) SaveImages(imageList map[string][]TwitchImageMeta) error {
	var wg sync.WaitGroup
	wg.Add(20)
	fn := func() {
		wg.Done()
	}
	c, res := crawler.spawnWorkers(20, fn)

	var imagesBeingProcessed int = 0
	for channelName, imageMetaSlice := range imageList {
		err := os.MkdirAll(filepath.FromSlash(crawler.WorkFolder+"/"+channelName), 0777)
		if err != nil {
			return err
		}
		for _, imageMeta := range imageMetaSlice {
			imageLink, err := url.Parse(imageMeta.Images["url"].(string))
			if err != nil {
				return err
			}
			log.Printf("link :%v", imageLink)
			imageExtension := path.Base(imageLink.Path)
			imageExtension = filepath.Ext(imageExtension)
			data := DownloadImageData{ImageUrl: imageMeta.Images["url"].(string),
				Path:      crawler.WorkFolder + "/" + channelName,
				ImageName: imageMeta.Regex + imageExtension}
			c <- data
			imagesBeingProcessed++
		}
	}
	for ind := imagesBeingProcessed; ind > 0; ind-- {
		err := <-res
		if err != nil {
			log.Printf("Err from channel:%v", err)
			return err
		}
	}
	close(c)
	close(res)
	return nil
}

type TwitchCrawler struct {
	ApiKey            string
	ApiLink           string
	ChannelsToProcess map[int]string
	WorkFolder        string
	CacheFileName     string
}

func (crawler *TwitchCrawler) GetImageList() (images map[string][]TwitchImageMeta, err error) {
	stat, err := os.Stat(crawler.CacheFileName)
	var bodyBytes []byte
	if err != nil || stat.ModTime().AddDate(0, 0, 1).Before(time.Now()) {
		if err != nil {
			log.Printf("Making api call, err:%v", err)
			err = nil
		} else {
			log.Printf("Making api call, mod time:%v", stat.ModTime())
		}
		client := &http.Client{}
		apiUrl := crawler.ApiLink + "/chat/emoticons"
		req, _ := http.NewRequest("GET", apiUrl, nil)
		req.Header.Set("Client-ID", crawler.ApiKey)
		req.Header.Set("Accept", "application/vnd.twitchtv.v5+json")
		response, err := client.Do(req)
		if response.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("Error while getting list %v", response.Status)
		}
		if err != nil {
			return nil, err
		}
		bodyBytes, _ = ioutil.ReadAll(response.Body)
		file, err := os.Create(crawler.CacheFileName)
		file.Write(bodyBytes)
		log.Print("Wrote response to a file")
	} else {
		log.Print("Using saved file")
		bodyBytes, err = ioutil.ReadFile(crawler.CacheFileName)
		if err != nil {
			return nil, err
		}
	}

	var f map[string][]interface{}
	images = make(map[string][]TwitchImageMeta)
	json.Unmarshal(bodyBytes, &f)
	//NA marshaling
	for _, imageMeta := range f["emoticons"] {
		converted := imageMeta.(map[string]interface{})
		meta := TwitchImageMeta{}
		meta.Regex = converted["regex"].(string)
		f4 := converted["images"].([]interface{})
		meta.Images = f4[0].(map[string]interface{})
		name, ok := crawler.ChannelsToProcess[int(meta.Images["emoticon_set"].(float64))]
		if ok {
			images[name] = append(images[name], meta)
		}
	}
	return images, err
}

type DownloadImageData struct {
	ImageUrl  string
	ImageName string
	Path      string
}

func (crawler *TwitchCrawler) DownloadImageFromUrl(data DownloadImageData, ch chan error) {
	log.Printf("Attempting to get image %s", data.ImageUrl)
	response, err := http.Get(data.ImageUrl)
	defer response.Body.Close()
	if err != nil {
		log.Print(err)
		ch <- err
		return
	}
	if response.StatusCode != http.StatusOK {
		ch <- fmt.Errorf("Error while getting file %v", response.Status)
		return
	}
	file, err := os.Create(filepath.FromSlash(data.Path + "/" + data.ImageName))
	if err != nil {
		log.Print(err)
		ch <- err
		return
	}
	_, err = io.Copy(file, response.Body)
	defer file.Close()
	if err != nil {
		log.Print(err)
		ch <- err
		return
	}
	ch <- nil
	return
}
