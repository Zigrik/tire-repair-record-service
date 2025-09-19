package api

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func AfterNow(t1, t2 time.Time) bool {
	truncT1 := t1.Truncate(24 * time.Hour)
	truncT2 := t2.Truncate(24 * time.Hour)

	return truncT1.After(truncT2)
}

// the required day of the current month is calculated when calculating from the end of the month
func daysInMonth(date time.Time, delta int) int {
	firstDayOfNextMonth := time.Date(date.Year(), date.Month()+1, 1, 0, 0, 0, 0, date.Location())
	rightDay := firstDayOfNextMonth.AddDate(0, 0, delta).Day()
	return rightDay
}

func nextDate(now time.Time, dstart string, repeat string) (string, error) {
	formatError := fmt.Errorf("unsupported 'repeat' format")
	start, err := time.Parse(dateForm, dstart)
	if err != nil {
		return "", fmt.Errorf("incorrect date format. Expected YYYYMMDD")
	}

	if len(repeat) == 0 {
		return "", fmt.Errorf("'repeat' cannot be an empty value")
	}

	repeatParam := strings.Split(repeat, " ")
	if len(repeatParam) > 3 {
		return "", formatError
	}

	dateNew := start
	switch repeatParam[0] {
	case "y":
		if len(repeatParam) != 1 {
			return "", formatError
		}
		for dateNew.Before(now) || !dateNew.After(start) {
			dateNew = dateNew.AddDate(1, 0, 0)
		}
	case "d":
		if len(repeatParam) != 2 {
			return "", formatError
		}
		days, err := strconv.Atoi(repeatParam[1])
		if err != nil || days > 400 || days < 1 {
			return "", formatError
		}
		for dateNew.Before(now) || !dateNew.After(start) {
			dateNew = dateNew.AddDate(0, 0, days)
		}
	case "w":
		if len(repeatParam) != 2 {
			return "", formatError
		}
		week := make(map[time.Weekday]bool)
		days := strings.Split(repeatParam[1], ",")
		for _, v := range days {
			day, err := strconv.Atoi(v)
			if err != nil || day > 7 || day < 1 {
				return "", formatError
			}
			if day == 7 {
				day = 0
			}
			week[time.Weekday(day)] = true
		}
		var correctDate bool
		for dateNew.Before(now) || !dateNew.After(start) || !correctDate {
			dateNew = dateNew.AddDate(0, 0, 1)
			correctDate = false
			if week[dateNew.Weekday()] {
				correctDate = true
			}
		}
	case "m":
		if len(repeatParam) < 2 || len(repeatParam) > 3 {
			return "", formatError
		}
		dayOfMonth := make(map[int]bool)

		days := strings.Split(repeatParam[1], ",")
		for _, v := range days {
			day, err := strconv.Atoi(v)
			if err != nil || day > 31 || day < -2 || day == 0 {
				return "", formatError
			}
			dayOfMonth[day] = true
		}

		month := make(map[time.Month]bool)
		if len(repeatParam) == 3 {
			monthNumbers := strings.Split(repeatParam[2], ",")
			for _, v := range monthNumbers {
				monthNumber, err := strconv.Atoi(v)
				if err != nil || monthNumber > 12 || monthNumber < 1 {
					return "", formatError
				}
				month[time.Month(monthNumber)] = true
			}
		}

		var correctDate bool
		for dateNew.Before(now) || !dateNew.After(start) || !correctDate {
			dateNew = dateNew.AddDate(0, 0, 1)
			correctDate = false
			if (len(repeatParam) == 3 && month[dateNew.Month()]) || len(repeatParam) == 2 {
				if dayOfMonth[dateNew.Day()] {
					correctDate = true
				}
				if dayOfMonth[-1] {
					if dateNew.Day() == daysInMonth(dateNew, -1) {
						correctDate = true
					}
				}
				if dayOfMonth[-2] {
					if dateNew.Day() == daysInMonth(dateNew, -2) {
						correctDate = true
					}
				}
			}
		}
	default:
		return "", formatError
	}
	return dateNew.Format(dateForm), nil
}

func nextDateHandler(res http.ResponseWriter, req *http.Request, logger *log.Logger) {
	if req.Method != http.MethodGet {
		logger.Printf("WARN: bad request")
		http.Error(res, "The request was not received using the GET method", http.StatusMethodNotAllowed)
		return
	}

	query := req.URL.Query()

	nowStr := query.Get("now")
	now, err := time.Parse(dateForm, nowStr)
	if err != nil {
		logger.Printf("WARN: Invalid 'now' parameter format. Expected YYYYMMDD")
		http.Error(res, "Invalid 'now' parameter format. Expected YYYYMMDD", http.StatusBadRequest)
		return
	}

	dstart := query.Get("date")
	if dstart == "" {
		logger.Printf("WARN: missing 'date' parameter")
		http.Error(res, "Missing 'date' parameter", http.StatusBadRequest)
		return
	}

	repeat := query.Get("repeat")

	nextDate, err := nextDate(now.AddDate(0, 0, 1), dstart, repeat)
	if err != nil {
		logger.Printf("WARN: nextDate error: %v", err.Error())
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	res.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := res.Write([]byte(nextDate)); err != nil {
		logger.Printf("WARN: failed to write response")
		http.Error(res, "Failed to write response", http.StatusInternalServerError)
	}
}
