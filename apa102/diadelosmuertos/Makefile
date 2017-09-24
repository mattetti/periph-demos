default: zero

zero:
	GOOS=linux GOARCH=arm GOARM=5 go build .

pi:
	GOOS=linux GOARCH=arm go build .

scp:
	scp diadelosmuertos pi@192.168.86.223:~