/*
Copyright (c) 2023 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package schedule

import (
	"github.com/pkg/errors"
	cron "github.com/robfig/cron/v3"
)

var schedulePresets = map[string]string{
	"weekdays": "* * * * 1-4",
	"anytime":  "* * * * *",
}

func SupportedSchedulePresets() []string {
	schedules := make([]string, len(schedulePresets))
	i := 0
	for schedule := range schedulePresets {
		schedules[i] = schedule
		i++
	}
	return schedules
}

func TranslateSchedule(schedule string) (string, error) {
	if schedule == "" {
		return "", errors.New("schedule cannot be empty")
	}

	// check if the schedule is one of the presets
	if schedulePreset, ok := schedulePresets[schedule]; ok {
		return schedulePreset, nil
	}

	// no preset - check if it is a valid cron expression
	cron_parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	_, err := cron_parser.Parse(schedule)
	if err != nil {
		return "", errors.Wrap(err, "invalid schedule")
	}

	return schedule, nil
}
