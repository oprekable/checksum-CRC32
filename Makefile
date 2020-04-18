SHELL = bash
# Branch we are working on
BRANCH := $(or $(lastword $(subst /, ,$(GITHUB_REF))),$(shell git rev-parse --abbrev-ref HEAD))
# Tag of the current commit, if any.  If this is not "" then we are building a release
RELEASE_TAG := $(shell git tag -l --points-at HEAD)
# Version of last release (may not be on this branch)
VERSION := $(shell cat VERSION)
# Last tag on this branch
LAST_TAG := $(shell git describe --tags --abbrev=0)
# If we are working on a release, override branch to master
ifdef RELEASE_TAG
	BRANCH := master
endif
TAG_BRANCH := -$(BRANCH)
BRANCH_PATH := branch/
# If building HEAD or master then unset TAG_BRANCH and BRANCH_PATH
ifeq ($(subst HEAD,,$(subst master,,$(BRANCH))),)
	TAG_BRANCH :=
	BRANCH_PATH :=
endif
# Make version suffix -DDD-gCCCCCCCC (D=commits since last relase, C=Commit) or blank
VERSION_SUFFIX := $(shell git describe --abbrev=8 --tags | perl -lpe 's/^v\d+\.\d+\.\d+//; s/^-(\d+)/"-".sprintf("%03d",$$1)/e;')
# TAG is current version + number of commits since last release + branch
TAG := $(VERSION)$(VERSION_SUFFIX)$(TAG_BRANCH)
NEXT_VERSION := $(shell echo $(VERSION) | perl -lpe 's/v//; $$_ += 0.1; $$_ = sprintf("v%.2f.0", $$_)')
ifndef RELEASE_TAG
	TAG := $(TAG)-beta
endif
GO_VERSION := $(shell go version)
GO_FILES := $(shell go list ./...)

# Pass in GOTAGS=xyz on the make command line to set build tags
ifdef GOTAGS
BUILDTAGS=-tags "$(GOTAGS)"
LINTTAGS=--build-tags "$(GOTAGS)"
endif

.PHONY: checksum_crc32 vars generate version

checksum_crc32: update
	go build -v --ldflags "-s -X github.com/oprekable/checksum-CRC32/fs.Version=$(TAG)" $(BUILDTAGS)
	mkdir -p `go env GOPATH`/bin/
	cp -av checksum-CRC32`go env GOEXE` `go env GOPATH`/bin/checksum-CRC32`go env GOEXE`.new
	mv -v `go env GOPATH`/bin/checksum-CRC32`go env GOEXE`.new `go env GOPATH`/bin/checksum-CRC32`go env GOEXE`

vars:
	@echo SHELL="'$(SHELL)'"
	@echo BRANCH="'$(BRANCH)'"
	@echo TAG="'$(TAG)'"
	@echo VERSION="'$(VERSION)'"
	@echo NEXT_VERSION="'$(NEXT_VERSION)'"
	@echo GO_VERSION="'$(GO_VERSION)'"

# Quick test
quicktest:
	go test $(BUILDTAGS) $(GO_FILES)

racequicktest:
	go test $(BUILDTAGS) -cpu=2 -race $(GO_FILES)

# Do source code quality checks
check:	checksum_crc32
	@echo "-- START CODE QUALITY REPORT -------------------------------"
	@golangci-lint run $(LINTTAGS) ./...
	@echo "-- END CODE QUALITY REPORT ---------------------------------"

# Get the build dependencies
build_dep:
	go run bin/get-github-release.go -extract golangci-lint golangci/golangci-lint 'golangci-lint-.*\.tar\.gz'

# Get the release dependencies
release_dep:
	go run bin/get-github-release.go -extract nfpm goreleaser/nfpm 'nfpm_.*_Linux_x86_64.tar.gz'
	go run bin/get-github-release.go -extract github-release aktau/github-release 'linux-amd64-github-release.tar.bz2'

# Update dependencies
update:
	GO111MODULE=on go get -u ./...
	GO111MODULE=on go mod tidy

# Tidy the module dependencies
tidy:
	GO111MODULE=on go mod tidy

clean:
	go clean ./...
	find . -name \*~ | xargs -r rm -f
	rm -f checksum-CRC32

tarball:
	git archive -9 --format=tar.gz --prefix=checksum-CRC32-$(TAG)/ -o build/checksum-CRC32-$(TAG).tar.gz $(TAG)

sign_upload:
	cd build && md5sum checksum-CRC32-v* | gpg --clearsign > MD5SUMS
	cd build && sha1sum checksum-CRC32-v* | gpg --clearsign > SHA1SUMS
	cd build && sha256sum checksum-CRC32-v* | gpg --clearsign > SHA256SUMS

check_sign:
	cd build && gpg --verify MD5SUMS && gpg --decrypt MD5SUMS | md5sum -c
	cd build && gpg --verify SHA1SUMS && gpg --decrypt SHA1SUMS | sha1sum -c
	cd build && gpg --verify SHA256SUMS && gpg --decrypt SHA256SUMS | sha256sum -c

upload_github:
	./bin/upload-github $(TAG)

cross:
	go run bin/cross-compile.go -release current $(BUILDTAGS) $(TAG)

log_since_last_release:
	git log $(LAST_TAG)..

compile_all:
	go run bin/cross-compile.go -compile-only $(BUILDTAGS) $(TAG)

tag:	doc
	@echo "Old tag is $(VERSION)"
	@echo "New tag is $(NEXT_VERSION)"
	echo "$(NEXT_VERSION)" > VERSION
	git tag -s -m "Version $(NEXT_VERSION)" $(NEXT_VERSION)
	@echo "Then commit all the changes"
	@echo git commit -m \"Version $(NEXT_VERSION)\" -a -v
	@echo "And finally run make retag before make cross etc"

retag:
	git tag -f -s -m "Version $(VERSION)" $(VERSION)

winzip:
	zip -9 checksum-CRC32-$(TAG).zip checksum-CRC32.exe

generate:
	@go generate ./...
	@echo "[OK] Files added to embed box!"
