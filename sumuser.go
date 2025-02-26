package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
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

/*type UserDailyStepData struct {
	header []string
	steps  int
}

type UserDailyActiveData struct {
	header []string
}

type UserDailySleepData struct {
	header []string
}

type UserDailyCaloriesData struct {
	header []string
}

type UserDailyHeartData struct {
	header []string
}

type UserDailyData struct {
	recordEntry []string
	id          string
	weekNumber  string
	user_id     string

	   stepsData    UserDailyStepData
	   activityData UserDailyActiveData
	   sleepData    UserDailySleepData
	   calorieData  UserDailyCaloriesData
	   heartData    UserDailyHeartData

}
*/

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

	//SummarizeSteps(currentWeekData, record[IDX_STEPS], &header)
	//SummarizeTimeActive(currentWeekData, record[IDX_TIME_ACTIVE], &header)
	//SummarizeSleep(currentWeekData, record[IDX_SLEEP])
	//SummarizeHeartRate(currentWeekData, record[IDX_HEART_RATE])
	//SummarizeCalorie(currentWeekData, record[IDX_CALORIES_IN])

	if !flatHeaderWritten {
		flatHeaderWritten = true
		csvWriter.Write(header)
	}

	WriteWeekSummary(&dailyRecord, csvWriter)
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

/*
func SummarizeCalorie(currentWeekData map[string]UserDailyData, s string) {
	fmt.Println("---------Summarizing Calorie Data--------- \n" + s)
}

func SummarizeHeartRate(currentWeekData map[string]UserDailyData, s string) {

	fmt.Println("--------Summarizing Heart Data--------- \n" + s)
}

func SummarizeSleep(currentWeekData map[string]UserDailyData, s string) {
	fmt.Println("---------Summarizing Sleep Data--------- \n" + s)
}

func SummarizeTimeActive(currentWeekData map[string]UserDailyData, s string) {
	fmt.Println("----------Summarizing Activity Data--------- \n" + s)
}

func SummarizeSteps(currentWeekData map[string]UserDailyData, record string, header *string) {
	// This doesn't do anything right now
}
*/

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
