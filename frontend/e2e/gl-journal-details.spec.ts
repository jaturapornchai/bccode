import { expect, test } from "@playwright/test";
import { resolve } from "node:path";
import { colorThemes, defaultColorTheme } from "../src/lib/theme-data";
import { emptyJournal } from "../src/lib/general-ledger";

let harness = "";
test.beforeAll(async () => {
  if(process.env.GL_DETAILS_ACTUAL_UI === "1") return;
  const { build } = await import("vite");
  const output = await build({ configFile: false, define: { __dirname: JSON.stringify("/") }, root: process.cwd(), resolve: { alias: { "@": resolve("src") } },
    plugins: [{ name: "report-test-entry", resolveId(id) { if (id === "report-test-entry") return id; }, load(id) { if (id === "report-test-entry") return 'import { createElement } from "react"; import { createRoot } from "react-dom/client"; import { GLJournals } from "' + resolve("src/app/gl/gl-journals.tsx").replaceAll("\\", "/") + '"; createRoot(document.getElementById("root")).render(createElement(GLJournals, {route:"/gl/journals"}));'; } }],
    build: { write: false, minify: false, rollupOptions: { input: "report-test-entry", output: { inlineDynamicImports: true } } } });
  const result = Array.isArray(output) ? output[0] : output;
  if (!("output" in result)) throw new Error("Missing test bundle");
  harness = result.output.filter(item => item.type === "chunk").map(item => item.code).join("\n");
});

