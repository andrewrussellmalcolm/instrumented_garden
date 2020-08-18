package main

import (
	"flag"
	"fmt"
	ttnsdk "github.com/TheThingsNetwork/go-app-sdk"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var (
	temperatureGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "temperature",
			Help: "temperature reported by a weather station over time in degrees Celcius",
		},
		[]string{"dev_id"})

	humidityGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "humidity",
			Help: "humidity reported by a weather station over time in percent",
		},
		[]string{"dev_id"})

	pressureGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "pressure",
			Help: "barometric reported by a weather station over time",
		},
		[]string{"dev_id"})

	intensityGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "intensity",
			Help: "light intensity reported by a weather station over time",
		},
		[]string{"dev_id"})

	batteryGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "battery",
			Help: "battery voltage reported by a weather station over time",
		},
		[]string{"dev_id"})

	addr = flag.String("prom-server-address", ":8080", "the address of the server for prometheus metrics")

	appID         = "loramini2"
	appAccessKey  = "ttn-account-v2.qg3uDMEdahYoMuUvOM2iWWI02OgoCg8wr53zRfGL1GM"
	sdkClientName = "lora-weather-station"
)

func main() {

	flag.Parse()
	prometheus.MustRegister(temperatureGauge)
	prometheus.MustRegister(humidityGauge)
	prometheus.MustRegister(pressureGauge)
	prometheus.MustRegister(intensityGauge)
	prometheus.MustRegister(batteryGauge)

	go func() {
		http.Handle("/metrics", promhttp.HandlerFor(
			prometheus.DefaultGatherer,
			promhttp.HandlerOpts{},
		))

		fmt.Printf("starting prometheus server on %s\n", *addr)

		err := http.ListenAndServe(*addr, nil)
		if err != nil {
			fmt.Printf("unable to start prometheus server\v")
		}
	}()

	config := ttnsdk.NewCommunityConfig(sdkClientName)

	client := config.NewClient(appID, appAccessKey)
	defer client.Close()

	devices, err := client.ManageDevices()
	if err != nil {
		fmt.Printf("could not get device manager\n")
	}

	// List the first 10 devices
	deviceList, err := devices.List(10, 0)
	if err != nil {
		fmt.Printf("could not get devices\n")
	}
	fmt.Printf("found device(s)\n")
	for _, device := range deviceList {
		fmt.Printf("%s\n", device.DevID)
	}

	fmt.Println()

	pubsub, err := client.PubSub()
	if err != nil {
		fmt.Printf("could not get application pub/sub\n")
	}

	defer pubsub.Close()

	allDevicesPubSub := pubsub.AllDevices()
	defer allDevicesPubSub.Close()
	uplink, err := allDevicesPubSub.SubscribeUplink()
	if err != nil {
		fmt.Printf("could not subscribe to uplink messages\n")
	}
	fmt.Printf("waiting for data\n")
	for message := range uplink {
		timestamp := time.Time(message.Metadata.Time).Format(time.RFC3339)
		measurements := strings.Split(string(message.PayloadRaw), ",")

		if len(measurements) < 4 {
			fmt.Printf("uplink message malformed, ignored\n")
			continue
		}

		fmt.Printf("measurement from %s at %s : %s %s %s %s %s\n", message.DevID, timestamp, measurements[0], measurements[1], measurements[2], measurements[3], measurements[4])

		temperature, err := strconv.ParseFloat(measurements[0], 64)

		if err == nil {
			temperatureGauge.With(prometheus.Labels{"dev_id": message.DevID}).Set(temperature)
		} else {
			fmt.Printf("error parsing temperature: %s\n", err)
		}

		humidity, err := strconv.ParseFloat(measurements[1], 64)

		if err == nil {
			humidityGauge.With(prometheus.Labels{"dev_id": message.DevID}).Set(humidity)
		} else {
			fmt.Printf("error parsing humidity: %s\n", err)
		}

		pressure, err := strconv.ParseFloat(measurements[2], 64)

		if err == nil {
			pressureGauge.With(prometheus.Labels{"dev_id": message.DevID}).Set(pressure)
		} else {
			fmt.Printf("error parsing pressure: %s\n", err)
		}

		intensity, err := strconv.ParseFloat(measurements[3], 64)

		if err == nil {
			intensityGauge.With(prometheus.Labels{"dev_id": message.DevID}).Set(intensity)
		} else {
			fmt.Printf("error parsing intensity: %s\n", err)
		}

		battery, err := strconv.ParseFloat(measurements[4], 64)

		if err == nil {
			batteryGauge.With(prometheus.Labels{"dev_id": message.DevID}).Set(battery)
		} else {
			fmt.Printf("error parsing battery: %s\n", err)
		}
	}
}
