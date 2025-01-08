package utility

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func FormatQuery(query string, parametersJson json.RawMessage) (string, error) {
	var parameters map[string]interface{}
	err := json.Unmarshal(parametersJson, &parameters)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal parameters: %w", err)
	}

	for key, val := range parameters {
		placeholder := fmt.Sprintf("$%s", key)
		var replacement string
		switch v := val.(type) {
		case string:
			if strings.HasPrefix(v, "now()") { // timestamp-like string
				rep, err := parseTimestamp(v)
				if err != nil {
					return "", fmt.Errorf("invalid timestamp format for %v: %v", key, val)
				}
				replacement = rep
			} else { // regular string
				replacement = fmt.Sprintf("'%s'", strings.ReplaceAll(v, "'", "''"))
			}
		case float64: // JSON numbers are unmarshaled as float64 by default
			if v == float64(int64(v)) { // Check if it's actually an integer value
				replacement = fmt.Sprintf("%d", int64(v))
			} else {
				replacement = fmt.Sprintf("%f", v)
			}
		case bool:
			replacement = strconv.FormatBool(v)
		default:
			return "", fmt.Errorf("unsupported parameter type for %v: %v", key, val)
		}
		query = strings.ReplaceAll(query, placeholder, replacement)
	}

	return query, nil
}

func parseTimestamp(input string) (string, error) {
	now := time.Now()

	if input == "now()" {
		return fmt.Sprintf("'%s'", now.Format("2006-01-02 15:04:05")), nil
	}

	if strings.HasPrefix(input, "now()-") {
		durationStr := strings.TrimPrefix(input, "now()-")
		timestamp := time.Now()

		if strings.HasSuffix(durationStr, "d") {
			days, err := strconv.Atoi(strings.TrimSuffix(durationStr, "d"))
			if err != nil {
				return "", fmt.Errorf("invalid day duration: %v", err)
			}
			timestamp = timestamp.AddDate(0, 0, -days)
		} else if strings.HasSuffix(durationStr, "w") {
			weeks, err := strconv.Atoi(strings.TrimSuffix(durationStr, "w"))
			if err != nil {
				return "", fmt.Errorf("invalid week duration: %v", err)
			}
			timestamp = timestamp.AddDate(0, 0, -7*weeks)
		} else if strings.HasSuffix(durationStr, "y") {
			years, err := strconv.Atoi(strings.TrimSuffix(durationStr, "y"))
			if err != nil {
				return "", fmt.Errorf("invalid year duration: %v", err)
			}
			timestamp = timestamp.AddDate(0, 0, -365*years)
		} else {
			duration, err := time.ParseDuration(strings.ReplaceAll(durationStr, " ", ""))
			if err != nil {
				return "", fmt.Errorf("invalid duration format: %v", err)
			}
			timestamp = timestamp.Add(-duration)
		}

		return fmt.Sprintf("'%s'", timestamp.Format("2006-01-02 15:04:05")), nil
	}
	return "", fmt.Errorf("unsupported timestamp format")
}
