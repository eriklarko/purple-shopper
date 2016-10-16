package downloader

import (
	"net/url"
	"os"
	"strings"
	"io"
	"crypto/sha1"
	"fmt"
	"io/ioutil"
	"net/http"
)

func DownloadImage(url *url.URL) (string, error) {
	res, error := http.Get(url.String())
	if error != nil {
		return "", error
	}

	data, error := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if error != nil {
		return "", error
	}

	file, error := os.Create("image_" + getImageName(url))
	defer file.Close()
	if error != nil {
		return "", error
	}

	_, error = file.Write(data)
	if error != nil {
		return "", error
	}
	file.Sync()
	return file.Name(), nil
}

func getImageName(url *url.URL) string {
	urlString := url.String()
	lastDotIndex := strings.LastIndex(urlString, "/")
	fileEnding := urlString[lastDotIndex + 1:]

	h := sha1.New()
	io.WriteString(h, urlString)
	return fmt.Sprintf("%x.%s", h.Sum(nil), fileEnding)
}
