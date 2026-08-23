package uat_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	redis "github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

const (
	defaultAPIURL   = "http://mainapi:8888"
	defaultMongoURI = "mongodb://mongodb:27017/?replicaSet=rs0"
	defaultRedis    = "redis:6379"
)

type localUATHarness struct {
	t            *testing.T
	ctx          context.Context
	apiURL       string
	http         *http.Client
	db           *mongo.Database
	redis        *redis.Client
	runID        string
	username     string
	password     string
	userUID      string
	holdingCode  string
	holdingUID   string
	companyCode  string
	companyUID   string
	branchOneUID string
	branchTwoUID string
	membershipID string
	createCode   string
	createdUID   string
	tokens       map[string]struct{}
	refreshes    map[string]struct{}
	sessionUIDs  map[string]struct{}
}

func TestLocalAuthPermissionMatrix(t *testing.T) {
	if os.Getenv("RUN_LOCAL_UAT") != "1" || os.Getenv("BCAI_UAT_LOCAL_DEV") != "1" {
		t.Skip("set RUN_LOCAL_UAT=1 and BCAI_UAT_LOCAL_DEV=1 only in confirmed Local/DEV")
	}

	h := newLocalUATHarness(t)
	t.Cleanup(h.cleanup)
	h.seed()

	h.expect("health", http.StatusOK, http.MethodGet, "/healthz", "", nil)
	h.expect("wrong password", http.StatusUnauthorized, http.MethodPost, "/login", "", bson.M{
		"username": h.username, "password": h.password + "x",
	})

	access, refresh := h.login("initial login")
	h.expect("verify login-only session", http.StatusOK, http.MethodGet, "/verify-token", access, nil)
	h.testHoldingBootstrap(access)

	exactWorkspace := bson.M{
		"holdingcode":  h.holdingCode,
		"businesscode": h.companyCode,
		"branchuid":    h.branchOneUID,
	}
	wrongBranch := bson.M{
		"holdingcode":  h.holdingCode,
		"businesscode": h.companyCode,
		"branchuid":    h.branchTwoUID,
	}
	h.expect("select exact Branch scope", http.StatusOK, http.MethodPost, "/select-holding", access, exactWorkspace)
	h.expect("read exact Branch scope", http.StatusOK, http.MethodGet, "/organization/branch/"+h.branchOneUID, access, nil)
	h.expect("deny other Branch read", http.StatusForbidden, http.MethodGet, "/organization/branch/"+h.branchTwoUID, access, nil)
	h.expect("deny other Branch selection", http.StatusForbidden, http.MethodPost, "/select-holding", access, wrongBranch)
	h.expect("raw GoAPI route removed", http.StatusNotFound, http.MethodGet, "/goapi/get", access, nil)
	h.expect("cross-Company result route removed", http.StatusNotFound, http.MethodPost, "/goapi/resultget", access, bson.M{})
	h.expect("cross-Company PDF route removed", http.StatusNotFound, http.MethodPost, "/goapi/genpdf", access, bson.M{})
	h.expect("cross-Company PDF history route removed", http.StatusNotFound, http.MethodGet, "/goapi/genpdf/history", access, nil)

	h.updateMembership(bson.M{"permissionversion": int64(2)})
	h.expect("permission version rejects old workspace", http.StatusForbidden, http.MethodGet, "/organization/branch/"+h.branchOneUID, access, nil)
	h.expect("permission change keeps Login", http.StatusOK, http.MethodGet, "/verify-token", access, nil)
	h.expect("permission change requires reselect", http.StatusOK, http.MethodPost, "/select-holding", access, exactWorkspace)

	h.updateMembership(bson.M{"isaccessdisabled": true, "permissionversion": int64(3)})
	h.expect("disabled Membership rejects old workspace", http.StatusForbidden, http.MethodGet, "/organization/branch/"+h.branchOneUID, access, nil)
	h.expect("disabled Membership keeps Login", http.StatusOK, http.MethodGet, "/verify-token", access, nil)
	h.updateMembership(bson.M{"isaccessdisabled": false, "permissionversion": int64(4)})
	h.expect("restored Membership requires reselect", http.StatusOK, http.MethodPost, "/select-holding", access, exactWorkspace)

	h.updateOne("organizationcompanies", bson.M{"isactive": false, "__v": int64(1)})
	h.expect("inactive Company rejects old workspace", http.StatusForbidden, http.MethodGet, "/organization/branch/"+h.branchOneUID, access, nil)
	h.expect("inactive Company keeps Login", http.StatusOK, http.MethodGet, "/verify-token", access, nil)
	h.updateOne("organizationcompanies", bson.M{"isactive": true, "__v": int64(2)})
	h.expect("reactivation does not restore workspace", http.StatusUnauthorized, http.MethodGet, "/organization/branch/"+h.branchOneUID, access, nil)
	h.expect("reactivated Company can be reselected", http.StatusOK, http.MethodPost, "/select-holding", access, exactWorkspace)

	rotatedAccess, rotatedRefresh := h.refresh("refresh rotation", refresh)
	h.expect("refresh replay rejected", http.StatusUnauthorized, http.MethodPost, "/refresh", "", bson.M{"token": refresh})
	h.expect("refresh replay revokes family", http.StatusUnauthorized, http.MethodGet, "/verify-token", rotatedAccess, nil)
	_ = rotatedRefresh

	oldAccess, oldRefresh := h.login("second login")
	latestAccess, latestRefresh := h.refresh("rotate before logout", oldRefresh)
	h.expect("logout with previous access", http.StatusOK, http.MethodPost, "/logout", oldAccess, nil)
	h.expect("logout revokes rotated access", http.StatusUnauthorized, http.MethodGet, "/verify-token", latestAccess, nil)
	h.expect("logout revokes rotated refresh", http.StatusUnauthorized, http.MethodPost, "/refresh", "", bson.M{"token": latestRefresh})
}

