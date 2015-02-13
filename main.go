package main

import (
	"encoding/json"
	"fmt"
	"github.com/lukegb/tflcountdown"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

const (
	LED_STRIP_LENGTH      = 150
	LED_STRIP_LENGTH_TIME = 15 * time.Minute
	LED_TIME              = LED_STRIP_LENGTH_TIME / LED_STRIP_LENGTH

	MIN_RUN_TIME  = 1 * time.Minute
	MIN_SAFE_TIME = 3 * time.Minute
	MAX_SAFE_TIME = 5 * time.Minute
)

var (
	INTERESTING_BUS_STOPS = []string{"15353"}

	TFL_URL = ""

	SPARK_URL             = ""
	SPARK_ACCESS_TOKEN    = ""
)

func performUpdate(api *tflcountdown.InstantAPI, req tflcountdown.Request) ([]bool, error) {
	leds := make([]bool, LED_STRIP_LENGTH)
	now := time.Now()

	msgChan, errChan := api.MakeRequest(req)
	for {
		select {
		case msg := <-msgChan:
			if msg, ok := msg.(tflcountdown.PredictionData); ok {
				b, err := json.Marshal(msg)
				estTime := msg.EstimatedTime

				if estTime.After(now.Add(LED_STRIP_LENGTH_TIME)) {
					//log.Println("SKIPMSG", string(b), err)
					continue // skip this
				}
				log.Println("MSG", string(b), err)

				// find location in leds
				timeFromNow := estTime.Sub(now)
				ledsFromEnd := timeFromNow / LED_TIME
				leds[LED_STRIP_LENGTH-int(ledsFromEnd)] = true
			}
		case err := <-errChan:
			if err == nil {
				return leds, nil
			} else {
				log.Fatalln("ERR", err)
				return nil, err
			}
		}
	}
}

func pushTo(endpoint, data string) error {
	fnUrl := fmt.Sprintf("%s%s", SPARK_URL, endpoint)
	log.Println("pushing to", fnUrl)
	resp, err := http.PostForm(fnUrl, url.Values{"data": {data}, "access_token": {SPARK_ACCESS_TOKEN}})
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func main() {
	if len(os.Args) != 4 {
		log.Println(os.Args[0], "<Spark Device URL>", "<Spark Access Token>", "<TFL Countdown Instant API>")
		os.Exit(1)
		return // just because
	}
	SPARK_URL = os.Args[1]
	SPARK_ACCESS_TOKEN = os.Args[2]
	TFL_URL = os.Args[3]

	api, err := tflcountdown.MakeInstantAPI(TFL_URL)
	if err != nil {
		panic(err)
	}

	log.Println("STARTING")

	fm := tflcountdown.NewDefaultFieldMap()
	fm.Add("EstimatedTime")
	fm.Add("StopID")
	fm.Add("VehicleID")
	req := tflcountdown.Request{
		StopID:     INTERESTING_BUS_STOPS,
		ReturnList: &fm,
	}

	for {
		leds, err := performUpdate(api, req)
		if err != nil {
			log.Println(err)
		}

		if true {
			fmt.Print("[")
			for _, x := range leds {
				if x {
					fmt.Print("x")
				} else {
					fmt.Print(" ")
				}
			}
			fmt.Println("]")
		}

		outdat := ""
		for n, x := range leds {
			timeFromNow := time.Duration(len(leds)-n) * LED_TIME
			outchar := "0"
			if x && timeFromNow < MIN_RUN_TIME {
				outchar = "1"
			} else if x && timeFromNow < MIN_SAFE_TIME {
				outchar = "2"
			} else if x && timeFromNow < MAX_SAFE_TIME {
				outchar = "3"
			} else if x {
				outchar = "4"
			}
			if n == len(leds)-1 {
				outchar = "5"
			}
			outdat += outchar
		}
		split1 := outdat[0:62]
		split2 := outdat[62:124]
		split3 := outdat[124:150]
		if err := pushTo("push1", split1); err != nil {
			log.Println("push1err", err)
		}
		if err := pushTo("push2", split2); err != nil {
			log.Println("push2err", err)
		}
		if err := pushTo("push3", split3); err != nil {
			log.Println("push3err", err)
		}
		log.Println(split1, split2, split3, len(split1), len(split2), len(split3), len(split1)+len(split2)+len(split3))

		time.Sleep(LED_TIME / 2)
	}
}
