# twBlueScan

TWSNMP FC 用の Bluetooth デバイスセンサー

[English](./README.md)

[![Godoc Reference](https://godoc.org/github.com/twsnmp/twBlueScan?status.svg)](http://godoc.org/github.com/twsnmp/twBlueScan)
[![Go Report Card](https://goreportcard.com/badge/twsnmp/twBlueScan)](https://goreportcard.com/report/twsnmp/twBlueScan)

## 概要

Linux マシンで周辺にある Bluetooth デバイスの情報を収集し、syslog または MQTT で TWSNMP FC などに送信するためのセンサープログラムです。

収集する情報は以下の通りです：

- デバイスのアドレス
- アドレスの種類 (random/public)
- 名前
- RSSI (信号強度)
- 製造元メーカー
- オムロンの環境センサーの情報
- SwitchBot のセンサー情報（温度・湿度など）

## 状態

- 2021/08/29: 開発開始
- 2021/09/02: v1.0.0 リリース
- 2021/09/12: v2.0.0 低レベルパッケージを bluewalker に変更
- 2021/09/27: v2.1.0 SwitchBot センサーとアクティブモードスキャンに対応
- 2026/03/20: v3.0.0 MQTT 送信対応、GoReleaser による自動リリース対応

## ビルド

ビルドは GoReleaser または Make で行います。

### GoReleaser によるビルド
```bash
goreleaser release --snapshot --clean
```

### Make によるビルド
```bash
$ make
```
以下のターゲットが指定できます：
- `all`: 全実行ファイルのビルド（amd64, arm, arm64）
- `clean`: ビルドした実行ファイルの削除
- `zip`: リリース用の ZIP ファイルを作成

実行ファイルは `dist` ディレクトリに作成されます。

## 実行方法

### 使用方法

```text
Usage of ./twBlueScan:
  -active
        アクティブスキャンモード
  -adapter string
        モニタリングする Bluetooth アダプター (デフォルト "hci0")
  -addr string
        アドレスからベンダーへのマップを作成
  -all
        すべての詳細を報告（プライベートアドレスを含む）
  -code string
        会社コードからベンダーへのマップを作成
  -debug
        デバッグモード
  -interval int
        syslog 送信間隔（秒） (デフォルト 600)
  -mqtt string
        MQTT ブローカーの宛先 (例: tcp://192.168.1.1:1883)
  -mqttClientID string
        MQTT クライアント ID (デフォルト "twBlueScan")
  -mqttPassword string
        MQTT パスワード
  -mqttTopic string
        MQTT トピック (デフォルト "twBlueScan")
  -mqttUser string
        MQTT ユーザー名
  -syslog string
        syslog 送信先リスト（カンマ区切り、例: 192.168.1.1:514）
```

### 環境変数による設定
各フラグは、`TWBLUESCAN_` をプレフィックスとした環境変数でも設定可能です（例: `TWBLUESCAN_SYSLOG`）。

### 動作環境
Linux 環境で `bluez` パッケージが必要です。
```bash
$ sudo apt update
$ sudo apt install bluez
```

Bluetooth デバイスが利用可能であることを確認してください：
```bash
# hcitool dev
Devices:
	hci0	00:E9:3A:89:8D:FE
```

### 起動例

```bash
# syslog 送信
./twBlueScan -adapter hci0 -syslog 192.168.1.1

# MQTT 送信 (アクティブスキャン有効)
./twBlueScan -active -mqtt tcp://192.168.1.1:1883 -mqttTopic myhome/ble
```

## 著作権
詳細は [./LICENSE](./LICENSE) を参照してください。

```text
Copyright 2021-2026 Masayuki Yamai
```
