default: zero

zero:
	GOOS=linux GOARCH=arm GOARM=5 go build .

pi:
	GOOS=linux GOARCH=arm go build .

scp:
	scp ht16k33leds pi@raspi-cam.local:~/