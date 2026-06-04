# Master Data Validation - Testing Guide

## Overview
This document describes how to test the strict validation feature for master data in the product import system.

## What Changed
The system now validates that all master data codes (Group, Brand, Design, etc.) exist in the system before allowing product creation or updates. If any master data is missing, the import will fail with a clear error message.

---

## Test Scenarios

### 1. ✅ Happy Path - All Master Data Exists

**Setup:**
1. Create master data in the system:
   - Group: CODE="GROUP001", Name="กลุ่มสินค้า A"
   - Brand: CODE="BRAND001", Name="แบรนด์ A"
   - Design: CODE="DESIGN001", Name="ดีไซน์ A"

**Test Steps:**
1. Create Excel file with product:
   ```
   Barcode     | Name         | GroupCode | BrandCode | DesignCode | Price
   1234567890  | สินค้า A     | GROUP001  | BRAND001  | DESIGN001  | 100
   ```
2. Upload Excel via `POST /productimport/upload`
3. Apply changes via `POST /productimport/{task-id}/apply`

**Expected Result:**
```json
{
  "success": true,
  "message": "Changes applied successfully with mode: AUTO"
}
```

Product should be created with:
- `groupcode`: "GROUP001"
- `groupguid`: <actual guid>
- `groupnames`: [{"code": "th", "name": "กลุ่มสินค้า A"}]
- Similarly for Brand and Design

---

### 2. ❌ Missing Single Master Data

**Setup:**
1. Create only:
   - Group: CODE="GROUP001"
   - Brand: CODE="BRAND001"
2. **Do NOT create** Design: CODE="DESIGN999"

**Test Steps:**
1. Create Excel with missing master data:
   ```
   Barcode     | Name         | GroupCode | BrandCode | DesignCode | Price
   1234567890  | สินค้า A     | GROUP001  | BRAND001  | DESIGN999  | 100
   ```
2. Upload and Apply

**Expected Result:**
```json
{
  "success": false,
  "message": "Completed with 1 errors out of 1 records",
  "errors": [
    "Barcode '1234567890': missing master data: Design code 'DESIGN999' not found"
  ]
}
```

Product should **NOT** be created.

---

### 3. ❌ Missing Multiple Master Data

**Setup:**
1. Create only:
   - Group: CODE="GROUP001"
2. **Do NOT create:**
   - Brand: CODE="BRAND999"
   - Design: CODE="DESIGN999"

**Test Steps:**
1. Create Excel with multiple missing codes:
   ```
   Barcode     | Name         | GroupCode | BrandCode | DesignCode | Price
   1234567890  | สินค้า A     | GROUP001  | BRAND999  | DESIGN999  | 100
   ```
2. Upload and Apply

**Expected Result:**
```json
{
  "success": false,
  "errors": [
    "Barcode '1234567890': missing master data: Brand code 'BRAND999' not found, Design code 'DESIGN999' not found"
  ]
}
```

Error message should list **all** missing master data.

---

### 4. ✅ Partial Success - Batch Import

**Setup:**
1. Create master data:
   - Group: CODE="GROUP001"
   - Brand: CODE="BRAND001"
2. **Do NOT create:**
   - Design: CODE="DESIGN999"

**Test Steps:**
1. Create Excel with 3 products:
   ```
   Barcode     | Name         | GroupCode | BrandCode | DesignCode | Price
   1111111111  | สินค้า 1     | GROUP001  | BRAND001  |            | 100
   2222222222  | สินค้า 2     | GROUP001  | BRAND001  | DESIGN999  | 200
   3333333333  | สินค้า 3     | GROUP001  | BRAND001  |            | 300
   ```
2. Upload and Apply

**Expected Result:**
```json
{
  "success": false,
  "message": "Completed with 1 errors out of 3 records",
  "errors": [
    "Barcode '2222222222': missing master data: Design code 'DESIGN999' not found"
  ]
}
```

- Product 1111111111: ✅ Created successfully
- Product 2222222222: ❌ Failed (missing Design)
- Product 3333333333: ✅ Created successfully

---

### 5. ❌ Update Product with Missing Master Data

**Setup:**
1. Create product with barcode "1234567890"
   - GroupCode: "GROUP001"
   - BrandCode: "BRAND001"
2. Create only GROUP002 (not BRAND999)

**Test Steps:**
1. Create Excel to update existing product:
   ```
   Barcode     | Name           | GroupCode | BrandCode | Price
   1234567890  | สินค้า Updated | GROUP002  | BRAND999  | 150
   ```
2. Upload and Apply (with UPDATE_ONLY or AUTO mode)

**Expected Result:**
```json
{
  "success": false,
  "errors": [
    "Barcode '1234567890': brand code 'BRAND999' not found in master data"
  ]
}
```

Original product should remain unchanged.

---

### 6. ✅ Empty Codes (Should Allow)

**Test Steps:**
1. Create Excel with empty codes:
   ```
   Barcode     | Name         | GroupCode | BrandCode | DesignCode | Price
   1234567890  | สินค้า A     |           |           |            | 100
   ```
2. Upload and Apply

**Expected Result:**
```json
{
  "success": true
}
```

Product created with:
- `groupcode`: ""
- `groupguid`: ""
- `groupnames`: null
- Empty codes are allowed and do not trigger validation

---

### 7. ❌ All 10 Master Data Types Missing

