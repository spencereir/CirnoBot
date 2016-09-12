package main

import (
	"bufio"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/otium/ytdl"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	dg              *discordgo.Session
	puush_logged_in bool

	markov   *Chain = NewChain(2)
	tracks   []string
	msgidlog map[string][]string = make(map[string][]string)
	zawarudo map[string]bool     = make(map[string]bool)
	msglog   map[string][]string = make(map[string][]string)
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	dg, _ = discordgo.New(os.Args[1])

	for _, coll := range COLLECTIONS {
		coll.Load()
	}

	file, _ := os.Open("tracks.dat")
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		tracks = append(tracks, scanner.Text())
	}

	fmt.Printf("Logging in to puush...\n")
	puush_logged_in = puushLogin()
	if puush_logged_in {
		fmt.Printf("Logged in\n")
	} else {
		fmt.Printf("Error logging in\n")
	}

	fmt.Printf("Building Markov chain...\n")
	//	f, _ := os.Open("corpus.dat")
	//markov.Build(f)
	fmt.Printf("Markov chain built\n")

	dg.AddHandler(messageCreate)

	dg.Open()

	dg.UpdateStatus(0, "War of the Human Tanks")

	ReadFromJSON()
	fmt.Println("Bot ready")

	var input string
	fmt.Scanln(&input)
	return
}

