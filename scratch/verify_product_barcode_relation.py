import requests
import json
import sys

sys.stdout.reconfigure(encoding='utf-8')

BASE_URL = "http://127.0.0.1:8888"

print("Starting Product <-> Barcode relation verification...")

# 1. Login or Register
print("\n1. Logging in / Registering...")
try:
    reg_payload = {
        "username": "uat_test_user@gmail.com",
        "password": "password123",
        "name": "UAT Test User"
    }
    reg_resp = requests.post(f"{BASE_URL}/register-username", json=reg_payload)
    print("Register response status:", reg_resp.status_code)

    resp = requests.post(f"{BASE_URL}/login", json={
        "username": reg_payload["username"],
        "password": reg_payload["password"]
    })
    login_data = resp.json()

    if not login_data.get("success") or "token" not in login_data:
        print("Login failed! Response:", login_data)
        sys.exit(1)

    token = login_data["token"]
    print("Logged in successfully.")
except Exception as e:
    print("Error during login:", e)
    sys.exit(1)

headers = {
    "Authorization": f"Bearer {token}",
    "Content-Type": "application/json"
}

# 2. Get shops list or Create a new shop
print("\n2. Getting/Creating a shop...")
holding_code = ""
try:
    resp = requests.get(f"{BASE_URL}/shop", headers=headers)
    shops = resp.json()
    if isinstance(shops, list) and len(shops) > 0:
        holding_code = shops[0]["holding_code"]
    elif isinstance(shops, dict) and isinstance(shops.get("data"), list) and len(shops["data"]) > 0:
        holding_code = shops["data"][0]["holding_code"]
    else:
        shop_payload = {
            "code": "UATSH2",
            "name": "UAT Shop 2",
            "names": [
                {"code": "th", "name": "ร้านทดสอบ UAT 2"},
                {"code": "en", "name": "UAT Shop 2"}
            ]
        }
        resp = requests.post(f"{BASE_URL}/shop", json=shop_payload, headers=headers)
        shop_data = resp.json()
        holding_code = shop_data["id"]

    print(f"Active Holding Code: {holding_code}")
except Exception as e:
    print("Error getting/creating shop:", e)
    sys.exit(1)

# 3. Request Token with holding_code bound
print("\n3. Binding token to Holding Code...")
try:
    resp = requests.post(f"{BASE_URL}/login", json={
        "username": reg_payload["username"],
        "password": reg_payload["password"],
        "holding_code": holding_code
    })
    shop_login_data = resp.json()
    token = shop_login_data["token"]
    headers["Authorization"] = f"Bearer {token}"
    print("Token successfully bound to Shop.")
except Exception as e:
    print("Error binding to shop:", e)
    sys.exit(1)

# 4. Create Product (สินค้าหลัก)
print("\n4. Creating a new Main Product...")
product_payload = {
    "code": "P-E2E-TEST",
    "names": [
        {"code": "th", "name": "สินค้าหลักทดสอบ E2E"},
        {"code": "en", "name": "Main Product E2E Test"}
    ],
    "item_type": 0,
    "vat_type": 0,
    "group_code": "GRP001",
    "group_names": [{"code": "th", "name": "กลุ่มทดสอบ"}],
    "brand_code": "BRAND001",
    "brandnames": [{"code": "th", "name": "ยี่ห้อทดสอบ"}],
    "categorycode": "CAT001",
    "category_names": [{"code": "th", "name": "หมวดทดสอบ"}],
    "manufacturers": [
        {
            "guid_fixed": "mfr-guid-1",
            "code": "MFR001",
            "names": [{"code": "th", "name": "ผู้ผลิตทดสอบ 1"}]
        },
        {
            "guid_fixed": "mfr-guid-2",
            "code": "MFR002",
            "names": [{"code": "th", "name": "ผู้ผลิตทดสอบ 2"}]
        }
    ],
    "suppliers": [
        {
            "guid_fixed": "spl-guid-1",
            "code": "SPL001",
            "names": [{"code": "th", "name": "ผู้จำหน่ายทดสอบ 1"}]
        }
    ]
}

product_guid = ""
try:
    resp = requests.post(f"{BASE_URL}/product", json=product_payload, headers=headers)
    print("Create Product Status:", resp.status_code)
    prod_data = resp.json()
    if resp.status_code not in [200, 201]:
        print("Create product failed:", prod_data)
        sys.exit(1)

    data_obj = prod_data.get("data", {})
    product_guid = data_obj.get("guidfixed") or data_obj.get("guid_fixed") or data_obj.get("guid") or ""
    print(f"Created Product GUID: {product_guid}")
