git_reversion := $(shell git rev-parse HEAD)
git_workdirty := $(shell git diff --quiet HEAD && echo "clean")
version := $(git_reversion)-$(git_workdirty)
options := -ldflags "-X main.VERSION=$(version)"

default:
	go build $(options)
	
android:
	GOOS=android GOARCH=amd64 \
		 CC=/data/ndk/bin/x86_64-linux-android-gcc \
		 CXX=/data/ndk/bin/x86_64-linux-android-g++ \
		 CGO_ENABLED=1 go build $(options)
