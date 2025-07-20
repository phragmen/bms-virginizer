package main

import (
	"bytes"
	"testing"
	"encoding/binary"
)


// helper to create an N‑bank synthetic dump whose CS1/CS2 are already valid
func makeBank(pattern byte) []byte {
	cfg := BMS_MP_Layout()
	bank := bytes.Repeat([]byte{pattern}, bankSize)
	// drop dummy VIN, ISN, keys so we see changes
	copy(bank[cfg.vinOff:], []byte("WBA0TEST0VIN00000"))
	for i := 0; i < cfg.isnLen; i++ { bank[cfg.isnOff+i] = 0xAA }
	for i := 0; i < cfg.keysLen; i++ { bank[cfg.keysOff+i] = 0xBB }
	// correct CRCs
	cs1 := crc16(bank[cfg.cs1Start : cfg.cs1End+1])
	cs2 := crc16(bank[cfg.cs2Start : cfg.cs2End+1])
	binary.BigEndian.PutUint16(bank[cfg.cs1Off:], cs1)
	binary.BigEndian.PutUint16(bank[cfg.cs2Off:], cs2)
	return bank
}

func TestVirginizer_AllBanks(t *testing.T) {
	cfg := BMS_MP_Layout()
	const banks = 6                     // emulate a 192 KB 1250 GS dump
	dump := bytes.Buffer{}
	for i := 0; i < banks; i++ {
		dump.Write(makeBank(byte(i)))   // each bank filled with 0x00,0x01…
	}

	if err := processBanks(dump.Bytes(), 0xFF, true, cfg); err != nil {
		t.Fatal(err)
	}

	// verify every bank got wiped + new CRCs are correct
	for b := 0; b < banks; b++ {
		off := b * bankSize
		bank := dump.Bytes()[off : off+bankSize]

		// VIN wiped?
		for i := 0; i < cfg.vinLen; i++ {
			if bank[cfg.vinOff+i] != 0xFF {
				t.Fatalf("bank %d VIN not wiped", b)
			}
		}
		// CS1/CS2 valid?
		cs1 := crc16(bank[cfg.cs1Start : cfg.cs1End+1])
		cs2 := crc16(bank[cfg.cs2Start : cfg.cs2End+1])
		if cs1 != binary.BigEndian.Uint16(bank[cfg.cs1Off:]) {
			t.Fatalf("bank %d CS1 mismatch", b)
		}
		if cs2 != binary.BigEndian.Uint16(bank[cfg.cs2Off:]) {
			t.Fatalf("bank %d CS2 mismatch", b)
		}
	}
}

func TestVirginizer_OnlyFirst(t *testing.T) {
	cfg := BMS_MP_Layout()
	dump := bytes.Buffer{}
	dump.Write(makeBank(0x11))
	dump.Write(makeBank(0x22)) // two‑bank 64 KB file

	if err := processBanks(dump.Bytes(), 0xFF, false, cfg); err != nil {
		t.Fatal(err)
	}

	b0 := dump.Bytes()[:bankSize]
	b1 := dump.Bytes()[bankSize:]

	// bank 0 VIN wiped, bank 1 preserved
	for i := 0; i < cfg.vinLen; i++ {
		if b0[cfg.vinOff+i] != 0xFF {
			t.Fatal("bank 0 VIN not wiped")
		}
		if b1[cfg.vinOff+i] == 0xFF {
			t.Fatal("bank 1 VIN should remain")
		}
	}
}
