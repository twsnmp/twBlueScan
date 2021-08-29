package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/muka/go-bluetooth/api"
	"github.com/muka/go-bluetooth/api/beacon"
	"github.com/muka/go-bluetooth/bluez/profile/adapter"
	"github.com/muka/go-bluetooth/bluez/profile/device"
	eddystone "github.com/suapapa/go_eddystone"
)

var beaconCount = 0

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
	total := int64(0)
	count := int64(0)
	for {
		select {
		case ev := <-discovery:
			checkBlueDevice(ev)
			count++
			total++
		case <-timer.C:
			syslogCh <- fmt.Sprintf("type=Stats,total=%d,count=%d,ps=%.2f,send=%d,beacon=%d,param=%s", total, count, float64(count)/float64(syslogInterval), syslogCount, beaconCount, adapterID)
			syslogCount = 0
			beaconCount = 0
			sendMonitor()
		case <-ctx.Done():
			log.Println("stop bluetooth scan")
			return
		}
	}
}

func checkBlueDevice(ev *adapter.DeviceDiscovered) {

	if ev.Type == adapter.DeviceRemoved {
		log.Printf("device removed: %s", ev.Path)
		return
	}

	dev, err := device.NewDevice1(ev.Path)
	if err != nil {
		log.Printf("%s: %s", ev.Path, err)
		return
	}
	if dev == nil {
		log.Printf("%s: not found", ev.Path)
		return
	}
	vendor := ""
	for k := range dev.Properties.ManufacturerData {
		if vendor != "" {
			vendor += ";"
		}
		if v, ok := codeToVendorMap[k]; ok {
			vendor += fmt.Sprintf("%s(0x%04x)", v, k)
		} else {
			vendor += fmt.Sprintf("unknown(0x%04x)", k)
		}
	}
	if vendor == "" && dev.Properties.AddressType == "public" {
		vendor = getVendorFromAddress(dev.Properties.Address)
	}
	log.Printf("find device addr=%s name=%s vendor=%s", dev.Properties.Address, dev.Properties.Name, vendor)
	syslogCh <- fmt.Sprintf("type=Device,address=%s,name=%s,rssi=%d,addrType=%s,vendor=%s",
		dev.Properties.Address, dev.Properties.Name, dev.Properties.RSSI,
		dev.Properties.AddressType, vendor)

	go func(ev *adapter.DeviceDiscovered) {
		err = handleBeacon(dev)
		if err != nil {
			log.Printf("%s: %s", ev.Path, err)
		}
	}(ev)
}

// handleBeacon : ビーコンの内容をチェックする
func handleBeacon(dev *device.Device1) error {
	b, err := beacon.NewBeacon(dev)
	if err != nil {
		return err
	}
	beaconUpdated, err := b.WatchDeviceChanges(context.Background())
	if err != nil {
		return err
	}
	isBeacon := <-beaconUpdated
	if !isBeacon {
		return nil
	}
	name := b.Device.Properties.Alias
	if name == "" {
		name = b.Device.Properties.Name
	}
	log.Printf("Found beacon %s %s", b.Type, name)
	beaconCount++
	if b.IsEddystone() {
		ed := b.GetEddystone()
		switch ed.Frame {
		case eddystone.UID:
			syslogCh <- fmt.Sprintf(
				"type=EddystoneUID,address=%s,name=%s,uid=%s,instance=%s,power=%d",
				b.Device.Properties.Address,
				name,
				ed.UID,
				ed.InstanceUID,
				ed.CalibratedTxPower,
			)
		case eddystone.TLM:
			syslogCh <- fmt.Sprintf(
				"type=EddystoneTLM,address=%s,name=%s,temp=%.0f,batt=%d,reboot=%d,advertising=%d,power=%d",
				b.Device.Properties.Address,
				name,
				ed.TLMTemperature,
				ed.TLMBatteryVoltage,
				ed.TLMLastRebootedTime,
				ed.TLMAdvertisingPDU,
				ed.CalibratedTxPower,
			)
		case eddystone.URL:
			syslogCh <- fmt.Sprintf(
				"type=EddystoneURL,address=%s,name=%s,url=%s,power=%d",
				b.Device.Properties.Address,
				name,
				ed.URL,
				ed.CalibratedTxPower,
			)
		}
	}
	if b.IsIBeacon() {
		ibeacon := b.GetIBeacon()
		syslogCh <- fmt.Sprintf(
			"type=IBeacon,address=%s,name=%s,uuid=%s,power=%d,major=%d,minor=%d",
			b.Device.Properties.Address,
			name,
			ibeacon.ProximityUUID,
			ibeacon.MeasuredPower,
			ibeacon.Major,
			ibeacon.Minor,
		)
	}
	return nil
}
