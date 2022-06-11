package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/Snawoot/transmission-resync/hub"
	"github.com/Snawoot/transmission-resync/spoke"

	"github.com/hekmon/transmissionrpc/v2"
	"github.com/spf13/viper"
)

var (
	home, _ = os.UserHomeDir()
	version = "undefined"
)

func run() int {
	fs := flag.NewFlagSet("transmission-resync", flag.ExitOnError)
	configFilename := fs.String("conf", filepath.Join(home, ".config", "transmission-resync.yaml"), "path to configuration file")
	targetHash := fs.String("hash", "", "target torrent hash")
	showVersion := fs.Bool("version", false, "show program version and exit")
	fs.Parse(os.Args[1:])

	if *showVersion {
		fmt.Println(version)
		return 0
	}

	lowerHash := strings.ToLower(*targetHash)
	if lowerHash == "" {
		log.Fatal("target hash is not specified! please use -hash option.")
	}

	viper.SetConfigType("yaml")
	viper.SetConfigFile(*configFilename)
	setDefaults(viper.GetViper())
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("unable to read config file: %s", err)
	}

	ctx := context.Background()
	if dur := viper.GetDuration("total_timeout"); dur != 0 {
		ctx1, cl := context.WithTimeout(ctx, dur)
		defer cl()
		ctx = ctx1
	}

	var chainCfg Chain
	if err := viper.UnmarshalKey("chain", &chainCfg); err != nil {
		log.Fatalf("unable to unmarshal `chain` section of config")
	}

	var spokes []hub.Spoke
	for _, chainItem := range chainCfg {
		spokes = append(spokes, spoke.NewSpoke(chainItem.Timeout, chainItem.Command))
	}
	hub := hub.NewHub(spokes)

	trpc, err := transmissionrpc.New(
		viper.GetString("rpc.host"),
		viper.GetString("rpc.user"),
		viper.GetString("rpc.password"),
		&transmissionrpc.AdvancedConfig{
			HTTPS:       viper.GetBool("rpc.https"),
			Port:        uint16(viper.GetUint32("rpc.port")),
			RPCURI:      viper.GetString("rpc.uri"),
			HTTPTimeout: viper.GetDuration("rpc.httptimeout"),
			UserAgent:   viper.GetString("rpc.useragent"),
			Debug:       viper.GetBool("rpc.debug"),
		},
	)
	if err != nil {
		log.Fatalf("unable to construct transmission RPC client: %v", err)
	}

	torrents, err := trpc.TorrentGetAll(context.Background())
	if err != nil {
		log.Fatalf("unable to get torrents: %v", err)
	}

	var needle *transmissionrpc.Torrent
	for i := range torrents {
		if torrents[i].HashString == nil {
			continue
		}
		if strings.ToLower(*torrents[i].HashString) == lowerHash {
			torrent := torrents[i]
			needle = &torrent
			break
		}
	}
	if needle == nil {
		log.Fatalf("requested torrent not found")
	}
	if needle.ID == nil {
		log.Fatalf("requested torrent has no ID!")
	}

	res, err := hub.Query(ctx, needle)
	if err != nil {
		log.Fatalf("error querying hub: %v", err)
	}

	log.Printf("resolved new torrent link: %q", res)
	if res == "" {
		log.Println("no torrent update was found!")
		return 3
	}

	log.Printf("removing old torrent with ID=%d...", *needle.ID)
	if err := trpc.TorrentRemove(
		ctx, transmissionrpc.TorrentRemovePayload{IDs: []int64{*needle.ID}},
	); err != nil {
		log.Fatalf("unable to remove old torrent: %v", err)
	}

	log.Printf("removed. Adding new torrent with link %q", res)
	falseVal := false
	if t, err := trpc.TorrentAdd(
		ctx, transmissionrpc.TorrentAddPayload{
			Filename: &res,
			Paused: &falseVal,
		},
	); err != nil {
		log.Fatalf("unable to add new torrent: %v", err)
	} else if t.ID != nil {
		log.Printf("Added torrent with ID %d", *t.ID)
	}

	return 0
}

func main() {
	log.Default().SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
	log.Default().SetPrefix("TRANSMISSION-RESYNC: ")
	os.Exit(run())
}
