package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

type dbStruct struct {
	DB     *sql.DB
	DBlock sync.RWMutex
	Busy   bool
}

type ApplicationRecord struct {
	AppName                     string `json:"appName"`
	AppAvailableVersion         []int  `json:"appAvailableVersion"`
	AppLatestVersion            int    `json:"appLatestVersion"`
	AppForceUpdateMiniumVersion int    `json:"appForceUpdateMiniumVersion"`
	DirectLink                  string `json:"directLink"`
	NoneDirectLink              string `json:"noneDirectLink"`
	Notice                      string `json:"notice"`
	LocalStoragePath            string `json:"localStoragePath"`
}

type ApplicationRecordWithoutPath struct {
	AppName                     string `json:"appName"`
	AppAvailableVersion         []int  `json:"appAvailableVersion"`
	AppLatestVersion            int    `json:"appLatestVersion"`
	AppForceUpdateMiniumVersion int    `json:"appForceUpdateMiniumVersion"`
	DirectLink                  string `json:"directLink"`
	NoneDirectLink              string `json:"noneDirectLink"`
	Notice                      string `json:"notice"`
}

var LocalDB *dbStruct

func InitDB() error {
	// Initialize and return the database connection
	db, err := sql.Open("sqlite3", "database.db")
	if err != nil {
		return err
	}
	LocalDB = &dbStruct{}
	LocalDB.DB = db
	LocalDB.DBlock = sync.RWMutex{}
	LocalDB.Busy = false
	// Create the applications table if it doesn't exist
	_, err = LocalDB.DB.Exec(`CREATE TABLE IF NOT EXISTS applications (
		AppName TEXT PRIMARY KEY,
		AppAvailableVersion TEXT,
		AppLatestVersion INTEGER,
		AppForceUpdateMiniumVersion INTEGER,
		DirectLink TEXT,
		NoneDirectLink TEXT,
		Notice TEXT,
		LocalStoragePath TEXT
	)`)
	if err != nil {
		return err
	}
	return nil
}

func CloseDB() error {
	LocalDB.DBlock.Lock()
	defer LocalDB.DBlock.Unlock()
	return LocalDB.DB.Close()
}

func AddRecord(record ApplicationRecord) error {
	LocalDB.DBlock.Lock()
	defer LocalDB.DBlock.Unlock()
	availableVersion, marshalErr := json.Marshal(record.AppAvailableVersion)
	if marshalErr != nil {
		return marshalErr
	}
	_, err := LocalDB.DB.Exec("INSERT INTO applications (AppName, AppAvailableVersion, AppLatestVersion, AppForceUpdateMiniumVersion, DirectLink, NoneDirectLink, Notice, LocalStoragePath) VALUES (?, ?, ?, ?, ?, ?, ?, ?)", record.AppName, availableVersion, record.AppLatestVersion, record.AppForceUpdateMiniumVersion, record.DirectLink, record.NoneDirectLink, record.Notice, record.LocalStoragePath)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func RemoveRecord(appName string) error {
	LocalDB.DBlock.Lock()
	defer LocalDB.DBlock.Unlock()
	_, err := LocalDB.DB.Exec("DELETE FROM applications WHERE AppName = ?", appName)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func GetRecord(appName string) (*ApplicationRecord, error) {
	LocalDB.DBlock.RLock()
	defer LocalDB.DBlock.RUnlock()
	row := LocalDB.DB.QueryRow("SELECT AppName, AppAvailableVersion, AppLatestVersion, AppForceUpdateMiniumVersion, DirectLink, NoneDirectLink, Notice, LocalStoragePath FROM applications WHERE AppName = ?", appName)
	var record ApplicationRecord
	var availableVersions string
	err := row.Scan(&record.AppName, &availableVersions, &record.AppLatestVersion, &record.AppForceUpdateMiniumVersion, &record.DirectLink, &record.NoneDirectLink, &record.Notice, &record.LocalStoragePath)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	// Convert the availableVersions string back to a slice of float32
	var versions []int
	err = json.Unmarshal([]byte(availableVersions), &versions)
	if err != nil {
		return nil, err
	}
	record.AppAvailableVersion = versions
	return &record, nil
}

func UpdateRecord(record ApplicationRecord) error {
	LocalDB.DBlock.Lock()
	defer LocalDB.DBlock.Unlock()
	availableVersion, marshalErr := json.Marshal(record.AppAvailableVersion)
	if marshalErr != nil {
		return marshalErr
	}
	_, err := LocalDB.DB.Exec("UPDATE applications SET AppAvailableVersion = ?, AppLatestVersion = ?, AppForceUpdateMiniumVersion = ?, DirectLink = ?, NoneDirectLink = ?, Notice = ?, LocalStoragePath = ? WHERE AppName = ?", availableVersion, record.AppLatestVersion, record.AppForceUpdateMiniumVersion, record.DirectLink, record.NoneDirectLink, record.Notice, record.LocalStoragePath, record.AppName)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}
