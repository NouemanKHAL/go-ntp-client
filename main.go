package main

import (
	"encoding/binary"
	"fmt"
	"math"
	"net"
	"os"
	"strings"
	"time"
)

const (
	defaultNTPServer = "pool.ntp.org:123"
	// The offset between NTP time epoch and UNIX time epoch is 70 years in seconds, this is need to convert from NTP to Unix tiestamp
	ntpTimeOffset = 2208988800
)

// NTP time epoch starts at 1 Jan 1900 00h00, this is needed to convert between NTP and Unix timestamps
var ntpEpoch = time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)

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

func UnixToNTP(t time.Time) (int64, int64) {
	sec := t.Sub(ntpEpoch)
	nanoPerSec := int64(1000000000)
	nsec := sec.Nanoseconds() / nanoPerSec
	frac := sec.Nanoseconds() % nanoPerSec
	return int64(nsec), int64(float64(frac) * 5.)
}

func NTPtoUnix(sec, frac uint32) (uint32, uint32) {
	return uint32(float64(sec) - ntpTimeOffset), uint32((float64(frac) / math.Pow(2, 32)))
}

func formatTimestamp[T ~float64 | ~int64 | ~uint32 | ~uint64](sec, frac T) string {
	sign := "+"
	if sec < 0 {
		sign = "-"
		sec = -sec
	}
	if frac < 0 {
		sign = "-"
		frac = -frac
	}
	return fmt.Sprintf("%s %v", sign, float64(sec)+(float64(frac)/math.Pow(2, 32)))
}

func queryNTPServer(host string) (NTPPayload, error) {
	p := NTPPayload{
		LiVnMode: 0x23, // Li: 0, Vn: 4, Mode: 3 (client)
	}
	conn, err := net.Dial("udp", host)
	if err != nil {
		return p, err
	}
	defer conn.Close()

	t1TmpS, t1TmpF := UnixToNTP(time.Now())
	t1Sec, t1Frac := int64(t1TmpS), int64(t1TmpF)

	p.TransmitSec = uint32(t1TmpS)
	p.TransmitFrac = uint32(t1TmpF)

	err = binary.Write(conn, binary.BigEndian, p)
	if err != nil {
		return p, err
	}

	var res NTPPayload
	err = binary.Read(conn, binary.BigEndian, &res)
	if err != nil {
		return res, err
	}

	t2Sec := int64(res.ReceiveSec)
	t2Frac := int64(res.ReceiveFrac)
	t3Sec := int64(res.TransmitSec)
	t3Frac := int64(res.TransmitFrac)

	t4TmpS, t4TmpF := UnixToNTP(time.Now())
	t4Sec, t4Frac := int64(t4TmpS), int64(t4TmpF)

	fmt.Printf("t1\t= %d.%d\n", t1Sec, t1Frac)
	fmt.Printf("t2\t= %d.%d\n", t2Sec, t2Frac)
	fmt.Printf("t3\t= %d.%d\n", t3Sec, t3Frac)
	fmt.Printf("t4\t= %d.%d\n", t4Sec, t4Frac)

	dSec := (t4Sec - t1Sec) - (t3Sec - t2Sec)
	dFrac := (t4Frac - t1Frac) - (t3Frac - t2Frac)

	offsetSec := float64((t2Sec-t1Sec)+(t3Sec-t4Sec)) / 2.
	offsetFrac := float64((t2Frac-t1Frac)+(t3Frac-t4Frac)) / 2.

	fmt.Printf("delay\t= %s\n", formatTimestamp(dSec, dFrac))
	fmt.Printf("offset\t= %s\n", formatTimestamp(offsetSec, offsetFrac))

	return res, err
}

func main() {
	host := defaultNTPServer
	if len(os.Args) > 1 {
		host = os.Args[1]
	}
	if !strings.HasSuffix(host, ":123") {
		host += ":123"
	}
	_, err := queryNTPServer(host)
	if err != nil {
		os.Stderr.WriteString(fmt.Sprintf("error querying the NTP host %s: %s", host, err))
		os.Exit(1)
	}
}
