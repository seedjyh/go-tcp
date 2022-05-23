// Package uuid 用于方便地生成UUID。长度32，小写16进制。
package uuid

import (
	"encoding/binary"
	"encoding/hex"
	"hash/crc32"
	"math/rand"
	"os"
	"time"
)

// uuid32Generator 生成长度32的16进制字符串。
// 生成规则：
// 1) 先按照下面方式生成16字节（128bit）的整数。
// 		[0]~[3]   4 bytes: 最近一次更新的unix秒（从1970-01-01开始）
// 		[4]~[6]   3 bytes: 机器码的低3字节。机器码是hostname的CRC32校验和。
// 		[7]~[8]   2 bytes: 进程号的低2字节。
// 		[9]~[11]  3 bytes: 秒内递增数的低3字节。
// 		[12]~[15] 4 bytes: 随机串，防猜测。
// 2) 然后每4个bit转化成一个16进制字符。
type uuid32Generator struct {
	latestSecond     int64
	increaseInSecond uint32
	machineCode      [3]byte
	pidCode          [2]byte
	idChan           chan string
}

func NewUUID32Generator() *uuid32Generator {
	gen := &uuid32Generator{
		latestSecond:     0,
		increaseInSecond: 0,
		machineCode:      getMachineCode(),
		pidCode:          getPidCode(),
		idChan:           make(chan string),
	}
	go gen.keepGenerating()
	return gen
}

func (g *uuid32Generator) Next() string {
	return <-g.idChan
}

func (g *uuid32Generator) keepGenerating() {
	for {
		g.updateTimePart()
		var buf [16]byte
		offset := 0
		// 4 bytes: latestSecond的低4字节
		binary.BigEndian.PutUint32(buf[offset:], uint32(g.latestSecond))
		offset += 4
		// 3 bytes: machineCode
		for _, b := range g.machineCode {
			buf[offset] = b
			offset++
		}
		// 2 bytes: pidCode
		for _, b := range g.pidCode {
			buf[offset] = b
			offset++
		}
		// 3 bytes: increaseInSecond的低3字节
		binary.BigEndian.PutUint32(buf[offset:], g.increaseInSecond) // 越界了1字节，不过后续会覆盖
		offset += 3
		// 4 bytes: 随机串，防猜测
		binary.BigEndian.PutUint32(buf[offset:], rand.Uint32())
		// 转16进制字符串
		g.idChan <- hex.EncodeToString(buf[:])
	}
}

func (g *uuid32Generator) updateTimePart() {
	now := time.Now().Unix()
	if g.latestSecond != now {
		g.latestSecond = now
		g.increaseInSecond = getIncreaseInSecond()
	} else {
		g.increaseInSecond++
	}
}

// getMachineCode 返回长度3的byte数组。内容是hostname的CRC32校验和的低3字节。
func getMachineCode() [3]byte {
	hostname, _ := os.Hostname()
	buf := [4]byte{}
	binary.BigEndian.PutUint32(buf[:], crc32.ChecksumIEEE([]byte(hostname)))
	return [3]byte{buf[1], buf[2], buf[3]}
}

// getPidCode 返回长度2的byte数组。内容是进程号的低2字节。
func getPidCode() [2]byte {
	pid := os.Getpid()
	pidCode := [2]byte{}
	binary.BigEndian.PutUint16(pidCode[:], uint16(pid))
	return pidCode
}

// getIncreaseInSecond 生成一秒内的随机递增起点
func getIncreaseInSecond() uint32 {
	return uint32(rand.Intn(100000000))
}
