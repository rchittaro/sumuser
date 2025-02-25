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

type UserDailyStepData struct {
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
	header       []string
	id           string
	weekNumber   string
	user_id      string
	stepsData    UserDailyStepData
	activityData UserDailyActiveData
	sleepData    UserDailySleepData
	calorieData  UserDailyCaloriesData
	heartData    UserDailyHeartData
}

func ProcessRecord(user_id string, record []string) {

	outputFile := "output/flat_user_" + user_id + ".csv"
	outFile, err := os.Create(outputFile)
	if err != nil {
		log.Fatal("Could not create output file: " + outputFile)
	}

	defer outFile.Close()
	csvWriter := csv.NewWriter(outFile)
	csvWriter.UseCRLF = false

	// Initialize a structure that represents all of the dates for this week.
	currentWeekData := GenerateMapForWeek(record)
	SummarizeSteps(currentWeekData, record[IDX_STEPS])
	//SummarizeTimeActive(currentWeekData, record[IDX_TIME_ACTIVE])
	//SummarizeSleep(currentWeekData, record[IDX_SLEEP])
	//SummarizeHeartRate(currentWeekData, record[IDX_HEART_RATE])
	//SummarizeCalorie(currentWeekData, record[IDX_CALORIES_IN])

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

func GenerateMapForWeek(record []string) map[string]UserDailyData {
	m := make(map[string]UserDailyData)

	if len(record[IDX_STEPS]) == 0 {
		panic("not done")
	}

	// Get a tokenized represenation of the steps record for all 7 days  ["value:0","dateTime:2021-01-16",.......,"value:0","dateTime:2021-01-17"]
	stepTokens := TokenizeSteps(record[IDX_STEPS])
	stepsTotal := 0

	for i := 1; i < len(stepTokens); i += 2 {
		dateTimeToken := strings.Split(stepTokens[i], ":")
		stepsValueToken := strings.Split(stepTokens[i-1], ":")

		v, err := strconv.Atoi(stepsValueToken[1])
		if err != nil || v < 0 {
			fmt.Println("Failed to convert steps value from string to int: ", stepsValueToken[i-1], err.Error())
		} else {
			stepsTotal += v
		}

		// Create a Date object out of the string
		tmpTime, err := time.Parse(time.DateOnly, dateTimeToken[1])
		if err != nil {
			log.Fatal("Could not parse time: ", err.Error())
		}

		var dailyData UserDailyData

		// Save the user and steps data now
		dailyData.user_id = record[IDX_USER_ID]
		dailyData.stepsData.steps = stepsTotal
		m[GetDateOnly(tmpTime)] = dailyData

		fmt.Println(dateTimeToken[1] + " becomes: " + GetDateOnly(tmpTime))
	}

	return m
}

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

func SummarizeSteps(currentWeekData map[string]UserDailyData, s string) {
	fmt.Println("--------Summarizing Steps Data-------- \n" + s)
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

	outputFile := "output/user_" + user_id + ".csv"
	outFile, err := os.Create(outputFile)
	if err != nil {
		log.Fatal("Could not create output file: " + outputFile)
	}

	defer outFile.Close()
	csvWriter := csv.NewWriter(outFile)
	csvWriter.UseCRLF = false

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

			// Save the user specific record to the new file...might be useful output
			csvWriter.Write(record)

			// Save to the flattened csv output file
			ProcessRecord(user_id, record)
		}
	}

	outFile.Close()
}
