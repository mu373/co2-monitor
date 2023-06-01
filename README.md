# co2monitor

## About
Get sensor values (CO2, humidity, and temperature) from [UD-CO2S](https://www.iodata.jp/product/tsushin/iot/ud-co2s/). The source code currently asumes that the device is available as `/dev/tty.usbmodemXXXX`.

## Run
```sh
# Build
go build co2monitor.go
# Run
./co2monitor
# To stop, press ctrl-c
```

## Outputs
```txt
$ ./co2monitor
{
  "co2": 699,
  "hum": 46.7,
  "temp": 30.8
}
{
  "co2": 699,
  "hum": 46.7,
  "temp": 30.8
}
{
  "co2": 699,
  "hum": 46.6,
  "temp": 30.7
}
```