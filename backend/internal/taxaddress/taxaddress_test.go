package taxaddress

import (
	"strings"
	"testing"
)

func TestKeysMatchFields(t *testing.T) {
	var a Address
	if len(a.Fields()) != len(Keys) {
		t.Fatalf("fields %d keys %d", len(a.Fields()), len(Keys))
	}
	for i, field := range a.Fields() {
		*field = Keys[i]
	}
	for key, value := range a.Values() {
		if key != value {
			t.Fatalf("field order mismatch: key %s holds %s", key, value)
		}
	}
	if len(a.Values()) != 13 {
		t.Fatalf("values %d", len(a.Values()))
	}
}

func TestNormalizeAndBlank(t *testing.T) {
	a := Address{No: "  88/12 ", Province: "  "}
	a.Normalize()
	if a.No != "88/12" || a.Province != "" {
		t.Fatalf("normalize: %+v", a)
	}
	if (Address{Road: "  "}).IsBlank() != true || (Address{Road: "พหลโยธิน"}).IsBlank() {
		t.Fatal("IsBlank")
	}
	if NormalizePhone(" 02-123-4567 ") != "02-123-4567" {
		t.Fatal("phone trim")
	}
}

func TestValidate(t *testing.T) {
	ok := Address{Building: strings.Repeat("ก", 40), Room: strings.Repeat("1", 20), Province: "ปทุมธานี", Postcode: "12120"}
	if err := Validate(ok, strings.Repeat("0", 50)); err != nil {
		t.Fatalf("valid address rejected: %v", err)
	}
	if err := Validate(Address{}, ""); err != nil {
		t.Fatalf("blank address rejected: %v", err)
	}
	cases := []struct {
		name   string
		addr   Address
		phone  string
		field  string
		reason string
		max    int
	}{
		{"building 41 runes (Thai counted as runes, not bytes)", Address{Building: strings.Repeat("ก", 41)}, "", "addr_building", ReasonTooLong, 40},
		{"room 21", Address{Room: strings.Repeat("1", 21)}, "", "addr_room", ReasonTooLong, 20},
		{"village 101", Address{Village: strings.Repeat("ม", 101)}, "", "addr_village", ReasonTooLong, 100},
		{"junction 101", Address{Junction: strings.Repeat("1", 101)}, "", "addr_junction", ReasonTooLong, 100},
		{"province 51", Address{Province: strings.Repeat("จ", 51)}, "", "addr_province", ReasonTooLong, 50},
		{"newline", Address{Road: "พหลโยธิน\nกม. 40"}, "", "addr_road", ReasonControl, 0},
		{"carriage return", Address{No: "88\r"}, "", "addr_no", ReasonControl, 0},
		{"tab", Address{Soi: "ลาดพร้าว\t71"}, "", "addr_soi", ReasonControl, 0},
		{"postcode 4 digits", Address{Postcode: "1212"}, "", "addr_postcode", ReasonPostcode, 0},
		{"postcode letters", Address{Postcode: "1212O"}, "", "addr_postcode", ReasonPostcode, 0},
		{"postcode Thai digits", Address{Postcode: "๑๒๑๒๐"}, "", "addr_postcode", ReasonPostcode, 0},
		{"postcode 6 digits", Address{Postcode: "121200"}, "", "addr_postcode", ReasonPostcode, 0},
		{"phone 51", Address{}, strings.Repeat("0", 51), PhoneKey, ReasonTooLong, 50},
		{"phone tab", Address{}, "02\t1234567", PhoneKey, ReasonControl, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			fe := AsFieldError(Validate(c.addr, c.phone))
			if fe == nil || fe.Field != c.field || fe.Reason != c.reason || fe.Max != c.max {
				t.Fatalf("got %+v want %s/%s/%d", fe, c.field, c.reason, c.max)
			}
		})
	}
}

