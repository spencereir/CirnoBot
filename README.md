# CirnoBot

CirnoBot is a catch-all Discord bot that provides some interesting features through a variety of techniques like machine learning. Cirno's functionality is split across multiple files, each of which will be discussed. First, all commands Cirno will currently respond to are listed here.

### Commands

Commands given to Cirno all start with either a name for the bot, or an exclamation mark. "N" will be used to denote a name for the bot in the list of commands to follow. The default names for Cirno are "@CirnoBot" and "Cirno" (all case insensitive). Text in \<triangle brackets\> is to be interpreted as something filled in depending on how the command should be executed. For example `roll <dice>` could be `roll 2d6`. Text in {braces} is optional, and is not required for the command to run. A pipe symbol, "|" denotes options to be chosen. For example, the command `a {b|c}` could be run as `a`, `a b`, or `a c`. Text in [square brackets] can be repeated any number of times. For example, `a [b|c]` could be executed as `a b`, `a b b c b`, etc..

`N roll {<options>} <dice> {op <dice>}` rolls a set of dice given. Look to the section on `roll.go` for a more in-depth look at the roll command.

`N generate stand {<genre>}` will generate a random stand, as inspired by the series "Jojo's Bizarre Adventure". Specifying a genre will restrict the song pool used to choose the stand name.

`!nineball|⑨` will return a standard 8-ball response.

`!farage` will return a random picture of Nigel Farage laughing.

`ZA WARUDO` will "freeze time" on the channel for 5 seconds, deleting all messages posted. It then reposts the messages after the time freeze.

`N add name <new name>` will make Cirno respond to the given new name (in addition to others). Names are channel-specific.

`N reorder <option> {[<option>]}` will reorder the list of words given using the Fisher-Yates shuffle.

`N delete <number>` deletes the previous given number of messages from the current channel. Delete can only delete messages that were sent during the current instance of CirnoBot.

`N choose{<X>} <option> {[<option>]}` will randomly choose an option from the list given. If X is specified, it will instead choose X options, with possibility of overlap.

`N roulette {<X> {<Y>}}` will play a game of Russian Roulette and report the results. If X is specified, the gun will have X bullets loaded. If Y is specified, the gun will have Y chambers.

`N rank <url>` will classify an image at the given URL according to the training sets. See the `classify.go` section for more information.

`N, <message>` will cause Cirno to respond to the given message using a Markov chain.

`N say <message>` will cause Cirno to reply with message.

`N recommend anime <MAL username>` will cause Cirno to reply with a list of the top 3 recommended anime for the given MyAnimeList user. See `anime_recommend.go` for more information.

`N puush <file URL> {<destination filename>}` will cause Cirno to upload the given file to puush, and returns a link to the puush'd file.

`N research <topic>` attempts to research a given topic by linking to a Wikipedia article about it.

`N brexit {leave|remain}` post a meme about the British Referendum. By specifying leave or remain, you can bias the meme to be about your preferred side.

`!god|oh god|ohgod|oh my god` will play a random sound from the Joseph Joestar "Oh my god" collection.

`!no|oh no|ohno` will play a random sound from the Joseph Joestar "Oh no" collection.

### puush.go

Todo

### anime_recommend.go

Todo. Note that the code for `train.go` has mysteriously gone missing, so instead the executable is bundled. Simply place this in the same directory as `users.dat`, and run it to generate a file `r.dat`. This will then be used for recommendations.

### classify.go

Todo

### markov.go

Todo

### roll.go

Todo

### sound.go

Credit for the sound goes to hammerandchisel's Airhorn Bot project, [available on GitHub.](https://github.com/hammerandchisel/airhornbot)

Todo

### stand.go

Todo