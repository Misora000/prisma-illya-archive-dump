package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

func main() {
	ctx := context.Background()

	list, err := getIllustJSON(ctx)
	if err != nil {
		log.Fatal(err)
	}

	for i, v := range list {
		// Detect the thumbnail url by the format `*-300x500.jpg`.
		isThumbnail, err := regexp.Match("-[0-9]{3}x[0-9]{3}.jpg", []byte(v.Eyecatch))
		if err != nil {
			log.Println(err)
			continue
		}

		if isThumbnail {
			// Covert the thumbnail url to the real url.
			seg := strings.Split(v.Eyecatch, "-")
			v.Eyecatch = strings.Join(seg[0:len(seg)-1], "-") + ".jpg"
		}

		fmt.Printf("(%03d/%03d) downloading %v ... ", i+1, len(list), v.Title)
		if err = downloadImg(ctx, v.Eyecatch, v.Title); err != nil {
			fmt.Printf("failed\n")
			log.Println(err)
			continue
		}
		fmt.Printf("done\n")
	}
}

type allIllustRsp struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Eyecatch string `json:"eyecatch"`
}

func getIllustJSON(ctx context.Context) ([]*allIllustRsp, error) {
	o := []*allIllustRsp{}

	rawText, err := getByHTTP(ctx, "https://prisma-illya.jp/portal/allillusts")
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(rawText, &o); err != nil {
		return nil, err
	}

	return o, nil
}

func getByHTTP(ctx context.Context, url string) ([]byte, error) {

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.WithContext(ctx)

	client := &http.Client{}
	rsp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer rsp.Body.Close()

	return ioutil.ReadAll(rsp.Body)
}

func downloadImg(ctx context.Context, url string, title string) error {
	dst := "dump/" + title + ".jpg"
	if _, err := os.Open(dst); !os.IsNotExist(err) {
		return err
	}

	buf, err := getByHTTP(ctx, url)
	if err != nil {
		return err
	}

	if err := os.Mkdir("dump", 0777); err != nil && !os.IsExist(err) {
		return err
	}

	return ioutil.WriteFile(dst, buf, 0644)
}
