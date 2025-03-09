GOOS=darwin GOARCH=arm64 go build -o bin/Apple/arm64/sumuser
GOOS=darwin GOARCH=amd64 go build -o bin/Apple/amd64/sumuser
GOOS=windows GOARCH=amd64 go build -o bin/Windows/amd64/sumuser
GOOS=windows GOARCH=386 go build -o bin/Windows/x86/sumuser
GOOS=linux GOARCH=amd64 go build -o bin/Linux/amd64/sumuser
