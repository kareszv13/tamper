package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	i2c "github.com/d2r2/go-i2c"
	rpio "github.com/stianeikeland/go-rpio"
)

type Configuration struct {
	BaseLine    int    `json:"basicVerbose"`
	Logger      bool   `json:"basicLogger"`
	BasicTimer  int    `json:"basicTimer"`
	MqttAddress string `json:"mqttAddress"`
	MqttTopic   string `json:"mqttTopic"`
}

var configuration Configuration

func main() {

	dat, _ := ioutil.ReadFile("conf.json")
	decoder := json.NewDecoder(bytes.NewBufferString(string(dat)))
	fmt.Println(string(dat))
	err := decoder.Decode(&configuration)

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

	err = rpio.Open()
	if err != nil {
		fmt.Println(err)
	}

	pin := rpio.Pin(12)
	pin.Output()

	for {

		data1, _ := i2c.ReadRegU8(0x80)

		i2c.WriteRegU8(0x80, data1+8)
		time.Sleep(time.Duration(100000) * time.Microsecond)

		prox1, _ := i2c.ReadRegU8(0x87)
		prox2, _ := i2c.ReadRegU8(0x88)
		proxiNum := uint16(prox1)*256 + uint16(prox2)
		fmt.Println(proxiNum)
		if proxiNum > 2150 {
			pin.High()
		} else {
			pin.Low()
		}
	}

	// Here goes code specific for sending and reading data
	// to and from device connected via I2C bus, like:
	//_, err := i2c.Write([]byte{0x1, 0xF3})
	//if err != nil { log.Fatal(err) }

}
