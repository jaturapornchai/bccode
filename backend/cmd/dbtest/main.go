package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"context"

	"github.com/ClickHouse/clickhouse-go/v2"
	_ "github.com/lib/pq"
)

type Config struct {
	ClickHouse struct {
		DatabaseName string `json:"databasename"`
		Host         string `json:"host"`
		Password     string `json:"password"`
		Port         string `json:"port"`
		User         string `json:"user"`
	} `json:"clickhouse"`
	PostgreSQL struct {
		DbName   string `json:"dbname"`
		Host     string `json:"host"`
		Password string `json:"password"`
		Port     string `json:"port"`
		SSLMode  string `json:"sslmode"`
		User     string `json:"user"`
	} `json:"postgresql"`
}

func main() {
	fmt.Println("==================================================")
	fmt.Println("   BC Ai Account Database Connection Tester")
	fmt.Println("==================================================")

	// 1. อ่านไฟล์ bootstrap.json
	configPath := filepath.Join(".", "bootstrap.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		configPath = filepath.Join("..", "bootstrap.json")
	}

	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Fatalf("ล้มเหลวในการอ่าน bootstrap.json: %v", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		log.Fatalf("ล้มเหลวในการ parse bootstrap.json: %v", err)
	}

	targetHoldingCode := "3E0aX0qsmeRr26TjCk3kRz5vBdv"

	// 2. ทดสอบ PostgreSQL (แยก Database ตาม Holding Code)
	fmt.Printf("\n[1] ทดสอบ PostgreSQL (Database แยกตาม holdingcode: %s)\n", targetHoldingCode)
	pgConnStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s connect_timeout=5",
		config.PostgreSQL.Host,
		config.PostgreSQL.Port,
		config.PostgreSQL.User,
		config.PostgreSQL.Password,
		targetHoldingCode, // ต่อตรงเข้า shop database
		config.PostgreSQL.SSLMode,
	)

	pgDb, err := sql.Open("postgres", pgConnStr)
	if err != nil {
		log.Printf("ERROR: ไม่สามารถเปิด PostgreSQL connection: %v\n", err)
	} else {
		defer pgDb.Close()

		// ลอง Ping ดู
		start := time.Now()
		err = pgDb.Ping()
		if err != nil {
			log.Printf("ERROR: ไม่สามารถเชื่อมต่อ (Ping) PostgreSQL: %v\n", err)
		} else {
			duration := time.Since(start)
			fmt.Printf("✓ เชื่อมต่อ PostgreSQL สำเร็จ (dbname=%s, host=%s) ในเวลา %v\n", targetHoldingCode, config.PostgreSQL.Host, duration)

			// สอบถามข้อมูลพื้นฐาน
			var dbName string
			err = pgDb.QueryRow("SELECT current_database()").Scan(&dbName)
			if err == nil {
				fmt.Printf("  - Active PostgreSQL Database: %s\n", dbName)
			}

			// ดึงรายชื่อ Table
			rows, err := pgDb.Query("SELECT table_name FROM information_schema.tables WHERE table_schema='public' LIMIT 5")
			if err == nil {
				defer rows.Close()
				fmt.Println("  - ตารางตัวอย่างใน PostgreSQL:")
				for rows.Next() {
					var tableName string
					if err := rows.Scan(&tableName); err == nil {
						fmt.Printf("    * %s\n", tableName)
					}
				}
			} else {
				log.Printf("  - ดึงรายชื่อ Table ล้มเหลว: %v\n", err)
			}
		}
	}

	// 3. ทดสอบ ClickHouse (รวม Database กลาง แต่แยกฟิลด์ holdingcode เพื่อทำ BI)
	fmt.Printf("\n[2] ทดสอบ ClickHouse (Database กลาง: %s)\n", config.ClickHouse.DatabaseName)

	chOpts := &clickhouse.Options{
		Addr: []string{fmt.Sprintf("%s:%s", config.ClickHouse.Host, config.ClickHouse.Port)},
		Auth: clickhouse.Auth{
			Database: config.ClickHouse.DatabaseName,
			Username: config.ClickHouse.User,
			Password: config.ClickHouse.Password,
		},
		DialTimeout: 5 * time.Second,
	}

	chConn, err := clickhouse.Open(chOpts)
	if err != nil {
		log.Printf("ERROR: ไม่สามารถเปิด ClickHouse connection: %v\n", err)
	} else {
		defer chConn.Close()

		start := time.Now()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err = chConn.Ping(ctx)
		if err != nil {
			log.Printf("ERROR: ไม่สามารถเชื่อมต่อ (Ping) ClickHouse: %v\n", err)
		} else {
			duration := time.Since(start)
			fmt.Printf("✓ เชื่อมต่อ ClickHouse สำเร็จ (dbname=%s, host=%s) ในเวลา %v\n", config.ClickHouse.DatabaseName, config.ClickHouse.Host, duration)

			// ดึงตารางที่มีอยู่ใน ClickHouse
			var tables []struct {
				Name string `ch:"name"`
			}
			query := fmt.Sprintf("SELECT name FROM system.tables WHERE database = '%s' LIMIT 5", config.ClickHouse.DatabaseName)
			err = chConn.Select(ctx, &tables, query)
			if err == nil {
				fmt.Println("  - ตารางตัวอย่างใน ClickHouse:")
				for _, t := range tables {
					fmt.Printf("    * %s\n", t.Name)
				}
			} else {
				log.Printf("  - ดึงตาราง ClickHouse ล้มเหลว: %v\n", err)
			}

			// ดึงสถิติจำนวนเอกสารแยกตาม holdingcode เพื่อแสดงการทำ BI หลายกิจการ
			var shopStats []struct {
				HoldingCode string `ch:"holdingcode"`
				Count       uint64 `ch:"cnt"`
			}
			statQuery := fmt.Sprintf("SELECT holdingcode, count() as cnt FROM %s.doc GROUP BY holdingcode LIMIT 10", config.ClickHouse.DatabaseName)
			err = chConn.Select(ctx, &shopStats, statQuery)
			if err == nil {
				fmt.Println("  - สถิติ BI (จำนวนแถวในตาราง doc แยกตาม holdingcode):")
				if len(shopStats) == 0 {
					fmt.Println("    * ยังไม่มีข้อมูลเอกสารในระบบ")
				}
				for _, stat := range shopStats {
					fmt.Printf("    * Holding Code: %s => จำนวนเอกสาร: %d แถว\n", stat.HoldingCode, stat.Count)
				}
			} else {
				// อาจจะยังไม่มีตาราง doc เลยทดลองหาจาก productbarcode แทน
				var barcodeStats []struct {
					HoldingCode string `ch:"holdingcode"`
					Count       uint64 `ch:"cnt"`
				}
				statQuery2 := fmt.Sprintf("SELECT holdingcode, count() as cnt FROM %s.productbarcode GROUP BY holdingcode LIMIT 10", config.ClickHouse.DatabaseName)
				err = chConn.Select(ctx, &barcodeStats, statQuery2)
				if err == nil {
					fmt.Println("  - สถิติ BI (จำนวนแถวในตาราง productbarcode แยกตาม holdingcode):")
					if len(barcodeStats) == 0 {
						fmt.Println("    * ยังไม่มีข้อมูลสินค้าในระบบ")
					}
					for _, stat := range barcodeStats {
						fmt.Printf("    * Holding Code: %s => จำนวนบาร์โค้ด: %d รายการ\n", stat.HoldingCode, stat.Count)
					}
				} else {
					log.Printf("  - ดึงข้อมูลสถิติ BI ล้มเหลว: %v\n", err)
				}
			}
		}
	}
	fmt.Println("\n==================================================")
}
