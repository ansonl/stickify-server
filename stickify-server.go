package main

import (
	"fmt"
	"net/http"
	"strconv"
	"encoding/json"
	"encoding/base64"
	"os"
	"time"
	"strings"
//	"html"
//	"io/ioutil"
)

var users map[string]string
var userStickies map[string][][]string
var userLastUpdate map[string]time.Time

var startTime = time.Now()

var userExpireSeconds = 60 * 60 * 24 * time.Second;


func parseLines(raw string) []string {
	rawString, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		fmt.Println("Base64 decode error:", err)
	}

	castedInput := string(rawString)
	return strings.Split(castedInput, "\n")
}

func aboutHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	fmt.Fprintf(w, "Sticky Server by Anson Liu\n")
	fmt.Fprintf(w, "User expiry duration (with no updates from Stickify Pusher) is " + userExpireSeconds.String())
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "https://ansonl.github.io/stickify-web-app", 301)
	
	/*
	fmt.Println(html.EscapeString(r.URL.Path))
	
	file := "/index.html"

	if (r.URL.Path != "/") {
		file = r.URL.Path
	} 
	if (r.URL.Path == "/favicon.ico") {
		return
	}

	data, err := ioutil.ReadFile("./pages" + file)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Fprintf(w, string(data))
	*/
	
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
			fmt.Println("Blank user, sending back error.");
			fmt.Fprintf(w, "1 Please supply a nickname.");
		}

		if (passcode == users[user]) {
			output, err := json.Marshal(userStickies[user]);
			if err != nil {
				fmt.Println(err)
			}
			fmt.Fprintf(w, string(output))
		} else {
		    if (userStickies[user] == nil) { //passcode does not match and user has no stickies, account probably not taken
		        fmt.Println("User " + user + " not found.");
		        fmt.Fprintf(w, "1 Nickname " + user + " not found.")
		    } else {
		        fmt.Println("Wrong PIN for user " + user);
		        fmt.Fprintf(w, "1 Wrong PIN")
		    }
		}


	} else {
		fmt.Fprintf(w, "1 Missing nickname/PIN")
	}
}

func updateHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("test")

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

		fmt.Println("User "+ user + " sent update request.")

		//check for blank user
		if (len(user) == 0){
			fmt.Println("Blank user, sending back error.");
			fmt.Fprintf(w, "Please supply a nickname.");
			return
		}

		if (len(r.Form["passcode"]) > 0) {
			passcode = r.Form["passcode"][0]
		} else {
			passcode = ""
		}

		//check if no passcode set, new user
		if users[user] == "" && len(passcode) > 0 {
			//set new passcode
			users[user] = passcode

			//initialize user map array, if user does not exist
			if (userStickies[user] == nil) {
			    fmt.Println("New user " + user)
				fmt.Println("User " + user + " did not have map, so made one for user.")
				userStickies[user] = make([][]string, 0)
			}

			

			//set use first entry
			data = r.Form["data"][0]
			userStickies[user] = append(userStickies[user], parseLines(data))
			fmt.Fprintf(w, "0")
			return

		//if passcode set, verify
		} else if passcode == users[user] {

			fmt.Println("Passcode for user" + user + "matches.")

			//set data var
			if len(r.Form["data"]) > 0 {
				data = r.Form["data"][0]
			} else {
				fmt.Println("No data sent for update.")
			}

			//check number
			if len(r.Form["number"]) > 0 {
				stickyNumber, err = strconv.Atoi(r.Form["number"][0])

				//fmt.Println("Update note index ", r.Form["number"])

                //update user last updated time
                userLastUpdate[user] = time.Now()

				if err != nil { //failed number parse, append new note
					fmt.Println("Error parsing number, append new note to user stickies array.")
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
	http.HandleFunc("/about", aboutHandler)
	http.HandleFunc("/", rootHandler)
	//http.ListenAndServe(":8080", nil)

	err := http.ListenAndServe(":"+os.Getenv("PORT"), nil)
	if err != nil {
		panic(err)
	}

	fmt.Println("Server ended on port " + os.Getenv("PORT"))
}

func main() {

	users = make(map[string]string)
	userStickies = make(map[string][][]string)
	userLastUpdate = make(map[string]time.Time)

	go server()

    //check for old users
	ticker := time.NewTicker(60 * time.Second)

	for {
		select {
		case <-ticker.C:
		    fmt.Println("Checking for old users.")
            for k, _ := range users {
             if time.Since(userLastUpdate[k]) > userExpireSeconds {
                fmt.Println("User " + k + " removed. Last updated " + time.Since(userLastUpdate[k]).String() + " ago.")
                delete(users, k)
                delete(userStickies, k)
                delete(userLastUpdate, k)
             }
            }
		}
	}
}
