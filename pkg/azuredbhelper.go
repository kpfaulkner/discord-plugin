package main

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/denisenkom/go-mssqldb"
	log "github.com/sirupsen/logrus"
	"time"
)

const (
	ActionUserJoinedGuild     int = 1
	ActionUserLeftGuild       int = 2
	ActionMessageCreated      int = 3
	ActionMembersInGuildCount int = 4
)

type NumberUsersPerTimeGroup struct {
	StartTime time.Time
	EndTime   time.Time
	Count     int
}

type TimeQueryEntry struct {
	TimeStamp time.Time
	Count     int64
}

type AzureDBHelper struct {
	db       *sql.DB
	config   Config
	database string
	username string
}

func NewAzureDBHelper(config Config, database string, username string) AzureDBHelper {
	a := AzureDBHelper{}
	a.config = config
	a.username = username
	a.database = database
	a.db, _ = a.ConnectToDB(database, username)
	return a
}

func (d *AzureDBHelper) ConnectToDB(database string, username string) (*sql.DB, error) {

	var connString string
	pw := d.config.Creds[username]
	connString = fmt.Sprintf("server=%s;user id=%s;password=%s;port=%s;database=%s;", d.config.Server, username, pw, d.config.Port, database)

	// Create connection pool
	db, err := sql.Open("sqlserver", connString)
	if err != nil {
		return nil, err
	}
	return db, nil
}

// GetNumberMessageForBuildBetweenTimes get number of messages between start and end. Group into groupingInMinutes chunks (eg 5 mins)
func (d *AzureDBHelper) GetNumberMessageForGuildBetweenTimes(guildID string, start time.Time, end time.Time, groupingInMinutes int) ([]TimeQueryEntry, error) {

	tsql := fmt.Sprintf("SELECT format(TimeStamp, 'yyyy-MM-dd HH:mm') AS RoundedTimeStamp, COUNT(*) as Count FROM Actions WHERE GuildID='%s' and Action=%d and TimeStamp between '%s' and '%s' GROUP BY  format(TimeStamp, 'yyyy-MM-dd HH:mm') ORDER BY 1", guildID, ActionMessageCreated,
		start.UTC().Format("2006-01-02 15:04:05"), end.UTC().Format("2006-01-02 15:04:05"))
	allResults, err := d.DoQueryAndRoundTimes(tsql)
	return allResults, err
}

// GetMemberCountForGuildBetweenTimes Returns the number of members in a guild...  each entry is the total (ie no summing/totalling please)
func (d *AzureDBHelper) GetMemberCountForGuildBetweenTimes(guildID string, start time.Time, end time.Time, groupingInMinutes int) ([]TimeQueryEntry, error) {

	tsql := fmt.Sprintf("SELECT TimeStamp, Count FROM Actions WHERE GuildID='%s' and Action=%d and TimeStamp between '%s' and '%s' order by timestamp desc ", guildID, ActionMembersInGuildCount,
		start.UTC().Format("2006-01-02 15:04:05"), end.UTC().Format("2006-01-02 15:04:05"))
	allResults, err := d.DoQuery(tsql)
	return allResults, err
}

// GetNumberJoinsForGuildBetweenTimes get number of joins between start and end. Group into groupingInMinutes chunks (eg 5 mins)
func (d *AzureDBHelper) GetNumberJoinsForGuildBetweenTimes(guildID string, start time.Time, end time.Time, groupingInMinutes int) ([]TimeQueryEntry, error) {

	tsql := fmt.Sprintf("SELECT format(TimeStamp, 'yyyy-MM-dd HH:mm') AS RoundedTimeStamp, COUNT(*) as Count FROM Actions WHERE GuildID='%s' and Action=%d and TimeStamp between '%s' and '%s' GROUP BY  format(TimeStamp, 'yyyy-MM-dd HH:mm') ORDER BY 1", guildID, ActionUserJoinedGuild,
		start.UTC().Format("2006-01-02 15:04:05"), end.UTC().Format("2006-01-02 15:04:05"))
	allResults, err := d.DoQueryAndRoundTimes(tsql)
	return allResults, err
}

// GetNumberLeftForGuildBetweenTimes get number of left actions between start and end. Group into groupingInMinutes chunks (eg 5 mins)
func (d *AzureDBHelper) GetNumberLeftForGuildBetweenTimes(guildID string, start time.Time, end time.Time, groupingInMinutes int) ([]TimeQueryEntry, error) {

	tsql := fmt.Sprintf("SELECT format(TimeStamp, 'yyyy-MM-dd HH:mm') AS RoundedTimeStamp, COUNT(*) as Count FROM Actions WHERE GuildID='%s' and Action=%d and TimeStamp between '%s' and '%s' GROUP BY  format(TimeStamp, 'yyyy-MM-dd HH:mm') ORDER BY 1", guildID, ActionUserLeftGuild,
		start.UTC().Format("2006-01-02 15:04:05"), end.UTC().Format("2006-01-02 15:04:05"))
	allResults, err := d.DoQueryAndRoundTimes(tsql)
	return allResults, err
}

// DoQueryAndRoundTimes execute query and round times.
func (d *AzureDBHelper) DoQuery(tsql string) ([]TimeQueryEntry, error) {

	//allResults := []NumberUsersPerTimeGroup{}
	allRawResults := []TimeQueryEntry{}

	ctx := context.Background()

	log.Infof("SQL is %s", tsql)

	// Execute query
	rows, err := d.db.QueryContext(ctx, tsql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Iterate through the result set.
	for rows.Next() {
		var timeStamp time.Time
		var count int64
		err := rows.Scan(&timeStamp, &count)
		if err != nil {
			return nil, err
		}

		data := TimeQueryEntry{TimeStamp: timeStamp, Count: count}
		allRawResults = append(allRawResults, data)
	}

	return allRawResults, nil
}

// DoQueryAndRoundTimes execute query and round times.
func (d *AzureDBHelper) DoQueryAndRoundTimes(tsql string) ([]TimeQueryEntry, error) {

	//allResults := []NumberUsersPerTimeGroup{}
	allRawResults := []TimeQueryEntry{}

	ctx := context.Background()

	log.Infof("SQL is %s", tsql)

	// Execute query
	rows, err := d.db.QueryContext(ctx, tsql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Iterate through the result set.
	for rows.Next() {
		var roundedTimeStamp string
		var count int64
		err := rows.Scan(&roundedTimeStamp, &count)
		if err != nil {
			return nil, err
		}

		timeStamp, _ := time.Parse("2006-01-02 15:04", roundedTimeStamp)

		data := TimeQueryEntry{TimeStamp: timeStamp, Count: count}
		allRawResults = append(allRawResults, data)
	}

	return allRawResults, nil
}
