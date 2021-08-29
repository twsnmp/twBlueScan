# twBlueScan
BlueTooth Device Sensor for TWSNMP

TWSNMPのためのBlueToothデバイスセンサー

[![Godoc Reference](https://godoc.org/github.com/twsnmp/twBlueScan?status.svg)](http://godoc.org/github.com/twsnmp/twBlueScan)
[![Go Report Card](https://goreportcard.com/badge/twsnmp/twBlueScan)](https://goreportcard.com/report/twsnmp/twBlueScan)

## Overview

Linuxマシンで周辺にあるBlueToothデバイスの情報を収集してsyslogで送信するためのセンサー
プログラムです。
収集する情報は

- デバイスのアドレス
- アドレスの種類(randam/public)
- 名前
- RSSI(信号の強度)
- 製造元メーカー
- ビーコン情報(iBeacon/Eddystone)

## Status

2021/8/29 開発を開始したばかりの最初のバージョンがとりあえず動作した。

## Build

ビルドはmakeで行います。
```
$make
```
以下のターゲットが指定できます。
```
  all        全実行ファイルのビルド（省略可能）
  docker     Docker Imageのビルド
  clean      ビルドした実行ファイルの削除
  zip        リリース用のZIPファイルを作成
```

```
$make
```
を実行すれば、Linux(amd64),Linux(arm)用の実行ファイルが、`dist`のディレクトリに作成されます。

Dockerイメージを作成するためには、
```
$make docker
```
を実行します。twssnmp/twpcaというDockerイメージが作成されます。

配布用のZIPファイルを作成するためには、
```
$make zip
```
を実行します。ZIPファイルが`dist/`ディレクトリに作成されます。

## Run

### 使用方法

```
Usage of /tmp/twBlueScan:
  -adapter string
    	monitor bluetooth adapter (default "hci0")
  -addr string
    	make address to vendor map
  -code string
    	make comapny code to vendor map
  -cpuprofile file
    	write cpu profile to file
  -interval int
    	syslog send interval(sec) (default 600)
  -memprofile file
    	write memory profile to file
  -syslog string
    	syslog destnation list
```

syslogの送信先はカンマ区切りで複数指定できます。:の続けてポート番号を
指定することもできます。

```
-syslog 192.168.1.1,192.168.1.2:5514
```


### 起動方法

起動するためには、モニタするアダプター(-adapter)とsyslogの送信先(-syslog)が必要です。

Linuxの環境では以下のコマンドで起動できます。（例はLinux場合）

```
#./twBlueScan -adapter hci0 -syslog 192.168.1.1
```

Docker環境では以下のコマンドを実行すれば起動できます。

```
#docker run --rm -d  --name twpcap  --net host twsnmp/twBlueScan  -adapter hci0 -syslog 192.168.1.1
```

## Copyright

see ./LICENSE

```
Copyright 2021 Masayuki Yamai
```