func newLocalUATHarness(t *testing.T) *localUATHarness {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	t.Cleanup(cancel)

	suffix := randomHex(t, 5)
	runID := "uatperm-" + suffix
	mongoURI := envOr("BCAI_UAT_MONGO_URI", defaultMongoURI)
	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI).SetServerSelectionTimeout(10*time.Second))
	if err != nil {
		t.Fatalf("connect Local MongoDB: %v", err)
	}
	if err := mongoClient.Ping(ctx, nil); err != nil {
		t.Fatalf("ping Local MongoDB: %v", err)
	}
	t.Cleanup(func() { _ = mongoClient.Disconnect(context.Background()) })

	redisClient := redis.NewClient(&redis.Options{Addr: envOr("BCAI_UAT_REDIS_ADDR", defaultRedis)})
	if err := redisClient.Ping(ctx).Err(); err != nil {
		t.Fatalf("ping Local Redis: %v", err)
	}
	t.Cleanup(func() { _ = redisClient.Close() })

	return &localUATHarness{
		t:            t,
		ctx:          ctx,
		apiURL:       strings.TrimRight(envOr("BCAI_UAT_API_URL", defaultAPIURL), "/"),
		http:         &http.Client{Timeout: 12 * time.Second},
		db:           mongoClient.Database("appdb"),
		redis:        redisClient,
		runID:        runID,
		username:     "u" + suffix + "user",
		password:     "UatOnly!" + suffix + "Safe",
		userUID:      runID + "-uid",
		holdingCode:  "u" + suffix,
		holdingUID:   runID + "-holdinguid",
		companyCode:  "UAT" + strings.ToUpper(suffix),
		companyUID:   runID + "-companyuid",
		branchOneUID: runID + "-branch-one-uid",
		branchTwoUID: runID + "-branch-two-uid",
		membershipID: runID + "-membershipuid",
		createCode:   "c" + suffix,
		tokens:       map[string]struct{}{},
		refreshes:    map[string]struct{}{},
		sessionUIDs:  map[string]struct{}{},
	}
}

