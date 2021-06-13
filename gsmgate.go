/* For my modem */

package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"mymodem/mymodem"
	"os"
	"strconv"
	"time"

	"github.com/alarmia/logthis"
	_ "github.com/go-sql-driver/mysql"
)

type SmsGate struct {
	Name     string `json:"name"`
	Address  string `json:"address"`
	Port     string `json:"port"`
	User     string `json:"user"`
	Pass     string `json:"pass"`
	Protocol string `json:"protocol"`
}

type Sms struct {
	Message, Source   string
	Id, IdSms, Number int
}

var smsgate SmsGate
var p Sms

func main() {
	logthis.Suffix = "MGSM"
	m := mymodem.New("/dev/ttyS0", 115200, "SMS GATE")
	smsGateConfig()
	for {
		err := checkSmsGateDb()
		if err == nil {
			s := strconv.Itoa(p.Number)
			s = "+48" + s
			a := m.SendSMS(p.Message, s)
			if a != "ERROR" {
				err = moveSmsInDb()
				logthis.Suffix = "MGSM"
				if err != nil {
					logthis.Logthis("ERROR", "CAN'T MOVE MESSAGE IN DB")
					time.Sleep(time.Second * 5)
				}
				logthis.Logthis("INFO", "A message was successfully sent to:"+s)
				time.Sleep(time.Second * 5)
			} else {
				logthis.Suffix = "MGSM"
				logthis.Logthis("ERROR", "UNABLE TO SEND MESSAGE TO NUMBER:"+s)
			}

		}
		time.Sleep(time.Second * 1)
	}
}

func smsGateConfig() {

	jsonFile, err := os.Open("confSMSGate.json") //Open json file
	if err != nil {
		log.Fatal("FATAL ERROR")
	}
	defer jsonFile.Close()                   // close json file when possible
	byteValue, _ := ioutil.ReadAll(jsonFile) // read opened json file as a byte array
	// initialize machine array
	json.Unmarshal(byteValue, &smsgate) //unmarshal byteArray with contains json file contnet into 'gate' witch is defined above
}

func checkSmsGateDb() (err error) {

	connStr := smsgate.User + ":" + smsgate.Pass + "@" + smsgate.Protocol + "(" + smsgate.Address +
		":" + smsgate.Port + ")/" + smsgate.Name // string constructor
	db, err := sql.Open("mysql", connStr) // connect to db
	if err != nil {
		logthis.Logthis("ERROR", "CAN'T OPEN DB")
	}
	defer db.Close() // close db connection when possible
	err = db.Ping()  // check db connection
	if err != nil {
		logthis.Logthis("ERROR", "CAN'T OPEN DB")
	}

	err = db.QueryRow("SELECT Id, Id_SMS, Number, Message, Source FROM ToSendSMS ORDER BY Id LIMIT 1;").Scan(&p.Id, &p.IdSms, &p.Number, &p.Message, &p.Source)
	if err != nil {
		err = errors.New("Empty")
		return
	}
	return nil
}

func moveSmsInDb() (err error) {

	connStr := smsgate.User + ":" + smsgate.Pass + "@" + smsgate.Protocol + "(" + smsgate.Address +
		":" + smsgate.Port + ")/" + smsgate.Name // string constructor
	db, err := sql.Open("mysql", connStr) // connect to db
	if err != nil {
		logthis.Logthis("ERROR", "CAN'T OPEN DB")
	}
	defer db.Close() // close db connection when possible
	err = db.Ping()  // check db connection
	if err != nil {
		logthis.Logthis("ERROR", "CAN'T OPEN DB")
	}
	stmt, err := db.Prepare("INSERT INTO MessSent VALUES (?,?,?,?,?)")
	if err != nil {
		logthis.Logthis("ERROR", "CAN'T OPEN DB")
	}
	_, err = stmt.Exec(p.Id, p.IdSms, p.Number, p.Message, p.Source)
	if err != nil {
		logthis.Logthis("ERROR", "CAN'T OPEN DB")
		return err
	}

	stmt, err = db.Prepare("DELETE FROM ToSendSMS WHERE Id=?")
	if err != nil {
		logthis.Logthis("ERROR", "CAN'T OPEN DB")
		return err
	}

	res, err := stmt.Exec(p.Id)
	if err != nil {
		logthis.Logthis("ERROR", "CAN'T OPEN DB")
		return err
	}
	_, err = res.RowsAffected()
	if err != nil {
		logthis.Logthis("ERROR", "CAN'T OPEN DB")
		return err
	}

	return err
}
