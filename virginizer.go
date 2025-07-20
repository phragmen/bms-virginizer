// virginizer.go  (multi‑bank, CRC‑safe)
//
// Build:  go build -o virginizer virginizer.go
//
// By default it wipes security data in *all* 0x8000‑byte banks;
// use -only-first if you want to keep mirrors unchanged.
//
// Example:
//   ./virginizer -in 1250gs_192k.bin -out virgin.bin       # wipe all banks
//   ./virginizer -in 1250gs_192k.bin -only-first           # wipe bank 0 only
//
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	log "github.com/sirupsen/logrus"
  "os"
)

func main() {
	bmsType := flag.String("type", "BMS-O", "device type: BMS-MP or BMS-O")
	inPath := flag.String("in", "", "input EEPROM dump (.bin)")
	outPath := flag.String("out", "virgin_eeprom.bin", "output file")
	fillByte := flag.Uint("fill", 0xFF, "byte used to blank security areas")
	onlyFirst := flag.Bool("only-first", false,
		"wipe VIN/ISN/keys in bank 0 only (mirrors left untouched), ignored for BMS-O")

	flag.Parse()

	if *inPath == "" {
		fmt.Printf("Usage: virginizer -type [bms_type] -in [input_file] -out [output_file] [OPTIONS]\n")
		flag.PrintDefaults()
		if *inPath == "" {
			log.Fatal("-in is required")
		}
		os.Exit(1)
	}

	// 1. read file --------------------------------------------------
	buf, err := ioutil.ReadFile(*inPath)
	if err != nil {
		log.Fatal(err)
	}
	if len(buf)%bankSize != 0 {
		log.Fatalf("file size %d is not a multiple of 0x%x", len(buf), bankSize)
	}
	banks := len(buf) / bankSize
	fmt.Printf("Detected %d bank(s) (%d KiB total)\n", banks, len(buf)/1024)


	cfg, err := buildLayout(*bmsType)
	if err != nil {
		log.Fatal(err)
	}

	if err = processBanks(buf, byte(*fillByte), !*onlyFirst || cfg.wipeAll, cfg); err != nil {
		log.Fatal(err)
	}

	// 3. write back ------------------------------------------------
	if err := ioutil.WriteFile(*outPath, buf, 0644); err != nil {
		log.Fatal(err)
	}
	fmt.Println("✓ Virgin file written to", *outPath)
}
