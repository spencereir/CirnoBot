package main

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"os"
)

var (
	r           [1000][1000]int
	idToAnime   map[string]AnimeData
	indToID     map[int]string
	idToInd     map[string]int
	numRatings  map[string]int
	rating      [1001](map[string]int)
	rates       []int
	predict     []float64
	num_predict []int
	s           []float64
	good        map[string]bool
	N           int
	M           int
	TOP_NUM     = 3
)

type UserInfo struct {
	UserID      string  `xml:"user_id"`
	Username    string  `xml:"user_name"`
	Watching    int     `xml:"user_watching"`
	Completed   int     `xml:"user_completed"`
	Onhold      int     `xml:"user_onhold"`
	Dropped     int     `xml:"user_dropped"`
	Plantowatch int     `xml:"user_plantowatch"`
	Dayswatched float64 `xml:"user_days_spent_watching"`
}

type AnimeData struct {
	AnimeID        string `xml:"series_animedb_id"`
	Title          string `xml:"series_title"`
	Synonyms       string `xml:"series_synonyms"`
	Type           int    `xml:"series_type"`
	Episodes       int    `xml:"series_episodes"`
	Status         int    `xml:"series_status"`
	Start          string `xml:"series_start"`
	End            string `xml:"series_end"`
	Image          string `xml:"series_image"`
	MyID           int    `xml:"my_id"`
	MyWatched      int    `xml:"my_watched_episodes"`
	MyStart        string `xml:"my_start_date"`
	MyFinish       string `xml:"my_finish_date"`
	MyScore        int    `xml:"my_score"`
	MyStatus       int    `xml:"my_status"`
	MyRewatching   string `xml:"my_rewatching"`
	MyRewatchingEp int    `xml:"my_rewatching_ep"`
	MyLastUpdated  string `xml:"my_last_updated"`
	MyTags         string `xml:"my_tags"`
}

type AnimeList struct {
	Info  UserInfo    `xml:"myinfo"`
	Anime []AnimeData `xml:"anime"`
}

type Pair struct {
	Key   string
	Value int
}

type PairList []Pair

func (p PairList) Len() int           { return len(p) }
func (p PairList) Less(i, j int) bool { return p[i].Value < p[j].Value }
func (p PairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func getAnimeList(username string) AnimeList {
	res, err := http.Get(fmt.Sprintf("http://myanimelist.net/malappinfo.php?u=%v&status=all&type=anime", username))
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	var al AnimeList
	xml.Unmarshal(b, &al)
	return al
}

func simil(x int) float64 {
	rx := 0.0
	ry := 0.0
	num_rated_x := 0
	num_rated_y := 0
	for i := 0; i < M; i++ {
		if rates[i] != 0 {
			num_rated_y++
			ry += float64(rates[i])
		}
		if r[x][i] != 0 {
			num_rated_x++
			rx += float64(r[x][i])
		}
	}

	rx /= float64(num_rated_x)
	ry /= float64(num_rated_y)

	fmt.Printf("%v %v\n", rx, ry)
	num := 0.0
	den1 := 0.0
	den2 := 0.0
	for i := 0; i < M; i++ {
		if rates[i] != 0 && r[x][i] != 0 {
			num += (float64(r[x][i]) - rx) * (float64(rates[i]) - ry)
			den1 += (float64(r[x][i]) - rx) * (float64(r[x][i]) - rx)
			den2 += (float64(rates[i]) - ry) * (float64(rates[i]) - ry)
		}
	}
	if den1*den2 == 0 {
		return 0
	}
	return num / math.Sqrt(den1*den2)
}

func RecommendAnime(username string, top int) string {
	f, err := os.Open("r.dat")
	TOP_NUM = top
	if err != nil {
		panic(fmt.Sprintf("Error: %v", err))
	}
	fmt.Fscanf(f, "%v %v\n", &N, &M)
	idToAnime = make(map[string]AnimeData)
	indToID = make(map[int]string)
	idToInd = make(map[string]int)

	for i := 0; i < N; i++ {
		for j := 0; j < M; j++ {
			fmt.Fscanf(f, "%v", &r[i][j])
		}
	}

	for i := 0; i < M; i++ {
		k := ""
		fmt.Fscanf(f, "%v", &k)
		indToID[i] = k
		idToInd[k] = i
	}
	f.Close()

	f, _ = os.Open("anime.dat")
	sc := bufio.NewScanner(f)

	for sc.Scan() {
		k := sc.Text()
		sc.Scan()
		a := AnimeData{}
		a.Title = sc.Text()
		idToAnime[k] = a
	}

	al := getAnimeList(username)
	rates = make([]int, M)
	predict = make([]float64, M)
	num_predict = make([]int, M)
	good = make(map[string]bool, M)
	s = make([]float64, N)
	rating[1000] = make(map[string]int)
	for _, a := range al.Anime {
		if a.MyStatus == 6 {
			good[a.AnimeID] = true
		} else {
			good[a.AnimeID] = false
		}
		rating[1000][a.AnimeID] = a.MyScore
	}
	for i := 0; i < M; i++ {

		rates[i] = rating[1000][indToID[i]]
	}
	for i := 0; i < N; i++ {
		s[i] = simil(i)
	}

	for i := 0; i < M; i++ {
		norm := 0.0
		var used []bool
		used = make([]bool, N)
		for j := 0; j < N; j++ {
			used[j] = false
		}
		for j := 0; j < TOP_NUM; j++ {
			most_similar := 0.0
			most_similar_ind := 0
			for k := 0; k < N; k++ {
				if r[k][i] != 0 && s[k] > most_similar && used[k] == false {
					most_similar = s[k]
					most_similar_ind = k
				}
			}
			norm += math.Abs(most_similar)
			used[most_similar_ind] = true
			predict[i] += most_similar * float64(r[most_similar_ind][i])
		}
		predict[i] /= norm
	}
	best := 0.0
	best_ind := 0
	for i := 0; i < M; i++ {
		if rates[i] == 0 {
			v, exists := good[indToID[i]]
			if exists && !v {
				continue
			}
			if predict[i] > best {
				best = predict[i]
				best_ind = i
			}
		}
	}
	best = 0.0
	sbest_ind := 0
	for i := 0; i < M; i++ {
		if rates[i] == 0 && i != best_ind {
			v, exists := good[indToID[i]]
			if exists && !v {
				continue
			}
			if predict[i] > best {
				best = predict[i]
				sbest_ind = i
			}
		}
	}
	best = 0.0
	tbest_ind := 0
	for i := 0; i < M; i++ {
		if rates[i] == 0 && i != best_ind && i != sbest_ind {
			v, exists := good[indToID[i]]
			if exists && !v {
				continue
			}
			if predict[i] > best {
				best = predict[i]
				tbest_ind = i
			}
		}
	}
	return fmt.Sprintf("%v\n%v\n%v", idToAnime[indToID[best_ind]].Title, idToAnime[indToID[sbest_ind]].Title, idToAnime[indToID[tbest_ind]].Title)
}
