"use client";

import { useEffect, useMemo, useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import type { LanguageCode } from "@/lib/i18n";
import {
  findThailandAddressMatchesByPostalCode,
  findThailandDistrict,
  findThailandProvince,
  findThailandSubdistrict,
  getThailandDistrictPostalCode,
  getThailandCountry,
  getThailandSubdistrictPostalCode,
  loadThailandAddressData,
  normalizeThaiPostalCode,
  onlyUniqueCodes,
  thailandAddressLabel,
  type ThailandAddressData,
  type ThailandAddressMatch,
  type ThailandDistrict,
  type ThailandProvince,
  type ThailandSubdistrict,
} from "@/lib/thailand-addresses";
import {
  type FormState,
  type SettingRecord,
  THAI_ADDRESS_SUBKEYS,
  getPathOrFlatValue,
  stringValue,
} from "../types";

// --- Read-only detail (used in SettingDetailPanel) ---

export function ThailandAddressReadOnlyDetail({
  backendUrl,
  form,
  language,
  prefix,
}: {
  backendUrl?: string;
  form: SettingRecord;
  language: LanguageCode;
  prefix: string;
}) {
  const [data, setData] = useState<ThailandAddressData | null>(null);
  const [loadError, setLoadError] = useState("");
  const countryCode = stringValue(getPathOrFlatValue(form, `${prefix}.countrycode`)) || "TH";
  const provinceCode = stringValue(getPathOrFlatValue(form, `${prefix}.provincecode`));
  const districtCode = stringValue(getPathOrFlatValue(form, `${prefix}.districtcode`));
  const subdistrictCode = stringValue(
    getPathOrFlatValue(form, `${prefix}.subdistrictcode`),
  );
  const postalCode = normalizeThaiPostalCode(
    getPathOrFlatValue(form, `${prefix}.zipcode`),
  );
  const labels = thailandAddressUi(language);

  useEffect(() => {
    if (countryCode !== "TH") return;
    let active = true;
    setLoadError("");
    loadThailandAddressData(backendUrl)
      .then((nextData) => {
        if (active) setData(nextData);
      })
      .catch((error: unknown) => {
        if (!active) return;
        setLoadError(error instanceof Error ? error.message : "load failed");
      });
    return () => {
      active = false;
    };
  }, [backendUrl, countryCode]);

  const country = data ? getThailandCountry(data) : undefined;
  const selectedProvince = data
    ? findThailandProvince(data, provinceCode)
    : undefined;
  const selectedDistrict = data
    ? findThailandDistrict(data, provinceCode, districtCode)
    : undefined;
  const selectedSubdistrict = data
    ? findThailandSubdistrict(
        data,
        provinceCode,
        districtCode,
        subdistrictCode,
      )
    : undefined;
  const loading = countryCode === "TH" && !country && !loadError;

  return (
    <section className="grid gap-2 rounded-2xl border border-border bg-background p-2 text-sm font-semibold md:col-span-2">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <span>{labels.title}</span>
        <span className="text-xs font-medium text-muted-foreground">
          {loading
            ? labels.loading
            : loadError
              ? labels.loadError
              : countryCode === "TH"
                ? `${country?.provinces.length ?? 0} ${labels.provinces}`
                : countryCode}
        </span>
      </div>
      <div className="grid gap-2 md:grid-cols-2">
        <ReadOnlyDetailValue
          label={labels.province}
          value={thailandAddressReadOnlyValue(
            provinceCode,
            selectedProvince
              ? thailandAddressLabel(selectedProvince.name, language)
              : "",
          )}
        />
        <ReadOnlyDetailValue
          label={labels.district}
          value={thailandAddressReadOnlyValue(
            districtCode,
            selectedDistrict
              ? thailandAddressLabel(selectedDistrict.name, language)
              : "",
          )}
        />
        <ReadOnlyDetailValue
          label={labels.subdistrict}
          value={thailandAddressReadOnlyValue(
            subdistrictCode,
            selectedSubdistrict
              ? thailandAddressLabel(selectedSubdistrict.name, language)
              : "",
          )}
        />
        <ReadOnlyDetailValue label={labels.postalCode} value={postalCode || "-"} />
      </div>
    </section>
  );
}

export function ReadOnlyDetailValue({
  label,
  value,
}: {
  label: string;
  value: string;
}) {
  return (
    <label className="grid gap-1 text-sm font-semibold">
      <span>{label}</span>
      <Input aria-readonly readOnly value={value || "-"} />
    </label>
  );
}

export function thailandAddressReadOnlyValue(code: string, label: string): string {
  if (code && label) return `${label} (${code})`;
  return code || label || "-";
}

// --- Editable field editor ---

export function ThailandAddressFieldEditor({
  backendUrl,
  copyFromPrefix,
  form,
  language,
  prefix,
  setForm,
}: {
  backendUrl?: string;
  copyFromPrefix?: string;
  form: FormState;
  language: LanguageCode;
  prefix: string;
  setForm: (update: FormState | ((current: FormState) => FormState)) => void;
}) {
  const countryKey = `${prefix}.countrycode`;
  const provinceKey = `${prefix}.provincecode`;
  const districtKey = `${prefix}.districtcode`;
  const subdistrictKey = `${prefix}.subdistrictcode`;
  const zipKey = `${prefix}.zipcode`;
  const [data, setData] = useState<ThailandAddressData | null>(null);
  const [loadError, setLoadError] = useState("");
  const countryCode = stringValue(form[countryKey]) || "TH";
  const provinceCode = stringValue(form[provinceKey]);
  const districtCode = stringValue(form[districtKey]);
  const subdistrictCode = stringValue(form[subdistrictKey]);
  const postalCode = normalizeThaiPostalCode(form[zipKey]);

  useEffect(() => {
    if (countryCode !== "TH") return;
    let active = true;
    setLoadError("");
    loadThailandAddressData(backendUrl)
      .then((nextData) => {
        if (active) setData(nextData);
      })
      .catch((error: unknown) => {
        if (!active) return;
        setLoadError(error instanceof Error ? error.message : "load failed");
      });
    return () => {
      active = false;
    };
  }, [backendUrl, countryCode]);

  const country = data ? getThailandCountry(data) : undefined;
  const postalMatches = useMemo(
    () =>
      data && postalCode.length === 5
        ? findThailandAddressMatchesByPostalCode(data, postalCode)
        : [],
    [data, postalCode],
  );
  const selectedProvince = data
    ? findThailandProvince(data, provinceCode)
    : undefined;
  const selectedDistrict = data
    ? findThailandDistrict(data, provinceCode, districtCode)
    : undefined;
  const selectedSubdistrict = data
    ? findThailandSubdistrict(
        data,
        provinceCode,
        districtCode,
        subdistrictCode,
      )
    : undefined;

  const provinces = filterThailandProvinces(
    country?.provinces ?? [],
    postalMatches,
  );
  const districts = filterThailandDistricts(
    selectedProvince?.districts ?? [],
    postalMatches,
    provinceCode,
  );
  const subdistricts = filterThailandSubdistricts(
    selectedDistrict?.subdistricts ?? [],
    postalMatches,
    provinceCode,
    districtCode,
  );

  function setAddress(nextValues: Partial<FormState>) {
    setForm({ ...form, ...nextValues });
  }

  function copyFromBilling() {
    if (!copyFromPrefix) return;
    const nextValues: Partial<FormState> = {
      [`${prefix}.address`]: form[`${copyFromPrefix}.address`],
    };
    for (const sub of THAI_ADDRESS_SUBKEYS) {
      nextValues[`${prefix}.${sub}`] = form[`${copyFromPrefix}.${sub}`];
    }
    setAddress(nextValues);
  }

  function chooseProvince(nextProvinceCode: string) {
    const nextValues: Partial<FormState> = {
      [provinceKey]: nextProvinceCode,
      [districtKey]: "",
      [subdistrictKey]: "",
      [zipKey]: "",
    };
    if (data && postalCode.length === 5 && nextProvinceCode) {
      const matches = postalMatches.filter(
        (item) => item.provinceCode === nextProvinceCode,
      );
      if (matches.length > 0) {
        nextValues[zipKey] = postalCode;
        const nextDistrictCode = singleThailandAddressCode(
          matches,
          (item) => item.districtCode,
        );
        if (nextDistrictCode) {
          nextValues[districtKey] = nextDistrictCode;
          const nextSubdistrictCode = singleThailandAddressCode(
            matches.filter((item) => item.districtCode === nextDistrictCode),
            (item) => item.subdistrictCode,
          );
          if (nextSubdistrictCode)
            nextValues[subdistrictKey] = nextSubdistrictCode;
        }
      }
    }
    setAddress(nextValues);
  }

  function chooseDistrict(nextDistrictCode: string) {
    const nextValues: Partial<FormState> = {
      [districtKey]: nextDistrictCode,
      [subdistrictKey]: "",
      [zipKey]: "",
    };
    if (data && provinceCode && nextDistrictCode) {
      const matches =
        postalCode.length === 5
          ? postalMatches.filter(
              (item) =>
                item.provinceCode === provinceCode &&
                item.districtCode === nextDistrictCode,
            )
          : [];
      nextValues[zipKey] =
        matches.length > 0
          ? postalCode
          : getThailandDistrictPostalCode(data, provinceCode, nextDistrictCode);
      if (postalCode.length === 5) {
        const nextSubdistrictCode = singleThailandAddressCode(
          matches,
          (item) => item.subdistrictCode,
        );
        if (nextSubdistrictCode)
          nextValues[subdistrictKey] = nextSubdistrictCode;
      }
    }
    setAddress(nextValues);
  }

  function chooseSubdistrict(nextSubdistrictCode: string) {
    const nextValues: Partial<FormState> = {
      [subdistrictKey]: nextSubdistrictCode,
      [zipKey]: "",
    };
    if (data && provinceCode && districtCode) {
      nextValues[zipKey] = nextSubdistrictCode
        ? getThailandSubdistrictPostalCode(
            data,
            provinceCode,
            districtCode,
            nextSubdistrictCode,
          )
        : getThailandDistrictPostalCode(data, provinceCode, districtCode);
    }
    setAddress(nextValues);
  }

  function applyPostalCode(rawValue: string) {
    const normalized = normalizeThaiPostalCode(rawValue);
    const nextValues: Partial<FormState> = { [zipKey]: normalized };
    if (data && normalized.length === 5) {
      const matches = findThailandAddressMatchesByPostalCode(data, normalized);
      const nextProvinceCode = singleThailandAddressCode(
        matches,
        (item) => item.provinceCode,
      );
      if (nextProvinceCode) {
        nextValues[provinceKey] = nextProvinceCode;
        const provinceMatches = matches.filter(
          (item) => item.provinceCode === nextProvinceCode,
        );
        const nextDistrictCode = singleThailandAddressCode(
          provinceMatches,
          (item) => item.districtCode,
        );
        nextValues[districtKey] = nextDistrictCode;
        if (nextDistrictCode) {
          nextValues[subdistrictKey] = singleThailandAddressCode(
            provinceMatches.filter(
              (item) => item.districtCode === nextDistrictCode,
            ),
            (item) => item.subdistrictCode,
          );
        } else {
          nextValues[subdistrictKey] = "";
        }
      } else {
        nextValues[provinceKey] = "";
        nextValues[districtKey] = "";
        nextValues[subdistrictKey] = "";
      }
    }
    setAddress(nextValues);
  }

  if (countryCode !== "TH") {
    return (
      <ThailandAddressFreeTextEditor
        form={form}
        language={language}
        prefix={prefix}
        setForm={setForm}
      />
    );
  }

  const loading = !country && !loadError;
  const labels = thailandAddressUi(language);

  return (
    <section className="grid gap-2 rounded-2xl border border-border bg-background p-2 text-sm font-semibold md:col-span-2">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <span>{labels.title}</span>
        <div className="flex items-center gap-2">
          {copyFromPrefix ? (
            <Button
              type="button"
              variant="outline"
              size="sm"
              className="h-7 text-xs"
              onClick={copyFromBilling}
            >
              {language === "th" ? "คัดลอกจากที่อยู่ออกใบกำกับภาษี" : "Copy from billing address"}
            </Button>
          ) : null}
          <span className="text-xs font-medium text-muted-foreground">
            {loading
              ? labels.loading
              : loadError
                ? labels.loadError
                : `${country?.provinces.length ?? 0} ${labels.provinces}`}
          </span>
        </div>
      </div>
      <div className="grid gap-2 md:grid-cols-2">
        <ThailandAddressSelect
          label={labels.province}
          value={provinceCode}
          disabled={loading || Boolean(loadError)}
          placeholder={labels.selectProvince}
          onChange={chooseProvince}
          options={provinces.map((province) => ({
            code: province.code,
            label: thailandAddressOptionLabel(province, language),
          }))}
        />
        <ThailandAddressSelect
          label={labels.district}
          value={districtCode}
          disabled={!selectedProvince || loading || Boolean(loadError)}
          placeholder={labels.selectDistrict}
          onChange={chooseDistrict}
          options={districts.map((district) => ({
            code: district.code,
            label: thailandAddressOptionLabel(district, language),
          }))}
        />
        <ThailandAddressSelect
          label={labels.subdistrict}
          value={subdistrictCode}
          disabled={!selectedDistrict || loading || Boolean(loadError)}
          placeholder={labels.selectSubdistrict}
          onChange={chooseSubdistrict}
          options={subdistricts.map((subdistrict) => ({
            code: subdistrict.code,
            label: thailandAddressOptionLabel(subdistrict, language),
          }))}
        />
        <label className="grid gap-1">
          <span>{labels.postalCode}</span>
          <Input
            inputMode="numeric"
            maxLength={5}
            value={postalCode}
            onChange={(event) => applyPostalCode(event.target.value)}
            placeholder="10200"
          />
        </label>
      </div>
      <p className="text-xs font-medium text-muted-foreground">
        {postalAddressHint(
          postalCode,
          postalMatches,
          selectedSubdistrict,
          language,
        )}
      </p>
    </section>
  );
}

// --- Free-text fallback for non-TH countries ---

export function ThailandAddressFreeTextEditor({
  form,
  language,
  prefix,
  setForm,
}: {
  form: FormState;
  language: LanguageCode;
  prefix: string;
  setForm: (update: FormState | ((current: FormState) => FormState)) => void;
}) {
  const labels = thailandAddressUi(language);
  const fields = [
    [`${prefix}.provincecode`, labels.province],
    [`${prefix}.districtcode`, labels.district],
    [`${prefix}.subdistrictcode`, labels.subdistrict],
    [`${prefix}.zipcode`, labels.postalCode],
  ] as const;
  return (
    <section className="grid gap-2 rounded-2xl border border-border bg-background p-2 text-sm font-semibold md:col-span-2">
      <div className="grid gap-2 md:grid-cols-2">
        {fields.map(([key, label]) => (
          <label className="grid gap-1" key={key}>
            <span>{label}</span>
            <Input
              value={String(form[key] ?? "")}
              onChange={(event) =>
                setForm({ ...form, [key]: event.target.value })
              }
            />
          </label>
        ))}
      </div>
    </section>
  );
}

// --- Shared select component ---

export function ThailandAddressSelect({
  disabled,
  label,
  onChange,
  options,
  placeholder,
  value,
}: {
  disabled?: boolean;
  label: string;
  onChange: (value: string) => void;
  options: Array<{ code: string; label: string }>;
  placeholder: string;
  value: string;
}) {
  return (
    <label className="grid gap-1">
      <span>{label}</span>
      <select
        className="min-h-10 w-full rounded-2xl border border-input bg-background px-3 text-sm text-foreground shadow-sm outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
        disabled={disabled}
        value={value}
        onChange={(event) => onChange(event.target.value)}
      >
        <option value="">{placeholder}</option>
        {options.map((option) => (
          <option key={option.code} value={option.code}>
            {option.label}
          </option>
        ))}
      </select>
    </label>
  );
}

// --- Helper functions ---

export function thailandAddressUi(language: LanguageCode) {
  if (language === "th") {
    return {
      title: "ที่อยู่ประเทศไทย",
      loading: "กำลังโหลดข้อมูลที่อยู่",
      loadError: "โหลดข้อมูลที่อยู่ไม่สำเร็จ",
      provinces: "จังหวัด",
      province: "จังหวัด",
      district: "อำเภอ/เขต",
      subdistrict: "ตำบล/แขวง",
      postalCode: "รหัสไปรษณีย์",
      selectProvince: "เลือกจังหวัด",
      selectDistrict: "เลือกอำเภอ/เขต",
      selectSubdistrict: "เลือกตำบล/แขวง",
    };
  }
  return {
    title: "Thailand address",
    loading: "Loading address data",
    loadError: "Address data failed to load",
    provinces: "provinces",
    province: "Province",
    district: "District",
    subdistrict: "Subdistrict",
    postalCode: "Postal code",
    selectProvince: "Select province",
    selectDistrict: "Select district",
    selectSubdistrict: "Select subdistrict",
  };
}

export function filterThailandProvinces(
  provinces: ThailandProvince[],
  matches: ThailandAddressMatch[],
): ThailandProvince[] {
  if (!matches.length) return provinces;
  const allowed = new Set(matches.map((item) => item.provinceCode));
  return provinces.filter((province) => allowed.has(province.code));
}

export function filterThailandDistricts(
  districts: ThailandDistrict[],
  matches: ThailandAddressMatch[],
  provinceCode: string,
): ThailandDistrict[] {
  if (!matches.length || !provinceCode) return districts;
  const allowed = new Set(
    matches
      .filter((item) => item.provinceCode === provinceCode)
      .map((item) => item.districtCode),
  );
  return districts.filter((district) => allowed.has(district.code));
}

export function filterThailandSubdistricts(
  subdistricts: ThailandSubdistrict[],
  matches: ThailandAddressMatch[],
  provinceCode: string,
  districtCode: string,
): ThailandSubdistrict[] {
  if (!matches.length || !provinceCode || !districtCode) return subdistricts;
  const allowed = new Set(
    matches
      .filter(
        (item) =>
          item.provinceCode === provinceCode &&
          item.districtCode === districtCode,
      )
      .map((item) => item.subdistrictCode),
  );
  return subdistricts.filter((subdistrict) => allowed.has(subdistrict.code));
}

export function singleThailandAddressCode<T>(
  rows: T[],
  getCode: (row: T) => string,
): string {
  const codes = onlyUniqueCodes(rows, getCode);
  return codes.length === 1 ? codes[0] : "";
}

export function thailandAddressOptionLabel(
  item: ThailandProvince | ThailandDistrict | ThailandSubdistrict,
  language: LanguageCode,
): string {
  return `${thailandAddressLabel(item.name, language)} (${item.code})`;
}

export function postalAddressHint(
  postalCode: string,
  matches: ThailandAddressMatch[],
  selectedSubdistrict: ThailandSubdistrict | undefined,
  language: LanguageCode,
): string {
  if (postalCode.length !== 5) {
    return language === "th"
      ? "เลือกจังหวัด > อำเภอ/เขต > ตำบล/แขวง แล้วระบบจะเติมรหัสไปรษณีย์ หรือกรอกรหัสไปรษณีย์ 5 หลักเพื่อกรองตัวเลือก"
      : "Select province > district > subdistrict to fill the postal code, or enter a 5-digit postal code to filter choices.";
  }
  if (!matches.length) {
    return language === "th"
      ? "ไม่พบข้อมูลจากรหัสไปรษณีย์นี้"
      : "No address found for this postal code.";
  }
  if (selectedSubdistrict) {
    return language === "th"
      ? `เลือกแล้ว: ${selectedSubdistrict.name.th} ${postalCode}`
      : `Selected: ${selectedSubdistrict.name.en || selectedSubdistrict.name.th} ${postalCode}`;
  }
  return language === "th"
    ? `พบ ${matches.length} ตำบล/แขวงจากรหัสนี้ เลือกตำบล/แขวงเพื่อยืนยัน`
    : `${matches.length} subdistricts found for this postal code. Select one to confirm.`;
}
