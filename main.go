package main

import (
	"os"
	"time"
	"fmt"
	"strconv"
	"net/http"
	"encoding/json"
	"github.com/carlescere/scheduler"
	"github.com/cloudfoundry-community/go-cfclient"
)

var ApiAddress			string
var Username			string
var Password			string
var MinutesInPastToQuery	int
var MinutesBetweenQueries	int
var NumberOfCrashesBeforeKill	int
var SkipSslValidation		bool
var DryRun			bool

type response struct{
	Message string
	Code uint8
}

func init() {
	fail := false
	if "" == os.Getenv("ApiAddress") {
		fmt.Fprintln(os.Stderr, "Environmental variable 'ApiAddress' not set")
		fail = true
	} else {
		ApiAddress = os.Getenv("ApiAddress")
	}
	if "" == os.Getenv("Username") {
		fmt.Fprintln(os.Stderr, "Environmental variable 'Username' not set")
		fail = true
	} else {
		Username = os.Getenv("Username")
	}
	if "" == os.Getenv("Password") {
		fmt.Fprintln(os.Stderr, "Environmental variable 'Password' not set")
		fail = true
	} else {
		Password = os.Getenv("Password")
	}
	var err error
	MinutesBetweenQueries, err = strconv.Atoi(os.Getenv("MinutesBetweenQueries"))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Can not parse environmental variable 'MinutesBetweenQueries' to an int")
		fail = true
	}
	MinutesInPastToQuery, err = strconv.Atoi(os.Getenv("MinutesInPastToQuery"))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Can not parse environmental variable 'MinutesInPastToQuery' to an int")
		fail = true
	}
	NumberOfCrashesBeforeKill, err = strconv.Atoi(os.Getenv("NumberOfCrashesBeforeKill"))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Can not parse environmental variable 'NumberOfCrashesBeforeKill' to an int")
		fail = true
	}
	SkipSslValidation, err = strconv.ParseBool(os.Getenv("SkipSslValidation"))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Can not parse environmental variable 'SkipSslValidation' to a boolean")
		fail = true
	}
	DryRun, err = strconv.ParseBool(os.Getenv("DryRun"))
	if err != nil {
		DryRun = false
	}
	if fail {
		os.Exit(1)
	}
}
func task() {
	c := &cfclient.Config {
		ApiAddress:		ApiAddress,
		Username:		Username,
		Password:		Password,
		SkipSslValidation:	SkipSslValidation,
	}
	client, err := cfclient.NewClient(c)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	var AppEventQuerys = []cfclient.AppEventQuery {
		cfclient.AppEventQuery {
			Filter: "timestamp",
			Operator: ">=",
			Value: time.Now().Add(time.Duration(-MinutesInPastToQuery) * time.Minute).UTC().Format(time.RFC3339),
		},
	}
	aeea, err := client.ListAppEventsByQuery("app.crash", AppEventQuerys)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(3)
	}
	var crashes = map[string]int{}
	for _, i := range aeea {
		crashes[i.Actor]++
	}
	for appGuid, crashCount := range crashes {
		a, _ := client.GetAppByGuid(appGuid)
		if a.State == "STARTED" {
			s, _ := client.GetSpaceByGuid(a.SpaceGuid)
			o, _ := client.GetOrgByGuid(a.SpaceData.Entity.OrganizationGuid)
			fmt.Printf("Org: %v, space: %v, app: %v, has crashed %v times in the past %v minutes ", o.Name, s.Name, a.Name, crashCount, MinutesInPastToQuery)
			if crashCount >= NumberOfCrashesBeforeKill {
				if !DryRun {
					aur := cfclient.AppUpdateResource{State: "STOPPED"}
					_, err := client.UpdateApp(appGuid, aur)
					if err != nil {
						fmt.Fprintln(os.Stderr, err)
					}
					fmt.Println("...killing!")
				} else {
					fmt.Println("...Lucky this is a dry run")
				}
			} else {
				fmt.Println("letting this app live.... for now")
			}
		}
	}
}
func rootHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response{Message: "Jason Voorhees", Code: 200})
}
func main() {
	job := func() { task() }
	port := os.Getenv("PORT")
	if len(port) < 1 {
		port = "8080"
	}
	scheduler.Every(MinutesBetweenQueries).Minutes().Run(job)
	http.HandleFunc("/", rootHandler)
	http.ListenAndServe(":"+port, nil)
}
