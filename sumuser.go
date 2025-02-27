package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"
)

const WEEKLY_LOG_FILE string = "data/weekly_logs.csv"

// Index of fields from the weekly logs
const IDX_ID int = 0
const IDX_USER_ID int = 1
const IDX_WEEK int = 2
const IDX_STEPS int = 3
const IDX_TIME_ACTIVE int = 4
const IDX_SLEEP int = 5
const IDX_HEART_RATE int = 6
const IDX_CALORIES_IN int = 7
const IDX_CREATED_AT int = 8
const IDX_UPDATED_AT int = 9

type DatesInWeek struct {
	weekNumber string
	dateTime   [7]time.Time
}

var startWeek DatesInWeek

func GetDateOnly(dateTime time.Time) string {

	y, m, d := dateTime.Date()
	return fmt.Sprintf("%d-%d-%d", y, int(m), d)
}

func StripAllSpaces(originalString string) string {
	return strings.ReplaceAll(originalString, " ", "")
}

func StripCurlyBrackets(originalString string) string {

	z := strings.ReplaceAll(originalString, "{", "")
	return strings.ReplaceAll(z, "}", "")
}

func StripSquareBrackets(originalString string) string {

	x := strings.ReplaceAll(originalString, "[", "")
	return strings.ReplaceAll(x, "]", "")
}

func StripByString(originalString string, stripString string) string {
	return strings.ReplaceAll(originalString, stripString, "")
}

func StripAllBrackets(originalString string) string {

	x := StripCurlyBrackets(originalString)
	return StripSquareBrackets(x)
}

func UpdateHeader(header []string, newHeader string) []string {

	if header == nil {
		log.Fatal("Header cannot be null")
	}

	return append(header, newHeader)
}

func NormalizeRecord(record []string) bool {
	// All this does is log bad data

	v, err := strconv.Atoi(record[IDX_WEEK])

	if err != nil || v < 0 {
		fmt.Println("Invalid Week: " + record[IDX_WEEK])
		return false
	}

	return true
}

// CalculateDateByWeek - Generates a string array of dates representing the week
func CalculateDateByWeek(startWeek *DatesInWeek) bool {

	found := false
	lineCt := 0

	file, err := os.Open(WEEKLY_LOG_FILE)
	if err != nil {
		log.Fatal("Could not open input file", err)
	}

	defer file.Close()
	csvReader := csv.NewReader(file)
	csvReader.TrimLeadingSpace = true

	for {
		record, err := csvReader.Read()
		lineCt++

		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("Input Error at line", lineCt, err)
			continue
		}

		// Normalize record..look for garbage data
		if !NormalizeRecord(record) {
			fmt.Println("Discarding Record at line")
		}

		// Hopefully there is a week 0, with a steps entry somewhere in the file. Fatal if not
		if !found && record[IDX_WEEK] == "0" && len(record[IDX_STEPS]) > 0 {

			startWeek.weekNumber = record[IDX_WEEK]

			// Get a tokenized represenation of the steps record for all 7 days  ["value:0","dateTime:2021-01-16",.......,"value:0","dateTime:2021-01-17",]
			stepTokens := TokenizeSteps(record[IDX_STEPS])

			var ct int = 0
			for i := 1; i < len(stepTokens); i += 2 {

				dateTimeToken := strings.Split(stepTokens[i], ":")

				// Create a Date object out of the string
				startWeek.dateTime[ct], err = time.Parse(time.DateOnly, dateTimeToken[1])
				if err != nil {
					log.Fatal("Could not parse time: ", err.Error())
				}

				fmt.Println(dateTimeToken[1] + " becomes: " + GetDateOnly(startWeek.dateTime[ct]))
			}

			found = true
		}
	}

	return found
}

type UserDailyRecord struct {
	dayRecord [7][]string
}

func WriteWeekSummary(dailyRecord *UserDailyRecord, csvWriter *csv.Writer) {

	for i := range len(dailyRecord.dayRecord) {
		csvWriter.Write(dailyRecord.dayRecord[i])
		csvWriter.Flush()
	}
}

var flatHeaderWritten bool = false