test("draft allocations and posted partial bank matches share the journal save flow",async({page},info)=>{
  const errors:string[]=[];page.on("pageerror",error=>errors.push(error.message));
  const posted={...emptyJournal(),id:"posted",version:1,docno:"RV1",description:"ตรวจธนาคาร",date:"2026-09-20",fiscalyear:"2026",status:"posted",lines:[{...emptyJournal().lines[0],accountcode:"1100",debit:"1000",credit:"0"},{...emptyJournal().lines[0],accountcode:"1000",debit:"0",credit:"1000"}],details:{bank_lines:[{line_number:1,bank_account_code:"B1",direction:1}]}};
  const draft={...emptyJournal(),id:"draft",version:1,docno:"JV1",description:"ทดสอบลูกหนี้",date:"2026-09-20",fiscalyear:"2026",lines:[{...emptyJournal().lines[0],accountcode:"1000",debit:"1000",credit:"0"},{...emptyJournal().lines[0],accountcode:"4000",debit:"0",credit:"1000"}]};
  const journals:Record<string,unknown>={draft,posted};const commands:Record<string,unknown>[]=[];
  const documents=[{id:"debt1",ledger:"ar",partner_code:"C1",document_no:"INV1",balance_side:1,amount:"600",remaining_amount:"600"},{id:"debt2",ledger:"ar",partner_code:"C1",document_no:"INV2",balance_side:1,amount:"400",remaining_amount:"400"},{id:"pay1",ledger:"ar",partner_code:"C1",document_no:"RECEIPT1",balance_side:2,amount:"200",remaining_amount:"200"}];
  await page.addInitScript(()=>localStorage.setItem("bc_auth",JSON.stringify({username:"test",backendUrl:location.origin})));
  await page.route("**/api/auth/refresh",route=>route.fulfill({json:{success:true,token:"test"}}));
  await page.route("**/api/gl/**",async route=>{const url=new URL(route.request().url());const path=url.pathname.replace("/api/gl/","");let data:unknown;
    if(path==="command"){const body=route.request().postDataJSON();commands.push(body);const old=journals[body.id] as typeof draft;const next={...old,...(body.action==="update"?body.journal:{}),version:old.version+1};journals[body.id]=next;data={id:body.id,version:next.version};}
    else if(path.startsWith("journals/"))data=journals[path.split("/")[1]];
    else if(path.startsWith("journal-reviews/"))data={journalid:path.split("/")[1],version:1,status:1,eventno:0,events:[]};
    else if(path==="journal-support"){const kind=url.searchParams.get("kind");const items=kind==="bank-accounts"?[{bank_account_code:"B1",bank_name:"ธนาคารทดสอบ",account_number:"123-456"}]:kind==="documents"?documents:kind==="statements"?[{id:"statement1",bank_account_code:"B1",transaction_date:"2026-09-20",bank_reference:"BANK001",amount:"600",remaining_amount:"600"}]:kind==="bank-lines"?[{journal_id:"posted",line_number:1,bank_account_code:"B1",docno:"RV1",amount:"1000",remaining_amount:"1000"}]:[];data={items,total:items.length,page:1,limit:30,sequence:1};}
    else {const items=path==="journals"?[draft,posted]:path==="fiscal-years"?[{id:"year-2026",isactive:true,closed:false,code:"2026",startdate:"2026-01-01",enddate:"2026-12-31",scale:2}]:path==="accounts"?["1000","1100","4000"].map(code=>({id:code,accountcode:code,names:[{code:"th",name:code}],isactive:true,allowposting:true,accounttype:code==="4000"?"income":"asset",normalbalance:code==="4000"?"credit":"debit"})):[];data={items,total:items.length,page:1,limit:1000,sequence:1};}
    await route.fulfill({json:{success:true,data}});
  });
  await page.route("**/details-test",route=>route.fulfill({contentType:"text/html",body:'<div id="root"></div><script type="module" src="/details-test.js"></script>'}));
  await page.route("**/details-test.js",route=>route.fulfill({contentType:"text/javascript",body:harness}));
  await page.goto(process.env.GL_DETAILS_ACTUAL_UI === "1" ? "http://127.0.0.1:3017/gl/journals" : "http://127.0.0.1:3017/details-test");
  await page.getByRole("row").filter({hasText:"JV1"}).getByRole("button",{name:"แก้ไข",exact:true}).click();
  const panel=page.getByRole("region",{name:"รายละเอียดตรวจสอบลูกหนี้ เจ้าหนี้ และธนาคาร"});
  const allocations=panel.locator("details").filter({has:page.locator("summary").filter({hasText:"จัดสรรบรรทัดบัญชีกับเอกสาร"})});
  await allocations.locator("summary").click();
  for(const [i,doc,amount] of [[0,"INV1","600"],[1,"INV2","400"]] as const){
    await allocations.getByRole("button",{name:"เพิ่มจัดสรรบรรทัดบัญชีกับเอกสาร",exact:true}).click();
    await allocations.getByRole("button",{name:"เอกสารที่จัดสรร",exact:true}).nth(i).click();
    const picker=page.getByRole("dialog",{name:"เลือกหลักฐานบัญชี"});
    await expect(picker).toBeVisible();
    await expect(picker.getByRole("textbox",{name:"ค้นหาหลักฐาน"})).toBeFocused();
    await page.keyboard.press("Tab");
    expect(await picker.evaluate(node=>({inside:node.contains(document.activeElement),active:document.activeElement?.outerHTML,open:(node as HTMLDialogElement).open}))).toMatchObject({inside:true});
    await picker.getByRole("button").filter({hasText:doc}).click();
    await expect(allocations.getByRole("button",{name:"เอกสารที่จัดสรร",exact:true}).nth(i)).toBeFocused();
    await allocations.getByRole("textbox",{name:"ยอดจัดสรร / จำนวนเงิน",exact:true}).nth(i).fill(amount);
    await allocations.getByRole("textbox",{name:"ยอดจัดสรร / จำนวนเงิน",exact:true}).nth(i).press("Tab");
  }
  if(process.env.GL_DETAILS_ACTUAL_UI === "1") {
    await page.setViewportSize({width:1600,height:1000});
    await panel.screenshot({path:info.outputPath("details-light.png")});
    await page.evaluate(vars=>{document.documentElement.dataset.theme="dark";for(const [key,value] of Object.entries(vars))document.documentElement.style.setProperty(key,value);},colorThemes.find(theme=>theme.id===defaultColorTheme)!.dark);
    await panel.screenshot({path:info.outputPath("details-dark.png")});
    await page.setViewportSize({width:390,height:844});
    await panel.screenshot({path:info.outputPath("details-mobile.png")});
    expect(await panel.evaluate(node=>node.scrollWidth<=node.clientWidth+1)).toBe(true);
    await page.setViewportSize({width:1600,height:1000});
    await page.evaluate(vars=>{document.documentElement.dataset.theme="light";for(const [key,value] of Object.entries(vars))document.documentElement.style.setProperty(key,value);},colorThemes.find(theme=>theme.id===defaultColorTheme)!.light);
  }
  await page.getByRole("row").filter({hasText:"RV1"}).click();
  await expect(page.getByRole("dialog")).toBeVisible();
  await page.getByRole("dialog").getByRole("button",{name:"ยกเลิก",exact:true}).last().click();
  await page.getByRole("button",{name:/บันทึกฉบับร่าง/}).last().click();
  await page.getByRole("dialog").getByRole("button",{name:"บันทึกฉบับร่าง",exact:true}).click();
  await expect.poll(()=>commands.length).toBe(1);
  const saved=commands[0].journal as {details:{allocations:{document_id:string;amount:string;line_number:number}[]}};
  expect(saved.details.allocations.map(row=>[row.document_id,row.amount,row.line_number])).toEqual([["debt1","600.00",1],["debt2","400.00",1]]);
  await page.getByRole("row").filter({hasText:"RV1"}).click();
  const bankLines=panel.locator("details").filter({has:page.locator("summary").filter({hasText:"บัญชีธนาคารของบรรทัด GL"})});

  await expect(bankLines.getByRole("button",{name:"บัญชีธนาคาร",exact:true})).toContainText("ธนาคารทดสอบ");
  await page.getByRole("button",{name:"เพิ่มผลกระทบยอด",exact:true}).click();
  const matches=panel.locator("details").filter({has:page.locator("summary").filter({hasText:"จับคู่บัญชีกับ Statement"})});
  await matches.locator("summary").click();
  await matches.getByRole("button",{name:"เพิ่มจับคู่บัญชีกับ Statement",exact:true}).click();
  await matches.getByRole("button",{name:"รายการ Statement",exact:true}).click();
  await page.getByRole("dialog",{name:"เลือกหลักฐานบัญชี"}).getByRole("button").filter({hasText:"BANK001"}).click();
  await matches.getByRole("button",{name:"รายการบัญชีธนาคารที่จะจับคู่",exact:true}).click();
  await page.getByRole("dialog",{name:"เลือกหลักฐานบัญชี"}).getByRole("button").filter({hasText:/บรรทัด 1/}).first().click();
  await matches.getByRole("textbox",{name:"ยอดจัดสรร / จำนวนเงิน",exact:true}).fill("600");
  await matches.getByRole("textbox",{name:"ยอดจัดสรร / จำนวนเงิน",exact:true}).press("Tab");
  const settlements=panel.locator("details").filter({has:page.locator("summary").filter({hasText:"ตัดยอดหลายบิล / ชำระบางส่วน"})});
  await settlements.locator("summary").click();
  await settlements.getByRole("button",{name:"เพิ่มตัดยอดหลายบิล / ชำระบางส่วน",exact:true}).click();
  await settlements.getByRole("button",{name:"บิลเพิ่มหนี้",exact:true}).click();
  await page.getByRole("dialog",{name:"เลือกหลักฐานบัญชี"}).getByRole("button").filter({hasText:"INV1"}).click();
  await settlements.getByRole("button",{name:"เอกสารลดหนี้ / ผลชำระ",exact:true}).click();
  await page.getByRole("dialog",{name:"เลือกหลักฐานบัญชี"}).getByRole("button").filter({hasText:"RECEIPT1"}).click();
  await settlements.getByRole("textbox",{name:"ยอดจัดสรร / จำนวนเงิน",exact:true}).fill("200");
  await settlements.getByRole("textbox",{name:"ยอดจัดสรร / จำนวนเงิน",exact:true}).press("Tab");
  await page.getByPlaceholder("จำเป็นสำหรับสร้างรายการกลับบัญชี").fill("ตรวจ Statement แล้ว");
  await page.getByRole("button",{name:"บันทึกผลกระทบยอด",exact:true}).click();
  await expect.poll(()=>commands.length).toBe(2);
  expect(commands[1]).toMatchObject({action:"reconcile",id:"posted",version:1,journal:{details:{matches:[{statement_line_id:"statement1",line_number:1,amount:"600.00"}],settlements:[{debt_document_id:"debt1",payment_document_id:"pay1",amount:"200.00",partner_code:"C1",ledger:"ar"}]}}});
  expect(Object.keys(commands[1].journal as object)).toEqual(["details"]);
  expect(errors).toEqual([]);
});
