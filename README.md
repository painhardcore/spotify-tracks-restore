# spotify-tracks-restore

# Why

> Everybody gangsta till spotify close your account because you are from *unwanted* country

Anyway, if you want to get back your tracks which you liked and not organized then into playlist - this tool is for you.

## Go is required
If don't have it yet - please visit https://go.dev/dl/

## Your YourLibrary.json from data dump
Access your Spotify account dashboard at https://www.spotify.com/. In the privacy settings, you'll find the option to request your data.

## Register your App and get tokens

Start by registering your application at the following page:

https://developer.spotify.com/my-applications/

You'll get a client ID and secret key for your application. An easy way to provide this data to your application is to set the SPOTIFY_ID and SPOTIFY_SECRET environment variables. 

RedirectUrl should be http://localhost:8080/callback

## How to use
Download dependancies
```shell
go mod download
```

If you want to keep it simple - create a playlist with name **restore** and put YourLibrary.json file in the same folder
```shell
SPOTIFY_ID=XXXX SPOTIFY_SECRET=YYYY go run main.go
```

Otherwise, you can define them via arguments like this: 

```shell
go run main.go --playlist myPlaylist --filepath /path/to/file.json
```
