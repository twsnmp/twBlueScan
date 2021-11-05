# twBlueScan

BlueTooth Device Sensor for TWSNMP FC  
TWSNMPのためのBlueToothデバイスセンサー  

[![Godoc Reference](https://godoc.org/github.com/twsnmp/twBlueScan?status.svg)](http://godoc.org/github.com/twsnmp/twBlueScan)
[![Go Report Card](https://goreportcard.com/badge/twsnmp/twBlueScan)](https://goreportcard.com/report/twsnmp/twBlueScan)

## Overview

Linuxマシンで周辺にあるBlueToothデバイスの情報を収集して  
syslogでTWSNMP FCなどに送信するためのセンサープログラムです。

収集する情報は

- デバイスのアドレス
- アドレスの種類(randam/public)
- 名前
- RSSI(信号の強度)
- 製造元メーカー
- オムロンの環境センサーの情報

## Status

2021/8/29  開発を開始したばかりの最初のバージョンがとりあえず動作した。  
2021/9/1   RC1テスト中（ビーコンの検知は断念）  
2021/9/2   v1.0.0 最初のリリース  
2021/9/11  v1.1.0 改善版（リリースしなかった）  
2021/9/12  v2.0.0 低レベルのパッケージbluewakerに変更  
2021/9/27  v2.1.0 SwitchBotのセンサーとアクティブモードスキャンに対応  

## Build

ビルドはmakeで行います。
```
$make
```
以下のターゲットが指定できます。
```
  all        全実行ファイルのビルド（省略可能）
  clean      ビルドした実行ファイルの削除
  zip        リリース用のZIPファイルを作成
```

```
$make
```
を実行すれば、Linux(amd64),Linux(arm)用の実行ファイルが、`dist`のディレクトリに作成されます。

配布用のZIPファイルを作成するためには、
```
$make zip
```
を実行します。ZIPファイルが`dist/`ディレクトリに作成されます。

## Run

### 使用方法

```
#twBlueScan -h
Usage of /tmp/twBlueScan:
  -active
    	active scan mode
  -adapter string
    	monitor bluetooth adapter (default "hci0")
  -addr string
    	make address to vendor map
  -code string
    	make company code to vendor map
  -cpuprofile file
    	write cpu profile to file
  -debug
    	debug mode
  -interval int
    	syslog send interval(sec) (default 600)
  -memprofile file
    	write memory profile to file
  -syslog string
    	syslog destnation list
```

syslogの送信先はカンマ区切りで複数指定できます。:に続けてポート番号を
指定することもできます。

```
-syslog 192.168.1.1,192.168.1.2:5514
```

### 動作環境

このプログラム起動するためには
Linux環境にbluezのインストールが必要です。
(v2.0.0以降ではhcitool devで確認できれば不要かもしれません。)

```
$sudo apt update
$sudo apt install bluez
```

BlueToothのデバイスが利用可能であることを以下のコマンドで確認してください。

```
# hcitool dev
Devices:
	hci0	00:E9:3A:89:8D:FE
```

### 起動方法

起動するためにはモニタするアダプター(-adapter)と  
syslogの送信先(-syslog)が必要です。
アダプターのデフォルトはhci0になっています。  
省略できるという意味です。  
SwitchBotの温度センサーを見つけるためには、  
-activeを指定してアクティブモードでスキャンしてください。  

以下のコマンドで起動できます。

```
#./twBlueScan -adapter hci0 -syslog 192.168.1.1
```

## TWSNMP FCのパッケージ

TWSNMP FCのパッケージにtwBlueScanが含まれています。  
Linux(amd64,arm)版があります。  
詳しくは、  
https://note.com/twsnmp/n/nc6e49c284afb  
を見てください。

## Copyright

see ./LICENSE

```
Copyright 2021 Masayuki Yamai
```
