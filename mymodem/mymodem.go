/*
--------------------------------------------------
Tested on:
Raspberry Pi 3b with Waveshare GSM SIM868 module
--------------------------------------------------
*/

package mymodem

import (
	"bufio"
	"log"
	"strings"
	"time"

	"github.com/alarmia/logthis"
	"github.com/tarm/serial"
)

type MyModem struct {
	ComPort  string
	BaudRate int
	Port     *serial.Port
	DeviceId string
}

// Modem config
func New(ComPort string, BaudRate int, DeviceId string) (modem *MyModem) {
	modem = &MyModem{ComPort: ComPort, BaudRate: BaudRate, DeviceId: DeviceId}
	return modem
}

// Primary func

func (m *MyModem) SendSMS(message string, number string) string {
	logthis.Suffix = "GSM"
	logthis.Logthis("INFO", "Start sending a message to the number: "+number)
	err := m.Connect()
	if err != nil {
		logthis.Logthis("ERROR", "Failure to send a message to the number: "+number+" - Modem not responding")
		return "ERROR"
	}

	a := m.initModem()
	if a == "ERROR" {
		logthis.Logthis("ERROR", "Failure to send a message to the number: "+number+" - Modem 'ERROR' response")
		return a
	}

	a = m.sendCommand("AT+CMGF=1\r\n", true)
	if a == "ERROR" {
		logthis.Logthis("ERROR", "Failure to send a message to the number: "+number+" - Modem 'ERROR' response for AT+CMGF=1 command")
		return a
	}

	m.sendCommand("AT+CMGS=\""+number+"\"\r", false)
	time.Sleep(time.Second * 1)
	m.sendCommand(message, false)
	a = m.sendCommand("\x1a\r\n", true)
	if a == "ERROR" {
		// My modem does not always return OK, I don't know why.
		logthis.Logthis("INFO", "Sending a message to the number: "+number+" - No confirmation")
		return "OK"
	}

	logthis.Logthis("INFO", "A message was successfully sent to: "+number)
	return "OK"
}

// Start connection
func (m *MyModem) Connect() (err error) {
	config := &serial.Config{Name: m.ComPort, Baud: m.BaudRate, ReadTimeout: time.Second * 15}
	m.Port, err = serial.OpenPort(config)
	return err
}

// Modem init
func (m *MyModem) initModem() string {
	a := m.sendCommand("AT\r\n", true) // Should return "OK"
	if a == "ERROR" {
		logthis.Logthis("ERROR", "No response to AT command")
		return "ERROR"
	}
	a = m.sendCommand("ATE0\r\n", true) // echo mode off [ATE1 - echo mode on]
	if a == "ERROR" {
		logthis.Logthis("ERROR", "No response to ATE command")
		return "ERROR"
	}
	return "OK"
}

/* ---> SENDING <--- */
// Sending command
func (m *MyModem) Send(command string) {
	m.Port.Flush()
	time.Sleep(time.Second / 2)
	_, err := m.Port.Write([]byte(command))
	if err != nil {
		log.Fatal(err)
	}
}

// Sending command - preparation
func (m *MyModem) sendCommand(command string, waitForOk bool) string {

	if command == "\x1a\r\n" {
		time.Sleep(time.Second * 1)
		m.Send(command)
		//time.Sleep(time.Second * 1)
	} else {
		m.Send(command)
	}
	if waitForOk {
		output := m.read()
		return output
	}
	return "OK"
}

// Modem answer
func (m *MyModem) read() string {
	var cok int // ok counter
	scanner := bufio.NewScanner(m.Port)

	for scanner.Scan() {
		cok = 0
		if strings.Contains(scanner.Text(), "OK") {
			cok = +1
			break
		} else if strings.Contains(scanner.Text(), "ERROR") {
			break
		} else if scanner.Text() == "" {
			continue
		} else {
			continue
		}
	}
	// Generating a reply message
	if cok >= 1 {
		return "OK"
	} else {
		return "ERROR"
	}
}
