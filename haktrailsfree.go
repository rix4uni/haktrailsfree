package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/pflag"
	"github.com/rix4uni/haktrailsfree/banner"
)

const (
	maxPages = 100
	delay    = 3400 * time.Millisecond
)

type Config struct {
	Domain    string
	ListFile  string
	CookieFile string
	UserAgent string
	Cookie    string
	silent   bool
	version   bool
	verbose   bool
}

func main() {
	config := parseFlags()
	
	// Load cookie and user agent from file
	if err := loadCookieAndUserAgent(config); err != nil {
		fmt.Fprintf(os.Stderr, "Error loading cookie file: %v\n", err)
		os.Exit(1)
	}

	domains := getDomains(config)
	
	for _, domain := range domains {
		processDomain(domain, config)
	}
}

func parseFlags() *Config {
	config := &Config{}
	
	pflag.StringVarP(&config.Domain, "domain", "d", "", "Single domain to process")
	pflag.StringVarP(&config.ListFile, "list", "l", "", "File containing list of domains")
	pflag.StringVarP(&config.CookieFile, "cookiefile", "c", "cookie.txt", "File containing curl command with cookies")
	pflag.BoolVar(&config.silent, "silent", false, "Silent mode.")
	pflag.BoolVar(&config.version, "version", false, "Print the version of the tool and exit.")
	pflag.BoolVar(&config.verbose, "verbose", false, "Enable verbose output for debugging purposes.")
	
	pflag.Parse()

	if config.version {
		banner.PrintBanner()
		banner.PrintVersion()
		os.Exit(0)
	}

	if !config.silent {
		banner.PrintBanner()
	}

	return config
}

func loadCookieAndUserAgent(config *Config) error {
	file, err := os.Open(config.CookieFile)
	if err != nil {
		return fmt.Errorf("failed to open cookie file: %v", err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("error reading cookie file: %v", err)
	}

	contentStr := string(content)

	// Extract user agent using regex
	userAgentRegex := regexp.MustCompile(`-H 'user-agent: ([^']*)'`)
	userAgentMatches := userAgentRegex.FindStringSubmatch(contentStr)
	if len(userAgentMatches) > 1 {
		config.UserAgent = userAgentMatches[1]
	}

	// Extract cookie using regex (looking for -b or --cookie)
	cookieRegex := regexp.MustCompile(`-b '([^']*)'`)
	cookieMatches := cookieRegex.FindStringSubmatch(contentStr)
	if len(cookieMatches) == 0 {
		// Try with --cookie flag
		cookieRegex = regexp.MustCompile(`--cookie '([^']*)'`)
		cookieMatches = cookieRegex.FindStringSubmatch(contentStr)
	}
	if len(cookieMatches) > 1 {
		config.Cookie = cookieMatches[1]
	}

	// Validate that we have both user agent and cookie
	if config.UserAgent == "" {
		return fmt.Errorf("user agent not found in curl command")
	}
	if config.Cookie == "" {
		return fmt.Errorf("cookie not found in curl command")
	}

	if config.verbose {
		fmt.Fprintf(os.Stderr, "Successfully extracted user agent and cookie from curl command\n")
	}
	return nil
}

func getDomains(config *Config) []string {
	var domains []string

	if config.Domain != "" {
		domains = append(domains, config.Domain)
	} else if config.ListFile != "" {
		file, err := os.Open(config.ListFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening list file: %v\n", err)
			os.Exit(1)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			domain := strings.TrimSpace(scanner.Text())
			if domain != "" {
				domains = append(domains, domain)
			}
		}

		if err := scanner.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "Error reading list file: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Check if we're reading from stdin
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			scanner := bufio.NewScanner(os.Stdin)
			for scanner.Scan() {
				domain := strings.TrimSpace(scanner.Text())
				if domain != "" {
					domains = append(domains, domain)
				}
			}
		} else {
			fmt.Fprintf(os.Stderr, "Error: either --domain, --list, or stdin must be provided\n")
			pflag.Usage()
			os.Exit(1)
		}
	}

	return domains
}

func processDomain(domain string, config *Config) {
	if config.verbose {
		fmt.Fprintf(os.Stderr, "Processing domain: %s\n", domain)
	}
	
	client := &http.Client{}
	regex := regexp.MustCompile(`href="/domain/([^/]+)/dns">`)

	for page := 1; page <= maxPages; page++ {
		url := fmt.Sprintf("https://securitytrails.com/list/apex_domain/%s?page=%d", domain, page)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating request for %s: %v\n", domain, err)
			continue
		}

		// Set headers
		setHeaders(req, config)

		resp, err := client.Do(req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching page %d for %s: %v\n", page, domain, err)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading response for %s: %v\n", domain, err)
			continue
		}

		matches := regex.FindAllStringSubmatch(string(body), -1)
		
		// Check exit conditions
		if len(matches) == 0 {
			fmt.Fprintf(os.Stderr, "Cookie Expired: please add new cookie\n")
			os.Exit(1)
		}

		if len(matches) == 1 {
			subdomain := matches[0][1]
			if subdomain == domain {
				if config.verbose {
					fmt.Fprintf(os.Stderr, "No subdomain found in this page\n")
				}
				return
			}
		}

		// Print subdomains, skipping the first line
		for i, match := range matches {
			if i == 0 {
				continue // skip the first match (like sed '1d')
			}
			if len(match) > 1 {
				fmt.Println(match[1])
			}
		}

		// Delay between requests
		if page < maxPages {
			time.Sleep(delay)
		}
	}
}

func setHeaders(req *http.Request, config *Config) {
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9,hi;q=0.8,en-IN;q=0.7")
	req.Header.Set("Cookie", config.Cookie)
	req.Header.Set("Dnt", "1")
	req.Header.Set("Priority", "u=0, i")
	req.Header.Set("Sec-Ch-Ua", `"Google Chrome";v="141", "Not?A_Brand";v="8", "Chromium";v="141"`)
	req.Header.Set("Sec-Ch-Ua-Arch", `"x86"`)
	req.Header.Set("Sec-Ch-Ua-Bitness", `"64"`)
	req.Header.Set("Sec-Ch-Ua-Full-Version", `"141.0.7390.108"`)
	req.Header.Set("Sec-Ch-Ua-Full-Version-List", `"Google Chrome";v="141.0.7390.108", "Not?A_Brand";v="8.0.0.0", "Chromium";v="141.0.7390.108"`)
	req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	req.Header.Set("Sec-Ch-Ua-Model", `""`)
	req.Header.Set("Sec-Ch-Ua-Platform", `"Windows"`)
	req.Header.Set("Sec-Ch-Ua-Platform-Version", `"19.0.0"`)
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("User-Agent", config.UserAgent)
}