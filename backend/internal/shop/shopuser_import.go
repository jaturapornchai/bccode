package shop

import (
	"bytes"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"errors"
	"net/http"
	"net/mail"
	"strings"

	"smlcloudplatform/internal/authentication/models"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/pkg/microservice"

	"github.com/xuri/excelize/v2"
)

// importUsersRequest is the JSON body for POST /holding/users/import. The file is sent as
// base64 (the mainapi microservice framework reads a JSON body, not multipart). commit=false
// returns a validated PREVIEW only; commit=true writes the valid rows.
type importUsersRequest struct {
	HoldingCode   string `json:"holdingcode"`
	Filename      string `json:"filename"`
	ContentBase64 string `json:"contentbase64"`
	Commit        bool   `json:"commit"`
}

// importUserRow is one parsed/validated row, returned for preview and updated on commit.
type importUserRow struct {
	Row        int    `json:"row"`
	Username   string `json:"username"`
	Email      string `json:"email"`
	Name       string `json:"name"`
	Role       uint8  `json:"role"`
	Position   string `json:"position"`
	Department string `json:"department"`
	Valid      bool   `json:"valid"`
	Message    string `json:"message"`
	Imported   bool   `json:"imported"`
}

type importUsersResult struct {
	Total    int             `json:"total"`
	Valid    int             `json:"valid"`
	Imported int             `json:"imported"`
	Failed   int             `json:"failed"`
	Rows     []importUserRow `json:"rows"`
}

// ImportHoldingUsers bulk-imports users into a holding from an uploaded .csv/.xlsx file.
// Two-phase: commit=false validates + returns a preview; commit=true persists valid rows via
// the existing SaveUserFullProfile path. Authorization is per-holding (OWNER/ADMIN), not the
// JWT-selected role.
func (h ShopMemberHttp) ImportHoldingUsers(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	holdingCode := userInfo.HoldingCode

	var req importUsersRequest
	if err := json.Unmarshal([]byte(ctx.ReadInput()), &req); err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}
	if strings.TrimSpace(req.HoldingCode) != "" {
		holdingCode = strings.TrimSpace(req.HoldingCode)
	}

	if err := h.svc.EnsureHoldingManager(holdingCode, authUsername); err != nil {
		ctx.Response(http.StatusOK, &common.ApiResponse{Success: false, Message: "permission denied"})
		return err
	}

	data, err := base64.StdEncoding.DecodeString(strings.TrimSpace(req.ContentBase64))
	if err != nil {
		h.ms.Logger.Error("ImportHoldingUsers base64 decode failed: " + err.Error())
		ctx.ResponseError(http.StatusBadRequest, "invalid file content")
		return err
	}
	const maxImportBytes = 5 << 20 // 5MB — guard against oversized / zip-bomb uploads before parsing
	if len(data) > maxImportBytes {
		ctx.ResponseError(http.StatusBadRequest, "ไฟล์ใหญ่เกินไป (สูงสุด 5MB)")
		return errors.New("import file too large")
	}

	rows, err := parseImportUsers(req.Filename, data)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	result := importUsersResult{Total: len(rows), Rows: rows}
	for i := range result.Rows {
		if result.Rows[i].Valid {
			result.Valid++
		}
	}

	if req.Commit {
		for i := range result.Rows {
			r := &result.Rows[i]
			if !r.Valid {
				continue
			}
			saveErr := h.svc.SaveUserFullProfile(holdingCode, authUsername, &models.UserRoleRequest{
				Username:        r.Username,
				Email:           r.Email,
				UserProfileName: r.Name,
				Role:            models.UserRole(r.Role),
				Position:        r.Position,
				Department:      r.Department,
			})
			if saveErr != nil {
				r.Imported = false
				r.Message = saveErr.Error()
				result.Failed++
			} else {
				r.Imported = true
				result.Imported++
			}
		}
	}

	ctx.Response(http.StatusOK, common.ApiResponse{Success: true, Data: result})
	return nil
}

