package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jamesnetherton/m3u"
	"gopkg.in/vansante/go-ffprobe.v2"
)

func main() {
	url := os.Args[1]
	playlist, err := m3u.Parse(url)

	if err == nil {
		for _, track := range playlist.Tracks {
			fmt.Println("Track name:", track.Name)
			fmt.Println("Track URI:", track.URI)
			fmt.Println("Track Resolutions:")
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			data, err := ffprobe.ProbeURL(ctx, track.URI)
			if err != nil {
				fmt.Printf("Error getting data: %v\n", err)
			} else {
				for _, stream := range data.StreamType(ffprobe.StreamVideo) {
					fmt.Printf(" - %dx%d\n", stream.Width, stream.Height)
				}
			}
			fmt.Println("Track Tags:")
			for i := range track.Tags {
				fmt.Println(" -", track.Tags[i].Name, "=>", track.Tags[i].Value)
			}
			fmt.Println("------------------------------")
		}
	} else {
		fmt.Println(err)
	}
}
