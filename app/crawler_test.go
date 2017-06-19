package app_test

import (
	"crypto/rand"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"testing"
	"twitch-crawler/app"
)

type CrawlerTestSuite struct {
	suite.Suite
	crawler app.TwitchCrawler
	server  httptest.Server
}

func (suite *CrawlerTestSuite) SetupSuite() {
	log.Print("Setting up suite")
	suite.server = *httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			suite.handleClient(w, r)
		}))
	suite.crawler = app.TwitchCrawler{ApiKey: "key",
		ApiLink:           suite.server.URL + "/images",
		ChannelsToProcess: map[int]string{10590: "deana"},
		WorkFolder:        "C:/test/images",
		CacheFileName:     "twitchApiResponse.json"} //test config
	os.MkdirAll(filepath.FromSlash(suite.crawler.WorkFolder),0777)
}

var (
	testImages = []string{
		"emoticon-69119-src-0d4b23ce12767ed8-28x28.png",
		"image2.png",
	}
)

func (suite *CrawlerTestSuite) TestDownloadImageFromUrl() {
	fileEndLocation := filepath.FromSlash(suite.crawler.WorkFolder)
	ch := make(chan error, 2)
	for _, testImage := range testImages {
		suite.crawler.DownloadImageFromUrl(app.DownloadImageData{suite.server.URL + "?file=" + testImage, testImage,
			fileEndLocation}, ch)
		err := <-ch
		suite.Assert().Nil(err)
		err = os.Remove(filepath.FromSlash(fileEndLocation +
			"/" +
			testImage))

		suite.Assert().Nil(err)
	}
}
func (suite *CrawlerTestSuite) TestGetImageList() {
	//clean call
	os.Remove(suite.crawler.CacheFileName)
	res, err := suite.crawler.GetImageList()
	suite.Assert().Nil(err)
	suite.Assert().NotEmpty(res, "GetImageList result was empty")
	_, found := res["deana"]
	suite.Assert().True(found, "deana message data not found")
	_, found = res["cmv"]
	suite.Assert().False(found, "extra channel data found")
}
func (suite *CrawlerTestSuite) TestSaveImages() {
	res, err := suite.crawler.GetImageList()
	serverLink, _ := url.Parse(suite.server.URL)
	for _, imageMetaSlice := range res {
		for _, imageMeta := range imageMetaSlice {
			imageLink, err := url.Parse(imageMeta.Images["url"].(string))
			suite.Assert().Nil(err)
			imageLink.Host = serverLink.Host
			imageLink.Scheme = serverLink.Scheme
			query := imageLink.Query()
			query.Set("file", path.Base(imageLink.Path))
			imageLink.RawQuery = query.Encode()
			imageMeta.Images["url"] = imageLink.String()
			log.Printf("new url : %v", imageLink)
		}
	}
	err = suite.crawler.SaveImages(res)
	suite.Assert().Nil(err)
	for channelName, imageMetaSlice := range res {
		for _, imageMeta := range imageMetaSlice {
			_, err = os.Stat(filepath.FromSlash(suite.crawler.WorkFolder +
				"/" + channelName +
				"/" + imageMeta.Regex +
				`.png`))
			suite.Assert().Nil(err)
		}
	}
}
func TestModelTestSuite(t *testing.T) {
	suite.Run(t, new(CrawlerTestSuite))
}
func (suite *CrawlerTestSuite) TearDownSuite() {
	log.Print("Closing test server")
	os.Remove(suite.crawler.WorkFolder)
	//suite.server.Close()
}

func (suite *CrawlerTestSuite) handleClient(response http.ResponseWriter, request *http.Request) {
	log.Printf("Received request %v", request.URL.Path)
	if request.URL.Query().Get("file") == "" {
		handleImageStructureApiRequest(response)
	} else {
		handleFileRequest(response, request)
	}
}
func handleFileRequest(response http.ResponseWriter, request *http.Request) {
	//First of check if Get is set in the URL
	Filename := request.URL.Query().Get("file")
	if Filename == "" {
		//Get not set, send a 400 bad request
		http.Error(response, "Get 'file' not specified in url.", 400)
		return
	}
	//Check if file exists and open
	log.Printf("Attempting to get file %v", Filename)
	File := make([]byte, 10000)
	rand.Read(File)
	FileSize := strconv.FormatInt(10000, 10) //Get file size as a string

	//Send the headers
	response.Header().Set("Content-Disposition", "attachment; filename="+Filename)
	response.Header().Set("Content-Type", "image/png")
	response.Header().Set("Content-Length", FileSize)
	response.Write(File)
	return
}
func handleImageStructureApiRequest(response http.ResponseWriter) {
	body, err := ioutil.ReadFile("testResponse.json")
	if err != nil {
		log.Printf("Error while creating test response: %v", err)
	}
	response.Write(body)
}
