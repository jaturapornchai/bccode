/**
 * Seed GL sample data for demo holding, company C01 (ร้านวัสดุก่อสร้างในไทย)
 * Usage:
 *   node scripts/seed-gl-local.mjs
 *   SEED_BASE=http://127.0.0.1:3000 node scripts/seed-gl-local.mjs
 */
import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import { fileURLToPath } from 'node:url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const root = path.resolve(__dirname, '..');

const origin = (process.env.SEED_BASE || 'http://127.0.0.1:3000').replace(/\/+$/, '');
const fixture = JSON.parse(fs.readFileSync(path.join(root, 'docs/examples/gl-thai-construction-2026.json'), 'utf8'));

const codeOf = (resource, value) => value[resource === 'accounts' ? 'accountcode' : resource === 'journals' ? 'docno' : 'code'];

function requestId(resource, code, action) {
  const h = crypto.createHash('sha256').update(`BC-GL-DEMO-BM69-v1/demo/C01/${resource}/${code}/${action}`).digest('hex');
  return `${h.slice(0, 8)}-${h.slice(8, 12)}-5${h.slice(13, 16)}-a${h.slice(17, 20)}-${h.slice(20, 32)}`;
}

const plan = [];
function add(resource, key, items) {
  for (const payload of (items || [])) {
    plan.push({ resource, key, payload, code: codeOf(resource, payload) });
  }
}

add('account-groups', 'master', fixture.accountgroups);
add('accounts', 'account', fixture.accounts);
add('fiscal-years', 'fiscalyear', [fixture.fiscalyear]);
add('periods', 'master', fixture.periods);
add('product-account-groups', 'master', fixture.productaccountgroups);
add('budgets', 'master', fixture.budgets);
add('forecast', 'master', fixture.forecast);
for (const entry of fixture.journals) {
  plan.push({ resource: 'journals', key: 'journal', payload: entry.journal, code: entry.journal.docno, post: entry.post });
}

let authorization = '';

async function call(method, endpoint, data) {
  const url = new URL(endpoint, origin);
  for (let attempt = 0; attempt < 10; attempt++) {
    const response = await fetch(url.href, {
      method,
      headers: {
        'x-bc-backend-url': `${origin}/backend/goapi`,
        'Content-Type': 'application/json',
        ...(authorization ? { Authorization: authorization } : {}),
      },
      ...(data === undefined ? {} : { body: JSON.stringify(data) }),
    });

    const body = await response.json().catch(() => null);

    if (response.status === 409 && body?.errorcode === 'GL_PROJECTION_PENDING' && method === 'GET') {
      await new Promise(r => setTimeout(r, 300));
      continue;
    }

    if (!response.ok || body?.success === false) {
      const err = new Error(`${method} ${url.pathname}: HTTP ${response.status} ${body?.message ?? body?.error ?? JSON.stringify(body)}`);
      err.status = response.status;
      err.body = body;
      throw err;
    }

    return body.data === undefined ? body : body.data;
  }
}

async function main() {
  console.log(`Starting GL sample data seeding on ${origin}...`);
  console.log(`Planned items: ${plan.length} records, ${fixture.journals.filter(x => x.post).length} posts.`);

  // 1. Login
  const login = await call('POST', '/api/auth/demo-login');
  if (typeof login.token !== 'string') throw new Error('No demo token');
  authorization = `Bearer ${login.token}`;
  console.log('✓ Logged in as demo user');

  // 2. Select holding
  await call('POST', '/api/workspace/select-holding', { holdingcode: 'demo', businesscode: 'C01' });

  // 3. Resolve branch 00000
  const holdings = await call('GET', '/api/workspace/holdings?holdingcode=demo&businesscode=C01');
  const holding = holdings.find(x => x.holdingcode?.toLowerCase() === 'demo');
  const company = holding?.companies?.find(x => x.code === 'C01');
  const branch = holding?.branches?.find(x => x.code === '00000' && [company?.guidfixed, company?.companyuid].filter(Boolean).some(id => x.companyguid === id || x.companyuid === id));
  if (!branch?.guidfixed) throw new Error('Branch 00000 not found for demo/C01');

  await call('POST', '/api/workspace/select-holding', { holdingcode: 'demo', businesscode: 'C01', branchuid: branch.guidfixed });
  console.log(`✓ Selected holding demo, company C01, branch 00000 (${branch.guidfixed})`);

  // 4. Seed records
  let createdCount = 0;
  let skippedCount = 0;
  let postedCount = 0;

  for (let i = 0; i < plan.length; i++) {
    const item = plan[i];
    const req = requestId(item.resource, item.code, 'create');
    const command = {
      resource: item.resource,
      action: 'create',
      requestid: req,
      [item.key]: item.payload,
    };

    try {
      const result = await call('POST', '/api/gl/command', command);
      createdCount++;
      
      if (item.post) {
        const postCmd = {
          resource: 'journals',
          action: 'post',
          requestid: requestId('journals', item.code, 'post'),
          id: result.id,
          version: result.version,
        };
        await call('POST', '/api/gl/command', postCmd);
        postedCount++;
      }
    } catch (err) {
      // If already created or duplicate
      if (err.body?.message && /exist|duplicate|ซ้ำ|มีอยู่แล้ว/i.test(err.body.message)) {
        skippedCount++;
      } else if (err.status === 409 || err.status === 200) {
        skippedCount++;
      } else {
        console.error(`Failed on ${item.resource} ${item.code}:`, err.message);
        throw err;
      }
    }

    if ((i + 1) % 10 === 0 || i === plan.length - 1) {
      console.log(`Progress: ${i + 1}/${plan.length} (created: ${createdCount}, skipped: ${skippedCount}, posted: ${postedCount})`);
    }
  }

  // 5. Verification
  console.log('\n--- Verification ---');
  const accounts = await call('GET', '/api/gl/accounts?limit=100');
  console.log(`Accounts in database: ${accounts.total} (items returned: ${accounts.items.length})`);

  const fiscalYears = await call('GET', '/api/gl/fiscal-years?limit=10');
  console.log(`Fiscal years in database: ${fiscalYears.total}`);

  const journals = await call('GET', '/api/gl/journals?limit=100');
  console.log(`Journals in database: ${journals.total}`);

  // Check Trial Balance report
  try {
    const tb = await call('GET', '/api/gl/reports/trialbalance?fiscalyear=2569');
    console.log(`✓ Trial balance generated successfully! Rows: ${tb.totalrows}, Debit: ${tb.totals?.debit}, Credit: ${tb.totals?.credit}`);
  } catch (tbErr) {
    console.warn(`Trial balance query note:`, tbErr.message);
  }

  console.log('\n✓ GL sample data seeding completed successfully!');
}

main().catch(err => {
  console.error('Seeding failed:', err);
  process.exit(1);
});
