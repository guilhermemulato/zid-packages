package autoupdate

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

const (
	StatePath      = "/var/db/zid-packages/auto-update.json"
	MinDays        = 0
	daySeconds     = 24 * time.Hour
	ScheduleHour   = 23
	ScheduleMinute = 59
)

type Entry struct {
	Version   string `json:"version"`
	FirstSeen int64  `json:"first_seen"`
	LastSeen  int64  `json:"last_seen"`
}

type State struct {
	Packages   map[string]Entry `json:"packages"`
	LastRunDay string           `json:"last_run_day"`
}

func Load() (State, error) {
	data, err := os.ReadFile(StatePath)
	if err != nil {
		if os.IsNotExist(err) {
			return State{Packages: map[string]Entry{}}, nil
		}
		return State{}, err
	}
	var st State
	if err := json.Unmarshal(data, &st); err != nil {
		return State{}, err
	}
	if st.Packages == nil {
		st.Packages = map[string]Entry{}
	}
	return st, nil
}

func Save(st State) error {
	if dir := filepath.Dir(StatePath); dir != "" {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return err
		}
	}
	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(StatePath, data, 0600)
}

func Update(st *State, key string, updateAvailable bool, remoteVersion string, now time.Time) (Entry, bool) {
	if st.Packages == nil {
		st.Packages = map[string]Entry{}
	}
	changed := false
	if !updateAvailable || remoteVersion == "" {
		if _, ok := st.Packages[key]; ok {
			delete(st.Packages, key)
			changed = true
		}
		return Entry{}, changed
	}
	entry, ok := st.Packages[key]
	if !ok || entry.Version != remoteVersion || entry.FirstSeen == 0 {
		entry = Entry{
			Version:   remoteVersion,
			FirstSeen: now.Unix(),
			LastSeen:  now.Unix(),
		}
		st.Packages[key] = entry
		return entry, true
	}
	return entry, changed
}

func Clear(st *State, key string) bool {
	if st.Packages == nil {
		return false
	}
	if _, ok := st.Packages[key]; ok {
		delete(st.Packages, key)
		return true
	}
	return false
}

func AgeDays(entry Entry, now time.Time) int {
	if entry.FirstSeen == 0 {
		return 0
	}
	start := time.Unix(entry.FirstSeen, 0)
	if now.Before(start) {
		return 0
	}
	return int(now.Sub(start) / daySeconds)
}

func Due(entry Entry, now time.Time) bool {
	if entry.FirstSeen == 0 {
		return false
	}
	start := time.Unix(entry.FirstSeen, 0)
	return now.Sub(start) >= time.Duration(MinDays)*daySeconds
}

func ThresholdDays() int {
	return MinDays
}

func DueAt(entry Entry, days int, loc *time.Location) time.Time {
	if entry.FirstSeen == 0 || days < 0 {
		return time.Time{}
	}
	if loc == nil {
		loc = time.Local
	}
	firstSeen := time.Unix(entry.FirstSeen, 0).In(loc)
	base := time.Date(firstSeen.Year(), firstSeen.Month(), firstSeen.Day(), ScheduleHour, ScheduleMinute, 0, 0, loc)
	return base.AddDate(0, 0, days)
}

func DueWithState(entry Entry, now time.Time, st State) bool {
	dueAt := DueAtWithState(entry, ThresholdDays(), now.Location(), st, now)
	if dueAt.IsZero() {
		return false
	}
	return !now.Before(dueAt)
}

func DueAtWithState(entry Entry, days int, loc *time.Location, st State, now time.Time) time.Time {
	dueAt := DueAt(entry, days, loc)
	if dueAt.IsZero() {
		return time.Time{}
	}
	if loc == nil {
		loc = time.Local
	}
	today := now.In(loc).Format("2006-01-02")
	if st.LastRunDay != "" && st.LastRunDay == today {
		next := time.Date(now.In(loc).Year(), now.In(loc).Month(), now.In(loc).Day(), ScheduleHour, ScheduleMinute, 0, 0, loc).AddDate(0, 0, 1)
		if dueAt.Before(next) {
			return next
		}
	}
	return dueAt
}

func ShouldRunNow(st State, now time.Time, hour, minute int) bool {
	if now.Hour() != hour || now.Minute() != minute {
		return false
	}
	today := now.Format("2006-01-02")
	return st.LastRunDay != today
}

func MarkRun(st *State, now time.Time) {
	st.LastRunDay = now.Format("2006-01-02")
}
