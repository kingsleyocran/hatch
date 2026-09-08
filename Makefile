.PHONY: build clean install setup dev

build:
	go build -o /tmp/hatch-build .
	cd vscode-extension && npm run compile

install: build
	mkdir -p ~/.hatch/bin
	cp /tmp/hatch-build ~/.hatch/bin/hatch 2>/dev/null || sudo cp /tmp/hatch-build ~/.hatch/bin/hatch
	chmod +x ~/.hatch/bin/hatch 2>/dev/null || sudo chmod +x ~/.hatch/bin/hatch

clean: build
	sudo pkill -9 -f "hatch start" 2>/dev/null || true
	sudo launchctl bootout system/com.hatch.daemon 2>/dev/null || true
	sudo rm -f /Library/LaunchDaemons/com.hatch.daemon.plist
	sudo rm -f /etc/resolver/test
	sudo security delete-certificate -c "Hatch Local Development CA" /Library/Keychains/System.keychain 2>/dev/null || true
	sudo rm -rf ~/.hatch /usr/local/var/hatch
	mkdir -p ~/.hatch/bin
	sudo cp /tmp/hatch-build ~/.hatch/bin/hatch
	sudo chmod +x ~/.hatch/bin/hatch
	@echo "CLEAN — binary rebuilt at ~/.hatch/bin/hatch"

setup: install
	sudo ~/.hatch/bin/hatch setup

dev: install
	@echo "Ready. Press F5 in VSCode."
