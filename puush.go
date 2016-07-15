package main

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"
)

var (
	session   string
	API_URL   string = "https://puush.me/api/"
	AUTH_URL  string = "https://puush.me/api/auth/"
	API_KEY   string = "B587BB2E757AE456C087AA054A378F69"
	UP_STRING string = "https://puush.me/api/up/"
)

func puushLogin() bool {
	r, err := http.PostForm(AUTH_URL, url.Values{"k": {API_KEY}})
	if err != nil {
		fmt.Println(err)
		return false
	}
	body, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	info := strings.Split(string(body), ",")
	if info[0] == "-1" {
		log.Fatal("Login failed:" + string(body))
		return false
	} else {
		session = info[1]
	}
	return true
}

func puush(filename string) string {
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	buf := new(bytes.Buffer)
	w := multipart.NewWriter(buf)
	kwriter, err := w.CreateFormField("k")
	if err != nil {
		fmt.Println(err)
		return ""
	}
	io.WriteString(kwriter, session)

	h := md5.New()
	h.Write(file)

	cwriter, err := w.CreateFormField("c")
	if err != nil {
		fmt.Println(err)
		return ""
	}
	io.WriteString(cwriter, fmt.Sprintf("%x", h.Sum(nil)))

	zwriter, err := w.CreateFormField("z")
	if err != nil {
		fmt.Println(err)
		return ""
	}
	io.WriteString(zwriter, "poop") // They must think their protocol is shit

	fwriter, err := w.CreateFormFile("f", filename)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	fwriter.Write(file)

	w.Close()

	req, err := http.NewRequest("POST", "http://puush.me/api/up", buf)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	body, _ := ioutil.ReadAll(res.Body)
	res.Body.Close()
	info := strings.Split(string(body), ",")
	if info[0] == "0" {
		return info[1]
	} else {
		log.Fatal("Upload failed:" + string(body))
	}
	return ""
}

func save(loc string) string {
	res, _ := http.Get(loc)
	b, _ := ioutil.ReadAll(res.Body)
	l := strings.Split(loc, "/")
	filename := l[len(l)-1]
	f, _ := os.Create(filename)
	f.Write(b)
	s := puush(filename)
	f.Close()
	os.Remove(filename)
	return s
}

func saveAs(loc, filename string) string {
	res, _ := http.Get(loc)
	b, _ := ioutil.ReadAll(res.Body)
	f, _ := os.Create(filename)
	f.Write(b)
	s := puush(filename)
	f.Close()
	os.Remove(filename)
	return s
}
