package main

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sort"
	"sync"
	"time"
)

var startTime = time.Now()

//Time in seconds for user registration and notes to expire if no activity (updates) occurs.
var userExpirySeconds = 60 * 60 * 24

var redisPool *redis.Pool
var maxConnections = 10
var maxIdleConnections = 2

var activeUserSetKey = "activeUsers"
var inactiveUsersSetKey = "inactiveUsers"

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
	fmt.Fprintf(w, fmt.Sprintf("User expiry duration (with no activity from Stickify Pusher) is %v seconds", userExpirySeconds))
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "https://github.com/stickify/", 301)

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

type UserAndScore struct {
	User string
	Score int
}

type UserAndScoreSlice []UserAndScore

func (slice UserAndScoreSlice) Len() int {
	return len(slice)
}

func (slice UserAndScoreSlice) Less(i, j int) bool {
	return slice[i].Score < slice[j].Score
}

func (slice UserAndScoreSlice) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

func createLeaderboard() string {
	var returnString = "Stats/Leaderboard\n"
	returnString += "---------------------\n"
	returnString += "Nickname\tStickies\n"
	c := redisPool.Get()
	defer c.Close()
	activeUsersSet, err := redis.Strings(c.Do("SMEMBERS", activeUserSetKey))
	if err != nil {
		fmt.Printf("SMEMBERS error: %v\n", err.Error())
	}
	
	var activeUserAndScore = make(UserAndScoreSlice, 0)

	for _, user := range activeUsersSet {
		userExists := checkUserExist(user, c)
		if userExists == true {
			userNoteCount, err := redis.Int(c.Do("LLEN", fmt.Sprintf("%v:notes", user)))
			if err != nil {
				fmt.Printf("LLEN error: %v\n", err.Error())
			}
			
			activeUserAndScore = append(activeUserAndScore, UserAndScore{user, userNoteCount})
		} else {
			moveUserFromActiveSet, err := redis.Int(c.Do("SMOVE", activeUserSetKey, inactiveUsersSetKey, user))
			if err != nil {
				fmt.Printf("SMOVE error: %v\n", err.Error())
			}
			if moveUserFromActiveSet != 1 {
				fmt.Printf("SMOVE returned: %v when moving from user from active set to inactive set\n", moveUserFromActiveSet)
			}
		}
	}

	sort.Sort(sort.Reverse(activeUserAndScore))

	for _, pair := range activeUserAndScore {
		user := pair.User[:1]
		user += "***"
		returnString += fmt.Sprintf("%v\t%v\n", user, pair.Score)
	}

	returnString += fmt.Sprintf("Active Users:\t%v\n", len(activeUserAndScore))

	return returnString
}

func uptimeHandler(w http.ResponseWriter, r *http.Request) {
	//bypass same origin policy
	w.Header().Set("Access-Control-Allow-Origin", "*")

	diff := time.Since(startTime)

	fmt.Fprintf(w, fmt.Sprintf("Uptime:\t%v\n",diff.String()))

	fmt.Fprintf(w, createLeaderboard())

	fmt.Println("Uptime requested")
}

func checkUserExist(user string, c redis.Conn) bool {
	userExists, err := redis.Int(c.Do("EXISTS", user))
	if err != nil {
		fmt.Printf("EXISTS error: %v\n", err.Error())
		return false
	}

	if userExists == 0 {
		return false
	} else if userExists == 1 {
		return true
	} else {
		fmt.Printf("Unknown return value of %v for EXISTS for User '%v'.\n", userExists)
		return false
	}
}

func checkUserPasscode(user string, passcode string, c redis.Conn) bool {
	getUserPasscodeDigestResult, err := redis.String(c.Do("GET", user))
	if err != nil {
		fmt.Printf("GET error: %v", err.Error())
		return false
	}

	passcodeDigest := fmt.Sprintf("%x", md5.Sum([]byte(passcode)))

	if getUserPasscodeDigestResult == passcodeDigest { //Passcode digests match
		return true
	} else { //Passcode digests do not match
		fmt.Printf("User '%v' passcode digests do not match.\n", user)
		return false
	}
}

