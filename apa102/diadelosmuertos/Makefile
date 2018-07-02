ip=192.168.86.52

default: zero

zero:
	GOOS=linux GOARCH=arm GOARM=5 go build .

pi:
	GOOS=linux GOARCH=arm go build .

scp:
	scp diadelosmuertos pi@$(ip)~
	scp pattern.json pi@$(ip):~
	scp diadelosmuertos.service pi@$(ip):~