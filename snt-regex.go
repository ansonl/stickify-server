package main

import (
	"fmt"
	"regexp"
	"net/http"
	"strconv"
	"encoding/json"
	"encoding/base64"
	"os"
	"time"
)

var users map[string]string
var userStickies map[string][][]string

var startTime = time.Now()

func parseLines(raw string) []string {

	//these work if you declare stirng epxlictly, uses multiline mode on the string. For some reason casting byte slice to string doesn't let Go find the \n characters even though the byte slice is the equivalent of the string!
	//get replace all \\tx1234 matches with nothing
	re01 := regexp.MustCompile(`(\\tx[0-9]*|\\ul(none)?|\\lang[0-9]|\\f[0-9]|\\fs[0-9]([0-1]|[3-9])|\\fs22\\par)`)
	//find RTF left and right double quotes
	re02 := regexp.MustCompile(`(\\[lr]dblquote)`)
	//find tabs and put correct tab character
	re03 := regexp.MustCompile(`(\\tab)`)
	//find first line that begins at \fs22 with a space
	re1 := regexp.MustCompile(`(?m:^.*(fs22\s)(.*)\\par.?$)`)
	//find all other entries, dicard the first line
	re2 := regexp.MustCompile(`(?m:^(.*)\\par.?$)`)

	//postprocessing for each string
	//strip any remaining \par
	re3 := regexp.MustCompile(`(\\par)`)
	//rawString := string("b'{\\rtf1\\ansi\\ansicpg1252\\deff0\\deflang1033{\\fonttbl{\\f0\\fnil\\fcharset0 Segoe Print;}{\\f1\\fnil Segoe Print;}}\r\n{\\*\\generator Msftedit 5.41.21.2510;}\\viewkind4\\uc1\\pard\\tx384\\tx768\\tx1152\\tx1536\\tx1920\\tx2304\\tx2688\\tx3072\\tx3456\\tx3840\\tx4224\\tx4608\\tx4992\\tx5376\\tx5760\\tx6144\\tx6528\\tx6912\\tx7296\\tx7680\\tx8064\\tx8448\\tx8832\\tx9216\\tx9600\\tx9984\\tx10368\\tx10752\\tx11136\\tx11520\\tx11904\\tx12288\\f0\\fs22 F 9/25 extramural duty section finalize, pull from duty sections\\par\r\n\\par\r\nU UPDATE BOOW, follow up bonitz\\par\r\n\\par\r\nU 9 /27 presidents circle finalize\\par\r\nU 9/27 comedy show finalize\\par\r\n\\par\r\nM CALL motor\\par\r\nW 9/30 GET  Vehicle pickup for game 3\\par\r\nW 9/30 get drag finalized\\par\r\nF 10/2 1515 4020 israel interview\\par\r\n\\par\r\n1900 T 10/6 MI 110 mgsp training\\par\r\n1900 T 10/13 CH100 mgsp training\\par\r\n1250 W 12/2 mitscher mgsp training\\par\r\n\\par\r\n\\lang9\\f1\\par\r\n}\r\n\x00ang9\\f1\\par\r\n}\r\n\x00\n\x00lang9\\f1\\par\r\n}\r\n\x00f1\\par\r\n}\r\n\x00e 3\\lang9\\f1\\par\r\n}\r\n\x00ang9\\f1\\par\r\n}\r\n\x00minar\\lang9\\f1\\par\r\n}\r\n\x00racter dev seminar\\lang9\\f1\\par\r\n}\r\n\x00'")

	/*
	//Modified regex to work with a single block of text with \n treated as normal character
	//junk entries end with \\f1 so have [^1] before end of regex2
	re1 := regexp.MustCompile(`(\\\\fs22.*)`)
	re2 := regexp.MustCompile(`(\\\\fs22|\\r\\n)(.*?)[^1]\\\\par`)
	*/
	/*
	firstResult := re1.FindAllStringSubmatch(castedInput, -1)
	if len(firstResult) > 0 && len(firstResult[0]) > 1 {
		//fmt.Println(firstResult[0][1]);
		//allResults = append(allResults, firstResult[0][1])

		//pass into regex2
		otherResult := re2.FindAllStringSubmatch(firstResult[0][1], -1)
		for i,result := range otherResult {

			//fmt.Println("R1", i, ": ", result[2])

			if len(otherResult[i]) > 1 {
				allResults = append(allResults, result[2])
			}
		}
	}
	*/

	rawString, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		fmt.Println("base64 decode error:", err)
	}

	castedInput := string(rawString)
	//fmt.Println(castedInput)

	castedInput = re01.ReplaceAllString(castedInput, "")
	castedInput = re02.ReplaceAllString(castedInput, "\"")
	castedInput = re03.ReplaceAllString(castedInput, "\t")


	allResults := make([]string, 0)

	firstResult := re1.FindAllStringSubmatch(castedInput, -1)
	if len(firstResult) > 0 && len(firstResult[0]) > 1 {
		//fmt.Println(firstResult[0][2])
		if len(firstResult[0]) > 2 {
			allResults = append(allResults, firstResult[0][2])
		}
	}

	otherResult := re2.FindAllStringSubmatch(castedInput, -1)
	for i,result := range otherResult {

		if i == 0 {
			continue;
		}

		//fmt.Println("R1", i, ": ", result[2])

		if len(otherResult[i]) > 1 {
			result[1] = re3.ReplaceAllString(result[1], "")
			allResults = append(allResults, result[1])
		}
	}

	for i, r := range allResults {
		fmt.Println("R", i, ": ", r)
	}
	return allResults
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Sticky Server v1 by Anson Liu")
}

