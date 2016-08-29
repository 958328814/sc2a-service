@SET PKG=github.com/DreamHacks/sc2a-service
@SET GOOS=linux
@SET GOARCH=amd64
@SET VERSION=0.1.2
@SET REPO=gcr.io/b2dev-1296/sc2a
@SET TAG=%REPO%:%VERSION%

docker run --rm -v "%GOPATH%":/root/go -e GOPATH="/root/go" -w "/root/go/src/%PKG%" golang make
docker build -t %TAG% .
docker tag %TAG% %REPO%:latest
REM docker run --rm -p 8080:8080 -v "%cd%/data/config.json":/root/config.json %TAG%
gcloud docker push %TAG%
gcloud docker push %REPO%:latest