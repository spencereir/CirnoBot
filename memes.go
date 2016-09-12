package main

import (
	"io/ioutil"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
)

var (
	nineball = []string{
		"It is certain",
		"It is decidedly so",
		"Without a doubt",
		"Yes, definitely",
		"You may rely on it",
		"As I see it, yes",
		"Most likely",
		"Outlook good",
		"Yes",
		"Signs point to yes",
		"Reply hazy try again",
		"Ask again later",
		"Better not tell you now",
		"Cannot predict now",
		"Concentrate and ask again",
		"Don't count on it",
		"My reply is no",
		"My sources say no",
		"Outlook not so good",
		"Very doubtful",
	}
	leave_memes = []string{
		"https://pbs.twimg.com/media/ClsbQajUoAAXKIx.jpg",
		"https://pbs.twimg.com/media/ClDk0CbUkAAs3aW.jpg",
	}
	remain_memes = []string{
		"https://pbs.twimg.com/media/CluCmwsWEAAh_UW.jpg",
		"https://pbs.twimg.com/media/CltNvdcWgAAevsU.jpg",
		"https://pbs.twimg.com/media/Cltto5QWIAAMHFw.jpg",
	}
	farage_laughing = []string{
		"http://static.independent.co.uk/s3fs-public/thumbnails/image/2013/05/03/21/farage.jpg",
		"https://gutsofabeggar.files.wordpress.com/2014/05/farage.jpeg",
		"http://www.thecommentator.com/system/articles/inner_pictures/000/005/835/original/o-NIGEL-FARAGE-LAUGHING-facebook.jpg?1431356974",
		"http://cdn2.theweek.co.uk/sites/theweek/files/farage-laughing-1.jpg",
		"https://sayitin500.files.wordpress.com/2014/10/image1.jpg",
		"http://www.thedrinksbusiness.com/wordpress/wp-content/uploads/2014/11/tumblr_inline_n6m8y86oc81r05kcc.jpg",
		"http://vadamagazine.com/wordpress/wp-content/uploads/2013/09/nigel-farage-headline.jpg",
	}
)

func farage() string {
	return farage_laughing[rand.Intn(len(farage_laughing))]
}

func xkcd(url string) string {
	res, _ := http.Get(url)
	b, _ := ioutil.ReadAll(res.Body)
	html := string(b)
	i_r := regexp.MustCompile("<img src=\"[\\.\\/\\w]+\"")
	i := i_r.FindAllString(html, -1)[1]
	i = i[10 : len(i)-1]
	a_r := regexp.MustCompile("<img src=\".+\"\\stitle=\"[0-9a-zA-Z!@#:$%;\\.,^&*()\\s]+\"")
	a := a_r.FindString(html)
	a = a[19+len(i) : len(a)-1]
	a = strings.Replace(a, "&#39;", "'", -1)
	return "http:" + i + "\nAlt text: " + a
}

func brexitmeme(remain, leave bool) string {
	if remain {
		return remain_memes[rand.Intn(len(remain_memes))]
	} else if leave {
		return leave_memes[rand.Intn(len(leave_memes))]
	} else {
		k := rand.Intn(len(remain_memes) + len(leave_memes))
		if k < len(remain_memes) {
			return remain_memes[k]
		} else {
			return leave_memes[k-len(remain_memes)]
		}
	}
}

func reorder(words []string) string {
	v := make([]int, len(words)-2)
	for i := 2; i < len(words); i++ {
		v[i-2] = i
	}
	for i := len(words) - 3; i >= 0; i-- {
		temp := v[i]
		j := rand.Intn(i + 1)
		v[i] = v[j]
		v[j] = temp
	}
	msg := ""
	for i := 2; i < len(words); i++ {
		if i > 2 {
			msg += " "
		}
		msg += words[v[i-2]]
	}
	return msg
}
