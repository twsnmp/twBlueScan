.PHONY: all test clean zip

### バージョンの定義
VERSION     := "v2.2.0"
COMMIT      := $(shell git rev-parse --short HEAD)
WD          := $(shell pwd)
### コマンドの定義
GO          = go
GO_BUILD    = $(GO) build
GO_TEST     = $(GO) test -v
GO_LDFLAGS  = -ldflags="-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT)"
ZIP          = zip

### ターゲットパラメータ
DIST = dist
SRC = ./main.go ./blueScan.go ./syslog.go ./vendor.go
TARGETS     = $(DIST)/twBlueScan $(DIST)/twBlueScan.arm
GO_PKGROOT  = ./...

### PHONY ターゲットのビルドルール
all: $(TARGETS)
test:
	env GOOS=$(GOOS) $(GO_TEST) $(GO_PKGROOT)
clean:
	rm -rf $(TARGETS) $(DIST)/*.zip
zip: $(TARGETS)
	cd dist && $(ZIP) twBlueScan_linux_amd64.zip twBlueScan
	cd dist && $(ZIP) twBlueScan_linux_arm.zip twBlueScan.arm

### 実行ファイルのビルドルール
$(DIST)/twBlueScan: $(SRC)
	env GO111MODULE=on GOOS=linux GOARCH=amd64 $(GO_BUILD) $(GO_LDFLAGS) -o $@
$(DIST)/twBlueScan.arm: $(SRC)
	env GO111MODULE=on GOOS=linux GOARCH=arm GOARM=7 $(GO_BUILD) $(GO_LDFLAGS) -o $@
