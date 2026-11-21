VERSION := 1.0.0
APP     := SATurn
BUNDLE  := bin/$(APP).app
DMG     := bin/$(APP)-$(VERSION)-darwin-arm64.dmg
REL     := bin/release-$(VERSION)
MACENV  := CGO_ENABLED=1 MACOSX_DEPLOYMENT_TARGET=14.0 CGO_CFLAGS=-mmacosx-version-min=14.0 CGO_LDFLAGS=-mmacosx-version-min=14.0
GOFILES := $(shell find . -name '*.go' -not -path './frontend/*' -not -path './bin/*')

.PHONY: help hooks frontend dist-stub bindings assets build app dmg install simulate test lint check check-attribution check-version release clean

help: ## list the commands
	@grep -E '^[a-z-]+:.*## ' $(MAKEFILE_LIST) | awk -F ':.*## ' '{printf "  make %-18s %s\n", $$1, $$2}'

hooks: ## install the commit-msg hook that rejects attribution lines
	install -m 0755 scripts/commit-msg .git/hooks/commit-msg

frontend: ## build the web UI into frontend/dist
	cd frontend && (test -d node_modules || npm ci) && npm run build

dist-stub:
	@mkdir -p frontend/dist && test -f frontend/dist/index.html || echo '<!doctype html>' > frontend/dist/index.html

bindings: dist-stub ## regenerate the TypeScript bindings after changing a Go service
	$(MACENV) wails3 generate bindings -ts -i -d frontend/bindings

assets: ## regenerate the sounds and icons (the results are committed)
	go run ./tools/gensounds internal/sound/wav
	go run ./tools/genicon build/appicon.png
	rm -rf bin/icons.iconset && mkdir -p bin/icons.iconset
	for s in 16 32 128 256 512; do \
		sips -z $$s $$s build/appicon.png --out bin/icons.iconset/icon_$${s}x$${s}.png >/dev/null; \
		d=$$((s*2)); sips -z $$d $$d build/appicon.png --out bin/icons.iconset/icon_$${s}x$${s}@2x.png >/dev/null; \
	done
	iconutil -c icns bin/icons.iconset -o build/icons.icns

build: frontend
	$(MACENV) go build -tags production -trimpath -buildvcs=false -ldflags="-w -s -X main.version=$(VERSION)" -o bin/$(APP)-gui .

app: build ## build bin/SATurn.app
	VERSION=$(VERSION) scripts/bundle.sh

dmg: app ## build the .dmg
	rm -rf bin/dmg && mkdir -p bin/dmg && cp -R $(BUNDLE) bin/dmg/ && ln -s /Applications bin/dmg/Applications
	hdiutil create -volname $(APP) -srcfolder bin/dmg -ov -format UDZO $(DMG) >/dev/null
	rm -rf bin/dmg && ls -lh $(DMG)

install: app ## replace /Applications/SATurn.app with this build and start it
	-pkill -x $(APP); sleep 1; pkill -9 -x $(APP) 2>/dev/null; true
	rm -rf /Applications/$(APP).app && cp -R $(BUNDLE) /Applications/
	open /Applications/$(APP).app

simulate: ## print a simulated day of reminders
	go run ./cmd/simulate

test: dist-stub ## Go and frontend tests
	$(MACENV) go vet ./...
	$(MACENV) go test ./...
	cd frontend && (test -d node_modules || npm ci) && npm run typecheck && npm test

lint:
	@test -z "$$(gofmt -l $(GOFILES))" || (echo "gofmt needed:"; gofmt -l $(GOFILES); exit 1)

check-attribution: ## no commit names an AI assistant as author or co-author
	@sh scripts/check-attribution.sh

check-version: ## VERSION matches the changelog and has release notes
	@v=$$(sed -n 's/^## \([0-9][0-9]*\.[0-9][0-9]*\.[0-9][0-9]*\) .*/\1/p' CHANGELOG.md | head -1); \
	[ "$$v" = "$(VERSION)" ] || { echo "CHANGELOG.md starts with $$v, the Makefile says $(VERSION)"; exit 1; }; \
	[ -f build/release-notes/$(VERSION).md ] || { echo "missing build/release-notes/$(VERSION).md"; exit 1; }; \
	echo "version $(VERSION) ok"

check: lint test check-attribution ## everything that must pass before a release

release: check check-version dmg ## build the release files into bin/release-$(VERSION)/
	rm -rf $(REL) && mkdir -p $(REL) && cp $(DMG) $(REL)/
	cd $(REL) && shasum -a 256 * > SHA256SUMS && cat SHA256SUMS

clean:
	rm -rf bin frontend/dist coverage.out
