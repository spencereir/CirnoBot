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
	information := []string{filename}
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println(err)
		return ErrorMessage("puush", "reading filename", information)
	}

	buf := new(bytes.Buffer)
	w := multipart.NewWriter(buf)
	kwriter, err := w.CreateFormField("k")
	if err != nil {
		fmt.Println(err)
		return ErrorMessage("puush", "writing k", information)
	}

	io.WriteString(kwriter, session)

	h := md5.New()
	h.Write(file)

	cwriter, err := w.CreateFormField("c")
	if err != nil {
		fmt.Println(err)
		return ErrorMessage("puush", "writing c", information)
	}
	io.WriteString(cwriter, fmt.Sprintf("%x", h.Sum(nil)))

	zwriter, err := w.CreateFormField("z")
	if err != nil {
		fmt.Println(err)
		return ErrorMessage("puush", "writing z", information)
	}
	io.WriteString(zwriter, "poop") // They must think their protocol is shit

	fwriter, err := w.CreateFormFile("f", filename)
	if err != nil {
		fmt.Println(err)
		return ErrorMessage("puush", "writing filename", information)
	}
	fwriter.Write(file)

	w.Close()

	req, err := http.NewRequest("POST", "http://puush.me/api/up", buf)
	if err != nil {
		fmt.Println(err)
		return ErrorMessage("puush", "querying the puu.sh API", information)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return ErrorMessage("puush", "executing the request", information)
	}
	body, _ := ioutil.ReadAll(res.Body)
	res.Body.Close()
	info := strings.Split(string(body), ",")
	if info[0] == "0" {
		return info[1]
	} else {
		information = append(information, string(body))
		return ErrorMessage("puush", "uploading", information)
	}
	return ""
}

func save(loc string) string {
	//Try common extensions
	extensions := []string{".webm", ".doc", ".docx", ".log", ".msg", ".odt", ".pages", ".rtf", ".tex", ".txt", ".wpd", ".wps", ".csv", ".dat", ".ged", ".key", ".keychain", ".pps", ".ppt", ".pptx", ".sdf", ".tar", ".tax2014", ".tax2015", ".vcf", ".xml", ".aif", ".iff", ".m3u", ".m4a", ".mid", ".mp3", ".mpa", ".wav", ".wma", ".3g2", ".3gp", ".asf", ".avi", ".flv", ".m4v", ".mov", ".mp4", ".mpg", ".rm", ".srt", ".swf", ".vob", ".wmv", ".3dm", ".3ds", ".max", ".obj", ".bmp", ".dds", ".gif", ".jpg", ".png", ".psd", ".pspimage", ".tga", ".thm", ".tif", ".tiff", ".yuv", ".ai", ".eps", ".ps", ".svg", ".indd", ".pct", ".pdf", ".xlr", ".xls", ".xlsx", ".accdb", ".db", ".dbf", ".mdb", ".pdb", ".sql", ".apk", ".app", ".bat", ".cgi", ".exe", ".gadget", ".jar", ".wsf", ".dem", ".gam", ".nes", ".rom", ".sav", ".dwg", ".dxf", ".gpx", ".kml", ".kmz", ".asp", ".aspx", ".cer", ".cfm", ".csr", ".css", ".htm", ".html", ".js", ".jsp", ".php", ".rss", ".xhtml", ".crx", ".plugin", ".fnt", ".fon", ".otf", ".ttf", ".cab", ".cpl", ".cur", ".deskthemepack", ".dll", ".dmp", ".drv", ".icns", ".ico", ".lnk", ".sys", ".cfg", ".ini", ".prf", ".hqx", ".mim", ".uue", ".7z", ".cbr", ".deb", ".gz", ".pkg", ".rar", ".rpm", ".sitx", ".tar", ".zip", ".zipx", ".bin", ".cue", ".dmg", ".iso", ".mdf", ".toast", ".vcd", ".c", ".class", ".cpp", ".cs", ".dtd", ".fla", ".h", ".java", ".lua", ".m", ".pl", ".py", ".sh", ".sln", ".swift", ".vb", ".vcxproj", ".xcodeproj", ".bak", ".tmp", ".crdownload", ".ics", ".msi", ".part", ".torrent"}
	for _, v := range extensions {
		if strings.Contains(loc, v) {
			return saveAs(loc, "puush_file"+v)
		}
	}
	return "I tried to find what kind of file this was, but couldn't Please specify a file extension like so: cirno puush <link> <filename with extension>"
}

func saveAs(loc, filename string) string {
	information := []string{loc, filename}
	res, err := http.Get(loc)
	if err != nil {
		return ErrorMessage("save", "getting response", information)
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return ErrorMessage("save", "parsing response", information)
	}
	f, err := os.Create(filename)
	if err != nil {
		return ErrorMessage("save", "creating f", information)
	}
	f.Write(b)
	s := puush(filename)
	err = f.Close()
	if err != nil {
		return ErrorMessage("save", "closing f", information)
	}
	err = os.Remove(filename)
	if err != nil {
		return ErrorMessage("save", "removing f", information)
	}
	return s
}