func getUserNotes(user string, c redis.Conn) ([][]string, error) {
	getUserNoteKeysResult, err := redis.Strings(c.Do("LRANGE", fmt.Sprintf("%v:notes", user), 0, -1))
	if err != nil {
		fmt.Printf("LRANGE error: %v", err.Error())
		return nil, err
	}

	//fmt.Printf("Note titles: %v\n",getUserNoteKeysResult)

	var userNotes [][]string
	userNotes = make([][]string, 0)

	for _, noteKey := range getUserNoteKeysResult {
		getUserNoteResult, err := redis.Strings(c.Do("LRANGE", noteKey, 0, -1))
		if err != nil {
			fmt.Printf("LRANGE error: %v", err.Error())
			return nil, err
		}
		userNotes = append(userNotes, getUserNoteResult)
	}

	return userNotes, nil
}

func getUser(user string, passcode string) string {
	//check for blank user
	if len(user) == 0 {
		fmt.Println("Blank user, sending back error.")
		return "1 Please supply a nickname."
	}

	c := redisPool.Get()
	defer c.Close()
	userExists := checkUserExist(user, c)

	if userExists == false { //User does not exist. Tell client that user does not exist.
		fmt.Printf("User '%v' does not exist\n", user)

		return "1 Nickname not found."
	} else { //User exists. Check passcode.
		fmt.Printf("User '%v' exists\n", user)

		authenticate := checkUserPasscode(user, passcode, c)

		if authenticate == true { //Passcode digests match
			//send stickies in json
			userNotes, err := getUserNotes(user, c)
			if err != nil {
				fmt.Println(err.Error())
				return err.Error()
			}
			output, err := json.Marshal(userNotes)
			if err != nil {
				fmt.Println(err.Error())
				return err.Error()
			}
			return string(output)
		} else { //Passcode digests do not match
			fmt.Printf("getUser - User '%v' passcode digests do not match.\n", user)
			return "1 Wrong PIN"
		}
	}

	return "1 Something went wrong."
}

func getUserHandler(w http.ResponseWriter, r *http.Request) {
	//bypass same origin policy
	w.Header().Set("Access-Control-Allow-Origin", "*")

	r.ParseForm()

	var user string
	var passcode string

	if len(r.Form["user"]) > 0 && len(r.Form["passcode"]) > 0 {
		user = r.Form["user"][0]
		passcode = r.Form["passcode"][0]

		fmt.Fprintf(w, getUser(user, passcode))

	} else {
		fmt.Fprintf(w, "1 Missing nickname/PIN")
	}
}

