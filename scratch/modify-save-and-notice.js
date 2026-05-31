const fs = require('fs');
const path = 'D:/bccode/frontend/src/app/menu/product-screen.tsx';
let content = fs.readFileSync(path, 'utf8');

// 1. Update bindBarcodeOnSave logic to not be fire-and-forget
const oldBindLogic = `      // Link Barcode if selected
      if (editorMode === "create" && bindBarcodeOnSave && data.data?.guidfixed) {
        const createdGuid = data.data.guidfixed;
        const createdCode = editProduct.code;
        try {
          const bcRes = await fetch(\`/api/product-barcode/\${encodeURIComponent(bindBarcodeOnSave.guidfixed)}\`, {
            headers: {
              Authorization: \`Bearer \${auth.token}\`,
              "x-bc-backend-url": auth.backendUrl,
            },
          });
          const bcJson = await bcRes.json();
          if (bcJson.success && bcJson.data) {
            const fullBarcode = bcJson.data;
            fullBarcode.item_guid = createdGuid;
            fullBarcode.itemcode = createdCode;

            await fetch(\`/api/product-barcode/\${encodeURIComponent(bindBarcodeOnSave.guidfixed)}\`, {
              method: "PUT",
              headers: {
                "Content-Type": "application/json",
                Authorization: \`Bearer \${auth.token}\`,
                "x-bc-backend-url": auth.backendUrl,
              },
              body: JSON.stringify({
                backendUrl: auth.backendUrl,
                data: fullBarcode,
              }),
            });
          }
        } catch (linkErr) {
          console.error("Failed to link barcode to new product", linkErr);
        }
      }`;

const newBindLogic = `      // Link Barcode if selected
      if (editorMode === "create" && bindBarcodeOnSave && data.data?.guidfixed) {
        const createdGuid = data.data.guidfixed;
        const createdCode = editProduct.code;

        const bcRes = await fetch(\`/api/product-barcode/\${encodeURIComponent(bindBarcodeOnSave.guidfixed)}\`, {
          headers: {
            Authorization: \`Bearer \${auth.token}\`,
            "x-bc-backend-url": auth.backendUrl,
          },
        });
        const bcJson = await bcRes.json();
        if (!bcRes.ok || !bcJson.success || !bcJson.data) {
          throw new Error(bcJson.message || "Failed to fetch barcode details for linking");
        }

        const fullBarcode = bcJson.data;
        fullBarcode.item_guid = createdGuid;
        fullBarcode.itemcode = createdCode;

        const putRes = await fetch(\`/api/product-barcode/\${encodeURIComponent(bindBarcodeOnSave.guidfixed)}\`, {
          method: "PUT",
          headers: {
            "Content-Type": "application/json",
            Authorization: \`Bearer \${auth.token}\`,
            "x-bc-backend-url": auth.backendUrl,
          },
          body: JSON.stringify({
            backendUrl: auth.backendUrl,
            data: fullBarcode,
          }),
        });
        const putJson = await putRes.json();
        if (!putRes.ok || !putJson.success) {
          throw new Error(putJson.message || "Failed to save barcode link");
        }
      }`;

if (content.includes(oldBindLogic)) {
  content = content.replace(oldBindLogic, newBindLogic);
} else {
  // Try fallback in case of formatting
  console.log("oldBindLogic not matched by exact string, trying substring/regex replacement...");
  // Let's do a replace based on key anchors
  const startIndex = content.indexOf('// Link Barcode if selected');
  const endIndex = content.indexOf('setNotice({ type: "success", text: text.saveSuccess });');
  if (startIndex !== -1 && endIndex !== -1) {
    content = content.substring(0, startIndex) + newBindLogic + '\n\n      ' + content.substring(endIndex);
    console.log("Replaced using substring indices!");
  } else {
    console.error("Could not find bind anchors!");
  }
}

// 2. Update notice type "info" color in JSX
const noticeStyleOld = `      {notice && (
        <div className={cn("px-4 py-2 text-sm flex items-center gap-2", notice.type === "success" ? "bg-emerald-500/10 text-emerald-500" : "bg-destructive/10 text-destructive")}>`;
const noticeStyleNew = `      {notice && (
        <div className={cn(
          "px-4 py-2 text-sm flex items-center gap-2 rounded-md border",
          notice.type === "success"
            ? "bg-emerald-500/10 text-emerald-500 border-emerald-500/20"
            : notice.type === "info"
              ? "bg-blue-500/10 text-blue-500 border-blue-500/20"
              : "bg-destructive/10 text-destructive border-destructive/20"
        )}>`;

if (content.includes(noticeStyleOld)) {
  content = content.replace(noticeStyleOld, noticeStyleNew);
} else {
  // Regex match fallback
  const noticeRegex = /\{notice && \(\s*<div className=\{cn\("px-4 py-2 text-sm flex items-center gap-2", notice\.type === "success" \? "bg-emerald-500\/10 text-emerald-500" : "bg-destructive\/10 text-destructive"\)\}>/;
  content = content.replace(noticeRegex, noticeStyleNew);
}

fs.writeFileSync(path, content, 'utf8');
console.log("Save and notice logic updated successfully!");
