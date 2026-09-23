import { execFileSync } from 'child_process';

/**
 * UAT data-layer checks against the central PostgreSQL DB (`bcai_projection`)
 * via `docker exec <container> psql`. execFileSync (no shell) so SQL quotes
 * survive on Windows. Override with PG_CONTAINER / PG_USER / PG_DB; the
 * password is never needed (psql runs inside the container over its socket).
 */
const PG_CONTAINER = process.env.PG_CONTAINER ?? 'postgres';
const PG_USER = process.env.PG_USER ?? 'postgres';
const PG_DB = process.env.PG_DB ?? 'bcai_projection';

/** Quote a test-generated value as a SQL string literal. */
export function sqlText(value: string): string {
  return `'${value.replace(/'/g, "''")}'`;
}

/** Run one statement; returns the last output line (unaligned, tuples only). */
export function pgQuery(sql: string): string {
  const args = ['exec', PG_CONTAINER, 'psql', '-U', PG_USER, '-d', PG_DB, '-v', 'ON_ERROR_STOP=1', '-At', '-c', sql];
  // docker exec can blip transiently — retry once
  for (let attempt = 0; ; attempt++) {
    try {
      return execFileSync('docker', args, { timeout: 45000 }).toString().trim().split('\n').pop() ?? '';
    } catch (err) {
      if (attempt >= 1) throw err;
      Atomics.wait(new Int32Array(new SharedArrayBuffer(4)), 0, 0, 1500); // sync sleep
    }
  }
}

/** `SELECT count(*) FROM <from> WHERE <where>` as a number. */
export function pgCount(from: string, where = 'true'): number {
  return Number(pgQuery(`SELECT count(*) FROM ${from} WHERE ${where}`));
}

/** Single row as parsed JSON (null when no row). */
export function pgRow<T = Record<string, unknown>>(sql: string): T | null {
  const out = pgQuery(`SELECT row_to_json(r) FROM (${sql}) r LIMIT 1`);
  return out ? (JSON.parse(out) as T) : null;
}
