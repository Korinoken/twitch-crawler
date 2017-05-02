package main

import (
	"github.com/stretchr/testify/suite"
	"log"
	"net/http/httptest"
	"testing"
	"net/http"
	"fmt"
	"io"
	"os"
	"strconv"
)

type CrawlerTestSuite struct {
	suite.Suite
	crawler                       TwitchCrawler
	server                        httptest.Server
	VariableThatShouldStartAtFive int
	testDataLocation              string
}

func (suite *CrawlerTestSuite) SetupSuite() {
	log.Print("Setting up suite")
	suite.testDataLocation = `c:\test\images`
	suite.server = *httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			suite.handleClient(w, r)
		}))
	suite.crawler = TwitchCrawler{"key",suite.server.URL}//test config
}

var (
	testImages = []string{
		"image1.png",
		"image2.png",
	}
)

func (suite *CrawlerTestSuite) TestDownloadImageFromUrl() {
	fileEndLocation := suite.testDataLocation + `\download`
	for _, testImage := range testImages {
		err := suite.crawler.downloadImageFromUrl(suite.server.URL+"?file="+testImage, testImage,
			fileEndLocation)
		suite.Assert().Nil(err)
		_, err = os.Stat(fileEndLocation + `\` + testImage)//check that file exists
		suite.Assert().Nil(err)
	}
}
func (suite *CrawlerTestSuite) TestGetImageList() {
	res, err := suite.crawler.getImageList()
	suite.Assert().NotEmpty(res)
	suite.Assert().Nil(err)
}
func TestModelTestSuite(t *testing.T) {
	suite.Run(t, new(CrawlerTestSuite))
}
func (suite *CrawlerTestSuite) TearDownSuite() {
	log.Print("Closing test server")
	//suite.server.Close()
}

func (suite *CrawlerTestSuite) handleClient(writer http.ResponseWriter, request *http.Request) {
	//First of check if Get is set in the URL
	Filename := request.URL.Query().Get("file")
	if Filename == "" {
		//Get not set, send a 400 bad request
		http.Error(writer, "Get 'file' not specified in url.", 400)
		return
	}
	fmt.Println("Client requests: " + Filename)

	//Check if file exists and open
	log.Printf("Attempting to get file %v", suite.testDataLocation+`\`+Filename)
	Openfile, err := os.Open(suite.testDataLocation + `\` + Filename)
	defer Openfile.Close() //Close after function return
	if err != nil {
		//File not found, send 404
		http.Error(writer, "File not found.", 404)
		return
	}

	//File is found, create and send the correct headers

	//Get the Content-Type of the file
	//Create a buffer to store the header of the file in
	FileHeader := make([]byte, 512)
	//Copy the headers into the FileHeader buffer
	Openfile.Read(FileHeader)
	//Get content type of file
	FileContentType := http.DetectContentType(FileHeader)

	//Get the file size
	FileStat, _ := Openfile.Stat()                     //Get info from file
	FileSize := strconv.FormatInt(FileStat.Size(), 10) //Get file size as a string

	//Send the headers
	writer.Header().Set("Content-Disposition", "attachment; filename="+Filename)
	writer.Header().Set("Content-Type", FileContentType)
	writer.Header().Set("Content-Length", FileSize)

	//Send the file
	//We read 512 bytes from the file already so we reset the offset back to 0
	Openfile.Seek(0, 0)
	io.Copy(writer, Openfile) //'Copy' the file to the client
	return
}
