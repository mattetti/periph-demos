# Dia de los Muertos

Setup to lit up a calavera decorated with 44 apa102 LEDs.

## Setup

- Make sure spi is enabled on your pi.
- Update the Makefile to point to the right ip.
- `make zero scp` to build and push the program + extra files (pattern.json and the example service)

## Connectivity

The apa102s must be connected via SPI. Connect the following pins:

- 5v (first pin opposite of the SD card)
- MOSI (data) (pin 10 on the side of the SD card)
- SCLK (clock) (pin 12 on the side of the SD card)
- ground (pin 13 on the side of the SD card)

## Run automatically

Setting up a service allows the calavera to lit up automatically when the device
is connected.

An example service is provided and automatically transfered when you run `make scp`: `diadelosmuertos.service`.
But you need to install that service.

The service must be installed in `/lib/systemd/system/`, ssh into your device, copy the file and change the permissions on it

```
sudo cp diadelosmuertos.service /lib/systemd/system/
sudo chmod 644 /lib/systemd/system/diadelosmuertos.service
```

Then load/enable the service:

```
sudo systemctl daemon-reload
sudo systemctl enable diadelosmuertos.service
```

Reboot so the service can load:

```
sudo reboot
```

To see the logs use the following commands:

```
journalctl -u diadelosmuertos.service
```

## Customize

You can customize the experience by using another pattern.json file or by passing the number of LEDs to use. Edit the service command and add `-n 24` to set the LED number to 24 for instance.

A couple important things, changing the service won't make your change immediate, for that you need to reload it and restart the service as shown below:

```
sudo systemctl daemon-reload
systemctl restart diadelosmuertos.service
```