func TestLineProvince(t *testing.T) {
	a := Address{Village: "พฤกษา 12", No: "88/12", Moo: "3", Soi: "รังสิต-นครนายก 46", Junction: "5", Road: "พหลโยธิน",
		Subdistrict: "คลองหนึ่ง", District: "คลองหลวง", Province: "ปทุมธานี", Postcode: "12120"}
	want := "หมู่บ้านพฤกษา 12 เลขที่ 88/12 ซอยรังสิต-นครนายก 46 แยก 5 หมู่ที่ 3 ถนนพหลโยธิน ตำบลคลองหนึ่ง อำเภอคลองหลวง จังหวัดปทุมธานี 12120"
	if got := a.Line(); got != want {
		t.Fatalf("got  %q\nwant %q", got, want)
	}
}

func TestLineBangkok(t *testing.T) {
	a := Address{Building: "สาทรซิตี้ทาวเวอร์", Room: "1201", Floor: "12", No: "175", Road: "สาทรใต้",
		Subdistrict: "ทุ่งมหาเมฆ", District: "สาทร", Province: "กรุงเทพมหานคร", Postcode: "10120"}
	want := "อาคารสาทรซิตี้ทาวเวอร์ ห้องเลขที่ 1201 ชั้นที่ 12 เลขที่ 175 ถนนสาทรใต้ แขวงทุ่งมหาเมฆ เขตสาทร กรุงเทพมหานคร 10120"
	if got := a.Line(); got != want {
		t.Fatalf("got  %q\nwant %q", got, want)
	}
	for _, name := range []string{"กรุงเทพฯ", "กทม.", " จังหวัดกรุงเทพมหานคร ", "จ.กรุงเทพ"} {
		if !IsBangkok(name) {
			t.Fatalf("%q must be Bangkok", name)
		}
	}
	if IsBangkok("นนทบุรี") || IsBangkok("") {
		t.Fatal("non-Bangkok detected as Bangkok")
	}
	if got := (Address{Subdistrict: "ลาดยาว", District: "จตุจักร", Province: "กทม."}).Line(); got != "แขวงลาดยาว เขตจตุจักร กทม." {
		t.Fatalf("abbreviated Bangkok: %q", got)
	}
}

func TestLineDoesNotDoublePrefixes(t *testing.T) {
	a := Address{Building: "อาคารชาญอิสสระ", Village: "หมู่บ้านเศรษฐกิจ", Room: "ห้อง 5", Floor: "ชั้น 3", No: "เลขที่ 9",
		Soi: "ตรอกจันทน์", Moo: "ม.4", Road: "ถ.พระราม 4", Subdistrict: "ตำบลบางพูด", District: "อ.ปากเกร็ด",
		Province: "จังหวัดนนทบุรี", Postcode: "11120"}
	want := "อาคารชาญอิสสระ หมู่บ้านเศรษฐกิจ ห้อง 5 ชั้น 3 เลขที่ 9 ตรอกจันทน์ ม.4 ถ.พระราม 4 ตำบลบางพูด อ.ปากเกร็ด จังหวัดนนทบุรี 11120"
	if got := a.Line(); got != want {
		t.Fatalf("got  %q\nwant %q", got, want)
	}
	// ผู้ใช้พิมพ์ แขวง/เขต มาแล้ว → ไม่เติม ตำบล/อำเภอ ซ้อน แม้จังหวัดไม่ใช่กรุงเทพฯ
	if got := (Address{Subdistrict: "แขวงบางนา", District: "เขตบางนา", Province: "กรุงเทพมหานคร"}).Line(); got != "แขวงบางนา เขตบางนา กรุงเทพมหานคร" {
		t.Fatalf("typed Bangkok prefixes: %q", got)
	}
}

func TestLineSkipsBlankParts(t *testing.T) {
	if got := (Address{}).Line(); got != "" {
		t.Fatalf("blank: %q", got)
	}
	if got := (Address{No: " 12 ", Road: "  ", Province: "เชียงใหม่"}).Line(); got != "เลขที่ 12 จังหวัดเชียงใหม่" {
		t.Fatalf("blank parts: %q", got)
	}
	if got := (Address{Road: "เอกชัย", Building: "Sathorn City"}).Line(); got != "อาคาร Sathorn City ถนนเอกชัย" {
		t.Fatalf("spacing: %q", got)
	}
}
