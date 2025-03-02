package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"
)

var OUTPUT_ROOT string = "./"
var OUTPUT_SLEEP_DIR string = "sleep"

const OUTPUT_DIR string = "output"
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

			return true
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

	activeHeader, err := SummarizeTimeActive(record, &dailyRecord)
	if err != nil {
		fmt.Println("Ignoring record ", err.Error())
		return
	}

	sleepHeader, err := SummarizeSleep(record, &dailyRecord)
	if err != nil {
		fmt.Println("Ignoring record: ", err.Error())
	}

	if !flatHeaderWritten {
		header = slices.Concat(header, activeHeader, sleepHeader)
		flatHeaderWritten = true
		csvWriter.Write(header)
	}

	WriteWeekSummary(&dailyRecord, csvWriter)
}

func SummarizeSleep(record []string, userDailyRecord *UserDailyRecord) ([]string, error) {

	if len(record[IDX_SLEEP]) == 0 {
		return nil, nil
	}

	sleepRecord := StripAllSpaces(record[IDX_SLEEP])
	sleepRecord = StripByString(sleepRecord, "\\")
	sleepRecord = StripByString(sleepRecord, "\"")

	file, err := os.Create(OUTPUT_SLEEP_DIR + "sleep_" + record[IDX_USER_ID] + "_week_" + record[IDX_WEEK] + ".txt")
	if err != nil {
		log.Fatal("sleep output file", err)
	}

	defer file.Close()

	fmt.Println(sleepRecord)
	file.WriteString(sleepRecord)
	file.Close()
	return nil, nil
}

func TokenizeActivity(activityRecord string) ([]string, []string, []string, []string, error) {

	// Do some basic normalization
	s := StripCurlyBrackets(activityRecord)
	s = StripAllSpaces(s)
	s = StripByString(s, "\"")
	s = StripByString(s, "very:")

	// Split up entry by activity type
	veryActive := strings.Split(s, ",fairly:")
	fairlyActive := strings.Split(veryActive[1], ",lightly:")
	lightlyActive := strings.Split(fairlyActive[1], ",sedentary:")

	if len(veryActive) != 2 || len(fairlyActive) != 2 || len(lightlyActive) != 2 {
		return nil, nil, nil, nil, errors.New("malformed activity record")
	}

	/*fmt.Println("=====Very Active=====\n" + veryActive[0])
	fmt.Println("=====Fairly Active=====\n" + fairlyActive[0])
	fmt.Println("=====Lightly Active=====\n" + lightlyActive[0])
	fmt.Println("=====Sedentary=====\n" + lightlyActive[1]) */

	// Isolate value:date by activity type
	veryActiveEntries := strings.Split(veryActive[0], ":[")
	fairlyActiveEntries := strings.Split(fairlyActive[0], ":[")
	lightlyActiveEntries := strings.Split(lightlyActive[0], ":[")
	sedentaryActiveEntries := strings.Split(lightlyActive[1], ":[")

	if len(veryActiveEntries) != 2 || len(fairlyActiveEntries) != 2 || len(lightlyActiveEntries) != 2 || len(sedentaryActiveEntries) != 2 {
		return nil, nil, nil, nil, errors.New("malformed activity section")
	}

	/*fmt.Println("----Very Active --- ", veryActiveEntries[1])
	fmt.Println("----Fairly Active --- ", fairlyActiveEntries[1])
	fmt.Println("----Lightly Active --- ", lightlyActiveEntries[1])
	fmt.Println("----Sedentary --- ", sedentaryActiveEntries[1]) */

	// Flatten into tokens
	veryActiveTokens := strings.Split(veryActiveEntries[1], ",")
	fairlyActiveTokens := strings.Split(fairlyActiveEntries[1], ",")
	lightlyActiveTokens := strings.Split(lightlyActiveEntries[1], ",")
	sedentaryTokens := strings.Split(sedentaryActiveEntries[1], ",")

	if len(veryActiveTokens) != 14 || len(fairlyActiveTokens) != 14 || len(lightlyActiveTokens) != 14 || len(sedentaryTokens) != 14 {
		return nil, nil, nil, nil, errors.New("malformed activity token")
	}

	return veryActiveTokens, fairlyActiveTokens, lightlyActiveTokens, sedentaryTokens, nil

}

