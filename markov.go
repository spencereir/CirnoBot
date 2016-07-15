package main

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"strings"
	_ "time"
)

type Prefix []string

//Choose the rarest nonzero word to base it off of
var ngram map[string]int = make(map[string]int)

func (p Prefix) String() string {
	return strings.Join(p, " ")
}

func (p Prefix) Shift(word string) {
	copy(p, p[1:])
	p[len(p)-1] = word
}

type Chain struct {
	chain     map[string][]string
	prefixLen int
}

func NewChain(prefixLen int) *Chain {
	return &Chain{make(map[string][]string), prefixLen}
}

func (c *Chain) Build(r io.Reader) {
	br := bufio.NewReader(r)
	p := make(Prefix, c.prefixLen)
	for {
		var s string
		if _, err := fmt.Fscan(br, &s); err != nil {
			break
		}
		ngram[strings.ToLower(s)]++
		key := p.String()
		c.chain[key] = append(c.chain[key], s)
		p.Shift(s)
	}
}

func (c *Chain) Generate(seed []string, n int) string {
	p := make(Prefix, c.prefixLen)
	for i, v := range seed {
		p[i] = v
	}
	var words []string
	words = append(words, p[0])
	words = append(words, p[1])
	for i := 0; i < n; i++ {
		choices := c.chain[p.String()]
		if len(choices) == 0 {
			if len(words) == 2 {
				return "I couldn't find anything for that starting seed. A larger corpus will allow for a greater range of responses."
			}
			break
		}
		next := choices[rand.Intn(len(choices))]
		words = append(words, next)
		p.Shift(next)
	}
	return strings.Join(words, " ")
}
