package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/pkg/browser"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
)

const redirectURI = "http://localhost:8080/callback"

// just making all things global
var (
	auth  = spotifyauth.New(spotifyauth.WithRedirectURL(redirectURI), spotifyauth.WithScopes(spotifyauth.ScopeUserReadPrivate, spotifyauth.ScopePlaylistModifyPrivate, spotifyauth.ScopePlaylistModifyPublic))
	ch    = make(chan *spotify.Client)
	state = "abc123"
)

// Library is just simple represntation of the dump file, which cares only about Tracks section
type Library struct {
	Tracks []struct {
		Artist string `json:"artist"`
		Album  string `json:"album"`
		Track  string `json:"track"`
		URI    string `json:"uri"`
	} `json:"tracks"`
}

func completeAuth(w http.ResponseWriter, r *http.Request) {
	tok, err := auth.Token(r.Context(), state, r)
	if err != nil {
		http.Error(w, "Couldn't get token", http.StatusForbidden)
		log.Fatal(err)
	}
	if st := r.FormValue("state"); st != state {
		http.NotFound(w, r)
		log.Fatalf("State mismatch: %s != %s\n", st, state)
	}

	// use the token to get an authenticated client
	client := spotify.New(auth.Client(r.Context(), tok))
	fmt.Fprintf(w, "Login Completed! Get back to your terminal")
	ch <- client
}

// addTracksToPlaylist adds tracks from Library to playlist using spotify API with 100 track batches
func AddTracksToPlaylist(client *spotify.Client, pl spotify.SimplePlaylist, lib Library) error {
	trackCount := len(lib.Tracks)

	for i := 0; i < trackCount; i += 100 {
		upperBound := i + 100
		if upperBound > trackCount {
			upperBound = trackCount
		}

		var trackURIs []spotify.ID
		for _, track := range lib.Tracks[i:upperBound] {
			trackID := strings.TrimPrefix(track.URI, "spotify:track:")
			trackURIs = append(trackURIs, spotify.ID(trackID))
		}
		_, err := client.AddTracksToPlaylist(context.Background(), pl.ID, trackURIs...)
		if err != nil {
			return err
		}

		log.Printf("Adding tracks %d to %d\n", i+1, upperBound)
	}
	log.Printf("Done, all %d were added", trackCount)
	return nil
}
func main() {
	// define flags for the CLI
	playlist := flag.String("playlist", "restore", "Name of the playlist")
	filePath := flag.String("filepath", "./YourLibrary.json", "Path of the file")

	// parse the flags
	flag.Parse()

	jsonFile, err := ioutil.ReadFile(*filePath)
	if err != nil {
		fmt.Println(err)
	}

	var library Library
	err = json.Unmarshal(jsonFile, &library)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Tracks num: %d", len(library.Tracks))

	// first start an HTTP server
	http.HandleFunc("/callback", completeAuth)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Got request for:", r.URL.String())
	})
	go func() {
		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			log.Fatal(err)
		}
	}()

	url := auth.AuthURL(state)

	err = browser.OpenURL(url)
	if err != nil {
		log.Printf("failed to open browser with error: %s. please do it manually", err)
	}
	log.Println("Please log in to Spotify by visiting the following page in your browser:", url)

	// wait for auth to complete
	client := <-ch

	// use the client to make calls that require authorization
	user, err := client.CurrentUser(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	log.Println("You are logged in as:", user.ID)
	pls, err := client.CurrentUsersPlaylists(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	var found bool
	for _, pl := range pls.Playlists {
		fmt.Println(pl.Name)
		if pl.Name == *playlist {
			found = true
			err := AddTracksToPlaylist(client, pl, library)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
	if !found {
		log.Printf("No playlist with name %s were found, make sure you have one\n", *playlist)
	}

}