func SummarizeTimeActive(record []string, userDailyRecord *UserDailyRecord) ([]string, error) {

	if len(record[IDX_STEPS]) == 0 {
		return nil, nil
	}

	header := []string{"activities-minutesVeryActive", "activities-minutesFairlyActive", "activities-minutesLightlyActive", "activities-minutesSedentary"}

	veryActiveTokens, fairlyActiveTokens, lightlyActiveTokens, sedentaryTokens, err := TokenizeActivity(record[IDX_TIME_ACTIVE])
	if err != nil {
		return nil, err
	}

	EmitEntryFromWeeklyTokens(userDailyRecord, veryActiveTokens, fairlyActiveTokens, lightlyActiveTokens, sedentaryTokens)
	return header, nil
}

func EmitValueDateFromTokenSlice(tokenSlice []string) ([]string, []string) {

	dateSlice := make([]string, 7)
	valSlice := make([]string, 7)

	for i, j := 1, 0; i < len(tokenSlice); i, j = i+2, j+1 {
		dateToken := strings.Split(tokenSlice[i], ":")
		valueToken := strings.Split(tokenSlice[i-1], ":")

		// Create a Date object out of the string
		tmpTime, err := time.Parse(time.DateOnly, strings.Trim(dateToken[1], "]"))
		if err != nil {
			log.Fatal("Could not parse time: ", err.Error())
		}

		dateSlice[j] = GetDateOnly(tmpTime)
		valSlice[j] = valueToken[1]

		fmt.Println("Date is: ", dateSlice[j], "Value is: ", valSlice[j])
	}

	return dateSlice, valSlice
}

func EmitEntryFromWeeklyTokens(userDailyRecord *UserDailyRecord, tokens ...[]string) {

	for _, tokenArg := range tokens {
		fmt.Println(tokenArg)
		dateSlice, valSlice := EmitValueDateFromTokenSlice(tokenArg)

		for i := 0; i < len(dateSlice); i++ {

			// Validate the dates and the assumption on the input format match. Fatal error otherwise
			if strings.Compare(userDailyRecord.dayRecord[i][0], dateSlice[i]) != 0 {
				panic("Dates do not match")
			}

			userDailyRecord.dayRecord[i] = append(userDailyRecord.dayRecord[i], valSlice[i])
		}
	}
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

	OUTPUT_ROOT = OUTPUT_ROOT + OUTPUT_DIR + "/" + user_id + "/"
	OUTPUT_SLEEP_DIR = OUTPUT_ROOT + "sleep/"

	err = os.MkdirAll(OUTPUT_SLEEP_DIR, 0755)
	if err != nil {
		log.Fatal("Could not create directory: ", OUTPUT_ROOT+OUTPUT_SLEEP_DIR)
	}

	masterUserOutFile := OUTPUT_ROOT + "user_" + user_id + ".csv"
	masterf, err := os.Create(masterUserOutFile)
	if err != nil {
		log.Fatal("Could not create output file: " + masterUserOutFile)
	}

	defer masterf.Close()
	csvWriter := csv.NewWriter(masterf)
	csvWriter.UseCRLF = false

	flatUserOutFile := OUTPUT_ROOT + "flat_user_" + user_id + ".csv"
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

			isHeaderEntryValid := func(x string) bool { return len(x) > 0 }

			i := 0
			for _, h := range record {
				if isHeaderEntryValid(h) {
					i++
				} else {
					break
				}
			}

			csvWriter.Write(record[:i])
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