func (h *localUATHarness) seed() {
	h.t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(h.password), bcrypt.MinCost)
	if err != nil {
		h.t.Fatalf("hash UAT password: %v", err)
	}
	now := time.Now().UTC()
	docs := []struct {
		collection string
		doc        bson.M
	}{
		{"users", bson.M{
			"uatrunid": h.runID, "guidfixed": h.userUID, "uid": h.userUID,
			"username": h.username, "email": h.username + "@example.invalid", "password": string(hash),
			"name": "UAT Permission User", "isdeleted": false, "__v": int64(0), "createdat": now,
		}},
		{"googleidentities", bson.M{
			"uatrunid": h.runID, "identityuid": h.runID + "-identity", "useruid": h.userUID,
			"issuer": "https://accounts.google.com", "subject": h.runID + "-subject",
			"verifiedemail": h.username + "@example.invalid", "isactive": true, "linkedat": now,
		}},
		{"shops", bson.M{
			"uatrunid": h.runID, "guidfixed": h.holdingUID, "holdinguid": h.holdingUID,
			"holdingcode": h.holdingCode, "isactive": true, "isdeleted": false, "__v": int64(0),
			"names": bson.A{bson.M{"code": "th", "name": "UAT Holding"}}, "createdat": now, "createdby": h.userUID,
		}},
		{"shopusers", bson.M{
			"uatrunid": h.runID, "membershipuid": h.membershipID, "holdinguid": h.holdingUID,
			"holdingcode": h.holdingCode, "useruid": h.userUID, "username": h.username,
			"role": 0, "permissionversion": int64(1), "isdeleted": false, "isaccessdisabled": false,
			"accessscopes": bson.A{bson.M{"scopetype": "branch", "companyuid": h.companyUID, "branchuid": h.branchOneUID}},
			"createdat":    now, "createdby": h.userUID, "__v": int64(0),
		}},
		{"organizationcompanies", bson.M{
			"uatrunid": h.runID, "holdingcode": h.holdingCode, "guidfixed": h.companyUID,
			"code": h.companyCode, "names": bson.A{bson.M{"code": "th", "name": "UAT Company"}},
			"isactive": true, "__v": int64(0), "createdat": now, "createdby": h.userUID,
		}},
		{"organizationbranches", bson.M{
			"uatrunid": h.runID, "holdingcode": h.holdingCode, "companyguid": h.companyUID,
			"guidfixed": h.branchOneUID, "code": "B1", "names": bson.A{bson.M{"code": "th", "name": "UAT Branch 1"}},
			"timezone": "Asia/Bangkok", "isactive": true, "__v": int64(0), "createdat": now, "createdby": h.userUID,
		}},
		{"organizationbranches", bson.M{
			"uatrunid": h.runID, "holdingcode": h.holdingCode, "companyguid": h.companyUID,
			"guidfixed": h.branchTwoUID, "code": "B2", "names": bson.A{bson.M{"code": "th", "name": "UAT Branch 2"}},
			"timezone": "Asia/Bangkok", "isactive": true, "__v": int64(0), "createdat": now, "createdby": h.userUID,
		}},
	}
	for _, item := range docs {
		if _, err := h.db.Collection(item.collection).InsertOne(h.ctx, item.doc); err != nil {
			h.t.Fatalf("seed %s: %v", item.collection, err)
		}
	}
}

func (h *localUATHarness) testHoldingBootstrap(access string) {
	h.t.Helper()
	status, body := h.request(http.MethodPost, "/create-holding", access, bson.M{
		"holdingcode": h.createCode,
		"names":       bson.A{bson.M{"code": "th", "name": "UAT Atomic Holding"}},
	})
	h.requireStatus("atomic Holding bootstrap", http.StatusOK, status, body)
	createdUID, _ := body["id"].(string)
	if strings.TrimSpace(createdUID) == "" {
		h.t.Fatal("atomic Holding bootstrap returned no stable ID")
	}
	h.createdUID = createdUID

	assertCount := func(collection string, filter bson.M, want int64) {
		h.t.Helper()
		got, err := h.db.Collection(collection).CountDocuments(h.ctx, filter)
		if err != nil || got != want {
			h.t.Fatalf("%s count = %d, want %d; err=%v", collection, got, want, err)
		}
	}
	assertCount("shops", bson.M{"holdingcode": h.createCode, "guidfixed": createdUID}, 1)
	assertCount("shopusers", bson.M{"holdingcode": h.createCode, "holdinguid": createdUID, "useruid": h.userUID, "role": 2}, 1)
	assertCount("organizationcodeclaims", bson.M{"entitytype": "holding", "scopeuid": "global", "normalizedcode": h.createCode, "entityuid": createdUID}, 1)
	assertCount("organizationaudits", bson.M{"action": "holding.created", "targetuid": createdUID}, 1)
	assertCount("outboxevents", bson.M{"eventtype": "holding.created", "aggregateuid": createdUID, "status": "PENDING"}, 1)

	// Local/DEV only: retire the exact fixture but keep its code claim, then prove
	// the historical business code cannot be reused and no partial bootstrap remains.
	_, _ = h.db.Collection("shops").DeleteMany(h.ctx, bson.M{"holdingcode": h.createCode, "guidfixed": createdUID})
	_, _ = h.db.Collection("shopusers").DeleteMany(h.ctx, bson.M{"holdingcode": h.createCode, "holdinguid": createdUID, "useruid": h.userUID})
	status, _ = h.request(http.MethodPost, "/create-holding", access, bson.M{
		"holdingcode": h.createCode,
		"names":       bson.A{bson.M{"code": "th", "name": "UAT Reused Holding"}},
	})
	if status >= 200 && status < 300 {
		h.t.Fatalf("retired Holding code was reused: status=%d", status)
	}
	assertCount("shops", bson.M{"holdingcode": h.createCode}, 0)
	assertCount("shopusers", bson.M{"holdingcode": h.createCode, "useruid": h.userUID}, 0)
}

