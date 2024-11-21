package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/rix4uni/haktrailsfree/banner"
)

func main() {
	// Parse command-line flags
	delay := flag.Int("delay", 3, "Delay between requests in seconds (not recommended to lower delay)")
	userAgent := flag.String("H", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36", "User-Agent header")
	cookieFilePath := flag.String("cf", "", "Path to cookie file (default: ~/.config/haktrailsfree/cookie.txt or ./cookie.txt)")
	version := flag.Bool("version", false, "Print the version of the tool and exit.")
	silent := flag.Bool("silent", false, "Silent mode.")
	flag.Parse()

	if *version {
		banner.PrintBanner()
		banner.PrintVersion()
		return
	}

	if !*silent {
		banner.PrintBanner()
	}

	// Determine cookie file path
	if *cookieFilePath == "" {
		// Check for ~/.config/haktrailsfree/cookie.txt
		homeDir, _ := os.UserHomeDir()
		defaultConfigPath := filepath.Join(homeDir, ".config", "haktrailsfree", "cookie.txt")
		if _, err := os.Stat(defaultConfigPath); err == nil {
			*cookieFilePath = defaultConfigPath
		} else if _, err := os.Stat("cookie.txt"); err == nil {
			// Fall back to ./cookie.txt
			*cookieFilePath = "cookie.txt"
		} else {
			fmt.Println("Error: cookie.txt file not found in ~/.config/haktrailsfree/cookie.txt and ./cookie.txt. use -cf flag for custom location.")
			return
		}
	}

	// Read cookies from the specified file
	cookieFile, err := os.Open(*cookieFilePath)
	if err != nil {
		fmt.Println("Error opening cookie file:", err)
		return
	}
	defer cookieFile.Close()

	cookies, err := io.ReadAll(cookieFile)
	if err != nil {
		fmt.Println("Error reading cookie file:", err)
		return
	}

	// Create a regex for extracting domains
	domainRegex := regexp.MustCompile(`href="/domain/([^/]+)`)

	// Create a map to store unique domains
	seenDomains := make(map[string]bool)

	// Process input domains
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		domain := scanner.Text()
		if domain == "" {
			continue
		}
		for page := 1; page <= 100; page++ {
			url := fmt.Sprintf("https://securitytrails.com/list/apex_domain/%s?page=%d", domain, page)

			// Create HTTP request
			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				fmt.Println("Error creating request:", err)
				continue
			}

			// Add headers
			req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
			req.Header.Set("Accept-Language", "en-US,en;q=0.9,hi;q=0.8,en-IN;q=0.7")
			req.Header.Set("Cache-Control", "max-age=0")
			req.Header.Set("DNT", "1")
			req.Header.Set("Priority", "u=0, i")
			req.Header.Set("Sec-CH-UA", `"Google Chrome";v="131", "Chromium";v="131", "Not_A Brand";v="24"`)
			req.Header.Set("Sec-CH-UA-Arch", `"x86"`)
			req.Header.Set("Sec-CH-UA-Bitness", `"64"`)
			req.Header.Set("Sec-CH-UA-Full-Version", `"131.0.6778.69"`)
			req.Header.Set("Sec-CH-UA-Full-Version-List", `"Google Chrome";v="131.0.6778.69", "Chromium";v="131.0.6778.69", "Not_A Brand";v="24.0.0.0"`)
			req.Header.Set("Sec-CH-UA-Mobile", `?0`)
			req.Header.Set("Sec-CH-UA-Model", `""`)
			req.Header.Set("Sec-CH-UA-Platform", `"Windows"`)
			req.Header.Set("Sec-CH-UA-Platform-Version", `"15.0.0"`)
			req.Header.Set("Sec-Fetch-Dest", "document")
			req.Header.Set("Sec-Fetch-Mode", "navigate")
			req.Header.Set("Sec-Fetch-Site", "none")
			req.Header.Set("Sec-Fetch-User", "?1")
			req.Header.Set("Upgrade-Insecure-Requests", "1")
			req.Header.Set("User-Agent", *userAgent)
			req.Header.Set("Cookie", string(cookies))

			// Perform the request
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				fmt.Println("Error performing request:", err)
				continue
			}
			defer resp.Body.Close()

			// Process the response
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Println("Error reading response body:", err)
				continue
			}

			// Extract domains and ensure uniqueness
			matches := domainRegex.FindAllStringSubmatch(string(body), -1)
			for _, match := range matches {
				extractedDomain := match[1]
				if !seenDomains[extractedDomain] {
					seenDomains[extractedDomain] = true
					fmt.Println(extractedDomain)
				}
			}

			// Delay between requests
			if *delay > 0 {
				time.Sleep(time.Duration(*delay) * time.Second)
			}
		}
	}

	if scanner.Err() != nil {
		fmt.Println("Error reading input:", scanner.Err())
	}
}