except Exception as e:
    print("Error creating product:", e)
    sys.exit(1)

# 5. Create Barcode linked to Main Product
print("\n5. Creating a new Product Barcode linked to the Main Product...")
barcode_payload = {
    "barcode": "B-E2E-TEST-001",
    "itemcode": "P-E2E-TEST",
    "item_guid": product_guid,
    "names": [
        {"code": "th", "name": "บาร์โค้ดทดสอบ E2E 1"},
        {"code": "en", "name": "Barcode E2E Test 1"}
    ],
    "price": 150.0,
    "unitcode": "PCS",
    "unitnames": [{"code": "th", "name": "ชิ้น"}],
}

barcode_guid = ""
try:
    resp = requests.post(f"{BASE_URL}/product/barcode", json=barcode_payload, headers=headers)
    print("Create Barcode Status:", resp.status_code)
    bar_data = resp.json()
    if resp.status_code not in [200, 201]:
        print("Create barcode failed:", bar_data)
        sys.exit(1)

    # Check return field for guid
    barcode_guid = bar_data.get("id") or bar_data.get("guidfixed") or bar_data.get("guid_fixed") or ""
    print(f"Created Barcode GUID: {barcode_guid}")
except Exception as e:
    print("Error creating barcode:", e)
    sys.exit(1)

# 6. Retrieve Barcode Detail and check fallback mappings
print("\n6. Retrieving Barcode info and validating fallback fields from Main Product...")
try:
    # Query barcode details
    resp = requests.get(f"{BASE_URL}/product/barcode/{barcode_guid}", headers=headers)
    print("Get Barcode Status:", resp.status_code)
    bar_info = resp.json()

    if resp.status_code != 200:
        print("Get barcode info failed:", bar_info)
        sys.exit(1)

    data = bar_info["data"]

    # Check Classifications
    print("\nChecking Classifications fallback:")
    print(f"- Group Code: {data.get('group_code')} (Expected: GRP001)")
    print(f"- Brand Code: {data.get('brand_code')} (Expected: BRAND001)")
    print(f"- Category Code: {data.get('categorycode')} (Expected: CAT001)")

    assert data.get('group_code') == "GRP001", "Group code mismatch"
    assert data.get('brand_code') == "BRAND001", "Brand code mismatch"
    assert data.get('categorycode') == "CAT001", "Category code mismatch"
    print("Classifications match successfully!")

    # Check Multi-select Manufacturers & Suppliers fallback
    print("\nChecking Manufacturers & Suppliers fallback:")
    manufacturers = data.get('manufacturers', [])
    suppliers = data.get('suppliers', [])

    print(f"- Manufacturers length: {len(manufacturers)} (Expected: 2)")
    print(f"- Suppliers length: {len(suppliers)} (Expected: 1)")

    assert len(manufacturers) == 2, "Manufacturers length mismatch"
    assert len(suppliers) == 1, "Suppliers length mismatch"

    m_codes = [m["code"] for m in manufacturers]
    s_codes = [s["code"] for s in suppliers]

    print(f"- Manufacturers codes: {m_codes} (Expected: ['MFR001', 'MFR002'])")
    print(f"- Suppliers codes: {s_codes} (Expected: ['SPL001'])")

    assert "MFR001" in m_codes and "MFR002" in m_codes, "Manufacturers codes mismatch"
    assert "SPL001" in s_codes, "Suppliers codes mismatch"
    print("Manufacturers & Suppliers fallback validated successfully!")

except Exception as e:
    print("Error verifying barcode details:", e)
    sys.exit(1)

# 7. Cleanup
print("\n7. Cleaning up test data...")
try:
    del_bar = requests.delete(f"{BASE_URL}/product/barcode/{barcode_guid}", headers=headers)
    print(f"- Deleted Barcode Status: {del_bar.status_code}")

    del_prod = requests.delete(f"{BASE_URL}/product/{product_guid}", headers=headers)
    print(f"- Deleted Product Status: {del_prod.status_code}")

except Exception as e:
    print("Error cleaning up:", e)

print("\nProduct <-> Barcode Relation Verification SUCCESSFUL!")
