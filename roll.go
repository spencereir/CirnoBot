package main

import (
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"strings"
)

func roll(words []string) string {
	if len(words) <= 2 || words[2] == "help" {
		msg := "Cirno's roll command allows you to roll dice, from simple to complex.\n"
		msg += "Cirno supports basic dice rolling: @CirnoBot roll 2d6 will roll 2 6-sided dice and report the total\n"
		msg += "The dice pools can be arbitrarily complex: @CirnoBot roll 2d6 + 4d10 - 3 will work exactly as you expect\n"
		msg += "Cirno also allows for opposed dice rolling: @CirnoBot roll 2d6 opp d12 will roll 2 6-sided dice in one pool, and one 12-sided dice in another, and compare the results\n"
		msg += "Thresholds are also an option: @CirnoBot roll thresh18 d20 + 4 will report a success if the roll is at least 18, or a failure otherwise\n"
		msg += "By enabling quiet mode, only the results will be printed, not the rolls themselves\n"
		msg += "Cirno also supports different games. The default is Dungeons and Dragons style, but for Exalted/Shadowrun, simply specify the game at the beginning, and Cirno will handle glitches, successes, etc.\n"
		msg += "Full command syntax: @CirnoBot roll [game] [options] pool1 [opp pool2]\n"
		msg += "(Optional portions noted in square brackets)\n"
		msg += "Games supported:\n"
		msg += "Dungeons and Dragons: leave game section blank\n"
		msg += "Shadowrun: use the game code \"sr\" or \"shadowrun\"\n"
		msg += "Exalted: use the game code \"ex\" or \"exalted\"\n\n"
		msg += "Note that options must be specified in the order that they are shown below\n"
		msg += "Options:\n"
		msg += "threshx: sets a threshold of x. For example, thresh32 makes the threshold 32. Threshold is interpreted differently depending on game\n"
		msg += "q or quiet: tells Cirno only to report the results, not the dice rolls themselves. Useful when rolling a lot of dice, or making a skill check in D&D.\n"
		return msg
	}
	startind := 2
	total := 0
	num := 0
	place := 0
	switched := false
	multidice := true

	//General dice parameters
	//Thresh may be interpreted based on the game specified
	thresh := -1
	err := error(nil)
	quiet := false
	opp := false
	total2 := 0

	//Shadowrun parameters
	sr := false
	srs := 0
	srf := 0
	opps := 0
	oppf := 0
	oppdr := 0
	dice_rolled := 0

	roll := ""

	//Check for roll options
	//First check game styles
	if words[startind] == "shadowrun" || words[startind] == "sr" || words[startind] == "exalted" || words[startind] == "ex" {
		sr = true
		startind++
	}
	if strings.HasPrefix(words[startind], "thresh") {
		if len(words[startind]) <= 6 {
			return "Thresh must have a parameter specified as the threshold, like thresh32 or thresh10."
		}
		thresh, err = strconv.Atoi(words[startind][6:])
		if err != nil {
			return "I excepted an integer after thresh, but instead got \"" + words[startind][6:] + "\"."
		}

		startind++
	}
	if words[startind] == "quiet" || words[startind] == "q" {
		quiet = true
		startind++
	}
	words[startind-1] = "+"
	for i := startind - 1; i+1 < len(words); i += 2 {
		if words[i] == "opp" || words[i] == "opponent" || words[i] == "opposed" {
			opp = true
			words[i] = "+"
		}
		place = 0
		num = 0
		multidice = false
		switched = false
		for j := 0; j < len(words[i+1]); j++ {
			if words[i+1][j] == 'd' {
				switched = true
			} else if words[i+1][j] < '0' || words[i+1][j] > '9' {
				return "I found an unexpected character; \"" + string(words[i+1][j]) + "\". The only dice format I can read is XdY where X and Y are integers"
			} else if switched {
				place = 10*place + int(words[i+1][j]-'0')
			} else {
				multidice = true
				num = 10*num + int(words[i+1][j]-'0')
			}
		}
		v := 0
		if multidice && switched {
			s := make([]int, 0)
			for i := 0; i < num; i++ {
				r := rand.Intn(place) + 1
				v += r
				if sr && r >= 5 {
					if !opp {
						srs++
					} else {
						opps++
					}
				} else if sr && r == 1 {
					if !opp {
						srf++
					} else {
						oppf++
					}
				}
				s = append(s, r)
			}
			if !quiet {
				sort.Ints(s)
				roll += words[i+1] + ": "
				for i := 0; i < num; i++ {
					if i > 0 {
						roll += ", "
					}
					roll += strconv.Itoa(s[i])
				}
				roll += "\n"
			}
			if !opp {
				dice_rolled += num
			} else {
				oppdr += num
			}
		} else if switched == true {
			r := rand.Intn(place) + 1
			if sr && r >= 5 {
				if !opp {
					srs++
				} else {
					opps++
				}
			} else if sr && r == 1 {
				if !opp {
					srf++
				} else {
					opps++
				}
			}
			v += r
			if !quiet {
				roll += words[i+1] + ": "
				roll += strconv.Itoa(r)
				roll += "\n"
			}
			if !opp {
				dice_rolled++
			} else {
				oppdr++
			}
		} else {
			v += num
		}
		if words[i] == "-" {
			if opp {
				total2 -= v
			} else {
				total -= v
			}
		} else if words[i] == "+" {
			if opp {
				total2 += v
			} else {
				total += v
			}
		} else {
			return "I wasn't sure how to interpret \"" + words[i] + "\". In the middle of rolls, I can only read plus and minus signs, as well as opposed checks (indicated by opp)."
		}
	}

	if sr {
		if srs == 1 {
			roll += "1 hit"
		} else {
			roll += fmt.Sprintf("%v hits", srs)
		}
		if opp {
			roll += " vs "
			if opps == 1 {
				roll += "1 hit"
			} else {
				roll += fmt.Sprintf("%v hits", opps)
			}
		}
		roll += "\n"
	} else {
		roll += strconv.Itoa(total)
		if opp {
			roll += " vs "
			roll += strconv.Itoa(total2)
		}
		roll += "\n"
	}

	if sr {
		if thresh != -1 {
			if srs >= thresh {
				roll += "Success\n"
				if srs == thresh+1 {
					roll += "1 net hit\n"
				} else {
					roll += fmt.Sprintf("%v net hits", srs-thresh)
				}
			} else {
				roll += "Failure\n"
			}
		} else if opp {
			if srs > opps {
				roll += "Success\n"
			} else if srs == opps {
				roll += "Tie\n"
			} else {
				roll += "Failure\n"
			}
		}
		if srf >= (dice_rolled+1)/2 {
			if srs == 0 {
				roll += "Critical Glitch\n"
			} else {
				roll += "Glitch\n"
			}
		}
		if opp && oppf >= (oppdr+1)/2 {
			if opps == 0 {
				roll += "Opponent Critical Glitch\n"
			} else {
				roll += "Opponent Glitch\n"
			}
		}
	} else if thresh != -1 {
		if total >= thresh {
			roll += "Success\n"
		} else {
			roll += "Failure\n"
		}
	} else if opp {
		if total > total2 {
			roll += "Success\n"
		} else if total == total2 {
			roll += "Tie\n"
		} else {
			roll += "Failure\n"
		}
	}
	return roll
}
