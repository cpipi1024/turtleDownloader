package main

import (
	"log"
	"os"

	"cpipi1024.com/turtleDownloader/utils/torrentfile"
)

func main() {
	inpath := os.Args[1]

	outPath := os.Args[2]

	tf, err := torrentfile.Open(inpath)

	if err != nil {
		log.Fatal(err)
	}

	err = tf.DownLoad(outPath)

	if err != nil {
		log.Fatal(err)
	}
}
