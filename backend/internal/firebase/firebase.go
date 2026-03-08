package firebase

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"

	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"
)

type IFirebaseAdapter interface {
	ValidateToken(token string) (*UserInfo, error)
}

type FirebaseAdapter struct {
	app *firebase.App
}

func NewFirebaseAdapter() IFirebaseAdapter {
	var app *firebase.App
	var err error

	// ตรวจสอบว่าอยู่ใน development mode หรือไม่
	goEnv := os.Getenv("GO_ENV")
	isDev := goEnv == "development" || goEnv == ""

	// ใน development mode: ข้าม TLS verify (global) เพื่อแก้ปัญหา proxy/antivirus intercept HTTPS
	// Firebase Auth SDK ใช้ internal HTTP client ที่ไม่รับ option.WithHTTPClient
	// จึงต้อง override http.DefaultTransport แทน
	if isDev {
		http.DefaultTransport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		fmt.Println("[Firebase] TLS skip verify: enabled (development mode)")
	}

	// ลำดับการตรวจสอบ credentials:
	// 1. GOOGLE_APPLICATION_CREDENTIALS (service account JSON) — ใช้ได้ทุก Firebase Admin operation
	// 2. FIREBASE_PROJECT_ID — ใช้เฉพาะ VerifyIDToken (ไม่ต้องใช้ service account)
	if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") != "" {
		app, err = firebase.NewApp(context.Background(), nil)
		if err != nil {
			fmt.Printf("[Firebase] เริ่มต้นด้วย service account ไม่สำเร็จ: %v\n", err)
		} else {
			fmt.Println("[Firebase] เริ่มต้นสำเร็จ (mode: service-account)")
		}
	} else if projectID := os.Getenv("FIREBASE_PROJECT_ID"); projectID != "" {
		config := &firebase.Config{ProjectID: projectID}
		app, err = firebase.NewApp(context.Background(), config, option.WithoutAuthentication())
		if err != nil {
			fmt.Printf("[Firebase] เริ่มต้นด้วย project ID '%s' ไม่สำเร็จ: %v\n", projectID, err)
		} else {
			fmt.Printf("[Firebase] เริ่มต้นสำเร็จ (project: %s, mode: token-verify-only)\n", projectID)
		}
	} else {
		fmt.Println("[Firebase] ไม่พบ GOOGLE_APPLICATION_CREDENTIALS หรือ FIREBASE_PROJECT_ID — token verification จะไม่ทำงาน")
	}

	return &FirebaseAdapter{
		app: app,
	}
}

func (f *FirebaseAdapter) ValidateToken(token string) (*UserInfo, error) {
	if f.app == nil {
		return nil, fmt.Errorf("Firebase ไม่ได้ถูกตั้งค่า — ต้องกำหนด GOOGLE_APPLICATION_CREDENTIALS หรือ FIREBASE_PROJECT_ID")
	}

	ctx := context.Background()
	client, err := f.app.Auth(ctx)
	if err != nil {
		return nil, fmt.Errorf("สร้าง Firebase Auth client ไม่สำเร็จ: %w", err)
	}

	authToken, err := client.VerifyIDToken(ctx, token)
	if err != nil {
		fmt.Printf("[Firebase] VerifyIDToken error: %v\n", err)
		return nil, fmt.Errorf("verify Firebase ID token ไม่สำเร็จ: %w", err)
	}

	userName := ""
	if authToken.Claims["name"] != nil {
		userName = authToken.Claims["name"].(string)
	} else {
		userName = authToken.Claims["email"].(string)
	}

	userInfo := &UserInfo{
		SignInProvider: authToken.Firebase.SignInProvider,
		Email:          authToken.Claims["email"].(string),
		UserId:         authToken.UID,
		Name:           userName,
	}

	return userInfo, nil
}
