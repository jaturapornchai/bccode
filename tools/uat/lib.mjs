// Shared helpers for the one-command UAT (`npm run uat`): API calls, psql checks after every step,
// exact money strings (Rule 20: integer satang, never float math) and a compact step log.
import { execFileSync } from 'node:child_process';

/** psql inside the postgres container (no password, no shell → quotes survive on Windows). */
export function makePsql({ container = 'postgres', user = 'postgres', db }) {
  return (sql, database = db) =>
    execFileSync('docker', ['exec', container, 'psql', '-U', user, '-d', database, '-v', 'ON_ERROR_STOP=1', '-At', '-F', '|', '-c', sql], {
      encoding: 'utf8',
      timeout: 45000,
    }).trim();
}

/** Quote a test-generated value as a SQL string literal. */
export const sqlText = (value) => `'${String(value).replace(/'/g, "''")}'`;

/** Satang (integer) → "1234.50". */
export const baht = (satang) => {
  const s = BigInt(satang);
  const neg = s < 0n;
  const abs = neg ? -s : s;
  return `${neg ? '-' : ''}${abs / 100n}.${String(abs % 100n).padStart(2, '0')}`;
};

/** Compare two decimal texts exactly ("1605" == "1605.00" == "1605.000"). */
export const sameDecimal = (a, b) => {
  const norm = (v) => {
    const t = String(v ?? '').trim();
    if (!/^-?\d+(\.\d+)?$/.test(t)) return `NaN:${t}`;
    let [i, f = ''] = t.replace(/^-/, '').split('.');
    i = i.replace(/^0+(?=\d)/, '');
    f = f.replace(/0+$/, '');
    const body = f ? `${i}.${f}` : i;
    return body === '0' ? '0' : (t.startsWith('-') ? '-' : '') + body;
  };
  return norm(a) === norm(b);
};

/** Money value for a request body: serialized as raw decimal text (1605.00), never via float. */
export const dec = (satang) => ({ __dec: baht(satang) });
const encode = (body) => JSON.stringify(body).replace(/\{"__dec":"(-?[0-9]+\.[0-9]+)"\}/g, '$1');

/** JSON HTTP client bound to a base URL + auth headers. */
export function makeApi(base, headers) {
  return async (method, path, body) => {
    const res = await fetch(base + path, {
      method,
      headers: { 'Content-Type': 'application/json', ...headers },
      body: body === undefined ? undefined : encode(body),
    });
    const text = await res.text();
    let json = null;
    try { json = text ? JSON.parse(text) : null; } catch { /* non-JSON body */ }
    return { status: res.status, json, text: text.slice(0, 300) };
  };
}

/** Row counts of the given tables, for "nothing else changed" checks. */
export function snapshot(psql, tables) {
  return Object.fromEntries(tables.map((t) => [t, Number(psql(`select count(*) from ${t}`))]));
}

export const diff = (a, b) => Object.fromEntries(Object.keys(a).map((k) => [k, b[k] - a[k]]).filter(([, v]) => v !== 0));
