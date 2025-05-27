package main

import (
	"CryptoLens_Backend/db"
	"CryptoLens_Backend/env"
	"CryptoLens_Backend/logger"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/robfig/cron/v3"
	"log"
	"net/http"
)

type Info struct {
	CT string `json:"ct"`
	CL string `json:"cl"`
	MT string `json:"mt"`
	MF string `json:"mf"`
	MU string `json:"mu"`
}

type OrangeNode struct {
	ID        int
	Name      string
	Host      string
	Port      string
	Deleted   bool
	CreatedAt string
}

var (
	authorization string
	dbwork        *sql.DB
)

func init() {
	initLogger()
	initDB()
	authorization = "Bearer " + env.GetToken()
}

func main() {
	c := cron.New()
	c.AddFunc("@every 10s", func() { fetchInfo(dbwork) })
	c.Start()

	select {}
}

func initDB() {
	var err error
	dbwork, err = db.InitDB()
	if err != nil {
		log.Fatal(err)
	}
}

func initLogger() {
	err := logger.Init("logs/app.log")
	if err != nil {
		log.Fatal(err)
	}
}

func fetchInfo(db *sql.DB) {
	var err error
	var rows *sql.Rows

	rows, err = db.Query("SELECT * FROM orange_node WHERE deleted <> true")
	if err != nil {
		logger.Log.Println("Error inserting to database:", err)
	}
	defer rows.Close()

	var nodes []OrangeNode

	for rows.Next() {
		var node OrangeNode

		err := rows.Scan(&node.ID, &node.Name, &node.Host, &node.Port, &node.Deleted, &node.CreatedAt)
		if err != nil {
			logger.Log.Println("Error scanning row:", err)
			continue
		}

		nodes = append(nodes, node)
	}

	if err = rows.Err(); err != nil {
		log.Println("Error during row iteration:", err)
	}

	for _, node := range nodes {
		apiUrl := fmt.Sprintf("%s:%s/getInfo", node.Host, node.Port)
		var req *http.Request
		req, err = http.NewRequest("GET", apiUrl, nil)
		if err != nil {
			logger.Log.Println("Error creating request:", err)
			continue
		}
		req.Header.Set("Authorization", authorization)
		client := &http.Client{}
		var resp *http.Response
		resp, err = client.Do(req)
		if err != nil {
			logger.Log.Println("Error making request:", err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			logger.Log.Println("Error: received non-200 response status:", resp.Status)
			continue
		}
		var info Info
		if err = json.NewDecoder(resp.Body).Decode(&info); err != nil {
			logger.Log.Println("Error decoding response:", err)
			continue
		}
		_, err = db.Exec("INSERT INTO orange_sys_data (orange_node_id, ct, cl, mt, mf, mu) VALUES ($1, $2, $3, $4, $5, $6)",
			node.ID, info.CT, info.CL, info.MT, info.MF, info.MU)
		if err != nil {
			logger.Log.Println("Error inserting to database:", err)
		}
	}
}