func updateStickies(user string, passcode string, stickyNumber int, stickyData string) string {
	//check for blank user
	if len(user) == 0 {
		fmt.Println("Blank user, sending back error.")
		return "1 Please supply a nickname."
	}

	passcodeDigest := fmt.Sprintf("%x", md5.Sum([]byte(passcode)))

	c := redisPool.Get()
	defer c.Close()
	userExists := checkUserExist(user, c)

	if userExists == false { //User does not exist. Create new user.
		fmt.Printf("User '%v' does not exist. Creating new user.\n", user)

		//Create new user.
		//SET user digest key/value to passcode digest to create new user.
		//EXPIRE the key in 24 hours
		setUserResult, err := redis.String(c.Do("SETEX", user, userExpirySeconds, passcodeDigest))
		if err != nil {
			fmt.Printf("SET error: %v\n", err.Error())
			return err.Error()
		}

		if setUserResult == "OK" {
			fmt.Printf("User '%v' created.\n", user)
		} else {
			fmt.Printf("SET result was '%v' when SET user '%v'.\n")
		}
		
		//SADD user to active user set
		addUserToSetResult, err := redis.Int(c.Do("SADD", activeUserSetKey, user))
		if err != nil {
			fmt.Printf("SADD error: %v\n", err.Error())
			return err.Error()
		}

		if addUserToSetResult == 1 {
			fmt.Printf("User '%v' added to active user set.\n", user)
		} else {
			fmt.Printf("SADD result was '%v' when SADD user to active user set '%v'.\n")
		}


	} else { //User exists. Check passcode.
		fmt.Printf("User '%v' exists\n", user)

		authenticate := checkUserPasscode(user, passcode, c)

		if authenticate == true {
			//reset EXPIRE for 24 hours for the user
			_, err := redis.Int(c.Do("EXPIRE", user, userExpirySeconds))
			if err != nil {
				fmt.Printf("EXPIRE user error: %v\n", err.Error())
				return err.Error()
			}

		} else { //Passcode digests do not match
			fmt.Printf("User '%v' passcode digests do not match.\n", user)
			return "1 Wrong PIN. Nickname may be used by someone else. Choose a new nickname. Nicknames expire after 24 hrs of no activity."
		}
	}

	//User now exists and passcode is correct. Time to update sticky notes

	//EXPIRE the user key in 24 hours
	_, err := redis.String(c.Do("SETEX", user, userExpirySeconds, passcodeDigest))
	if err != nil {
		fmt.Printf("SETEX error: %v", err.Error())
		return err.Error()
	}

	if stickyNumber == 0 {
		/*
			First DEL entire list.
			Next RPUSH zeroth note onto list to create it.

			DEL the note title key
			for each note line, RPUSH to list at note title key
		*/

		_, err := redis.Int(c.Do("DEL", fmt.Sprintf("%v:notes", user)))
		if err != nil {
			fmt.Printf("DEL error: %v", err.Error())
			return err.Error()
		}

		_, err = redis.Int(c.Do("RPUSH", fmt.Sprintf("%v:notes", user), fmt.Sprintf("%v:notes:%v", user, stickyNumber)))
		if err != nil {
			fmt.Printf("RPUSH error: %v", err.Error())
			return err.Error()
		}

		_, err = redis.Int(c.Do("DEL", fmt.Sprintf("%v:notes:%v", user, stickyNumber)))
		if err != nil {
			fmt.Printf("DEL error: %v", err.Error())
			return err.Error()
		}

		var args []interface{}
		args = append(args, fmt.Sprintf("%v:notes:%v", user, stickyNumber))
		noteLines := parseLines(stickyData)
		for _, line := range noteLines {
			args = append(args, line)
		}
		_, err = c.Do("RPUSH", args...)
		if err != nil {
			fmt.Printf("RPUSH error: %v", err.Error())
			return err.Error()
		}

	} else {
		/*
			LLEN to get length of list
			if sufficient space, LSET
			else
			RPUSH the difference between the note index and LLEN
			RPUSH the note title

			DEL the note title key
			for each note line, RPUSH to list at note title key
		*/

		userNotesLen, err := redis.Int(c.Do("LLEN", fmt.Sprintf("%v:notes", user)))
		if err != nil {
			fmt.Printf("LLEN error: %v", err.Error())
			return err.Error()
		}

		if userNotesLen < stickyNumber+1 {
			var i int
			for i = stickyNumber - userNotesLen; i > 0; i-- {
				_, err := redis.Int(c.Do("RPUSH", fmt.Sprintf("%v:notes", user), ""))
				if err != nil {
					fmt.Printf("RPUSH error: %v", err.Error())
					return err.Error()
				}
			}

			_, err := redis.Int(c.Do("RPUSH", fmt.Sprintf("%v:notes", user), fmt.Sprintf("%v:notes:%v", user, stickyNumber)))
			if err != nil {
				fmt.Printf("RPUSH error: %v", err.Error())
				return err.Error()
			}
		} else { //sufficient space
			setUserNoteTitleResult, err := redis.String(c.Do("LSET", fmt.Sprintf("%v:notes", user), stickyNumber, fmt.Sprintf("%v:notes:%v", user, stickyNumber)))
			if err != nil {
				fmt.Printf("LSET error: %v", err.Error())
				return err.Error()
			}

			if setUserNoteTitleResult != "OK" {
				fmt.Printf("LSET user note title got result: %v", setUserNoteTitleResult)
				return setUserNoteTitleResult
			}
		}

		_, err = redis.Int(c.Do("DEL", fmt.Sprintf("%v:notes:%v", user, stickyNumber)))
		if err != nil {
			fmt.Printf("DEL error: %v", err.Error())
			return err.Error()
		}

		noteLines := parseLines(stickyData)
		for _, line := range noteLines {
			c.Send("RPUSH", fmt.Sprintf("%v:notes:%v", user, stickyNumber), line)
		}
		_, err = c.Do("")
		if err != nil {
			fmt.Printf("RPUSH error: %v", err.Error())
			return err.Error()
		}
	}

	//reset EXPIRE for 24 hours for the user notes title
	_, err = redis.Int(c.Do("EXPIRE", fmt.Sprintf("%v:notes", user), userExpirySeconds))
	if err != nil {
		fmt.Printf("EXPIRE user notes title error: %v", err.Error())
		return err.Error()
	}
	//reset EXPIRE for 24 hours for the user notes
	_, err = redis.Int(c.Do("EXPIRE", fmt.Sprintf("%v:notes:%v", user, stickyNumber), userExpirySeconds))
	if err != nil {
		fmt.Printf("EXPIRE user notes error: %v", err.Error())
		return err.Error()
	}

	return "0"
}