func ProcessRecord(user_id string, record []string, csvWriter *csv.Writer) {
	var dailyRecord UserDailyRecord

	// Initialize a structure that represents all of the dates for this week.
	header := GenerateMapForWeek(record, &dailyRecord)

	activeHeader := SummarizeTimeAcitive(record, &dailyRecord)
	header = slices.Concat(header, activeHeader)

	//SummarizeSteps(currentWeekData, record[IDX_STEPS], &header)
	//SummarizeSleep(currentWeekData, record[IDX_SLEEP])
	//SummarizeHeartRate(currentWeekData, record[IDX_HEART_RATE])
	//SummarizeCalorie(currentWeekData, record[IDX_CALORIES_IN])

	if !flatHeaderWritten {
		flatHeaderWritten = true
		csvWriter.Write(header)
	}

	WriteWeekSummary(&dailyRecord, csvWriter)
}

/*
{"very": {"activities-minutesVeryActive": [{"value": "0", "dateTime": "2021-08-20"}, {"value": "0", "dateTime": "2021-08-21"}, {"value": "5", "dateTime": "2021-08-22"}, {"value": "0", "dateTime": "2021-08-23"}, {"value": "0", "dateTime": "2021-08-24"}, {"value": "11", "dateTime": "2021-08-25"}, {"value": "0", "dateTime": "2021-08-26"}]},
"fairly": {"activities-minutesFairlyActive": [{"value": "0", "dateTime": "2021-08-20"}, {"value": "0", "dateTime": "2021-08-21"}, {"value": "3", "dateTime": "2021-08-22"}, {"value": "0", "dateTime": "2021-08-23"}, {"value": "0", "dateTime": "2021-08-24"}, {"value": "13", "dateTime": "2021-08-25"}, {"value": "0", "dateTime": "2021-08-26"}]},
"lightly": {"activities-minutesLightlyActive": [{"value": "254", "dateTime": "2021-08-20"}, {"value": "294", "dateTime": "2021-08-21"}, {"value": "233", "dateTime": "2021-08-22"}, {"value": "227", "dateTime": "2021-08-23"}, {"value": "165", "dateTime": "2021-08-24"}, {"value": "174", "dateTime": "2021-08-25"}, {"value": "332", "dateTime": "2021-08-26"}]},
 "sedentary": {"activities-minutesSedentary": [{"value": "631", "dateTime": "2021-08-20"}, {"value": "665", "dateTime": "2021-08-21"}, {"value": "701", "dateTime": "2021-08-22"}, {"value": "840", "dateTime": "2021-08-23"}, {"value": "621", "dateTime": "2021-08-24"}, {"value": "470", "dateTime": "2021-08-25"}, {"value": "781", "dateTime": "2021-08-26"}]}}

"activities-minutesLightlyActive:[value:19,dateTime:2021-01-22,value:209,dateTime:2021-01-23,value:188,dateTime:2021-01-24,value:232,dateTime:2021-01-25,value:315,dateTime:2021-01-26,value:212,dateTime:2021-01-27,value:218,dateTime:2021-01-28],sedentary:activities-minutesSedentary:[value:905,dateTime:2021-01-22,value:1173,dateTime:2021-01-23,value:731,dateTime:2021-01-24,value:550,dateTime:2021-01-25,value:512,dateTime:2021-01-26,value:455,dateTime:2021-01-27,value:583,dateTime:2021-01-28]"


*/

func TokenizeActivity(activityRecord string) {
	s := StripCurlyBrackets(activityRecord)
	s = StripAllSpaces(s)
	s = StripByString(s, "\"")
	s = StripByString(s, "very:")

	veryActive := strings.Split(s, ",fairly:")
	fairlyActive := strings.Split(veryActive[1], ",lightly:")
	lightlyActive := strings.Split(fairlyActive[1], ",sedentary:")

	/*fmt.Println("=====Very Active=====\n" + veryActive[0])
	fmt.Println("=====Fairly Active=====\n" + fairlyActive[0])
	fmt.Println("=====Lightly Active=====\n" + lightlyActive[0])
	fmt.Println("=====Sedentary=====\n" + lightlyActive[1]) */

	veryActiveEntries := strings.Split(veryActive[0], ":[")
	fairlyActiveEntries := strings.Split(fairlyActive[0], ":[")
	lightlyActiveEntries := strings.Split(lightlyActive[0], ":[")
	sedentaryActiveEntries := strings.Split(lightlyActive[1], ":[")

	fmt.Println("----Very Active --- ", veryActiveEntries[1])
	fmt.Println("----Fairly Active --- ", fairlyActiveEntries[1])
	fmt.Println("----Lightly Active --- ", lightlyActiveEntries[1])
	fmt.Println("----Sedentary --- ", sedentaryActiveEntries[1])

}