**Test Steps:**
1. Create Excel with all types set to non-existent codes:
   ```
   Barcode     | GroupCode  | GroupsuboneCode | GroupsubtwoCode | BrandCode  | DesignCode | ModelCode  | PatternCode | GradeCode  | CategoryCode | ClassCode
   1234567890  | GROUP999   | SUB999          | SUB2999         | BRAND999   | DESIGN999  | MODEL999   | PATTERN999  | GRADE999   | CAT999       | CLASS999
   ```
2. Upload and Apply

**Expected Result:**
Error message should list all 10 missing codes:
```
Barcode '1234567890': missing master data:
Group code 'GROUP999' not found,
Groupsubone code 'SUB999' not found,
Groupsubtwo code 'SUB2999' not found,
Brand code 'BRAND999' not found,
Design code 'DESIGN999' not found,
Model code 'MODEL999' not found,
Pattern code 'PATTERN999' not found,
Grade code 'GRADE999' not found,
Category code 'CAT999' not found,
Class code 'CLASS999' not found
```

---

## Manual Testing with Postman

### Step 1: Upload Excel
```http
POST /productimport/upload
Content-Type: multipart/form-data
Authorization: Bearer <token>

file: <excel-file>
```

Response:
```json
{
  "success": true,
  "id": "TASK_ABC123"
}
```

### Step 2: Check Task Status (Optional)
```http
GET /productimport/task/TASK_ABC123/status
```

### Step 3: Apply Changes
```http
POST /productimport/TASK_ABC123/apply
Content-Type: application/json

{
  "import_mode": "AUTO",
  "force_update": false
}
```

### Step 4: Check Results
If successful:
```json
{
  "success": true,
  "message": "Successfully processed 100 records"
}
```

If errors:
```json
{
  "success": false,
  "message": "Completed with 5 errors out of 100 records",
  "errors": [
    "Barcode '1234567890': missing master data: Group code 'GROUP999' not found",
    ...
  ]
}
```

---

## Testing with Different Import Modes

### INSERT_ONLY Mode
```json
{
  "import_mode": "INSERT_ONLY"
}
```
- Only validates master data for new products
- Skips existing products

### UPDATE_ONLY Mode
```json
{
  "import_mode": "UPDATE_ONLY"
}
```
- Only validates master data for updated products
- Skips new products

### AUTO Mode (Default)
```json
{
  "import_mode": "AUTO"
}
```
- Validates for both new and updated products

---

## Expected Error Message Format

### For Insert (New Product):
```
barcode {barcode}: missing master data: {list of missing codes}
```

Example:
```
barcode 1234567890: missing master data: Group code 'GROUP999' not found, Brand code 'BRAND888' not found
```

### For Update (Existing Product):
```
barcode {barcode}: {field} code '{code}' not found in master data
```

Example:
```
barcode 1234567890: brand code 'BRAND999' not found in master data
```

---

## Verification Checklist

After running tests, verify:

- [ ] Import succeeds when all master data exists
- [ ] Import fails with clear error when master data is missing
- [ ] Error message lists all missing codes
- [ ] Partial success works (some pass, some fail)
- [ ] Empty codes are allowed (no validation)
- [ ] All 10 master data types are validated:
  - [ ] Group
  - [ ] Groupsubone
  - [ ] Groupsubtwo
  - [ ] Brand
  - [ ] Design
  - [ ] Model
  - [ ] Pattern
  - [ ] Grade
  - [ ] Category
  - [ ] Class
- [ ] Update mode validates correctly
- [ ] Batch processing shows correct error count
- [ ] Successful products are created despite errors in batch

---

## Database Verification

After successful import, verify in MongoDB:

```javascript
// Check product was created with master data
db.productbarcode.findOne({barcode: "1234567890"})

// Should have:
{
  barcode: "1234567890",
  groupcode: "GROUP001",
  groupguid: "550e8400-...",  // Not empty
  groupnames: [{code: "th", name: "กลุ่มสินค้า A"}],  // Not null
  brandcode: "BRAND001",
  brandguid: "550e8400-...",  // Not empty
  brandnames: [{code: "th", name: "แบรนด์ A"}],  // Not null
}
```

For failed imports:
```javascript
// Product should NOT exist
db.productbarcode.findOne({barcode: "FAILED_BARCODE"})
// Result: null
```

---

## Performance Testing

Test with large Excel files:

1. **Small batch** (10 products):
   - All with valid master data → Should complete in < 1 second

2. **Medium batch** (100 products):
   - 50 valid, 50 with missing master data → Should complete in < 5 seconds
   - Should report 50 errors clearly

3. **Large batch** (1000 products):
   - Use async mode: `?async=true`
   - Check progress via `/task/{id}/status`
   - Verify error summary shows first 10 errors

---

## Troubleshooting

### Issue: "Missing master data" but code exists in system

**Check:**
1. Code is exactly the same (case-sensitive)
2. Code belongs to correct holding_code
3. Master data is not soft-deleted

### Issue: Import succeeds but master data fields are empty

**This means:**
- You're using old code (before validation was added)
- Update to latest code with strict validation

### Issue: Too many errors, cannot see all

**Solution:**
- System shows first 10 errors by default
- Check logs for complete error list
- Fix master data and retry

---

## Rollback Plan

If validation causes issues:

1. **Temporary fix**: Import master data first
2. **If needed**: Can modify validation to be warnings instead of errors
3. **Contact**: Development team for assistance

---

## Additional Notes

- Validation only applies when using cache (batch mode)
- Fallback mode (no cache) also validates but slower
- Empty/null codes are always allowed
- Validation happens before database write (no partial data)
