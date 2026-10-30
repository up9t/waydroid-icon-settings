package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

var logger = slog.New(slog.NewTextHandler(os.Stdout, nil))

func main() {
	start := time.Now()

	logger.Info("Start!")

	defer func() {
		logger.Info("Done!",
			"elapsed", time.Since(start),
		)
	}()

	hideFlag := flag.Bool("hide", true, "")

	flag.Parse()

	expectHidden := *hideFlag

	logger.Info("Set hiden",
		"hide", expectHidden,
	)

	home, err := os.UserHomeDir()
	// home := os.Getenv("HOME") // this is on linux

	if err != nil {
		logger.Error("Failed to retrieve home directory",
			"reason", err,
		)
		return
	}

	directory := filepath.Join(home, ".local/share/applications")
	entries, err := os.ReadDir(directory)

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			logger.Info("Directory not found, nothing to do")
		} else {
			logger.Error("Failed to read destination directory",
				"directory", directory,
				"reason", err,
			)
			return
		}
	}

	targetPrefix := "waydroid"

	targetFilenames := make(chan string, 10)

	// workerCount := 4
	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		for filename := range targetFilenames {
			path := filepath.Join(directory, filename)
			file, err := os.OpenFile(path, os.O_RDONLY, 0)

			if err != nil {
				logger.Error("Failed to open file",
					"filename", path,
					"reason", err,
				)
				continue
			}

			scanner := bufio.NewScanner(file)

			builder := strings.Builder{}

			start := false
			update := false
			hasDetected := false

			for scanner.Scan() {
				if err := scanner.Err(); err != nil {
					logger.Error("Failed to scan line",
						"reason", err,
					)
					break
				}

				line := strings.TrimSpace(scanner.Text())

				if line == "[Desktop Entry]" {
					start = true
				}

				// remove top lines if any
				if !start && line == "" {
					continue
				}

				// if NoDisplay entry found, modify the line
				// if not found until the "" then create new entry for NoDisplay
				if strings.HasPrefix(line, "NoDisplay") {
					s := strings.Split(line, "=")
					isHidden := s[1] == "true"
					update = expectHidden != isHidden
					hasDetected = true
				}

				if update {
					update = false
					s := fmt.Sprintf("NoDisplay=%s\n", strconv.FormatBool(expectHidden))
					builder.WriteString(s)
					continue
				}

				if !hasDetected && line == "" {
					hasDetected = true
					s := fmt.Sprintf("NoDisplay=%s\n", strconv.FormatBool(expectHidden))
					builder.WriteString(s)
				}

				builder.WriteString(line)
				builder.WriteString("\n")
			}
			file.Close()

			// rewrite step
			file, err = os.OpenFile(path, os.O_WRONLY|os.O_TRUNC, 0)

			if err != nil {
				logger.Error("Failed to open file",
					"filename", path,
					"reason", err,
				)
				return
			}
			_, err = file.WriteString(builder.String())

			if err != nil {
				logger.Error("Failed to write file",
					"reason", err,
				)
				file.Close()
			}

			file.Close()
		}
	}()

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()

		if strings.HasPrefix(filename, targetPrefix) {
			targetFilenames <- filename
		}
	}

	close(targetFilenames)

	wg.Wait()
}