func uptimeHandler(w http.ResponseWriter, r *http.Request) {
    //bypass same origin policy
	w.Header().Set("Access-Control-Allow-Origin", "*")

	diff := time.Since(startTime)

	fmt.Fprintf(w, "" + "Uptime:\t" + diff.String())
	fmt.Println("Uptime requested")
}

func getUserHandler(w http.ResponseWriter, r *http.Request) {
	//bypass same origin policy
	w.Header().Set("Access-Control-Allow-Origin", "*")

	r.ParseForm()

	var user string
	var passcode string

	if len(r.Form["user"]) > 0 && len(r.Form["passcode"]) > 0{
		user = r.Form["user"][0]
		passcode = r.Form["passcode"][0]

		//check for blank user
		if (len(user) == 0){
			fmt.Println("blank user, sending back error");
			fmt.Fprintf(w, "Please supply a nickname.");
		}

		if (passcode == users[user]) {
			output, err := json.Marshal(userStickies[user]);
			if err != nil {
				fmt.Println(err)
			}
			fmt.Fprintf(w, string(output))
		} else {
			fmt.Fprintf(w, "Wrong passcode")
		}


	} else {
		fmt.Fprintf(w, "Missing nickname/passcode")
	}
}

func updateHandler(w http.ResponseWriter, r *http.Request) {
	//bypass same origin policy
	w.Header().Set("Access-Control-Allow-Origin", "*")

	r.ParseForm()
	//fmt.Println(r.Form)

	var user string
	var passcode string
	var stickyNumber int
	var data string
	var err error


	if len(r.Form["user"]) > 0 {
		user = r.Form["user"][0]

		fmt.Println("user: ", user)

		//check for blank user
		if (len(user) == 0){
			fmt.Println("blank user, sending back error");
			fmt.Fprintf(w, "Please supply a nickname.");
			return
		}



		//check if no passcode set, new user
		if users[user] == "" && len(r.Form["passcode"]) > 0 {
			//set new passcode
			passcode = r.Form["passcode"][0]
			users[user] = passcode

			//initialize user map array
			userStickies[user] = make([][]string, 0)

			fmt.Println("new user ", user)

			//set use first entry
			data = r.Form["data"][0]
			userStickies[user] = append(userStickies[user], parseLines(data))
			fmt.Fprintf(w, "0")
			return

		//if passcode set, verify
		} else if len(r.Form["passcode"]) > 0 && r.Form["passcode"][0] == users[user] {

			fmt.Println("passcode matches")

			//set data var
			if len(r.Form["data"]) > 0 {
				data = r.Form["data"][0]
			} else {
				fmt.Println("no data")
			}

			//check number
			if len(r.Form["number"]) > 0 {
				stickyNumber, err = strconv.Atoi(r.Form["number"][0])

				if err != nil { //failed number parse, append new note
					fmt.Println("Error parsing number, append new note")
					userStickies[user] = append(userStickies[user], parseLines(data))
					fmt.Fprintf(w, "0")
					return

				} else { //successful number parse, set user map index to data
					fmt.Println("Successful parsing number, set to index ", stickyNumber)
					if stickyNumber < len(userStickies[user]) {
						userStickies[user][stickyNumber] = parseLines(data)
						fmt.Fprintf(w, "0")
						return
					} else {
						for len(userStickies[user]) - 1 < stickyNumber {
							userStickies[user] = append(userStickies[user], make([]string, 0))
						}
						userStickies[user][stickyNumber] = parseLines(data)
						fmt.Fprintf(w, "0")
						return
					}

				}
			}

			if len(r.Form["data"]) > 0 {
				data = r.Form["data"][0]
			} else {
				fmt.Println("no data")
			}
			userStickies[user] = append(userStickies[user], parseLines(data))
			fmt.Fprintf(w, "0")

		} else {
			fmt.Fprintf(w, "Wrong passcode for nickname " + user + ".\n" + user + " may be taken by someone else. Fine time to choose another nickname.")
			return
		}
	} else {
		fmt.Fprintf(w, "1")
		return
	}

	//fmt.Println("assigned to",user + strconv.Itoa(stickyNumber))
	//strconv.Itoa(users[user])
	//stickies[user + strconv.Itoa(stickyNumber)] = parseLines(data)

	fmt.Fprintf(w, "0")
}

func server() {
	http.HandleFunc("/update", updateHandler)
	http.HandleFunc("/getUser", getUserHandler)
	http.HandleFunc("/uptime", uptimeHandler)
	http.HandleFunc("/", rootHandler)
	//http.ListenAndServe(":8080", nil)

	err := http.ListenAndServe(":"+os.Getenv("PORT"), nil)
	if err != nil {
		panic(err)
	}

	fmt.Println("listening on port " + os.Getenv("PORT"))
}

func main() {

	users = make(map[string]string)
	userStickies = make(map[string][][]string)

	go server()

	ticker := time.NewTicker(60 * 60 * 24 * time.Second)

	for {
		select {
		case <-ticker.C:
			users = make(map[string]string)
			userStickies = make(map[string][][]string)
		}
	}
}
