package main

import (
	"context"
	"fmt"
	"log"
	"runtime"
	"strings"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/muka/go-bluetooth/api"
	"github.com/muka/go-bluetooth/bluez/profile/adapter"
	"github.com/muka/go-bluetooth/bluez/profile/device"
)

// startBlueScan : start packet capchare
func startBlueScan(ctx context.Context) {
	defer api.Exit()
	a, err := adapter.GetAdapter(adapterID)
	if err != nil {
		log.Panic(err)
	}

	log.Println("Flush cached devices")
	err = a.FlushDevices()
	if err != nil {
		log.Panic(err)
	}

	log.Println("Start discovery")
	discovery, cancel, err := api.Discover(a, nil)
	if err != nil {
		log.Panic(err)
	}
	defer cancel()

	timer := time.NewTicker(time.Second * time.Duration(syslogInterval))
	defer timer.Stop()
	total := 0
	new := 0
	remove := 0
	for {
		select {
		case ev := <-discovery:
			if ev.Type == adapter.DeviceRemoved {
				remove++
			} else {
				checkBlueDevice(ev.Path)
				new++
			}
			total++
		case <-timer.C:
			sendMonitor()
			a.FlushDevices()
			syslogCh <- fmt.Sprintf("type=Stats,total=%d,count=%d,remove=%d,send=%d,param=%s",
				total, new, remove, syslogCount, adapterID)
			log.Printf("total=%d count=%d remove=%d NumGoroutine=%d", total, new, remove, runtime.NumGoroutine())
			syslogCount = 0
			new = 0
			remove = 0
		case <-ctx.Done():
			log.Println("stop bluetooth scan")
			return
		}
	}
}

func checkBlueDevice(p interface{}) {
	path, ok := p.(dbus.ObjectPath)
	if !ok {
		return
	}
	dev, err := device.NewDevice1(path)
	if err != nil {
		log.Printf("%s: %s", path, err)
		return
	}
	if dev == nil {
		log.Printf("%s: not found", path)
		return
	}
	if dev.Properties.RSSI >= 0 {
		log.Printf("%s:%s: RSSI = 0", path, dev.Properties.Address)
		return
	}
	vendor := ""
	md := ""
	for k := range dev.Properties.ManufacturerData {
		if vendor != "" {
			vendor += ";"
		}
		if v, ok := codeToVendorMap[k]; ok {
			vendor += fmt.Sprintf("%s(0x%04x)", v, k)
		} else {
			vendor += fmt.Sprintf("unknown(0x%04x)", k)
		}
		i, ok := dev.Properties.ManufacturerData[k]
		if !ok {
			continue
		}
		if v, ok := i.(dbus.Variant); ok {
			if ba, ok := v.Value().([]uint8); ok && len(ba) > 0 {
				if md != "" {
					md += ";"
				}
				md += fmt.Sprintf("%04x:%0x", k, ba)
			}
		}
	}
	if vendor == "" {
		vendor = getVendorFromAddress(dev.Properties.Address)
	}
	syslogCh <- fmt.Sprintf("type=Device,address=%s,name=%s,rssi=%d,addrType=%s,vendor=%s,md=%s",
		dev.Properties.Address, dev.Properties.Name, dev.Properties.RSSI,
		dev.Properties.AddressType, vendor, md)
	if strings.HasPrefix(dev.Properties.Name, "Rbt") {
		checkOMRONEnv(dev)
	}
	log.Printf("device addr=%s name=%s vendor=%s", dev.Properties.Address, dev.Properties.Name, vendor)
}

// OMRONSセンサーのデータ
// https://omronfs.omron.com/ja_JP/ecb/products/pdf/CDSC-016A-web1.pdf
// P60
// https://armadillo.atmark-techno.com/howto/armadillo_2JCIE-BU01_GATT
// 01     Data Type
// c5     連番
// a9 09  温度 0.01℃
// cd 1a  湿度 0.01%
// 0d 00  照度 1lx
// 26 6c 0f 00 気圧 1hPa
// 3d 13  騒音 0.01dB
// 07 00  eTVOC 1ppb
// c3 01  二酸化炭素 1ppm
// ff

func checkOMRONEnv(dev *device.Device1) {
	if dev == nil || dev.Properties == nil || dev.Properties.ManufacturerData == nil {
		log.Printf("checkOMRONEnv no data")
	}
	i, ok := dev.Properties.ManufacturerData[0x02d5]
	if !ok {
		log.Printf("checkOMRONEnv no ManufacturerData")
	}
	if v, ok := i.(dbus.Variant); ok {
		if ba, ok := v.Value().([]uint8); ok && len(ba) > 18 && ba[0] == 1 {
			seq := int(ba[1])
			temp := float64(int(ba[3])*256+int(ba[2])) * 0.01
			hum := float64(int(ba[5])*256+int(ba[4])) * 0.01
			lx := int(ba[7])*256 + int(ba[6])
			press := float64(int(ba[11])*(256*256*256)+int(ba[10])*(256*256)+int(ba[9])*256+int(ba[8])) * 0.001
			sound := float64(int(ba[13])*256+int(ba[12])) * 0.01
			v := int(ba[15])*256 + int(ba[14])
			co2 := int(ba[17])*256 + int(ba[16])
			log.Printf("omron seq=%d,temp=%.02f,hum=%.02f,lx=%d,press=%.02f,sound=%.02f,eTVOC=%d,eCO2=%d",
				seq, temp, hum, lx, press, sound, v, co2,
			)
			syslogCh <- fmt.Sprintf("type=OMRONEnv,address=%s,name=%s,rssi=%d,seq=%d,temp=%.02f,hum=%.02f,lx=%d,press=%.02f,sound=%.02f,eTVOC=%d,eCO2=%d",
				dev.Properties.Address, dev.Properties.Name, dev.Properties.RSSI,
				seq, temp, hum, lx, press, sound, v, co2,
			)
		}
	}
}
