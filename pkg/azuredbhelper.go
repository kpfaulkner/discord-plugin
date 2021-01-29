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
	ActionUserJoinedGuild int = 1
	ActionUserLeftGuild   int = 2
	ActionMessageCreated  int = 3
)

type NumberUsersPerTimeGroup struct {
	StartTime time.Time
	EndTime   time.Time
	Count     int
}

type TimeQueryEntry struct {
	TimeStamp time.Time
	Count     int
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
func (d *AzureDBHelper) GetNumberMessageForBuildBetweenTimes(guildID string, start time.Time, end time.Time, groupingInMinutes int) ([]TimeQueryEntry, error) {

	//allResults := []NumberUsersPerTimeGroup{}
	allRawResults := []TimeQueryEntry{}

	ctx := context.Background()
	tsql := fmt.Sprintf("SELECT TimeStamp, Count, FROM Actions where GuildID='%s' and Action=%d and TimeStamp between '%s' and '%s'", guildID, ActionMessageCreated,
		start.UTC().Format("2006-01-02 15:04:05"), end.UTC().Format("2006-01-02 15:04:05"))

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
		var count int
		err := rows.Scan(&timeStamp, &count)
		if err != nil {
			return nil, err
		}

		data := TimeQueryEntry{TimeStamp: timeStamp, Count: count}
		allRawResults = append(allRawResults, data)
	}

	return allRawResults, nil
}
