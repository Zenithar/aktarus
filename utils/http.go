package utils

import (
	"errors"
	"fmt"
	"html"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"
)

func GetPage(url string) (string, error) {
	client := newHttpTimeoutClient()

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.Printf("Couldn't build http request: %s", err.Error())
		return "", err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 Leader-1/Mighty, Mighty GoBot")

	resp, err := client.Do(req)
	if err != nil {
		logger.Printf("Couldn't perform http request: %s", err.Error())
		return "", err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Printf("Couldn't read http response body: %s", err.Error())
		return "", err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		logger.Printf("HTTP response code: %d", resp.StatusCode)
		err = errors.New(fmt.Sprintf("Bad HTTP response code: %d", resp.StatusCode))
	}

	return string(body), err
}

func GetPageWithAuth(url string, user string, pass string) (string, error) {
	client := newHttpTimeoutClient()

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.Printf("Couldn't build http request: %s", err.Error())
		return "", err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 Leader-1/Mighty, Mighty GoBot")
	req.SetBasicAuth(user, pass)

	resp, err := client.Do(req)
	if err != nil {
		logger.Printf("Couldn't perform http request: %s", err.Error())
		return "", err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Printf("Couldn't read http response body: %s", err.Error())
		return "", err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		logger.Printf("HTTP response code: %d", resp.StatusCode)
		err = errors.New(fmt.Sprintf("Bad HTTP response code: %d", resp.StatusCode))
	}

	return string(body), err
}

func timeoutDialer(cTimeout time.Duration, rwTimeout time.Duration) func(net, addr string) (c net.Conn, err error) {
	return func(netw, addr string) (net.Conn, error) {
		conn, err := net.DialTimeout(netw, addr, cTimeout)
		if err != nil {
			return nil, err
		}
		conn.SetDeadline(time.Now().Add(rwTimeout))
		return conn, nil
	}
}

// Sets up a 2 second timeout http client because waiting longer for a page request is too costly
func newHttpTimeoutClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			Dial: timeoutDialer(time.Second*2, time.Second*2),
		},
	}
}

// Extracts an URL out of a string
func ExtractURL(str string) (url string, err error) {
	// Assume errors by default
	err = errors.New("No URL found")

	// Capture links and post their titles, etc
	start := strings.Index(str, "http://")

	// Try https if no http match
	if start == -1 {
		start = strings.Index(str, "https://")
	}

	// Found a link... maybe
	if start > -1 {
		url = strings.SplitN(str[start:], " ", 2)[0]
		// String isn't just a protocol
		if len(url) > 9 && !strings.HasSuffix(url, "://") {
			err = nil
		} else {
			url = ""
		}
	}
	return
}

func ExtractTitle(url string) (title string, err error) {
	client := newHttpTimeoutClient()
	resp, err := client.Get(url)

	if err != nil {
		if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
			logger.Printf("title extraction request timed out")
		} else {
			logger.Printf("Failed to GET %s due to %s", url, err.Error())
		}
		return
	}

	// Make sure we close our response reader like a good citizen
	defer resp.Body.Close()

	// No point in parsing response if it wasn't a success
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		// Read 10kb of the body
		body := make([]byte, 0)
		var count int

		for {
			more_body := make([]byte, 10240) // Read in up to 10kb increments
			count, err = resp.Body.Read(more_body)
			body = append(body, more_body[:count]...)
			if len(body) >= 10240 { // Got 10kb, so break out
				break
			}

			// Couldn't read the page for some reason, maybe EOF
			if err != nil {
				if len(body) == 0 {
					logger.Printf("Failed to read page response for %s due to %s", url, err.Error())
					return
				}
				break
			}
		}

		pageData := string(body)

		start := strings.Index(strings.ToLower(pageData), "<title")
		end := strings.Index(strings.ToLower(pageData), "</title>")

		switch {
		case start > -1 && end > -1: // Found a title tag, get it's contents
			title = pageData[start:end]
			title = pageData[start+strings.Index(title, ">")+1 : end]
		case start > -1: // If for some reason within 10kb we get a chopped off <title>, use the remainder
			title = pageData[start:]
			title = pageData[start+strings.Index(title, ">")+1:]
		default: // No title tags to speak off, use the mime type instead to be helpful
			title = fmt.Sprintf("unknown [%s]", resp.Header.Get("Content-Type"))
		}
	} else {
		title = resp.Status // Lets just return the status text for non-successful responses
	}

	title = html.UnescapeString(strings.TrimSpace(title))
	return
}