func (h *localUATHarness) login(name string) (string, string) {
	h.t.Helper()
	status, body := h.request(http.MethodPost, "/login", "", bson.M{"username": h.username, "password": h.password})
	h.requireStatus(name, http.StatusOK, status, body)
	return h.rememberTokens(body)
}

func (h *localUATHarness) refresh(name, refresh string) (string, string) {
	h.t.Helper()
	status, body := h.request(http.MethodPost, "/refresh", "", bson.M{"token": refresh})
	h.requireStatus(name, http.StatusOK, status, body)
	return h.rememberTokens(body)
}

func (h *localUATHarness) rememberTokens(body map[string]interface{}) (string, string) {
	h.t.Helper()
	access, _ := body["token"].(string)
	refresh, _ := body["refresh"].(string)
	if access == "" || refresh == "" {
		h.t.Fatal("authentication response omitted access or refresh token")
	}
	h.tokens[access] = struct{}{}
	h.refreshes[refresh] = struct{}{}
	if sessionUID, err := h.redis.HGet(h.ctx, "auth-"+access, "sessionuid").Result(); err == nil && sessionUID != "" {
		h.sessionUIDs[sessionUID] = struct{}{}
	}
	return access, refresh
}

func (h *localUATHarness) request(method, path, access string, payload interface{}) (int, map[string]interface{}) {
	h.t.Helper()
	var body bytes.Buffer
	if payload != nil {
		if err := json.NewEncoder(&body).Encode(payload); err != nil {
			h.t.Fatalf("encode %s %s: %v", method, path, err)
		}
	}
	req, err := http.NewRequestWithContext(h.ctx, method, h.apiURL+path, &body)
	if err != nil {
		h.t.Fatalf("build %s %s: %v", method, path, err)
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if access != "" {
		req.Header.Set("Authorization", "Bearer "+access)
	}
	resp, err := h.http.Do(req)
	if err != nil {
		h.t.Fatalf("call %s %s: %v", method, path, err)
	}
	defer resp.Body.Close()
	decoded := map[string]interface{}{}
	_ = json.NewDecoder(resp.Body).Decode(&decoded)
	return resp.StatusCode, decoded
}

func (h *localUATHarness) requireStatus(name string, want int, got int, _ map[string]interface{}) {
	h.t.Helper()
	if got != want {
		h.t.Fatalf("%s status = %d, want %d", name, got, want)
	}
}

func (h *localUATHarness) expect(name string, want int, method, path, access string, payload interface{}) {
	h.t.Helper()
	status, body := h.request(method, path, access, payload)
	h.requireStatus(name, want, status, body)
}

func (h *localUATHarness) updateMembership(fields bson.M) {
	h.t.Helper()
	result, err := h.db.Collection("shopusers").UpdateOne(h.ctx, bson.M{"uatrunid": h.runID}, bson.M{"$set": fields})
	if err != nil || result.MatchedCount != 1 {
		h.t.Fatalf("update Membership fixture: matched=%d err=%v", result.MatchedCount, err)
	}
}

func (h *localUATHarness) updateOne(collection string, fields bson.M) {
	h.t.Helper()
	result, err := h.db.Collection(collection).UpdateOne(h.ctx, bson.M{"uatrunid": h.runID}, bson.M{"$set": fields})
	if err != nil || result.MatchedCount != 1 {
		h.t.Fatalf("update %s fixture: matched=%d err=%v", collection, result.MatchedCount, err)
	}
}

func (h *localUATHarness) cleanup() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	marker := bson.M{"uatrunid": h.runID}
	for _, collection := range []string{"users", "googleidentities", "shopusers", "shops", "organizationcompanies", "organizationbranches"} {
		_, _ = h.db.Collection(collection).DeleteMany(ctx, marker)
	}
	_, _ = h.db.Collection("shopuseraccesslogs").DeleteMany(ctx, bson.M{"$or": bson.A{bson.M{"username": h.username}, bson.M{"holdingcode": h.holdingCode}, bson.M{"holdingcode": h.createCode}}})
	_, _ = h.db.Collection("authaudits").DeleteMany(ctx, bson.M{"useruid": h.userUID})
	_, _ = h.db.Collection("organizationaudits").DeleteMany(ctx, bson.M{"$or": bson.A{bson.M{"actoruid": h.userUID}, bson.M{"targetuid": h.createdUID}}})
	_, _ = h.db.Collection("outboxevents").DeleteMany(ctx, bson.M{"aggregateuid": h.createdUID})
	_, _ = h.db.Collection("organizationcodeclaims").DeleteMany(ctx, bson.M{"$or": bson.A{bson.M{"entityuid": h.createdUID}, bson.M{"claimedby": h.userUID}}})

	keys := map[string]struct{}{"user:" + h.username: {}}
	for token := range h.tokens {
		keys["auth-"+token] = struct{}{}
	}
	for refresh := range h.refreshes {
		keys["refresh-"+refresh] = struct{}{}
		keys["refresh-used-"+refresh] = struct{}{}
	}
	for sessionUID := range h.sessionUIDs {
		keys["session-"+sessionUID] = struct{}{}
		keys["session-revoked-"+sessionUID] = struct{}{}
	}
	var cursor uint64
	for {
		batch, next, err := h.redis.Scan(ctx, cursor, "*", 200).Result()
		if err != nil {
			break
		}
		for _, key := range batch {
			if strings.Contains(key, h.runID) {
				keys[key] = struct{}{}
				continue
			}
			if kind, _ := h.redis.Type(ctx, key).Result(); kind == "hash" {
				values, _ := h.redis.HVals(ctx, key).Result()
				for _, value := range values {
					if strings.Contains(value, h.runID) {
						keys[key] = struct{}{}
						break
					}
				}
			}
		}
		cursor = next
		if cursor == 0 {
			break
		}
	}
	if len(keys) > 0 {
		keyList := make([]string, 0, len(keys))
		for key := range keys {
			keyList = append(keyList, key)
		}
		_ = h.redis.Del(ctx, keyList...).Err()
	}

	for _, collection := range []string{"users", "googleidentities", "shopusers", "shops", "organizationcompanies", "organizationbranches"} {
		if count, err := h.db.Collection(collection).CountDocuments(ctx, marker); err != nil || count != 0 {
			h.t.Errorf("cleanup %s count = %d, want 0; err=%v", collection, count, err)
		}
	}
}

func randomHex(t *testing.T, size int) string {
	t.Helper()
	raw := make([]byte, size)
	if _, err := rand.Read(raw); err != nil {
		t.Fatalf("generate UAT marker: %v", err)
	}
	return hex.EncodeToString(raw)
}

func envOr(name, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}

func TestLocalUATRequiresExplicitEnvironmentConfirmation(t *testing.T) {
	if os.Getenv("RUN_LOCAL_UAT") == "1" && os.Getenv("BCAI_UAT_LOCAL_DEV") == "1" {
		return
	}
	if defaultAPIURL == "" || defaultMongoURI == "" || defaultRedis == "" {
		t.Fatal(fmt.Errorf("local UAT defaults must be explicit"))
	}
}
