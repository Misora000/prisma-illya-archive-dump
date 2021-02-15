package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/Misora000/easyhtml"
)

func main() {
	ctx := context.Background()
	illustPages := []string{}

	for i := 1; i <= 23; i++ {
		p := fmt.Sprintf("https://prisma-illya.jp/portal/illust/page/%d", i)
		if r, err := dumpIllustPageURL(ctx, p); err == nil {
			illustPages = append(illustPages, r...)
		} else {
			log.Println(err)
		}
	}

	for _, u := range illustPages {
		title, illustURL, err := dumpIllustURL(ctx, "https://prisma-illya.jp"+u)
		if err != nil {
			log.Println(err)
			continue
		}

		fmt.Printf("downloading %v ... ", title)
		if err = downloadImg(ctx, illustURL, title); err != nil {
			fmt.Printf("failed\n")
			log.Println(err)
			continue
		}
		fmt.Printf("done\n")
	}
}

func dumpIllustPageURL(ctx context.Context, URL string) ([]string, error) {
	o := []string{}

	body, err := getByHTTP(ctx, URL)
	if err != nil {
		return nil, err
	}

	z := easyhtml.NewTokenizer(body)
	for {
		attr, eof := z.JumpToClass("a", "c-card--link")
		if eof {
			break
		}
		if u, exists := attr["href"]; exists {
			o = append(o, u)
		}
	}

	return o, nil
}

func dumpIllustURL(ctx context.Context, URL string) (string, string, error) {
	title := ""
	imgURL := ""

	body, err := getByHTTP(ctx, URL)
	if err != nil {
		return title, imgURL, err
	}

	z := easyhtml.NewTokenizer(body)
	attr, eof := z.JumpToClass("img", "u-sz_w_100-sp")
	if eof {
		return "", "", fmt.Errorf("EOF")
	}

	if u, exists := attr["src"]; exists {
		imgURL = u
	}

	if _, eof = z.JumpToClass("h1", "c-entry-header__title"); eof {
		return title, imgURL, nil
	}
	title, _ = z.GetNextText()

	return title, imgURL, nil
}

func getByHTTP(ctx context.Context, url string) (io.Reader, error) {

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

	buf, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil, err
	}
	rsp.Body.Close()

	return bytes.NewReader(buf), nil
}

func downloadImg(ctx context.Context, url string, title string) error {
	dst := "dump/" + title + ".jpg"
	if _, err := os.Open(dst); !os.IsNotExist(err) {
		return err
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.WithContext(ctx)

	client := &http.Client{}
	rsp, err := client.Do(req)
	if err != nil {
		return err
	}

	buf, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return err
	}
	rsp.Body.Close()

	if err := os.Mkdir("dump", 0777); err != nil && !os.IsExist(err) {
		return err
	}

	return ioutil.WriteFile(dst, buf, 0644)
}
