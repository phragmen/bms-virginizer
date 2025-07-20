// virginizer_internal.go  (same package as main, no exports needed)
package main

import (
	"encoding/binary"
	"errors"
	"fmt"
  "strings"
)


const bankSize = 0x8000 // 32 KiB

// ---------- CRC‑16/CCITT‑F0 ----------
func crc16(data []byte) uint16 {
	var crc uint16 = 0xFFFF
	for _, b := range data {
		crc ^= uint16(b) << 8
		for i := 0; i < 8; i++ {
			if crc&0x8000 != 0 {
				crc = (crc << 1) ^ 0x1021
			} else {
				crc <<= 1
			}
		}
	}
	return crc
}

// ---------- layout for one 32 KiB block ----------
type layout struct {
	vinOff, vinLen   int
	isnOff, isnLen   int
	keysOff, keysLen int
	cs1Start, cs1End int
	cs1Off           int
	cs2Start, cs2End int
	cs2Off           int
  wipeAll          bool
}

func BMS_MP_Layout() *layout {
	return &layout{
		vinOff:  0x20, vinLen: 17,
		isnOff:  0x100, isnLen: 16,
		keysOff: 0x140, keysLen: 40,
		cs1Start: 0x0000, cs1End: 0x07FD, cs1Off: 0x07FE,
		cs2Start: 0x0800, cs2End: 0x0FFD, cs2Off: 0x0FFE,
    wipeAll: false,
	}
}

func BMS_O_Layout() *layout {
    return &layout{
        vinOff:  0x40,  vinLen: 17,
        isnOff:  0x120, isnLen: 16,
        keysOff: 0x180, keysLen: 48, // 6 key slots
        cs1Start: 0x0000, cs1End: 0x07FD, cs1Off: 0x07FE,
        cs2Start: 0x0800, cs2End: 0x0FFD, cs2Off: 0x0FFE,
        wipeAll: true,
    }
}

func buildLayout(bmsType string) (*layout, error) {
  switch(strings.ToUpper(bmsType)) {
    case "BMS-MP", "BMSMP", "MP":
      return BMS_MP_Layout(), nil
    case "BMS-O", "BMSO", "O":
      return BMS_O_Layout(), nil
    default:
      return nil, fmt.Errorf("unknown BMS type '%s'", bmsType)
  }
}

// processBanks mutates buf in‑place.
// Returns error if size isn't N×0x8000 or offsets run past bank.
func processBanks(buf []byte, fillByte byte, wipeAll bool, cfg *layout) error {

  if len(buf)%bankSize != 0 {
		return fmt.Errorf("file size %d is not a multiple of 0x%x", len(buf), bankSize)
	}
	banks := len(buf) / bankSize
	fmt.Printf("Detected %d bank(s) (%d KiB total)\n", banks, len(buf)/1024)

  for b := 0; b < banks; b++ {
		base := b * bankSize
		block := buf[base : base+bankSize]

		// 2a. wipe security areas
		if wipeAll || b == 0 {
			wipe := func(off, ln int) {
				for i := off; i < off+ln && i < len(block); i++ {
					block[i] = fillByte
				}
			}
			wipe(cfg.vinOff, cfg.vinLen)
			wipe(cfg.isnOff, cfg.isnLen)
			wipe(cfg.keysOff, cfg.keysLen)
		}

		// 2b. fix checksums
		updateCRC := func(start, end, where int) error {
			if end >= len(block) || where+1 >= len(block) {
				return errors.New("CRC offset outside bank")
			}
			crc := crc16(block[start : end+1])
			binary.BigEndian.PutUint16(block[where:], crc)
			return nil
		}
		if err := updateCRC(cfg.cs1Start, cfg.cs1End, cfg.cs1Off); err != nil {
			return err
		}
		if err := updateCRC(cfg.cs2Start, cfg.cs2End, cfg.cs2Off); err != nil {
			return err
		}
	}

  return nil

}
