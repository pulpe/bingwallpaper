package main

import (
	"bufio"
	"io"
	"log"
	"math/rand/v2"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

func getWalls() ([]string, map[string]string, error) {
	resp, err := http.Get("https://raw.githubusercontent.com/niumoo/bing-wallpaper/main/bing-wallpaper.md")
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	wallKeys := make([]string, 0)
	walls := make(map[string]string)

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "|", 2)
		if len(parts) != 2 {
			continue
		}

		wallDate := strings.TrimSpace(parts[0])

		re := regexp.MustCompile(`\[([^\]]+)\]\((https?://[^\)]+)\)`)
		matches := re.FindStringSubmatch(parts[1])
		if len(matches) != 3 {
			continue
		}

		// wallDesc := matches[1]
		wallUrl := matches[2]

		walls[wallDate] = wallUrl
		wallKeys = append(wallKeys, wallDate)
	}

	return wallKeys, walls, nil
}

func setWallPlasma(wallFile string) error {
	cmd := exec.Command("plasma-apply-wallpaperimage", wallFile)
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func main() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	wallDir := filepath.Join(homeDir, ".local", "share", "bingwallpaper")

	err = os.MkdirAll(wallDir, 0755)
	if err != nil {
		log.Fatal(err)
	}

	wallKeys, walls, err := getWalls()
	if err != nil {
		log.Fatal(err)
	}

	var d, w string

	args := os.Args
	if len(args) > 1 && args[1] == "--random" {
		d = wallKeys[rand.IntN(len(wallKeys))]
		w = walls[d]
	} else {
		d = time.Now().Format(time.DateOnly)
		w = walls[d]
	}

	wallFile := filepath.Join(wallDir, d+".jpg")
	_, err = os.Stat(wallFile)
	if err != nil {
		out, err := os.Create(wallFile)
		if err != nil {
			log.Fatal(err)
		}
		defer out.Close()

		resp, err := http.Get(w)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()

		_, err = io.Copy(out, resp.Body)
		if err != nil {
			log.Fatal(err)
		}
	}

	// TODO: windows & different linux DEs
	err = setWallPlasma(wallFile)
	if err != nil {
		log.Fatal(err)
	}

	entries, err := os.ReadDir(wallDir)
	if err != nil {
		log.Fatal(err)
	}

	for _, entry := range entries {
		if entry.Name() != d+".jpg" {
			err := os.Remove(filepath.Join(wallDir, entry.Name()))
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}
