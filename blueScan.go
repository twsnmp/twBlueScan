package main

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"gitlab.com/jtaimisto/bluewalker/hci"
	"gitlab.com/jtaimisto/bluewalker/host"
)

type BluetoothDeviceEnt struct {
	Address     string
	AddressType string
	Name        string
	MinRSSI     int
	MaxRSSI     int
	RSSI        int
	Info        string
	Count       int
	Code        uint16
	EnvData     []byte
	FirstTime   int64
	LastTime    int64
}

func (d *BluetoothDeviceEnt) String() string {
	return fmt.Sprintf("type=Device,address=%s,name=%s,rssi=%d,min=%d,max=%d,addrType=%s,vendor=%s,info=%s,ft=%s,lt=%s",
		d.Address, d.Name, d.RSSI, d.MinRSSI, d.MaxRSSI,
		d.AddressType, getVendor(d), d.Info,
		time.Unix(d.FirstTime, 0).Format(time.RFC3339),
		time.Unix(d.LastTime, 0).Format(time.RFC3339),
	)
}

var deviceMap sync.Map
var total = 0
var skip = 0

// startBlueScan : start scan
func startBlueScan(ctx context.Context) {
	if err := exec.Command("hciconfig", adapter, "down").Run(); err != nil {
		log.Panic(err)
	}
	raw, err := hci.Raw(adapter)
	if err != nil {
		log.Fatalln("Raw", err)
	}
	h := host.New(raw)
	if err = h.Init(); err != nil {
		log.Fatalln("Init", err)
	}
	reportCh, err := h.StartScanning(active, nil)
	if err != nil {
		log.Fatalln("startScan", err)
	}
	timer := time.NewTicker(time.Second * time.Duration(syslogInterval))
	defer timer.Stop()
	for {
		select {
		case report := <-reportCh:
			checkBlueDevice(report)
		case <-timer.C:
			sendMonitor()
			sendReport()
		case <-ctx.Done():
			h.StopScanning()
			h.Deinit()
			log.Println("stop bluetooth scan")
			return
		}
	}
}

func checkBlueDevice(r *host.ScanReport) {
	rssi := int(r.Rssi)
	if rssi == 0 {
		skip++
		return
	}
	total++
	now := time.Now().Unix()
	addr := r.Address.String()
	if v, ok := deviceMap.Load(addr); ok {
		if d, ok := v.(*BluetoothDeviceEnt); ok {
			d.RSSI = rssi
			if d.RSSI > d.MaxRSSI {
				d.MaxRSSI = d.RSSI
			}
			if d.RSSI < d.MinRSSI {
				d.MinRSSI = d.RSSI
			}
			checkDeviceInfo(d, r)
			d.Count++
			d.LastTime = now
			return
		} else {
			deviceMap.Delete(addr)
		}
	}
	d := &BluetoothDeviceEnt{
		Address:   addr,
		RSSI:      int(r.Rssi),
		MinRSSI:   int(r.Rssi),
		MaxRSSI:   int(r.Rssi),
		Count:     1,
		FirstTime: now,
		LastTime:  now,
	}
	checkDeviceInfo(d, r)
	deviceMap.Store(addr, d)
}

func getVendor(d *BluetoothDeviceEnt) string {
	if d.Code != 0x0000 {
		if v, ok := codeToVendorMap[d.Code]; ok {
			return fmt.Sprintf("%s(0x%04x)", v, d.Code)
		}
		return fmt.Sprintf("unknown(0x%04x)", d.Code)
	}
	return getVendorFromAddress(d.Address)
}

var uuidMap sync.Map

func checkDeviceInfo(d *BluetoothDeviceEnt, r *host.ScanReport) {
	if d.AddressType == "" {
		d.AddressType = getAddrType(r.Address)
	}
	name := ""
	info := ""
	code := uint16(0x0000)
	for _, a := range r.Data {
		switch a.Typ {
		case hci.AdFlags:
			if len(a.Data) == 1 {
				info += getInfoFromFlag(int(a.Data[0]))
			}
		case hci.AdCompleteLocalName, hci.AdShortenedLocalName:
			name = string(a.Data)
		case hci.AdManufacturerSpecific:
			if len(a.Data) < 2 {
				continue
			}
			code = uint16(a.Data[1])*256 + uint16(a.Data[0])
			if code == 0x02d5 && len(a.Data) >= 18 {
				d.EnvData = a.Data[2:]
				// } else if code == 0x004c {
				// 	if debug {
				// 		log.Println("apple md", d.Address, a.String())
				// 	}
				// } else if code == 0x012d {
				// 	if debug {
				// 		log.Println("sony md", d.Address, a.String())
				// 	}
				// } else {
				// 	if debug {
				// 		log.Println("other md", d.Address, a.String())
				// 	}
			}
		case hci.AdTxPower:
		case hci.AdComplete128BitService:
			if id, err := uuid.FromBytes(a.Data); err == nil {
				if _, ok := uuidMap.Load(d.Address + id.String()); !ok {
					uuidMap.Store(d.Address+id.String(), true)
					if debug {
						log.Println("uuid", d.Address, id.String())
					}
				}
			} else {
				log.Println(err)
			}
		case hci.AdServiceData:
			if len(a.Data) == 8 && a.Data[0] == 0 && a.Data[1] == 0x0d && a.Data[2] == 0x54 {
				d.EnvData = a.Data[:]
			}
		default:
			if debug {
				log.Println("unknown", d.Address, a.String())
			}
		}
	}
	if name != "" {
		d.Name = name
	}
	if info != "" {
		d.Info = info
	}
	if code != 0x0000 {
		d.Code = code
	}
}

