package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type ServerList struct {
	Servers []Server `json:"servers"`
}

type Server struct {
	ID     string   `json:"id"`
	Names  []string `json:"names"`
	Pastes []Paste  `json:"pastes"`
}

type Paste struct {
	Code     string `json:"code"`
	Response string `json:"response"`
}

var (
	servers      ServerList
	name         map[string][]string          = make(map[string][]string)
	paste        map[string]map[string]string = make(map[string]map[string]string)
	guild_to_ind map[string]int               = make(map[string]int)
	NUM_GUILDS                                = 0
)

//Writes the main ServerList to file
func WriteToJSON() {
	b, err := json.Marshal(servers)
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	f, err := os.Create("guild_info.json")
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	f.Write(b)
}

func ReadFromJSON() {
	f, err := os.Open("guild_info.json")
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	b, err := ioutil.ReadAll(f)
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	json.Unmarshal(b, &servers)
	for _, server := range servers.Servers {
		guild_to_ind[server.ID] = NUM_GUILDS
		NUM_GUILDS++
		name[server.ID] = server.Names
		paste[server.ID] = make(map[string]string)
		for _, p := range server.Pastes {
			paste[server.ID][p.Code] = p.Response
		}
	}
}

func AddNewServer(ID string) {
	newServer := Server{}
	newServer.ID = ID
	paste[ID] = make(map[string]string)
	newServer.Names = []string{"@cirnobot", "cirno"}
	servers.Servers = append(servers.Servers, newServer)
	WriteToJSON()
	name[ID] = newServer.Names
	guild_to_ind[ID] = NUM_GUILDS
	NUM_GUILDS++
}

func makePaste(a, b string) Paste {
	p := Paste{}
	p.Code = a
	p.Response = b
	return p
}
