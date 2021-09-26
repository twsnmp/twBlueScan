package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"strings"
	"syscall"
	"time"
)

var version = "v1.0.0"
var commit = ""
var syslogDst = ""
var adapter = ""
var syslogInterval = 300
var cpuprofile string
var memprofile string
var codeToVendor string
var addrToVendor string
var debug bool
var active bool

func init() {
	flag.StringVar(&syslogDst, "syslog", "", "syslog destnation list")
	flag.StringVar(&adapter, "adapter", "hci0", "monitor bluetooth adapter")
	flag.IntVar(&syslogInterval, "interval", 600, "syslog send interval(sec)")
	flag.StringVar(&cpuprofile, "cpuprofile", "", "write cpu profile to `file`")
	flag.StringVar(&memprofile, "memprofile", "", "write memory profile to `file`")
	flag.StringVar(&codeToVendor, "code", "", "make company code to vendor map")
	flag.StringVar(&addrToVendor, "addr", "", "make address to vendor map")
	flag.BoolVar(&debug, "debug", false, "debug mode")
	flag.BoolVar(&active, "active", false, "active scan mode")
	flag.VisitAll(func(f *flag.Flag) {
		if s := os.Getenv("TWBLUESCAN_" + strings.ToUpper(f.Name)); s != "" {
			f.Value.Set(s)
		}
	})
	flag.Parse()
}

type logWriter struct {
}

func (writer logWriter) Write(bytes []byte) (int, error) {
	return fmt.Print(time.Now().Format("2006-01-02T15:04:05.999 ") + string(bytes))
}

func main() {
	log.SetFlags(0)
	log.SetOutput(new(logWriter))
	if cpuprofile != "" {
		f, err := os.Create(cpuprofile)
		if err != nil {
			log.Fatalf("could not create CPU profile: %v", err)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatalf("could not start CPU profile: %v", err)
		}
		defer pprof.StopCPUProfile()
	}
	if memprofile != "" {
		f, err := os.Create(memprofile)
		if err != nil {
			log.Fatalf("could not create memory profile: %v", err)
		}
		defer f.Close()
		runtime.GC() // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatalf("could not write memory profile:%v", err)
		}
	}
	if codeToVendor != "" {
		makeCodeToVendor()
		return
	}
	if addrToVendor != "" {
		makeAddressToVendor()
		return
	}
	log.Printf("version=%s", fmt.Sprintf("%s(%s)", version, commit))
	if adapter == "" {
		log.Fatalln("no monitor adapter")
	}
	if syslogDst == "" {
		log.Fatalln("no syslog distenation")
	}
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())
	go startSyslog(ctx)
	go startBlueScan(ctx)
	<-quit
	syslogCh <- "quit by signal"
	time.Sleep(time.Second * 1)
	log.Println("quit by signal")
	cancel()
	time.Sleep(time.Second * 2)
}
