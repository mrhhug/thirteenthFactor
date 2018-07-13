package main

import (
	"os"
	"time"
	"fmt"
	"strings"
	"strconv"
	"net/http"
	"encoding/json"
	"github.com/robfig/cron"
	"github.com/cloudfoundry-community/go-cfclient"
)

var ApiAddress					string
var Username					string
var Password					string
var MinutesInPastToQuery			int
var CronString					string
var NumberOfCrashesBeforeKill			int
var DelayedActionCronString			string
var DelayedActionMinutesInPastToQuery		int
var DelayedActionNumberOfCrashesBeforeKill	int
var DelayedActionOrgsGuid			[]string
var DelayedActionSpacesGuid			[]string
var DelayedActionAppsGuid			[]string
var SkipSslValidation				bool
var DryRun					bool

type response struct{
	ApiAddress				string
	Username				string
	MinutesInPastToQuery			int
	CronString				string
	NumberOfCrashesBeforeKill		int
	DelayedActionCronString			string
	DelayedActionMinutesInPastToQuery	int
	DelayedActionNumberOfCrashesBeforeKill	int
	DelayedActionOrgsGuid			[]string
	DelayedActionSpacesGuid			[]string
	DelayedActionAppsGuid			[]string
	SkipSslValidation			bool
	DryRun					bool

}
func rootHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response{
		ApiAddress: ApiAddress,
		Username: Username,
		MinutesInPastToQuery: MinutesInPastToQuery,
		CronString: CronString,
		NumberOfCrashesBeforeKill: NumberOfCrashesBeforeKill,
		DelayedActionCronString: DelayedActionCronString,
		DelayedActionNumberOfCrashesBeforeKill: DelayedActionNumberOfCrashesBeforeKill,
		DelayedActionMinutesInPastToQuery: DelayedActionMinutesInPastToQuery,
		DelayedActionOrgsGuid: DelayedActionOrgsGuid,
		DelayedActionSpacesGuid: DelayedActionSpacesGuid,
		DelayedActionAppsGuid: DelayedActionAppsGuid,
		SkipSslValidation: SkipSslValidation,
		DryRun: DryRun,
	})
}

func init() {
	fail := false
	if "" == os.Getenv("ApiAddress") {
		fmt.Fprintln(os.Stderr, "Environmental variable 'ApiAddress' not set")
		fail = true
	} else {
		ApiAddress = os.Getenv("ApiAddress")
	}
	if "" == os.Getenv("CFServiceAccountUsername") {
		fmt.Fprintln(os.Stderr, "Environmental variable 'CFServiceAccountUsername' not set")
		fail = true
	} else {
		Username = os.Getenv("CFServiceAccountUsername")
	}
	if "" == os.Getenv("CFServiceAccountPassword") {
		fmt.Fprintln(os.Stderr, "Environmental variable 'CFServiceAccountPassword' not set")
		fail = true
	} else {
		Password = os.Getenv("CFServiceAccountPassword")
	}
	var err error
	if "" == os.Getenv("CronString") {
		fmt.Fprintln(os.Stderr, "Environmental variable 'CronString' not set")
		fail = true
	} else {
		CronString = os.Getenv("CronString")
	}
	DelayedActionCronString = os.Getenv("DelayedActionCronString")
	MinutesInPastToQuery, err = strconv.Atoi(os.Getenv("MinutesInPastToQuery"))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Can not parse environmental variable 'MinutesInPastToQuery' to an int")
		fail = true
	}
	DelayedActionMinutesInPastToQuery, err = strconv.Atoi(os.Getenv("DelayedActionMinutesInPastToQuery"))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Can not parse environmental variable 'DelayedActionMinutesInPastToQuery' to an int")
	}
	DelayedActionNumberOfCrashesBeforeKill, err = strconv.Atoi(os.Getenv("DelayedActionNumberOfCrashesBeforeKill"))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Can not parse environmental variable 'DelayedActionNumberOfCrashesBeforeKill' to an int")
	}
	NumberOfCrashesBeforeKill, err = strconv.Atoi(os.Getenv("NumberOfCrashesBeforeKill"))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Can not parse environmental variable 'NumberOfCrashesBeforeKill' to an int")
		fail = true
	}
	SkipSslValidation, err = strconv.ParseBool(os.Getenv("SkipSslValidation"))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Can not parse environmental variable 'SkipSslValidation' to a boolean")
	}
	DryRun, err = strconv.ParseBool(os.Getenv("DryRun"))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Can not parse environmental variable 'DryRun' to a boolean")
	}
	DelayedActionOrgsGuid = strings.Split(strings.Replace(os.Getenv("DelayedActionOrgsGuid"), " ", "", -1), ",")
	DelayedActionSpacesGuid = strings.Split(strings.Replace(os.Getenv("DelayedActionSpacesGuid"), " ", "", -1), ",")
	DelayedActionAppsGuid = strings.Split(strings.Replace(os.Getenv("DelayedActionAppsGuid"), " ", "", -1), ",")
	if fail {
		os.Exit(1)
	}
}
func stringInSlice(a string, list []string) bool {
        for _, b := range list {
                if b == a {
                        return true
                }
        }
	return false
}
func task(banner string,
		localNumberOfCrashesBeforeKill int,
		localMinutesInPastToQuery int,
		localDelayedActionOrgsGuid []string,
		localDelayedActionSpacesGuid []string,
		localDelayedActionAppsGuid []string) {
	fmt.Println(banner)
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
			Value: time.Now().Add(time.Duration(-localMinutesInPastToQuery) * time.Minute).UTC().Format(time.RFC3339),
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
			if crashCount >= localNumberOfCrashesBeforeKill {
				if !DryRun {
					//kinda had to do guids since app names and space names can be reused
					if stringInSlice(appGuid, localDelayedActionOrgsGuid) ||
							stringInSlice(a.SpaceGuid, localDelayedActionSpacesGuid) ||
							stringInSlice(a.SpaceData.Entity.OrganizationGuid, DelayedActionOrgsGuid) {
						fmt.Println("... is in a DelayedAction org, space, or app. ")
					} else {
						aur := cfclient.AppUpdateResource{State: "STOPPED"}
						_, err := client.UpdateApp(appGuid, aur)
						if err != nil {
							fmt.Fprintln(os.Stderr, err)
						}
						fmt.Println("...killing!")
					}
				} else {
					fmt.Println("...Lucky this is a dry run")
				}
			} else {
				fmt.Println("letting this app live.... for now")
			}
		}
	}
}

func typicalTask() {
	task("Regularly Scheduled task", NumberOfCrashesBeforeKill, MinutesInPastToQuery, DelayedActionOrgsGuid, DelayedActionSpacesGuid, DelayedActionAppsGuid)
}
func delayedAction() {
	task("DelayedAction task", DelayedActionNumberOfCrashesBeforeKill, DelayedActionMinutesInPastToQuery, []string{}, []string{}, []string{})
}
func main() {
	port := os.Getenv("PORT")
	if len(port) < 1 {
		port = "8080"
	}
	c := cron.New()
	c.AddFunc(CronString, func() {typicalTask()})
	c.AddFunc(DelayedActionCronString, func() {delayedAction()})
	c.Start()
	http.HandleFunc("/", rootHandler)
	http.ListenAndServe(":"+port, nil)
}
