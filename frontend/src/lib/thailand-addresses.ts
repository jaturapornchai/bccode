import type { LanguageCode } from "@/lib/i18n";

export type ThailandAddressName = {
  th: string;
  en: string;
};

export type ThailandSubdistrict = {
  code: string;
  name: ThailandAddressName;
  postalCode: string;
};

export type ThailandDistrict = {
  code: string;
  name: ThailandAddressName;
  subdistricts: ThailandSubdistrict[];
};

export type ThailandProvince = {
  code: string;
  name: ThailandAddressName;
  districts: ThailandDistrict[];
};

export type ThailandCountry = {
  code: "TH";
  name: ThailandAddressName;
  provinces: ThailandProvince[];
};

export type ThailandAddressData = {
  version: number;
  generatedAt: string;
  source: {
    name: string;
    url: string;
    rawUrl: string;
    license: string;
    copyright: string;
    sourceRows: number;
  };
  countries: ThailandCountry[];
};

export type ThailandAddressMatch = {
  countryCode: "TH";
  provinceCode: string;
  provinceName: ThailandAddressName;
  districtCode: string;
  districtName: ThailandAddressName;
  subdistrictCode: string;
  subdistrictName: ThailandAddressName;
  postalCode: string;
};

const thailandAddressDataPromises = new Map<string, Promise<ThailandAddressData>>();

export function loadThailandAddressData(
  backendUrl?: string,
): Promise<ThailandAddressData> {
  const normalizedBackendUrl = backendUrl?.trim() ?? "";
  const cacheKey = normalizedBackendUrl || "__missing_backend_url__";
  const existing = thailandAddressDataPromises.get(cacheKey);
  if (existing) return existing;

  const endpoint = normalizedBackendUrl
    ? `/api/address/thailand?backendUrl=${encodeURIComponent(normalizedBackendUrl)}`
    : "/api/address/thailand";
  const promise = fetch(endpoint, {
    cache: "force-cache",
  })
    .then(async (response) => {
      if (!response.ok) throw new Error("Thailand address data load failed");
      return (await response.json()) as ThailandAddressData;
    })
    .catch((error: unknown) => {
      thailandAddressDataPromises.delete(cacheKey);
      throw error;
    });
  thailandAddressDataPromises.set(cacheKey, promise);
  return promise;
}

export function thailandAddressLabel(
  name: ThailandAddressName,
  language: LanguageCode,
): string {
  return language === "th" ? name.th || name.en : name.en || name.th;
}

export function getThailandCountry(
  data: ThailandAddressData,
): ThailandCountry | undefined {
  return data.countries.find((country) => country.code === "TH");
}

export function findThailandProvince(
  data: ThailandAddressData,
  provinceCode: string,
): ThailandProvince | undefined {
  return getThailandCountry(data)?.provinces.find(
    (province) => province.code === provinceCode,
  );
}

export function findThailandDistrict(
  data: ThailandAddressData,
  provinceCode: string,
  districtCode: string,
): ThailandDistrict | undefined {
  return findThailandProvince(data, provinceCode)?.districts.find(
    (district) => district.code === districtCode,
  );
}

export function findThailandSubdistrict(
  data: ThailandAddressData,
  provinceCode: string,
  districtCode: string,
  subdistrictCode: string,
): ThailandSubdistrict | undefined {
  return findThailandDistrict(data, provinceCode, districtCode)?.subdistricts.find(
    (subdistrict) => subdistrict.code === subdistrictCode,
  );
}

export function findThailandAddressMatchesByPostalCode(
  data: ThailandAddressData,
  postalCode: string,
): ThailandAddressMatch[] {
  const normalized = normalizeThaiPostalCode(postalCode);
  if (normalized.length !== 5) return [];
  const country = getThailandCountry(data);
  if (!country) return [];
  const matches: ThailandAddressMatch[] = [];
  for (const province of country.provinces) {
    for (const district of province.districts) {
      for (const subdistrict of district.subdistricts) {
        if (subdistrict.postalCode !== normalized) continue;
        matches.push({
          countryCode: "TH",
          provinceCode: province.code,
          provinceName: province.name,
          districtCode: district.code,
          districtName: district.name,
          subdistrictCode: subdistrict.code,
          subdistrictName: subdistrict.name,
          postalCode: normalized,
        });
      }
    }
  }
  return matches;
}

export function getThailandDistrictPostalCode(
  data: ThailandAddressData,
  provinceCode: string,
  districtCode: string,
): string {
  const district = findThailandDistrict(data, provinceCode, districtCode);
  const postalCodes = onlyUniqueCodes(
    district?.subdistricts ?? [],
    (subdistrict) => subdistrict.postalCode,
  );
  return postalCodes.length === 1 ? postalCodes[0] : "";
}

export function getThailandSubdistrictPostalCode(
  data: ThailandAddressData,
  provinceCode: string,
  districtCode: string,
  subdistrictCode: string,
): string {
  return (
    findThailandSubdistrict(data, provinceCode, districtCode, subdistrictCode)
      ?.postalCode ?? ""
  );
}

export function normalizeThaiPostalCode(value: unknown): string {
  return String(value ?? "")
    .replace(/\D/g, "")
    .slice(0, 5);
}

export function onlyUniqueCodes<T>(
  rows: T[],
  getCode: (row: T) => string,
): string[] {
  return Array.from(new Set(rows.map(getCode).filter(Boolean)));
}
