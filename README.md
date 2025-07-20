# bmsmp-virginizer
BMSMP and BMSO ECU Virginizer.

_Example with BMW-Explorer:_  
`Extra options → Renew EMS (BMS-MP)` → software asks _“Unlock X\_EMS?”_ → **Yes** → wait until progress bar reaches 100 % and get a prompt to **save the renewed file**. [help.auto-explorer.com](https://help.auto-explorer.com/help15/)

If your tool does **not** have an auto-renew button you can edit the EEPROM in a hex editor and run a BMS-MP checksum calculator, but that’s only recommended if you already know the byte layout.

Or you can use this tool.

### What you’ll need

| Item                                                                                                                                               | Notes                                           |
| -------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------- |
| Bench power supply, 13.8 V/5 A                                                                                                                     | Keep voltage rock-steady while flashing.        |
| **TriCore-capable programmer** (one of): BMW-Explorer + RMI interface, AutoHex II/HexProg, Xhorse VVDI BIM Tool Pro, Ktag/Kess3 in boot-mode, etc. | Needs BMS-MP protocol (TC1797/1796 GPT).        |
| BMS-MP breakout harness or ECU jig                                                                                                                 | Gives +BATT, IGN, GND, CAN L/H, GPT0/1.         |
| ISTA-P or ISTA+ with ICOM or ENET lead                                                                                                             | For “Control-unit replacement” after reinstall. |
| Valid key fob (or two) registered to the bike                                                                                                      | For post-flash teach-in.                        |


### How it works (step-by-step)

A full BMS‑MP image is simply 2–8 identical 32 KiB banks concatenated:
| ECU variant               | Total dump size | # of 32 KiB banks | Notes                             |
| ------------------------- | --------------- | ----------------- | --------------------------------- |
| Early BMS‑KP              |    64 KiB       |  2                | CS1/CS2 repeat every 0x2000       |
| BMS‑MP mid‑gen            |  128 KiB        |  4                | common on F‑series twins          |
| BMS‑MP “6‑block” (R 1250) |  192 KiB        |  6                | VIN/ISN only meaningful in bank 0 |
| Rare BMS‑MP large         |  256 KiB        |  8                | mostly K‑1600 GT/GTL              |

Algorithm

| Step                 | What the code does                                                                                                    | Adjustable via flag                    |
| -------------------- | --------------------------------------------------------------------------------------------------------------------- | -------------------------------------- |
| **1. Read**          | Loads the raw dump into RAM (no size limit).                                                                          | `-in`                                  |
| **2. Blank**         | Overwrites the VIN, ISN/secret key and five Keyless-Ride slots with `0xFF` (or any byte you choose) in each/first block.    | `-fill` `-only-first`            |
| **3 & 4. Checksums** | Re-computes CS1 and CS2 with the industry-standard CRC-16/CCITT-F0 polynomial 0x1021 and writes them back big-endian in each/first block. | `-only-first`      |
| **5. Write**         | Saves the modified dump ready for flashing in boot-mode.                                                              | `-out`                                 |

The default offsets match the **32 k × 6 block (192 k) BMS-MP** layout seen on R-/K-series bikes: VIN at 0x20, ISN at 0x100, key table at 0x140, CS1 at 0x07FE and CS2 at 0x0FFE. Adjust them if your dump is the 64 k or 128 k variant (the tool works with any size because the ranges are fully tunable).  
The two-checksum scheme (`CS1` + `CS2`) is standard on late Bosch bike ECUs; repair tools such as _Checksum-Tool v1.x_ do exactly the same maths before writing back the file. [bmw-az.info](https://bmw-az.info/checksum-software/34-checksum-tool-v10.html?utm_source=chatgpt.com)

The CRC-16/CCITT implementation follows the reference polynomial 0x1021 (`x^16 + x^12 + x^5 + 1`), the same one listed by the ISO/ITU spec and explained in the CRC article on Wikipedia. [Wikipedia](https://en.wikipedia.org/wiki/Cyclic_redundancy_check?utm_source=chatgpt.com)

## How BMS‑O differs from BMS‑MP

| Aspect               | BMS‑MP (ME17.2.4/17.2.41)                                  | **BMS‑O** (ME17.2.42)                                                                                                                  |
| -------------------- | ---------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------- |
| Typical bikes        | R‑1200 LC, S1000R/RR up to Euro 4                          | R‑1250 family, late S1000RR, Euro 5                                                                                                    |
| EEPROM size in dumps | 64 k, 128 k, 192 k                                         | **256 k** (8 × 0x8000 banks)                                                                                                           |
| Security bytes       | VIN @ 0x20, ISN @ 0x100, 5 × 8‑byte keys @ 0x140           | **VIN @ 0x40, ISN @ 0x120, 6 × 8‑byte keys @ 0x180**                                                                                   |
| Checksums per bank   | CS1 @ 0x07FE (0x0000‑07FD) <br> CS2 @ 0x0FFE (0x0800‑0FFD) | **Same CS1/CS2** *plus* a second mirror block later in flash on some sub‑versions (safe to ignore if you virginise every 0x8000 bank). |
| Read/write options   | Boot‑mode GPT or J‑TAG                                     | Boot‑mode GPT **or** new OBD‑only protocol 724 in K‑Suite / KESSv2([ECU Edit Tuning forum][1])                                         |

[1]: https://www.ecuedit.com/chip-tuning-tools-t19005/page100 "Chip Tuning Tools : ECU Tuning Hardware - Page 11 | ecuedit.com"


The checksum routine (CRC‑16/CCITT‑F0) is **identical**—the proof is that forum threads ask for a _single_ CS1/CS2 calculator that works for **both** BMS‑MP and BMS‑O units[nefariousmotorsports.com](https://nefariousmotorsports.com/forum/index.php?topic=19009.0&utm_source=chatgpt.com)—so you only need to point the tool at the right bytes.



## Bench wiring for the ECU BMS-MP

1.  **Remove** the ECU from the donor bike and place on an ESD mat.

2.  **Wire the breakout**:

    -   BATT = pin 1, IGN = pin 28, GND = 8/9, CAN L/H = 6/7 (check your loom).

    -   Connect GPT0/GPT1 to programmer if your tool needs them.

3.  **Power up** the bench supply **before** you connect the USB cable to avoid brown-outs.

4.  In your programmer, choose _BMS-MP → Boot/Service mode → Read P-FLASH & TRICORE EEPROM_.

5.  Save **two backups**: full flash (4 MB) and EEPROM (64–256 kB). Label them clearly—those dumps are your only way back if something goes wrong.


### Bench wiring for the ECU BMS-O

| ECU pin    | Signal                  | Bench note                              |
| ---------- | ----------------------- | --------------------------------------- |
|  **1, 15** |  +BATT (12 V)           | ≥13.0 V / 5 A PSU                       |
|  **2, 4**  |  GND                    | Battery‑negative and tool‑ground        |
|  **18**    |  IGN (KL15)             | Switched 12 V (use tool “IGN” line)     |
|  **16**    |  CAN‑L                  | For ID handshake                        |
|  **17**    |  CAN‑H                  | —                                       |
|  **54**    |  GPT‑1                  | TriCore handshake                       |
|  **55**    |  GPT‑0                  | TriCore handshake                       |
| (optional) |  BOOT / CNF1 pad on PCB | Only if you fall back to full boot mode |

A ready‑made “F32GN037” (or F32GN071 for Dimsport) harness puts the above on a keyed mini‑AMP plug, so no solder is required. [ecu.design](https://ecu.design/ecu-pinout/pinout-bosch-me17-2-4-2-bmso-irom-tc1793-egpt-bmw-motorrad/?utm_source=chatgpt.com)

_Use Bench/Factory mode unless the ECU is totally dead._  
Pin 1/15 = +12 V, 2/4 = GND, 18 = IGN, 16/17 = CAN, 54/55 = GPT.


## Reading the EEPROM – step by step (Flex example) for BMS-O

> The click‑labels are the same for HexProg, KESS3 Bench, Autotuner … only the UI names differ.

1.  **Power & loom**

    -   13.8 V bench supply on pins 1/15, IGN wire on pin 18, grounds on 2/4.

    -   Clip CAN & GPT wires as in the table above.

2.  **Select protocol** → _Bike → BMW → BMS‑O ME17.2.42 → Bench/Factory_

3.  **Identify ECU** – the tool checks micro (TC1793) and software ID.

4.  **Read → Internal EEPROM** (≈ 4 s) and **Read → Internal Flash** (≈ 80 s).  
    _Save both to disk!_  
    Flex calls this a “**Complete Read**”; HexProg’s button is “**Backup**”. [magicmotorsport.us](https://magicmotorsport.us/new-solutions-bosch-medc17-2021/)

5.  **Unplug power before disconnecting** to avoid brown‑out corruption.


You now have a 0x40000‑byte EEPROM dump (8 × 0x8000 banks). Run Go virginizer to wipe the VIN/ISN/key slots and fix the eight CS1/CS2 checksums.

## If Bench mode fails for BMS-O → Boot / BSL fallback

1.  Remove the aluminium lid (Torx T20) and flip the board.

2.  Solder or pogo‑pin CNF1 to ground (forces BSL).

3.  Apply 5 V to GPT0/GPT1 test pads plus the usual 12 V / GND.

4.  Choose _Infineon TC17xx → Boot mode_ in your tool and repeat Read / Write.  
    This bypasses a corrupted boot sector or a badly tuned CAN watchdog.


## Virginize EEPROM

Run this utility to create a new dump.
```
go build -o virginizer virginizer.go
./virginizer -in 1250gs_eeprom_192k.bin -out virgin_eeprom.bin
```

## Write the virgin image back for BMS-MP

1.  In the programmer choose _Write EEPROM only_ (P-FLASH stays untouched).

2.  Verify after write; voltage must stay > 13 V all the way.

3.  Disconnect power, wait 10 s, reconnect and do a _Quick ID_—the ECU should now report _**no VIN**_ and zero keys.


##  Writing it back for BMS-O on the bench

1.  Re‑connect exactly as for the read.

2.  **Write → EEPROM** (Flex) or **Restore → EEPROM** (HexProg).

    -   Write takes ≈ 5 s; the tool auto‑verifies CRC‑16 on each 0x8000 bank.

3.  Cycle IGN; perform a fresh **ECU ID**—it should now report _no VIN_.


## Flashing & pairing afterwards

1.  **Boot-mode flash** the `virgin_eeprom.bin` back into the your donor BMS-MP.

2.  In ISTA: **Control‑unit replacement → DME** (writes your VIN).
    * install ECU on the bike
    * connect **ISTA**
    * `Vehicle management → Control-unit replacement → DME`
    * ISTA writes your *real* VIN/ISN.

3.  Run _Keyless Ride → Teach-in key_ to pair each fob (hold it to the rear mud-guard antenna for ~60 s).

4.  Teach in keys → throttle adaptation ride.

    * Reset throttle adaptations and do the 3 000 rpm / idle routine.


### Why mirrors must stay valid

The TriCore bootloader can jump to _any_ 0x8000‑aligned sector if the previous one fails CRC while powering‑up, so every mirror needs its own good CS1/CS2. Leaving banks 1‑n with stale CRCs will brick the ECU the moment it “fails over” during start‑up.


### Quick sanity check for BMS-MP
After running the tool, open the output in a hex viewer:

```
00000020  FF FF FF ... (17 × FF)          // VIN wiped
00000100  FF … FF (16 bytes)              // ISN wiped
00000140  FF … FF (40 bytes)              // 5 key slots wiped
000007FE  xx xx   (CS1)   │ 0x0000‑07FD
00000FFE  yy yy   (CS2)   │ 0x0800‑0FFD
...
000087FE  xx'xx'  (bank 1 CS1)            │ 0x8000‑87FD
```

Where xx xx, yy yy, xx'xx'… are freshly‑calculated CRC‑16 values, different from the donor file.

### Sanity check for BMS-O before you flash

-   **VIN/ISN blanked?** Open the output in a hex viewer—bytes at 0x40‑0x50 and 0x120‑0x12F should be `FF`.

-   **Per‑bank CRC good?** Spot‑check a couple of banks:  
    `crc16( bank[0x0000:0x07FD] ) == uint16(bank[0x07FE:0x07FF])` and same for CS2.  
    (The unit tests you added earlier will already fail if that isn’t the case.)

-   **File size multiple of 0x8000?** Should be 0x40000 (256 k) for an 8‑bank BMS‑O.


## Re-install and code on the motorcycle

1.  Bolt the ECU back, reconnect battery.

2.  Hook up ISTA: **Vehicle management → Control-unit replacement → DME (BMS-MP)**.

    -   ISTA writes the bike’s VIN into the now-empty EEPROM and synchronises ISN with the X \_SLZ.

3.  **Teach in keys**: _Service functions → Keyless Ride → Teach-in key_. Hold each fob against the rear-mudguard antenna until the dash acknowledges it (about 60 s).

4.  Run _Drive → Adaptations → Reset throttle values_ and then perform the 3 000 rpm / idle routine.


At this point the bike should crank and start with your own fob.


## After-care & tips

-   **Throttle adaptation ride**: 3 rd gear, steady 3 000 rpm for 10 s, then idle for 1 min—this smooths cold starts.

-   Keep your _original backups_ and the virgin file; future updates with ISTA or dealer tools will proceed normally because the ECU now shows your VIN.

-   If ISTA complains that the DME is “from another vehicle”, the VIN write step failed—repeat _Control-unit replacement_.

-   A virgin ECU still counts its internal runtime; if you want odometer parity you’ll need an odometer correction session on the cluster.


#### Legal & security note

Virginising or immobiliser-off work must be backed by **proof of ownership** (title/registration). Many jurisdictions treat unmatched VINs or erased IMMO data as evidence of tampering. Always have documentation ready if a dealer or inspection authority asks.

Author of this program did not use any reverse-engineering tools, all those offsets were guessed.
If you have the correct numbers please submit the pull request.

Usage AS-IS, program is not tested on any kind of hardware or dumps. All this open source software repository is only for educational purposes for technical professionals.