var flagNames = []struct {
	flag int
	name string
}{
	{hci.AdFlagLimitedDisc, "LE Limited"},
	{hci.AdFlagGeneralDisc, "LE General"},
	{hci.AdFlagNoBrEdr, "No BR/EDR"},
	{hci.AdFlagLeBrEdrController, "LE & BR/EDR (controller)"},
	{hci.AdFlagLeBrEdrHost, "LE & BR/EDR (host)"},
}

func getInfoFromFlag(flag int) string {
	ret := ""
	for _, f := range flagNames {
		if (flag & f.flag) == f.flag {
			if ret != "" {
				ret += ";"
			}
			ret += f.name
		}
	}
	return ret
}

func getAddrType(addr hci.BtAddress) string {
	ret := addr.Atype.String()
	if addr.Atype == hci.LeRandomAddress {
		ret += "("
		if addr.IsNonResolvable() {
			ret += "non-resolvable"
		} else if addr.IsResolvable() {
			ret += "resolvable"
		} else if addr.IsStatic() {
			ret += "static"
		} else {
			ret += "??"
		}
		ret += ")"
	}
	return ret
}

// OMRONS????????????????????????
// https://omronfs.omron.com/ja_JP/ecb/products/pdf/CDSC-016A-web1.pdf
// P60
// https://armadillo.atmark-techno.com/howto/armadillo_2JCIE-BU01_GATT
// 01     Data Type
// c5     ??????
// a9 09  ?????? 0.01???
// cd 1a  ?????? 0.01%
// 0d 00  ?????? 1lx
// 26 6c 0f 00 ?????? 1hPa
// 3d 13  ?????? 0.01dB
// 07 00  eTVOC 1ppb
// c3 01  ??????????????? 1ppm
// ff

func sendOMRONEnv(d *BluetoothDeviceEnt) {
	seq := int(d.EnvData[1])
	temp := float64(int(d.EnvData[3])*256+int(d.EnvData[2])) * 0.01
	hum := float64(int(d.EnvData[5])*256+int(d.EnvData[4])) * 0.01
	lx := int(d.EnvData[7])*256 + int(d.EnvData[6])
	press := float64(int(d.EnvData[11])*(256*256*256)+int(d.EnvData[10])*(256*256)+int(d.EnvData[9])*256+int(d.EnvData[8])) * 0.001
	sound := float64(int(d.EnvData[13])*256+int(d.EnvData[12])) * 0.01
	v := int(d.EnvData[15])*256 + int(d.EnvData[14])
	co2 := int(d.EnvData[17])*256 + int(d.EnvData[16])
	if debug {
		log.Printf("omron seq=%d,temp=%.02f,hum=%.02f,lx=%d,press=%.02f,sound=%.02f,eTVOC=%d,eCO2=%d",
			seq, temp, hum, lx, press, sound, v, co2)
	}
	syslogCh <- fmt.Sprintf("type=OMRONEnv,address=%s,name=%s,rssi=%d,seq=%d,temp=%.02f,hum=%.02f,lx=%d,press=%.02f,sound=%.02f,eTVOC=%d,eCO2=%d",
		d.Address, d.Name, d.RSSI,
		seq, temp, hum, lx, press, sound, v, co2,
	)
}

//  0x00 0d 54 10 e4 07 9a 37
func sendSwitchBotEnv(d *BluetoothDeviceEnt) {
	bat := int(d.EnvData[4] & 0x7f)
	temp := float64(int(d.EnvData[5]&0x0f))/10.0 + float64(d.EnvData[6]&0x7f)
	if (d.EnvData[6] & 0x80) != 0x80 {
		temp *= -1.0
	}
	hum := float64(int(d.EnvData[7] & 0x7f))
	if debug {
		log.Printf("switchbot temp=%.02f,hum=%.02f,bat=%d", temp, hum, bat)
	}
	syslogCh <- fmt.Sprintf("type=SwitchBotEnv,address=%s,name=%s,rssi=%d,temp=%.02f,hum=%.02f,bat=%d",
		d.Address, d.Name, d.RSSI,
		temp, hum, bat,
	)
}

var lastSendTime int64

func sendReport() {
	count := 0
	new := 0
	remove := 0
	omron := 0
	swbot := 0
	now := time.Now().Unix()
	deviceMap.Range(func(k, v interface{}) bool {
		d, ok := v.(*BluetoothDeviceEnt)
		if !ok {
			return true
		}
		if d.LastTime < now-3600*1 {
			deviceMap.Delete(k)
			remove++
			return true
		}
		count++
		if d.LastTime < lastSendTime {
			return true
		}
		if d.FirstTime > lastSendTime {
			new++
		}
		if strings.HasPrefix(d.Name, "Rbt") && len(d.EnvData) >= 18 && d.EnvData[0] == 1 {
			sendOMRONEnv(d)
			omron++
		}
		if len(d.EnvData) == 8 && d.EnvData[0] == 0 && d.EnvData[1] == 0x0d && d.EnvData[2] == 0x54 {
			sendSwitchBotEnv(d)
			swbot++
		}
		syslogCh <- d.String()
		return true
	})
	syslogCh <- fmt.Sprintf("type=Stats,total=%d,count=%d,new=%d,remove=%d,send=%d,param=%s",
		total, count, new, remove, syslogCount, adapter)
	if debug {
		log.Printf("total=%d skip=%d count=%d new=%d remove=%d omron=%d swbot=%d send=%d",
			total, skip, count, new, remove, omron, swbot, syslogCount)
	}
	syslogCount = 0
	lastSendTime = now
}