func SummarizeTimeAcitive(record []string, userDailyRecord *UserDailyRecord) []string {

	if len(record[IDX_STEPS]) == 0 {
		return nil
	}

	header := []string{"activities-minutesVeryActive", "activities-minutesFairlyActive", "activities-minutesLightlyActive", "activities-minutesSedentary"}

	TokenizeActivity(record[IDX_TIME_ACTIVE])
	//fmt.Println("==========Active Record============\n" + record[IDX_TIME_ACTIVE])

	return header
}

func TokenizeSteps(stepsRecord string) []string {

	// Strip the record of all brackets then tokenize so we can get to the 7 dates that make up the week
	s := StripAllBrackets(stepsRecord)
	s = StripAllSpaces(s)
	s = StripByString(s, "\"")
	fmt.Println(s)
	stepTokens := strings.Split(s, ",")
	return stepTokens
}

func GenerateMapForWeek(record []string, dailyRecord *UserDailyRecord) []string {

	if len(record[IDX_STEPS]) == 0 {
		panic("not done")
	}

	header := []string{"Date", "user_id", "id", "Week", "Total Steps"}

	// We do the steps in this function
	// Get a tokenized represenation of the steps record for all 7 days  ["value:0","dateTime:2021-01-16",.......,"value:0","dateTime:2021-01-17"]
	stepTokens := TokenizeSteps(record[IDX_STEPS])

	for i, j := 1, 0; i < len(stepTokens); i, j = i+2, j+1 {
		dateTimeToken := strings.Split(stepTokens[i], ":")
		stepsValueToken := strings.Split(stepTokens[i-1], ":")

		// Create a Date object out of the string
		tmpTime, err := time.Parse(time.DateOnly, dateTimeToken[1])
		if err != nil {
			log.Fatal("Could not parse time: ", err.Error())
		}

		dailyRecord.dayRecord[j] = make([]string, 5)
		dailyRecord.dayRecord[j][0] = GetDateOnly(tmpTime)
		dailyRecord.dayRecord[j][1] = record[IDX_USER_ID]
		dailyRecord.dayRecord[j][2] = record[IDX_ID]
		dailyRecord.dayRecord[j][3] = record[IDX_WEEK]
		dailyRecord.dayRecord[j][4] = stepsValueToken[1]

		fmt.Println(dateTimeToken[1] + " becomes: " + dailyRecord.dayRecord[j][0])
	}

	return header
}

func main() {

	// First scan the file to initialize our week and date range
	if !CalculateDateByWeek(&startWeek) {
		// We couldn't figure it out
		log.Fatal("Could not establish week and date ranges. ")
	}

	weeklyLogFile, err := os.Open(WEEKLY_LOG_FILE)
	if err != nil {
		log.Fatal("Could not open input file", err)
	}

	defer weeklyLogFile.Close()

	csvReader := csv.NewReader(weeklyLogFile)
	csvReader.TrimLeadingSpace = true

	// Get user_id from user
	user_id := "55"
	//fmt.Print("Enter user_id: ")
	//fmt.Scan(&user_id)

	masterUserOutFile := "output/user_" + user_id + ".csv"
	masterf, err := os.Create(masterUserOutFile)
	if err != nil {
		log.Fatal("Could not create output file: " + masterUserOutFile)
	}

	defer masterf.Close()
	csvWriter := csv.NewWriter(masterf)
	csvWriter.UseCRLF = false

	flatUserOutFile := "output/flat_user_" + user_id + ".csv"
	flatf, err := os.Create(flatUserOutFile)
	if err != nil {
		log.Fatal("Could not create output file: " + flatUserOutFile)
	}

	defer flatf.Close()
	csvFlatWriter := csv.NewWriter(flatf)
	csvFlatWriter.UseCRLF = false

	header := true
	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("Error:", err)
			continue
		}

		if header {
			csvWriter.Write(record)
			header = false
		} else if record[IDX_USER_ID] == user_id {

			// Save the user specific record to the master file...might be useful output
			csvWriter.Write(record)

			// Process record for the flatten per user CSV
			ProcessRecord(user_id, record, csvFlatWriter)
		}
	}

	flatf.Close()
	masterf.Close()
}
