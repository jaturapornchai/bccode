export type CurrencyPresetForm = {
  code: string;
  name: string;
  symbol: string;
};

export type CurrencySymbolPreset = {
  code: string;
  name: string;
  isoName: string;
  symbol: string;
  aliases?: readonly string[];
};

export const currencyPresetSource = {
  standard: "ISO 4217",
  authority: "SIX Financial Information",
  list: "List One: Current Currency & Funds",
  published: "2026-01-01",
} as const;

export const currencySymbolPresets: readonly CurrencySymbolPreset[] = [
  { code: "THB", name: "Thai Baht", isoName: "Baht", symbol: "฿", aliases: ["Baht"] },
  { code: "VND", name: "Vietnamese Dong", isoName: "Dong", symbol: "₫", aliases: ["Dong"] },
  { code: "LAK", name: "Lao Kip", isoName: "Lao Kip", symbol: "₭" },
  { code: "KHR", name: "Cambodian Riel", isoName: "Riel", symbol: "៛", aliases: ["Riel"] },
  { code: "MMK", name: "Myanmar Kyat", isoName: "Kyat", symbol: "K", aliases: ["Kyat"] },
  { code: "MYR", name: "Malaysian Ringgit", isoName: "Malaysian Ringgit", symbol: "RM" },
  { code: "SGD", name: "Singapore Dollar", isoName: "Singapore Dollar", symbol: "S$" },
  { code: "IDR", name: "Indonesian Rupiah", isoName: "Rupiah", symbol: "Rp", aliases: ["Rupiah"] },
  { code: "PHP", name: "Philippine Peso", isoName: "Philippine Peso", symbol: "₱" },
  { code: "BND", name: "Brunei Dollar", isoName: "Brunei Dollar", symbol: "B$" },
  { code: "USD", name: "US Dollar", isoName: "US Dollar", symbol: "$" },
  { code: "EUR", name: "Euro", isoName: "Euro", symbol: "€" },
  { code: "JPY", name: "Japanese Yen", isoName: "Yen", symbol: "¥", aliases: ["Yen"] },
  { code: "CNY", name: "Chinese Yuan", isoName: "Yuan Renminbi", symbol: "¥", aliases: ["Yuan Renminbi"] },
  { code: "GBP", name: "Pound Sterling", isoName: "Pound Sterling", symbol: "£" },
  { code: "INR", name: "Indian Rupee", isoName: "Indian Rupee", symbol: "₹" },
  { code: "KRW", name: "Korean Won", isoName: "Won", symbol: "₩", aliases: ["Won"] },
  { code: "HKD", name: "Hong Kong Dollar", isoName: "Hong Kong Dollar", symbol: "HK$" },
  { code: "AUD", name: "Australian Dollar", isoName: "Australian Dollar", symbol: "A$" },
  { code: "CHF", name: "Swiss Franc", isoName: "Swiss Franc", symbol: "CHF" },
  { code: "BRL", name: "Brazilian Real", isoName: "Brazilian Real", symbol: "R$" },
  { code: "CAD", name: "Canadian Dollar", isoName: "Canadian Dollar", symbol: "C$" },
] as const;

export function applyCurrencySymbolPreset<TForm extends CurrencyPresetForm>(
  form: TForm,
  preset: CurrencySymbolPreset,
  keepCode: boolean,
): TForm {
  return {
    ...form,
    code: keepCode ? form.code : preset.code,
    name: preset.name,
    symbol: preset.symbol,
  };
}

export function filterCurrencySymbolPresets(query: string): CurrencySymbolPreset[] {
  const needle = query.trim().toLowerCase();
  if (!needle) return [...currencySymbolPresets];
  return currencySymbolPresets.filter((preset) =>
    `${preset.code} ${preset.name} ${preset.isoName} ${preset.aliases?.join(" ") ?? ""} ${preset.symbol}`.toLowerCase().includes(needle),
  );
}

export function findCurrencySymbolPreset(form: CurrencyPresetForm): CurrencySymbolPreset | undefined {
  const code = form.code.trim().toUpperCase();
  const name = form.name.trim().toLowerCase();
  const symbol = form.symbol.trim();
  return currencySymbolPresets.find((preset) =>
    preset.code === code &&
    isKnownCurrencyName(preset, name) &&
    preset.symbol === symbol,
  );
}

function isKnownCurrencyName(preset: CurrencySymbolPreset, normalizedName: string): boolean {
  const names = [preset.name, preset.isoName, ...(preset.aliases ?? [])].map((value) => value.toLowerCase());
  return names.includes(normalizedName);
}
