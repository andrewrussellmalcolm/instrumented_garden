package main

import (
	"os"
	"fmt"
	"time"
	"strings"
	ttnsdk "github.com/TheThingsNetwork/go-app-sdk"
)

const (
	sdkClientName = "lora-weather-station"
)

func main() {

	f, err := os.OpenFile("weather.dat", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(err)
	}

	defer f.Close()

	appID := "loramini2"
	appAccessKey := "ttn-account-v2.qg3uDMEdahYoMuUvOM2iWWI02OgoCg8wr53zRfGL1GM"

	config := ttnsdk.NewCommunityConfig(sdkClientName)

	client := config.NewClient(appID, appAccessKey)
	defer client.Close()

	devices, err := client.ManageDevices()
	if err != nil {
		fmt.Printf("lora-weather-station: could not get device manager\n")
	}

	// List the first 10 devices
	deviceList, err := devices.List(10, 0)
	if err != nil {
		fmt.Printf("lora-weather-station: could not get devices\n")
	}
	fmt.Printf("lora-weather-station: found device(s)\n")
	for _, device := range deviceList {
		fmt.Printf("%s\n", device.DevID)
	}

	pubsub, err := client.PubSub()
	if err != nil {
		fmt.Printf("lora-weather-station: could not get application pub/sub\n")
	}

	defer pubsub.Close()

	allDevicesPubSub := pubsub.AllDevices()
	defer allDevicesPubSub.Close()
	uplink, err := allDevicesPubSub.SubscribeUplink()
	if err != nil {
		fmt.Printf("lora-weather-station: could not subscribe to uplink messages")
	}
	fmt.Printf("waiting for data\n")
	for message := range uplink {
		timestamp := time.Time(message.Metadata.Time).Unix()
		measurements := strings.Split(string(message.PayloadRaw), ",")

		if len(measurements)<4 {
			fmt.Printf("lora-weather-station: uplink message malformed, ignored")
			continue
		}

		data := fmt.Sprintf("%s %d %s %s %s %s\n", message.HardwareSerial, timestamp,measurements[0],measurements[1],measurements[2],measurements[3],,measurements[4])		

		fmt.Printf(data)

		if _, err := f.WriteString(data); err != nil {
			fmt.Printf("lora-weather-station: could not write data")
		}
	}
}
