package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os/user"
	"time"

	i2c "github.com/d2r2/go-i2c"
	rpio "github.com/stianeikeland/go-rpio"
	"github.com/yosssi/gmq/mqtt"
	"github.com/yosssi/gmq/mqtt/client"
)

type Configuration struct {
	BaseLine    int    `json:"BaseLine"`
	Logger      bool   `json:"Logger"`
	BasicTimer  int    `json:"BasicTimer"`
	MqttAddress string `json:"MqttAddress"`
	MqttTopic   string `json:"MqttTopic"`
}

var configuration Configuration
var beforebol bool

func proxiMeas(i2c *i2c.I2C, cli *client.Client) bool {
	data1, _ := i2c.ReadRegU8(0x80)

	i2c.WriteRegU8(0x80, data1+8)
	time.Sleep(time.Duration(10) * time.Millisecond)
	for {
		data1, _ := i2c.ReadRegU8(0x80)
		if (data1 & 0x20) == 0x00 {
			break
		}
		if configuration.Logger {
			fmt.Println(data1)
		}
		time.Sleep(time.Duration(10) * time.Millisecond)
	}
	prox1, _ := i2c.ReadRegU8(0x87)
	prox2, _ := i2c.ReadRegU8(0x88)
	proxiNum := uint16(prox1)*256 + uint16(prox2)
	if configuration.Logger {
		fmt.Println(data1, prox1, prox2)
		fmt.Println(proxiNum)

	}
	actualbol := false
	if int(proxiNum) > configuration.BaseLine {
		actualbol = true
	}

	if actualbol != beforebol {
		user, err := user.Current()
		if err != nil {
			log.Fatal(err)
		}

		str := "door close"
		if !actualbol && beforebol {
			str = "door open!!!!44!!!!NÉGY!!"
		}
		err = cli.Publish(&client.PublishOptions{
			QoS:       mqtt.QoS0,
			TopicName: []byte("log"),
			Message:   []byte(user.Name + ":" + str),
		})
		if err != nil {
			fmt.Println("ohooo")
			panic(err)
		}
	}
	return actualbol
}

func main() {

	dat, _ := ioutil.ReadFile("conf.json")
	decoder := json.NewDecoder(bytes.NewBufferString(string(dat)))
	fmt.Println(string(dat))
	err := decoder.Decode(&configuration)

	// Create an MQTT Client.
	cli := client.New(&client.Options{
		// Define the processing of the error handler.
		ErrorHandler: func(err error) {
			fmt.Println(err)
		},
	})
	// Terminate the Client.
	defer cli.Terminate()
	user, err := user.Current()
	// Connect to the MQTT Server.
	err = cli.Connect(&client.ConnectOptions{
		Network:  "tcp",
		Address:  configuration.MqttAddress,
		ClientID: []byte(user.Name + "rpitamper"),
	})
	if err != nil {
		panic(err)
	}

	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Println(configuration)
	// Create new connection to I2C bus on 2 line with address 0x27
	i2c, err := i2c.NewI2C(0x13, 1)
	if err != nil {
		fmt.Println(err)
	}
	// Free I2C connection on exit
	defer i2c.Close()
	i2c.WriteRegU8(0x82, 0x7)
	data, _ := i2c.ReadRegU8(0x83)
	i2c.WriteRegU8(0x83, (data&0xC0)+15)

	err = rpio.Open()
	if err != nil {
		fmt.Println(err)
	}

	pin := rpio.Pin(12)
	pin.Output()

	if configuration.BasicTimer == 0 {
		for {
			bol := proxiMeas(i2c, cli)
			if bol {
				pin.High()
			} else {
				pin.Low()
			}
			time.Sleep(time.Duration(100) * time.Millisecond)
		}
	} else {
		meterTicker := time.NewTicker(time.Second * time.Duration(configuration.BasicTimer))
		go func() {
			for t := range meterTicker.C {
				fmt.Println(t)
				bol := proxiMeas(i2c, cli)
				if bol {
					pin.High()
				} else {
					pin.Low()
				}
			}
		}()
		for {
			time.Sleep(time.Duration(100) * time.Hour)
		}
	}

	// Here goes code specific for sending and reading data
	// to and from device connected via I2C bus, like:
	//_, err := i2c.Write([]byte{0x1, 0xF3})
	//if err != nil { log.Fatal(err) }

}
