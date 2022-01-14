
HOST?=bed.fritz.box
build:
	env GOOS=linux GOARCH=arm GOARM=5 go build -o bed cmd/main.go

deploy: build
	-ssh pi@${HOST} "killall bed"
	scp bed pi@${HOST}:/home/pi/
	ssh -t pi@${HOST} "./bed"