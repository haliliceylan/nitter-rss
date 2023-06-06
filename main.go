package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
	"regexp"
)

var (
	hostFlag = flag.String("host", "", "HTTP server IP address")
	portFlag = flag.Int("port", 8080, "HTTP server port")
	httpFlag = flag.Bool("http", false, "Run as HTTP server")
	fileFlag = flag.String("file", "", "File to write RSS content to")
)

var nitterHosts = []string{
	
	"nitter.us.projectsegfau.lt",
	
	"nitter.in.projectsegfau.lt",
}

func main() {
	flag.Parse()

	if *httpFlag {
		http.HandleFunc("/", handleUsername)

		address := fmt.Sprintf("%s:%d", *hostFlag, *portFlag)
		fmt.Printf("Listening on %s\n", address)

		err := http.ListenAndServe(address, nil)
		if err != nil {
			fmt.Printf("Failed to start HTTP server: %s\n", err.Error())
		}
	} else {
		if flag.NArg() == 0 {
			fmt.Fprintln(os.Stderr, "username argument is required")
			os.Exit(1)
		}

		username := flag.Arg(0)

		body, err := fetchRss(username)
		if err != nil {
			fmt.Printf("Failed to fetch RSS feed: %s\n", err.Error())
			os.Exit(1)
		}

		if *fileFlag == "" {
			fmt.Println(string(body))
		} else {
			err = ioutil.WriteFile(*fileFlag, body, 0644)
			if err != nil {
				fmt.Printf("Failed to write RSS content to file: %s\n", err.Error())
				os.Exit(1)
			}

			fmt.Printf("RSS content written to %s\n", *fileFlag)
		}
	}
}

func handleUsername(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Path[1:]

	body, err := fetchRss(username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/rss+xml")
	w.Write(body)
}

func fetchRss(username string) ([]byte, error) {
	var body []byte
	var err error
	var hostIndex int
	for i := 0; i < len(nitterHosts); i++ {
		hostIndex = (getNitterIndex() + i) % len(nitterHosts)
		host := nitterHosts[hostIndex]

		url := fmt.Sprintf("https://%s/%s/rss", host, username)
		resp, err := http.Get(url)
		if err != nil {
			// Try next host on error
			continue
		}

		if resp.StatusCode != http.StatusOK {
			// Try next host on error status
			resp.Body.Close()
			continue
		}

		defer resp.Body.Close()
		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			// Try next host on error
			continue
		}

		break // Success!
	}

	if err != nil {
		return nil, err
	}

	// Replace nitter host with Twitter host in status URLs
	bodyStr := string(body)
	bodyStr = strings.ReplaceAll(bodyStr, nitterHosts[hostIndex], "twitter.com")

	// experimental
	m := regexp.MustCompile(`\/\/.*\/(\w+)\/status\/(\d+)`)
	Str := "//twitter.com/${1}/status/$2"
	bodyStr = m.ReplaceAllString(bodyStr, Str)

	return []byte(bodyStr), nil
}

func getNitterIndex() int {
	currentTime := time.Now().Unix()
	periodDuration := int64(len(nitterHosts) * 300) // 300 seconds = 5 minutes
	periodIndex := currentTime / periodDuration
	return int(periodIndex % int64(len(nitterHosts)))
}
