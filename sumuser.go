package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
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
	dates      [7]string
	dateTime   [7]time.Time
}

var startWeek DatesInWeek

func StripCurlyBrackets(originalString string) string {

	z := strings.ReplaceAll(originalString, "{", "")
	return strings.ReplaceAll(z, "}", "")
}

func StripSquareBrackets(originalString string) string {

	x := strings.ReplaceAll(originalString, "[", "")
	return strings.ReplaceAll(x, "]", "")
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

// CalculateDateByWeek - Generates a string array of dates representing the week
func CalculateDateByWeek() bool {

	file, err := os.Open(WEEKLY_LOG_FILE)
	if err != nil {
		log.Fatal("Could not open input file", err)
	}

	defer file.Close()
	csvReader := csv.NewReader(file)
	csvReader.TrimLeadingSpace = true

	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("Error:", err)
			continue
		}

		// Hopefully there is a week 0, with a steps entry somewhere in the file. Fatal if not
		if record[IDX_WEEK] == "0" && len(record[IDX_STEPS]) > 0 {

			startWeek.weekNumber = record[IDX_WEEK]

			// Strip the record of all brackets then tokenize so we can get to the 7 dates that make up the week
			stepsRecord := StripAllBrackets(record[IDX_STEPS])
			fmt.Println(stepsRecord)
			stepTokens := strings.Split(stepsRecord, ",")

			var ct int = 0
			for i := 1; i < len(stepTokens); i += 2 {
				dateTimeToken := strings.Split(stepTokens[i], ":")
				startWeek.dates[ct] = dateTimeToken[1]
			}

			return true
		}
	}

	return false
}

func ProcessSteps(record string) string {

	return record
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

	weekNumber := record[IDX_WEEK]

}

func main() {

	// First scan the file to initialize our week and date range
	if !CalculateDateByWeek() {
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