func updateHandler(w http.ResponseWriter, r *http.Request) {
	//bypass same origin policy
	w.Header().Set("Access-Control-Allow-Origin", "*")

	r.ParseForm()

	var user string
	var passcode string
	var stickyNumber int
	var stickyData string
	var err error

	if len(r.Form["user"]) > 0 && len(r.Form["passcode"]) > 0 && len(r.Form["number"]) > 0 && len(r.Form["data"]) > 0 {
		user = r.Form["user"][0]
		passcode = r.Form["passcode"][0]
		stickyNumber, err = strconv.Atoi(r.Form["number"][0])
		if err != nil {
			fmt.Fprintf(w, err.Error())
		}
		stickyData = r.Form["data"][0]

		fmt.Fprintf(w, updateStickies(user, passcode, stickyNumber, stickyData))

	} else {
		fmt.Fprintf(w, "Missing required parameters 'user' 'passcode' 'number' 'data'.")
	}
}

func server(wg *sync.WaitGroup) {
	//Called by stickify pusher client
	http.HandleFunc("/update", updateHandler)
	//Called by stickify viewer client
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

	wg.Done()
}

func createRedisPool() *redis.Pool {
	pool := redis.NewPool(func() (redis.Conn, error) {
		c, err := redis.DialURL(os.Getenv("REDIS_URL"))

		if err != nil {
			log.Println(err)
			return nil, err
		}

		return c, err
	}, maxIdleConnections)
	pool.TestOnBorrow = func(c redis.Conn, t time.Time) error {
        if time.Since(t) < time.Minute {
            return nil
        }
        _, err := c.Do("PING")
        return err
    }

	pool.MaxActive = maxConnections
	pool.IdleTimeout = time.Second * 10
	return pool
}

func main() {
	//Setup redis connection pool
	redisPool = createRedisPool()

	/*
	//Test sequence for sample note with three lines
	fmt.Println(updateStickies("test", "123", 0, "c29tZSBub3RlcyBtYXRlcmlhbA0KZmFkc2Zhc2QNCnZjY3p2eHg="))
	fmt.Println(updateStickies("anson", "123", 0, "c29tZSBub3RlcyBtYXRlcmlhbA0KZmFkc2Zhc2QNCnZjY3p2eHg="))
	fmt.Println(updateStickies("anson", "123", 1, "c29tZSBub3RlcyBtYXRlcmlhbA0KZmFkc2Zhc2QNCnZjY3p2eHg="))
	//fmt.Println(getUser("test", "123"))
	c := redisPool.Get()
	userNotes, err := getUserNotes("test", c)
	c.Close()
	if err != nil {
		fmt.Println(err.Error())
	}
	output, err := json.Marshal(userNotes)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Printf("Error: %v Notes: %v\n", err, string(output))
	*/

	//start server and wait
	var wg sync.WaitGroup
	wg.Add(1)
	go server(&wg)
	wg.Wait()

	redisPool.Close()
}
