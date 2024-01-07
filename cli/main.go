package main

import (
	"flag"
	"io/fs"
	"log"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
)

const SRC_FILE_SUFFIX = ".md"
const H1_HEADER_PREFIX = "#"

var flagBaseDir string
var flagServerId string

func init() {
	flag.StringVar(&flagBaseDir, "basedir", "", "base dir to look for docs")
	flag.StringVar(&flagServerId, "serverid", "", "server id aka guild id")
}

func validFile(f fs.DirEntry) bool {
	return strings.HasSuffix(f.Name(), SRC_FILE_SUFFIX)
}

func sortFileNameAscend(files []fs.DirEntry) {
	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() < files[j].Name()
	})
}

// TODO: test
func fileHeaderString(n string) string {
	// Trims the '<n>-' name from the filename to create a markdown header
	n = strings.TrimSuffix(n, ".md")
	index := strings.Index(n, "-")
	if index >= 0 {
		beforedash := n[:index]
		log.Printf("HEADER %s", beforedash)
		if _, err := strconv.Atoi(beforedash); err == nil {
			// looks like a number, only return after the dash
			return H1_HEADER_PREFIX + " " + strings.TrimPrefix(n[index:], "-")
		}
	}
	return H1_HEADER_PREFIX + " " + n
}

func main() {
	// base dir
	// discord auth token
	// only
	flag.Parse()

	dg, err := discordgo.New("Bot " + os.Getenv("TOKEN"))
	dg.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Println("INFO: bot ready")
	})

	err = dg.Open()
	if err != nil {
		log.Fatalf("ERROR: Cannot open the session: %v", err)
	}
	defer dg.Close()

	log.Printf("INFO: state %+v", dg.State)

	if len(dg.State.Guilds) == 0 {
		log.Fatalf("ERROR: no guilds")
	}

	var foundGuild *discordgo.Guild

	for _, guild := range dg.State.Guilds {
		if flagServerId == guild.ID {
			foundGuild = guild
			break
		}
	}

	log.Printf("INFO: found guild %s", foundGuild.ID)

	srcDirs, err := os.ReadDir(flagBaseDir)
	if err != nil {
		log.Fatalf("ERROR: could not read base dir: %v", err)
	}

	for _, srcDir := range srcDirs {
		if !srcDir.IsDir() {
			log.Printf("DEBUG: skipping %s, not a dir", srcDir.Name())
			continue
		}

		channelId := srcDir.Name()
		channel, err := dg.Channel(channelId)
		if err != nil {
			log.Printf("ERROR: could not get channel from id %s: %v", channelId, err)
			continue
		}

		log.Printf("INFO: got channel %v", channel)

		p := path.Join(flagBaseDir, srcDir.Name())
		files, err := os.ReadDir(p)
		if err != nil {
			log.Fatal(err)
		}

		sortFileNameAscend(files)

		// TODO: more than 100 messages?
		channelMessages, err := dg.ChannelMessages(channelId, 100, "", "", "")
		if err != nil {
			log.Printf("ERROR: could not get channel messages from id %s: %v", channelId, err)
			continue
		}

		// messages that can be edited by this bot session
		candidateMessages := []*discordgo.Message{}
		for _, msg := range channelMessages {
			if dg.State.User.ID == msg.Author.ID {
				candidateMessages = append(candidateMessages, msg)
			}
		}

		numCandidateMessages := len(candidateMessages)
		log.Printf("INFO: %d candidate messages in channel", numCandidateMessages)

		filei := 0
		for _, f := range files {
			if !validFile(f) {
				log.Printf("INFO: %s not a valid file name", f.Name())
				continue
			}
			log.Printf("INFO: processing '%s' valid file", f.Name())
			fp := path.Join(flagBaseDir, srcDir.Name(), f.Name())
			filecontent, err := os.ReadFile(fp)
			if err != nil {
				log.Printf("ERROR: '%s' valid file", f.Name())
			}

			var b strings.Builder
			b.WriteString(fileHeaderString(f.Name()))
			b.WriteString("\n")
			b.WriteString(string(filecontent))

			if filei < numCandidateMessages {
				// an edit
				// messages come in oldest last so move from last index
				msg := candidateMessages[numCandidateMessages-filei-1]
				log.Printf("INFO: editing message index:, %d, message id: %s, file name: '%s'", filei, msg.ID, f.Name())
				_, err := dg.ChannelMessageEdit(channelId, msg.ID, b.String())
				if err != nil {
					log.Printf("ERROR: error editing message '%v'", err)
				}
			} else {
				// a create
				log.Printf("INFO: creating message '%s'", f.Name())
				_, err := dg.ChannelMessageSend(channelId, b.String())
				if err != nil {
					log.Printf("ERROR: error sending message '%v'", err)
				}
			}
			filei += 1
		}
	}

	log.Println("INFO: done")
}
