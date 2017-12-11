#!/bin/bash

if [[ $(gofmt -l $(find . -type f -name '*.go' -not -path './vendor/*')) ]]; then
    echo -e "\033[0;31mTurned out that not everything is formatted according to gofmt.\n\033[0;31mPlease execute gofmt manually or configure your editor/IDE to do so for you. Thanks! :)\033[0m"
    exit 1
fi
