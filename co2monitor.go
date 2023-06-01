package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"regexp"
	"syscall"
	"time"

	"encoding/json"
	"strings"

	"github.com/tarm/serial"
)

type SensorData struct {
	Co2  int     `json:"co2"`
	Hum  float64 `json:"hum"`
	Temp float64 `json:"temp"`
}

func main() {

	serialPortPath := getDevicePath()

	// シリアルポートの設定
	c := &serial.Config{
		Name: serialPortPath, // 使用するシリアルポートのパスを指定
		Baud: 115200,         // ボーレートを指定
	}

	// シリアルポートを開く
	s, err := serial.OpenPort(c)
	if err != nil {
		log.Fatal(err)
	}
	defer s.Close()

	// Ctrl+Cが実行された場合にSTPコマンドを実行
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("Ctrl+C detected. Sending stop command...")
		exitCmd := "STP\r\n"
		_, err := s.Write([]byte(exitCmd))
		time.Sleep(500 * time.Millisecond)
		if err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()

	// STAコマンドによってログが開始している場合があるので、最初にまず止まる（リセット）
	time.Sleep(100 * time.Millisecond)
	cmd := "STP\r\n"
	_, err = s.Write([]byte(cmd))
	if err != nil {
		log.Fatal(err)
	}

	// STAコマンドでログ開始
	time.Sleep(500 * time.Millisecond)
	cmd = "STA\r\n"
	_, err = s.Write([]byte(cmd))
	if err != nil {
		log.Fatal(err)
	}

	buf := make([]byte, 27) // 25byte + CRLF = 27 byte

	for {
		n, err := s.Read(buf)
		if err != nil {
			log.Fatal(err)
		}

		data := string(buf[:n])

		// fmt.Println(data) // 受信データをコンソールに表示

		// 改行文字（CRLF）を検出した場合の処理
		if strings.Contains(data, "\r\n") {
			if strings.HasPrefix(data, "CO2") {
				line := strings.Split(data, "\r\n")[0]
				if line != "" {
					// 受信データを解析してJSON形式に変換
					sensorData, err := parseSensorData(line)
					if err != nil {
						log.Println("Failed to parse sensor data:", err)
						continue
					}

					// JSON形式のデータを表示
					jsonData, err := json.MarshalIndent(sensorData, "", "  ")
					if err != nil {
						log.Println("Failed to marshal JSON data:", err)
						continue
					}
					fmt.Println(string(jsonData))
				}

			}
		}

	}

}

// 受信データを解析してSensorData構造体に変換
func parseSensorData(data string) (*SensorData, error) {
	// データをキーと値に分割してマップに格納
	dataMap := make(map[string]string)
	pairs := strings.Split(data, ",")
	for _, pair := range pairs {
		keyValue := strings.Split(pair, "=")
		if len(keyValue) != 2 {
			return nil, fmt.Errorf("invalid data format")
		}
		dataMap[keyValue[0]] = keyValue[1]
	}

	// SensorData構造体に値を詰め替え
	sensorData := &SensorData{}
	co2, err := parseIntValue(dataMap["CO2"])
	if err != nil {
		return nil, fmt.Errorf("failed to parse CO2 value: %w", err)
	}
	sensorData.Co2 = co2

	hum, err := parseFloatValue(dataMap["HUM"])
	if err != nil {
		return nil, fmt.Errorf("failed to parse HUM value: %w", err)
	}
	sensorData.Hum = hum

	temp, err := parseFloatValue(dataMap["TMP"])
	if err != nil {
		return nil, fmt.Errorf("failed to parse TMP value: %w", err)
	}
	sensorData.Temp = temp

	return sensorData, nil
}

// 文字列を整数値に変換
func parseIntValue(s string) (int, error) {
	var value int
	_, err := fmt.Sscanf(s, "%d", &value)
	if err != nil {
		return 0, err
	}
	return value, nil
}

// 文字列を浮動小数点数に変換
func parseFloatValue(s string) (float64, error) {
	var value float64
	_, err := fmt.Sscanf(s, "%f", &value)
	if err != nil {
		return 0, err
	}
	return value, nil
}

func getDevicePath() string {

	// /dev/ディレクトリ内のデバイス一覧を取得
	files, err := ioutil.ReadDir("/dev/")
	if err != nil {
		log.Fatal(err)
	}

	// tty.usbで始まるデバイスパスを抽出
	devicePath := ""
	for _, file := range files {
		if matched, _ := regexp.MatchString("^tty\\.usbmodem", file.Name()); matched {
			devicePath = fmt.Sprintf("/dev/%s", file.Name())
			break
		}
	}

	return devicePath
}