// parseImportUsers decodes a .csv or .xlsx user list. Header row is matched case-insensitively
// in Thai or English: usercode/username (required), email (optional), name, role, position, department.
func parseImportUsers(filename string, data []byte) ([]importUserRow, error) {
	var records [][]string
	lower := strings.ToLower(strings.TrimSpace(filename))
	switch {
	case strings.HasSuffix(lower, ".csv"), strings.HasSuffix(lower, ".txt"):
		reader := csv.NewReader(bytes.NewReader(data))
		reader.FieldsPerRecord = -1
		recs, err := reader.ReadAll()
		if err != nil {
			return nil, errors.New("อ่านไฟล์ CSV ไม่สำเร็จ")
		}
		records = recs
	case strings.HasSuffix(lower, ".xlsx"), strings.HasSuffix(lower, ".xls"):
		f, err := excelize.OpenReader(bytes.NewReader(data))
		if err != nil {
			return nil, errors.New("อ่านไฟล์ Excel ไม่สำเร็จ")
		}
		defer f.Close()
		sheets := f.GetSheetList()
		if len(sheets) == 0 {
			return nil, errors.New("ไฟล์ Excel ไม่มีชีตข้อมูล")
		}
		recs, err := f.GetRows(sheets[0])
		if err != nil {
			return nil, errors.New("อ่านชีต Excel ไม่สำเร็จ")
		}
		records = recs
	default:
		return nil, errors.New("รองรับเฉพาะไฟล์ .csv หรือ .xlsx")
	}

	if len(records) < 2 {
		return nil, errors.New("ไฟล์ต้องมีหัวตารางและข้อมูลอย่างน้อย 1 แถว")
	}
	if len(records) > 2001 {
		return nil, errors.New("นำเข้าได้สูงสุด 2000 แถวต่อครั้ง")
	}

	idx := mapImportColumns(records[0])
	if idx.usercode < 0 {
		return nil, errors.New("ไม่พบคอลัมน์ รหัสผู้ใช้/usercode ในหัวตาราง")
	}

	seenUserCodes := map[string]bool{}
	seenEmails := map[string]bool{}
	out := make([]importUserRow, 0, len(records)-1)
	for i := 1; i < len(records); i++ {
		rec := records[i]
		cell := func(c int) string {
			if c >= 0 && c < len(rec) {
				return strings.TrimSpace(rec[c])
			}
			return ""
		}
		usercode := utils.NormalizeUsername(cell(idx.usercode))
		email := utils.NormalizeEmail(cell(idx.email))
		row := importUserRow{
			Row:        i + 1,
			Username:   usercode,
			Email:      email,
			Name:       cell(idx.name),
			Position:   cell(idx.position),
			Department: cell(idx.department),
			Role:       parseImportRole(cell(idx.role)),
		}
		switch {
		case usercode == "" && email == "" && row.Name == "" && cell(idx.role) == "":
			continue // skip fully-blank rows silently
		case usercode == "":
			row.Valid, row.Message = false, "ไม่มีรหัสผู้ใช้"
		case email != "" && !isValidEmail(email):
			row.Valid, row.Message = false, "อีเมลไม่ถูกต้อง"
		case seenUserCodes[usercode]:
			row.Valid, row.Message = false, "รหัสผู้ใช้ซ้ำในไฟล์"
		case email != "" && seenEmails[email]:
			row.Valid, row.Message = false, "อีเมลซ้ำในไฟล์"
		default:
			seenUserCodes[usercode] = true
			if email != "" {
				seenEmails[email] = true
			}
			row.Valid = true
		}
		out = append(out, row)
	}
	if len(out) == 0 {
		return nil, errors.New("ไม่พบข้อมูลผู้ใช้ในไฟล์")
	}
	return out, nil
}

type importColIndex struct{ usercode, email, name, role, position, department int }

func mapImportColumns(header []string) importColIndex {
	idx := importColIndex{usercode: -1, email: -1, name: -1, role: -1, position: -1, department: -1}
	for i, h := range header {
		key := strings.ToLower(strings.TrimSpace(h))
		switch {
		case idx.usercode < 0 && containsAny(key, "usercode", "username", "รหัสผู้ใช้", "ชื่อเข้าสู่ระบบ"):
			idx.usercode = i
		case idx.email < 0 && containsAny(key, "email", "อีเมล"):
			idx.email = i
		case idx.name < 0 && containsAny(key, "name", "ชื่อ", "นามสกุล"):
			idx.name = i
		case idx.role < 0 && containsAny(key, "role", "สิทธิ", "ระดับ", "บทบาท"):
			idx.role = i
		case idx.position < 0 && containsAny(key, "position", "ตำแหน่ง"):
			idx.position = i
		case idx.department < 0 && containsAny(key, "department", "แผนก", "dept"):
			idx.department = i
		}
	}
	return idx
}

func containsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

func isValidEmail(s string) bool {
	_, err := mail.ParseAddress(s)
	return err == nil
}

// parseImportRole maps a cell to a role. Accepts 0/1/2, English (user/admin/owner) or Thai
// (ผู้ใช้/แอดมิน/เจ้าของ). Defaults to ROLE_USER (least privilege) when blank or unknown.
func parseImportRole(s string) uint8 {
	v := strings.ToLower(strings.TrimSpace(s))
	switch {
	case v == "2" || containsAny(v, "owner", "เจ้าของ"):
		return uint8(models.ROLE_OWNER)
	case v == "1" || containsAny(v, "admin", "แอดมิน", "ผู้ดูแล"):
		return uint8(models.ROLE_ADMIN)
	default:
		return uint8(models.ROLE_USER)
	}
}
