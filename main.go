package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/jamesnetherton/m3u"
	md "github.com/nao1215/markdown"
	"github.com/schollz/progressbar/v3"
	"github.com/urfave/cli/v2"
	"gopkg.in/vansante/go-ffprobe.v2"
)

var Version = "development"

func main() {
	app := &cli.App{
		Name:    "m3u-checker",
		Usage:   "Checks availability of streams in an m3u playlist and generates report",
		Version: Version,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "playlist",
				Required: true,
				Usage:    "Path (or URL) of m3u playlist file.",
			},
			&cli.StringFlag{
				Name:     "file",
				Required: false,
				Usage:    "Name of the report file. By default reports to stdout.",
			},
		},
		Action: func(cCtx *cli.Context) error {
			if _, err := exec.LookPath("ffprobe"); err != nil {
				return fmt.Errorf("ffprobe is not found in PATH")
			}

			url := cCtx.String("playlist")
			playlist, err := m3u.Parse(url)
			if err != nil {
				return err
			}

			bar := progressbar.NewOptions(len(playlist.Tracks),
				progressbar.OptionEnableColorCodes(true),
				progressbar.OptionSetDescription("Checking streams..."),
				progressbar.OptionSetTheme(progressbar.Theme{
					Saucer:        "[green]=[reset]",
					SaucerHead:    "[green]>[reset]",
					SaucerPadding: " ",
					BarStart:      "[",
					BarEnd:        "]",
				}))

			var rows [][]string
			for _, track := range playlist.Tracks {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				data, err := ffprobe.ProbeURL(ctx, track.URI)
				var resolution string
				if err != nil {
					resolution = "Not available"
				} else {
					var sb strings.Builder
					for index, stream := range data.StreamType(ffprobe.StreamVideo) {
						if stream.Width > 0 && stream.Height > 0 {
							if index > 0 {
								sb.WriteString(", ")
							}
							sb.WriteString(fmt.Sprintf("%dx%d", stream.Width, stream.Height))
						}
					}
					resolution = sb.String()
				}
				rows = append(rows, []string{track.Name, resolution, track.URI})
				bar.Add(1)
			}

			var markdown *md.Markdown
			if cCtx.IsSet("file") {
				f, err := os.Create(cCtx.String("file"))
				if err != nil {
					return err
				}
				defer f.Close()
				markdown = md.NewMarkdown(f)
			} else {
				markdown = md.NewMarkdown(os.Stdout)
			}

			markdown.
				H2(url).
				CustomTable(md.TableSet{
					Header: []string{"Name", "Resolutions", "URL"},
					Rows:   rows,
				},
					md.TableOptions{
						AutoWrapText:      false,
						AutoFormatHeaders: true,
					},
				).
				Build()

			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
