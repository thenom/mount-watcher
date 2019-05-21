package main

import (
	"bufio"
	"flag"
	"github.com/docker/docker/pkg/mount"
	"log"
	"os"
	"strings"
)

type mountLine struct {
	source     string
	mountPoint string
	filesystem string
	options    []string
	dump       string
	fsck       string
}

var fstabs []mountLine
var includeNoAuto bool
var includeSwap bool
var postScript string
var preScript string
var singleMountPoint string

func main() {
	// setup the sysv args first
	flag.BoolVar(&includeNoAuto, "includeNoAuto", false, "Process fstab lines that have noauto in the options")
	flag.BoolVar(&includeSwap, "includeSwap", false, "Process fstab lines that are for swap")
	flag.StringVar(&preScript, "preScript", "", "The script to run before the mount check")
	flag.StringVar(&postScript, "postScript", "", "The script to run after the mount check")
	flag.StringVar(&singleMountPoint, "singleMountPoint", "", "Check only this mount point")
	flag.Parse()

	// read in fstab
	fstab, err := os.Open("/etc/fstab")
	if err != nil {
		log.Fatal(err)
	}
	defer fstab.Close()

	scanner := bufio.NewScanner(fstab)
	for scanner.Scan() {
		thisLine := scanner.Text()

		if len(thisLine) > 0 && thisLine[:1] != "#" {
			appendFstab(thisLine)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	log.Println("Fstab readin complete")

	// finally check the built fstabs var and see if they are mounted
	for _, v := range fstabs {
		log.Println("Checking mount point:", v.mountPoint)
		isMounted, err := mount.Mounted(v.mountPoint)
		if err != nil {
			log.Println("Error occured checking mount:", err)
		}
		log.Println("--- mounted:", isMounted)

	}
}

func appendFstab(line string) {
	// split the raw fstab line
	splitLine := strings.Fields(line)

	// assign the values to the struct
	newLine := mountLine{
		source:     splitLine[0],
		mountPoint: splitLine[1],
		filesystem: splitLine[2],
		options:    strings.Split(splitLine[3], ","),
		dump:       splitLine[4],
		fsck:       splitLine[5],
	}

	// check to see if this is in the singleMountPoint sys arg
	if singleMountPoint == "" {
		// update the fstabs
		if newLine.filesystem == "swap" {
			if !includeSwap {
				return
			}
		} else {
			if !includeNoAuto {
				for _, v := range newLine.options {
					if v == "noauto" {
						return
					}
				}
			}
		}
	} else {
		if singleMountPoint != newLine.mountPoint {
			return
		}
	}
	fstabs = append(fstabs, newLine)

	log.Println("Added new fstab line:", newLine)
}
