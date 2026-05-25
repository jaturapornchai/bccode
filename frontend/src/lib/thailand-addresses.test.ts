import { readFileSync } from "node:fs";
import { join } from "node:path";
import { describe, expect, it } from "vitest";
import {
  findThailandAddressMatchesByPostalCode,
  findThailandDistrict,
  findThailandProvince,
  findThailandSubdistrict,
  getThailandDistrictPostalCode,
  getThailandCountry,
  getThailandSubdistrictPostalCode,
  normalizeThaiPostalCode,
  onlyUniqueCodes,
  thailandAddressLabel,
  type ThailandAddressData,
} from "./thailand-addresses";

const thailandData = JSON.parse(
  readFileSync(
    join(process.cwd(), "../backend/assets/address/thailand-addresses.json"),
    "utf8",
  ),
) as ThailandAddressData;

describe("Thailand address data", () => {
  it("contains Thailand administrative hierarchy and postal codes", () => {
    const country = getThailandCountry(thailandData);
    expect(country?.code).toBe("TH");
    expect(country?.provinces).toHaveLength(77);
    expect(country?.provinces.flatMap((province) => province.districts)).toHaveLength(
      928,
    );
    expect(
      country?.provinces.flatMap((province) =>
        province.districts.flatMap((district) => district.subdistricts),
      ),
    ).toHaveLength(7436);
  });

  it("finds Bangkok Phra Nakhon by official hierarchy codes", () => {
    expect(findThailandProvince(thailandData, "10")?.name.th).toBe(
      "กรุงเทพมหานคร",
    );
    expect(findThailandDistrict(thailandData, "10", "1001")?.name.th).toBe(
      "พระนคร",
    );
    expect(
      findThailandSubdistrict(thailandData, "10", "1001", "100101")?.postalCode,
    ).toBe("10200");
  });

  it("returns district postal code only when every subdistrict uses the same code", () => {
    expect(getThailandDistrictPostalCode(thailandData, "10", "1001")).toBe(
      "10200",
    );
    expect(getThailandDistrictPostalCode(thailandData, "95", "9501")).toBe("");
  });

  it("returns exact postal code by selected subdistrict", () => {
    expect(
      getThailandSubdistrictPostalCode(thailandData, "95", "9501", "950104"),
    ).toBe("95160");
  });

  it("finds address candidates by postal code", () => {
    const matches = findThailandAddressMatchesByPostalCode(thailandData, "10200");
    expect(matches.length).toBeGreaterThan(1);
    expect(onlyUniqueCodes(matches, (item) => item.provinceCode)).toEqual(["10"]);
    expect(matches[0]?.postalCode).toBe("10200");
  });

  it("normalizes postal code input to five digits", () => {
    expect(normalizeThaiPostalCode(" 10-200 abc")).toBe("10200");
    expect(normalizeThaiPostalCode("10200123")).toBe("10200");
  });

  it("uses Thai labels for Thai UI and English labels otherwise", () => {
    const province = findThailandProvince(thailandData, "10");
    expect(province).toBeTruthy();
    expect(thailandAddressLabel(province!.name, "th")).toBe("กรุงเทพมหานคร");
    expect(thailandAddressLabel(province!.name, "en")).toBe("Bangkok");
  });
});
