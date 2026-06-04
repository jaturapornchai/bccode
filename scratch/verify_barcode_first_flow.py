import requests
import json
import sys

def main():
    print("Starting Barcode-First E2E Flow Verification...")

    BASE_URL = "http://127.0.0.1:8888"

    # 1. Login / Register
    print("\n1. Logging in / Registering...")
    try:
        reg_payload = {
            "username": "uat_test_user_bf@gmail.com",
            "password": "password123",
            "name": "UAT Test User BF"
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

    # 2. Get active shop
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
                "code": "UATSH3",
                "name": "UAT Shop 3",
                "names": [
                    {"code": "th", "name": "ร้านทดสอบ UAT 3"},
                    {"code": "en", "name": "UAT Shop 3"}
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

    # 4. Create an independent barcode (unlinked)
    print("\n4. Creating unlinked barcode...")
    barcode_payload = {
        "barcode": "B-E2E-BF-001",
        "names": [
            {"code": "th", "name": "บาร์โค้ดสร้างก่อนสินค้าหลัก"},
            {"code": "en", "name": "Barcode Created Before Product"}
        ],
        "price": 99.0,
        "unitcode": "PCS",
        "unitnames": [{"code": "th", "name": "ชิ้น"}],
        "item_guid": "",   # Unlinked
        "itemcode": "",    # Unlinked
        "group_code": "GRP_QUICK",
        "group_names": [{"code": "th", "name": "กลุ่มด่วน"}],
        "brand_code": "BRAND_QUICK",
        "brandnames": [{"code": "th", "name": "แบรนด์ด่วน"}]
    }

    barcode_guid = ""
    try:
        resp = requests.post(f"{BASE_URL}/product/barcode", json=barcode_payload, headers=headers)
        print("Create Barcode Status:", resp.status_code)
        bar_data = resp.json()
        if resp.status_code not in [200, 201]:
            print("Create barcode failed:", bar_data)
            sys.exit(1)

        barcode_guid = bar_data.get("id") or bar_data.get("guidfixed") or bar_data.get("guid_fixed") or ""
        print(f"Created Barcode GUID: {barcode_guid}")
    except Exception as e:
        print("Error creating barcode:", e)
        sys.exit(1)

    # 5. Create a Main Product copying data from the barcode
    print("\n5. Creating new Main Product Master...")
    product_payload = {
        "code": "P-BF-001",
        "names": barcode_payload["names"],
        "item_type": 0,
        "vat_type": 0,
        "group_code": "GRP_LINKED_MASTER",
        "group_names": [{"code": "th", "name": "กลุ่มสินค้าหลักเชื่อมแล้ว"}],
        "brand_code": "BRAND_LINKED_MASTER",
        "brandnames": [{"code": "th", "name": "ยี่ห้อสินค้าหลักเชื่อมแล้ว"}],
        "categorycode": "CAT_LINKED_MASTER",
        "category_names": [{"code": "th", "name": "หมวดสินค้าหลักเชื่อมแล้ว"}],
        "manufacturers": [
            {
                "guid_fixed": "mfr-guid-bf1",
                "code": "MFR_BF1",
                "names": [{"code": "th", "name": "ผู้ผลิตใหม่"}]
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

    # 6. Link/Update the barcode to the new Product Master
    print("\n6. Linking barcode to the new Product Master...")
    try:
        # Get full barcode info first
        resp = requests.get(f"{BASE_URL}/product/barcode/{barcode_guid}", headers=headers)
        full_barcode = resp.json()["data"]

        # Link it
        full_barcode["item_guid"] = product_guid
        full_barcode["itemcode"] = product_payload["code"]

        update_resp = requests.put(f"{BASE_URL}/product/barcode/{barcode_guid}", json=full_barcode, headers=headers)
        print("Update Barcode Link Status:", update_resp.status_code)
        if update_resp.status_code not in [200, 201]:
            print("Link update failed:", update_resp.text)
            sys.exit(1)
    except Exception as e:
        print("Error linking barcode:", e)
        sys.exit(1)

    # 7. Retrieve barcode and verify classifications are pulled from the linked Product Master
    print("\n7. Verifying dynamic classifications mapping...")
    try:
        resp = requests.get(f"{BASE_URL}/product/barcode/{barcode_guid}", headers=headers)
        data = resp.json()["data"]

        print("\nChecking Classifications fallback (overwriting local values):")
        print(f"- Group Code: {data.get('group_code')} (Expected: GRP_LINKED_MASTER)")
        print(f"- Brand Code: {data.get('brand_code')} (Expected: BRAND_LINKED_MASTER)")
        print(f"- Category Code: {data.get('categorycode')} (Expected: CAT_LINKED_MASTER)")
        print(f"- Manufacturers: {[m['code'] for m in data.get('manufacturers', [])]} (Expected: ['MFR_BF1'])")

        assert data.get('group_code') == "GRP_LINKED_MASTER", "Group Code mapping mismatch!"
        assert data.get('brand_code') == "BRAND_LINKED_MASTER", "Brand Code mapping mismatch!"
        assert data.get('categorycode') == "CAT_LINKED_MASTER", "Category Code mapping mismatch!"
        assert len(data.get('manufacturers', [])) == 1 and data.get('manufacturers', [])[0]["code"] == "MFR_BF1", "Manufacturers mapping mismatch!"
        print("Dynamic mapping fallback verified successfully!")
    except Exception as e:
        print("Error verifying details:", e)
        sys.exit(1)

    # 8. Cleanup
    print("\n8. Cleaning up test data...")
    try:
        del_bar = requests.delete(f"{BASE_URL}/product/barcode/{barcode_guid}", headers=headers)
        print(f"- Deleted Barcode Status: {del_bar.status_code}")

        del_prod = requests.delete(f"{BASE_URL}/product/{product_guid}", headers=headers)
        print(f"- Deleted Product Status: {del_prod.status_code}")
    except Exception as e:
        print("Error cleaning up:", e)

    print("\nBarcode-First Flow E2E Verification SUCCESSFUL!")

if __name__ == "__main__":
    main()
