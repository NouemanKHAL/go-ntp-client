package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"time"
)

const (
	defaultNTPServer = "pool.ntp.org:123"

	// The offset between NTP time epoch and UNIX time epoch is 70 years in seconds
	ntpTimeOffset = 2208988800
)

func parseTimestamp(t uint64) (uint32, uint32) {
	seconds := uint32(t - ntpTimeOffset)
	fract := uint32((t * 1e9) >> 32)
	return seconds, fract
}

type NTPPayload struct {
	LiVnMode       uint8
	Stratum        uint8
	Poll           int8
	Precision      int8
	RootDelay      uint32
	RootDispersion uint32
	ReferenceID    uint32
	ReferenceSec   uint32
	ReferenceFrac  uint32
	OriginSec      uint32
	OriginFrac     uint32
	ReceiveSec     uint32
	ReceiveFrac    uint32
	TransmitSec    uint32
	TransmitFrac   uint32
}

func getNTPTime(t time.Time) (uint32, uint32) {
	sec := t.Unix() + ntpTimeOffset
	frac := (t.Add(time.Second * ntpTimeOffset).UnixNano()) % (sec * 1e9)
	return uint32(sec), uint32(frac)
}

func queryNTPServer(host string) (NTPPayload, error) {
	t1Sec, t1Frac := getNTPTime(time.Now())
	p := NTPPayload{
		LiVnMode:     0x1b,
		TransmitSec:  t1Sec,
		TransmitFrac: t1Frac,
	}
	conn, err := net.Dial("udp", host)
	if err != nil {
		return p, err
	}
	defer conn.Close()

	err = binary.Write(conn, binary.BigEndian, p)
	if err != nil {
		return p, err
	}

	var res NTPPayload
	err = binary.Read(conn, binary.BigEndian, &res)
	if err != nil {
		return res, err
	}

	t2Sec := res.ReceiveSec
	t2Frac := res.ReceiveFrac
	t3Sec := res.TransmitSec
	t3Frac := res.TransmitFrac
	t4Sec, t4Frac := getNTPTime(time.Now())

	fmt.Printf("t1 = %d.%d\n", t1Sec, t1Frac)
	fmt.Printf("t2 = %d.%d\n", t2Sec, t2Frac)
	fmt.Printf("t3 = %d.%d\n", t3Sec, t3Frac)
	fmt.Printf("t4 = %d.%d\n", t4Sec, t4Frac)

	return res, err
}

func main() {
	host := defaultNTPServer
	if len(os.Args) > 1 {
		host = os.Args[1]
	}
	_, err := queryNTPServer(host)
	if err != nil {
		os.Stderr.WriteString(fmt.Sprintf("error quering the NTP host %s: %s", host, err))
		os.Exit(1)
	}
}
