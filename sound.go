package main

import (
	"encoding/binary"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"io"
	"math/rand"
	"os"
	"time"
)

var (
	queues         map[string]chan *Play = make(map[string]chan *Play)
	MAX_QUEUE_SIZE                       = 6
	BITRATE                              = 128
	cancelRequest                        = false
	volume                               = 1.0
)

type Play struct {
	GuildID   string
	ChannelID string
	UserID    string
	Sound     *Sound
	Next      *Play
	Forced    bool
}

type SoundCollection struct {
	Prefix     string
	Commands   []string
	Sounds     []*Sound
	ChainWith  *SoundCollection
	soundRange int
}

type Sound struct {
	Name      string
	Weight    int
	PartDelay int
	buffer    [][]byte
}

var OHGOD *SoundCollection = &SoundCollection{
	Prefix: "ohgod",
	Commands: []string{
		"!god",
		"oh god",
		"ohgod",
		"oh my god",
	},
	Sounds: []*Sound{
		createSound("god1", 1, 1000),
		createSound("god2", 1, 1000),
		createSound("god3", 1, 1000),
		createSound("god4", 1, 1000),
		createSound("god5", 1, 1000),
		createSound("god6", 1, 1000),
	},
}

var OHNO *SoundCollection = &SoundCollection{
	Prefix: "ohno",
	Commands: []string{
		"!no",
		"oh no",
		"ohno",
	},
	Sounds: []*Sound{
		createSound("no1", 1, 1000),
		createSound("no2", 1, 1000),
		createSound("no3", 1, 1000),
		createSound("no4", 1, 1000),
		createSound("no5", 1, 1000),
		createSound("no6", 1, 1000),
	},
}

var SONG *SoundCollection = &SoundCollection{
	Prefix: "song",
	Commands: []string{
		"!music",
	},
	Sounds: []*Sound{
		createSound("1", 1, 1000),
		createSound("2", 1, 1000),
		createSound("3", 1, 1000),
		createSound("4", 1, 1000),
	},
}

var ZAWARUDO *SoundCollection = &SoundCollection{
	Prefix:   "zw",
	Commands: []string{},
	Sounds: []*Sound{
		createSound("1", 1, 1000),
		createSound("2", 1, 1000),
	},
}

var COLLECTIONS []*SoundCollection = []*SoundCollection{
	OHGOD,
	OHNO,
	ZAWARUDO,
	SONG,
}

func createSound(Name string, Weight int, PartDelay int) *Sound {
	return &Sound{
		Name:      Name,
		Weight:    Weight,
		PartDelay: PartDelay,
		buffer:    make([][]byte, 0),
	}
}

func (sc *SoundCollection) Load() {
	fmt.Printf("Loading collection %v\n", sc.Prefix)
	for _, sound := range sc.Sounds {
		sc.soundRange += sound.Weight
		sound.Load(sc)
	}
	fmt.Printf("Loaded\n")
}

func (s *SoundCollection) Random() *Sound {
	var (
		i      int
		number int = rand.Intn(s.soundRange)
	)
	for _, sound := range s.Sounds {
		i += sound.Weight
		if number < i {
			return sound
		}
	}
	return nil
}

func (s *Sound) Load(c *SoundCollection) error {
	path := fmt.Sprintf("audio/%v_%v.dca", c.Prefix, s.Name)

	file, err := os.Open(path)

	if err != nil {
		fmt.Println("error opening dca file :", err)
		return err
	}

	var opuslen int16

	for {
		// read opus frame length from dca file
		err = binary.Read(file, binary.LittleEndian, &opuslen)

		// If this is the end of the file, just return
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return nil
		}

		if err != nil {
			fmt.Println("error reading from dca file :", err)
			return err
		}

		// read encoded pcm from dca file
		InBuf := make([]byte, opuslen)
		err = binary.Read(file, binary.LittleEndian, &InBuf)

		// Should not be any end of file errors
		if err != nil {
			fmt.Println("error reading from dca file :", err)
			return err
		}

		// append encoded pcm data to the buffer
		s.buffer = append(s.buffer, InBuf)
	}
}

