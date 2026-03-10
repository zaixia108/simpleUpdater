package db

import (
	"database/sql"
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
	_, err := LocalDB.DB.Exec("INSERT INTO applications (AppName, AppAvailableVersion, AppLatestVersion, AppForceUpdateMiniumVersion, DirectLink, NoneDirectLink, Notice) VALUES (?, ?, ?, ?, ?, ?, ?)", record.AppName, fmt.Sprintf("%v", record.AppAvailableVersion), record.AppLatestVersion, record.AppForceUpdateMiniumVersion, record.DirectLink, record.NoneDirectLink, record.Notice)
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
	row := LocalDB.DB.QueryRow("SELECT AppName, AppAvailableVersion, AppLatestVersion, AppForceUpdateMiniumVersion, DirectLink, NoneDirectLink, Notice FROM applications WHERE AppName = ?", appName)
	var record ApplicationRecord
	var availableVersions string
	err := row.Scan(&record.AppName, &availableVersions, &record.AppLatestVersion, &record.AppForceUpdateMiniumVersion, &record.DirectLink, &record.NoneDirectLink, &record.Notice)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	// Convert the availableVersions string back to a slice of float32
	var versions []int
	_, err = fmt.Sscanf(availableVersions, "%v", &versions)
	if err != nil {
		return nil, err
	}
	record.AppAvailableVersion = versions
	return &record, nil
}

func UpdateRecord(record ApplicationRecord) error {
	LocalDB.DBlock.Lock()
	defer LocalDB.DBlock.Unlock()
	_, err := LocalDB.DB.Exec("UPDATE applications SET AppAvailableVersion = ?, AppLatestVersion = ?, AppForceUpdateMiniumVersion = ?, DirectLink = ?, NoneDirectLink = ?, Notice = ? WHERE AppName = ?", fmt.Sprintf("%v", record.AppAvailableVersion), record.AppLatestVersion, record.AppForceUpdateMiniumVersion, record.DirectLink, record.NoneDirectLink, record.Notice, record.AppName)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}