func ErrorMessage(location, context string, information []string) string {
	return "Error with " + context + " in function " + location + ". Send this message to Spencer. Also give him this information: " + strings.Join(information, ";") + "."
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	ch_id := m.ChannelID
	author := m.Author
	channel, _ := dg.State.Channel(ch_id)
	guild, _ := dg.State.Guild(channel.GuildID)

	msgidlog[ch_id] = append(msgidlog[ch_id], m.Message.ID)
	_, exists := name[guild.ID]
	if !exists {
		AddNewServer(guild.ID)
		fmt.Printf("Adding new server w/ name %v", guild.ID)
	}
	//no self replying
	if author.Bot == true {
		return
	}

	content := m.ContentWithMentionsReplaced()
	cl := strings.ToLower(content)
	words := strings.Fields(cl)
	wordsCase := strings.Fields(content)

	if len(words) == 0 {
		return
	} else if content == "ZA WARUDO" && !zawarudo[ch_id] {
		sound := ZAWARUDO.Sounds[0]
		go enqueuePlay(m.Author, guild, ZAWARUDO, sound)
		zawarudo[ch_id] = true
		s.ChannelMessageDelete(ch_id, m.Message.ID)
		s.ChannelMessageSend(ch_id, "ZA WARUDO\nTOKI WO TOMARE\nhttp://i.imgur.com/zXAPhxP.png")
		time.AfterFunc(5*time.Second, func() {
			sound = ZAWARUDO.Sounds[1]
			go enqueuePlay(m.Author, guild, ZAWARUDO, sound)
			s.ChannelMessageSend(ch_id, "Soshite, toki wa ugoki dasu")
			for _, v := range msglog[ch_id] {
				s.ChannelMessageSend(ch_id, v)
			}
			zawarudo[ch_id] = false
			msglog[ch_id] = []string{}
		})
		return
	} else if zawarudo[ch_id] {
		msglog[ch_id] = append(msglog[ch_id], author.Username+": "+content)
		s.ChannelMessageDelete(ch_id, m.Message.ID)
		return
	} else if cl == "!farage" {
		s.ChannelMessageSend(ch_id, farage())
		return
	} else if cl == "!stop" || cl == "stop music" {
		cancelRequest = true
	} else if strings.Contains(cl, "buses") && strings.Contains(cl, "gensokyo") {
		msg := "There are no buses in Gensokyo\n"
		msg += "https://www.youtube.com/watch?v=5wFDWP5JwSM"
		s.ChannelMessageSend(ch_id, msg)
		return
	} else if cl == "!nineball" || cl == "⑨" {
		s.ChannelMessageSend(ch_id, nineball[rand.Intn(20)])
	} else if words[0] == "!stand" {
		if len(words) == 1 {
			s.ChannelMessageSend(ch_id, stand(""))
		} else {
			s.ChannelMessageSend(ch_id, stand(words[1]))
		}
	} else if isName(guild.ID, words[0]) {
		if strings.HasPrefix(cl[len(words[0]):], " who is the strongest") {
			msg := "Eye'm the strongest\n"
			msg += "http://puu.sh/pJmUU/1fd25b3783.jpg"
			s.ChannelMessageSend(ch_id, msg)
		} else if strings.HasPrefix(cl[len(words[0]):], " add name") {
			name[guild.ID] = append(name[guild.ID], words[3])
			s.ChannelMessageSend(ch_id, "I will now respond to "+words[3])
			servers.Servers[guild_to_ind[guild.ID]].Names = append(servers.Servers[guild_to_ind[guild.ID]].Names, words[3])
			WriteToJSON()
		} else if strings.HasPrefix(cl[len(words[0]):], " xkcd") {
			if len(words) == 2 {
				s.ChannelMessageSend(ch_id, xkcd("http://xkcd.com/"))
			} else {
				_, err := strconv.Atoi(words[2])
				if err != nil {
					s.ChannelMessageSend(ch_id, "http://imgs.xkcd.com/comics/"+strings.Join(words[2:], "_")+".png")
				} else {
					s.ChannelMessageSend(ch_id, xkcd("http://xkcd.com/"+words[2]+"/"))
				}
			}
		} else if strings.HasPrefix(cl[len(words[0]):], " copy") {
			paste[guild.ID][words[3]] = words[2]
			servers.Servers[guild_to_ind[guild.ID]].Pastes = append(servers.Servers[guild_to_ind[guild.ID]].Pastes, makePaste(words[3], words[2]))
			WriteToJSON()
		} else if strings.HasPrefix(cl[len(words[0]):], " paste") {
			if paste[guild.ID] == nil {
				s.ChannelMessageSend(ch_id, "I couldn't find any pastes in this guild")
			} else {
				v, found := paste[guild.ID][words[2]]
				if !found {
					s.ChannelMessageSend(ch_id, "I couldn't find a paste with code "+words[2])
				} else {
					s.ChannelMessageSend(ch_id, v)
				}
			}
		} else if strings.HasPrefix(cl[len(words[0]):], " reorder") {
			s.ChannelMessageSend(ch_id, reorder(wordsCase))
		} else if strings.HasPrefix(cl[len(words[0]):], " delete") {
			k, _ := strconv.Atoi(words[2])
			for i := len(msgidlog[ch_id]) - 1; i >= len(msgidlog[ch_id])-1-k; i-- {
				s.ChannelMessageDelete(ch_id, msgidlog[ch_id][i])
			}
			msgidlog[ch_id] = msgidlog[ch_id][:len(msgidlog[ch_id])-k]
		} else if strings.HasPrefix(cl[len(words[0]):], " roll") {
			s.ChannelMessageSend(ch_id, roll(words))
		} else if len(words) > 1 && strings.HasPrefix(words[1], "choose") && len(words[1]) > 6 {
			msg := ""
			v, _ := strconv.Atoi(words[1][6:])
			for i := 0; i < v; i++ {
				if i > 0 {
					msg += ", "
				}
				msg += words[rand.Intn(len(words)-2)+2]
			}
			s.ChannelMessageSend(ch_id, msg)
		} else if len(words) > 1 && words[1] == "choose" {
			s.ChannelMessageSend(ch_id, words[rand.Intn(len(words)-2)+2])
		} else if strings.HasPrefix(cl[len(words[0]):], " roulette") {
			if len(words) == 4 {
				j, _ := strconv.Atoi(words[2])
				k, _ := strconv.Atoi(words[3])
				if rand.Intn(k) < j {
					s.ChannelMessageSend(ch_id, "You lost")
				} else {
					s.ChannelMessageSend(ch_id, "*click*")
				}
			} else if len(words) == 3 {
				k, _ := strconv.Atoi(words[2])
				if rand.Intn(k) < 0 {
					s.ChannelMessageSend(ch_id, "You lost")
				} else {
					s.ChannelMessageSend(ch_id, "*click*")
				}
			} else {
				if rand.Intn(6) == 0 {
					s.ChannelMessageSend(ch_id, "You lost")
				} else {
					s.ChannelMessageSend(ch_id, "*click*")
				}
			}
		} else if len(wordsCase) > 2 && strings.HasPrefix(cl[len(words[0]):], " rank") {
			s.ChannelMessageSend(ch_id, classify(wordsCase[2]))
		} else if len(words) > 2 && strings.HasPrefix(cl[len(words[0])-1:], ",") {
			best := 100000
			best_i := 0
			for i := 1; i < len(words)-1; i++ {
				if ngram[wordsCase[i]] < best && ngram[wordsCase[i]] != 0 {
					best_i = i
					best = ngram[wordsCase[i]]
				}
			}
			var seed []string = make([]string, 2)
			seed[0] = wordsCase[best_i]
			seed[1] = wordsCase[best_i+1]
			s.ChannelMessageSend(ch_id, markov.Generate(seed, 100))
		} else if strings.HasPrefix(cl[len(words[0]):], " say") {
			msg := ""
			for i := 2; i < len(words); i++ {
				msg += words[i]
				msg += " "
			}
			s.ChannelMessageSend(ch_id, msg)
		} else if len(words) >= 4 && strings.HasPrefix(cl[len(words[0]):], " recommend anime") {
			if len(words) == 5 {
				k, _ := strconv.Atoi(words[4])
				s.ChannelMessageSend(ch_id, RecommendAnime(wordsCase[3], k))
			} else {
				s.ChannelMessageSend(ch_id, RecommendAnime(wordsCase[3], 3))
			}
		} else if strings.HasPrefix(cl[len(words[0]):], " generate stand") {
			if len(words) == 4 {
				s.ChannelMessageSend(ch_id, stand(words[3]))
			} else {
				s.ChannelMessageSend(ch_id, stand(""))
			}
		} else if strings.HasPrefix(cl[len(words[0]):], " generate stanza") {
			stanzas := []string{"Although its scent still lingers on \\ the form of a flower has scattered away", "For whom will the glory \\ of this world remain unchanged?", "Arriving today at the yonder side \\ of the deep mountains of evanescent existence", "We shall never allow ourselves to drift away \\ intoxicated, in the world of shallow dreams.", "Two roads diverged in a yellow wood, \\ And sorry I could not travel both \\ And be one traveler, long I stood \\ And looked down one as far as I could \\ To where it bent in the undergrowth;", "Then took the other, as just as fair, \\ And having perhaps the better claim, \\ Because it was grassy and wanted wear; \\ Though as for that the passing there \\ Had worn them really about the same,", "And both that morning equally lay \\ In leaves no step had trodden black. \\ Oh, I kept the first for another day! \\ Yet knowing how way leads on to way, \\ I doubted if I should ever come back.", "I shall be telling this with a sigh \\ Somewhere ages and ages hence: \\ Two roads diverged in a wood, and I— \\ I took the one less traveled by, \\ And that has made all the difference.", "The way a crow \\ Shook down on me \\ The dust of snow \\ From a hemlock tree", "Has given my heart \\ A change of mood \\ And saved some part \\ Of a day I had rued."}
			s.ChannelMessageSend(ch_id, stanzas[rand.Intn(len(stanzas))])
		} else if len(words) > 2 && strings.HasPrefix(cl[len(words[0]):], " puush") {
			if !puush_logged_in {
				s.ChannelMessageSend(ch_id, "Apologies, but I'm currently having trouble connecting to Puush. Contact Spencer if this problem persists.\n")
			} else if strings.Contains(wordsCase[2], "youtu.be") || strings.Contains(wordsCase[2], "youtube.") {
				vid, _ := ytdl.GetVideoInfo(wordsCase[2])
				a, _ := vid.GetDownloadURL(vid.Formats[0])
				request := fmt.Sprintf("%v", a)
				url := "http://tinyurl.com/api-create.php?url=" + request
				res, _ := http.Get(url)
				b, _ := ioutil.ReadAll(res.Body)
				s.ChannelMessageSend(ch_id, string(b))
			} else if len(words) < 4 {
				s.ChannelMessageSend(ch_id, save(wordsCase[2]))
			} else {
				s.ChannelMessageSend(ch_id, saveAs(wordsCase[2], wordsCase[3]))
			}
		} else if strings.HasPrefix(cl[len(words[0]):], " research") {
			if len(words) < 3 {
				s.ChannelMessageSend(ch_id, "Usage: <Name> research <topic>")
			} else {
				url := "https://en.wikipedia.org/wiki/" + strings.Join(wordsCase[2:], "_")
				url = strings.Replace(url, "?", "%3F", -1)
				res, err := http.Get(url)
				if err != nil {
					s.ChannelMessageSend(ch_id, "I couldn't find anything about "+strings.Join(wordsCase[2:], " "))
					return
				}
				b, _ := ioutil.ReadAll(res.Body)
				if strings.Contains(string(b), "Wikipedia does not have an article with this exact name.") {
					s.ChannelMessageSend(ch_id, "I couldn't find anything about "+strings.Join(wordsCase[2:], " "))
					return
				}
				s.ChannelMessageSend(ch_id, url)
			}
		} else if strings.HasPrefix(cl[len(words[0]):], " brexit") {
			if len(words) < 3 {
				s.ChannelMessageSend(ch_id, brexitmeme(false, false))
			} else if words[2] == "remain" {
				s.ChannelMessageSend(ch_id, brexitmeme(true, false))
			} else {
				s.ChannelMessageSend(ch_id, brexitmeme(false, true))
			}
		}
	} else {
		for _, coll := range COLLECTIONS {
			if scontains(cl, coll.Commands) {
				var sound *Sound
				fmt.Print("Enqueuing play")
				go enqueuePlay(m.Author, guild, coll, sound)
				return
			}
		}
	}
}

func isName(guildID, n string) bool {
	for _, v := range name[guildID] {
		if n == v {
			return true
		}
	}
	return false
}
