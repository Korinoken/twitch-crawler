package main

import (
	"encoding/json"
	"github.com/urfave/cli"
	"io/ioutil"
	"log"
	"os"
	cr "twitch-crawler/app"
)

func main() {
	app := cli.NewApp()
	app.Name = "TwitchCrawler"
	app.Usage = "Starting Twitch crawler app"

	app.Commands = []cli.Command{
		{
			Name:  "start",
			Usage: "Start crawler",
			Action: func(c *cli.Context) error {
				data, err := ioutil.ReadFile("config.json")
				if err != nil {
					log.Fatal(err)
					return err
				}
				crawler := cr.TwitchCrawler{}
				json.Unmarshal(data, &crawler)
				log.Printf("Loaded config: %v", crawler)
				list, err := crawler.GetImageList()
				if err != nil {
					log.Fatal(err)
					return err
				}
				err = crawler.SaveImages(list)
				if err != nil {
					log.Fatal(err)
					return err
				}
				return nil
			},
		},
	}
	app.Run(os.Args)
}