// Plays this sound over the specified VoiceConnection
func (s *Sound) Play(vc *discordgo.VoiceConnection) {
	vc.Speaking(true)
	defer vc.Speaking(false)

	for _, buff := range s.buffer {
		vc.OpusSend <- buff
		if cancelRequest {
			cancelRequest = false
			break
		}
	}
}

// Attempts to find the current users voice channel inside a given guild
func getCurrentVoiceChannel(user *discordgo.User, guild *discordgo.Guild) *discordgo.Channel {
	for _, vs := range guild.VoiceStates {
		if vs.UserID == user.ID {
			channel, _ := dg.State.Channel(vs.ChannelID)
			return channel
		}
	}
	return nil
}

// Prepares a play
func createPlay(user *discordgo.User, guild *discordgo.Guild, coll *SoundCollection, sound *Sound) *Play {
	// Grab the users voice channel
	channel := getCurrentVoiceChannel(user, guild)
	if channel == nil {
		return nil
	}

	// Create the play
	play := &Play{
		GuildID:   guild.ID,
		ChannelID: channel.ID,
		UserID:    user.ID,
		Sound:     sound,
		Forced:    true,
	}

	// If we didn't get passed a manual sound, generate a random one
	if play.Sound == nil {
		play.Sound = coll.Random()
		play.Forced = false
	}

	// If the collection is a chained one, set the next sound
	if coll.ChainWith != nil {
		play.Next = &Play{
			GuildID:   play.GuildID,
			ChannelID: play.ChannelID,
			UserID:    play.UserID,
			Sound:     coll.ChainWith.Random(),
			Forced:    play.Forced,
		}
	}

	return play
}

// Prepares and enqueues a play into the ratelimit/buffer guild queue
func enqueuePlay(user *discordgo.User, guild *discordgo.Guild, coll *SoundCollection, sound *Sound) {
	play := createPlay(user, guild, coll, sound)
	if play == nil {
		return
	}

	// Check if we already have a connection to this guild
	//   yes, this isn't threadsafe, but its "OK" 99% of the time
	_, exists := queues[guild.ID]

	if exists {
		if len(queues[guild.ID]) < MAX_QUEUE_SIZE {
			queues[guild.ID] <- play
		}
	} else {
		queues[guild.ID] = make(chan *Play, MAX_QUEUE_SIZE)
		playSound(play, nil)
	}
}

// Play a sound
func playSound(play *Play, vc *discordgo.VoiceConnection) (err error) {

	if vc == nil {
		vc, err = dg.ChannelVoiceJoin(play.GuildID, play.ChannelID, false, false)
		// vc.Receive = false
		if err != nil {
			delete(queues, play.GuildID)
			return err
		}
	}

	// If we need to change channels, do that now
	if vc.ChannelID != play.ChannelID {
		vc.ChangeChannel(play.ChannelID, false, false)
		time.Sleep(time.Millisecond * 125)
	}

	// Sleep for a specified amount of time before playing the sound
	time.Sleep(time.Millisecond * 32)

	// Play the sound
	play.Sound.Play(vc)

	// If this is chained, play the chained sound
	if play.Next != nil {
		playSound(play.Next, vc)
	}

	// If there is another song in the queue, recurse and play that
	if len(queues[play.GuildID]) > 0 {
		play := <-queues[play.GuildID]
		playSound(play, vc)
		return nil
	}

	// If the queue is empty, delete it
	time.Sleep(time.Millisecond * time.Duration(play.Sound.PartDelay))
	delete(queues, play.GuildID)
	vc.Disconnect()
	return nil
}

func scontains(key string, options []string) bool {
	for _, item := range options {
		if item == key {
			return true
		}
	}
	return false
}
