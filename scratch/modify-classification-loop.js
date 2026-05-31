const fs = require('fs');
const path = 'D:/bccode/frontend/src/app/menu/product-screen.tsx';
let content = fs.readFileSync(path, 'utf8');

// Define classification tab loop structure
const loopContent = `              {/* Tab: classification */}
              {productTab === "classification" && (
                <div className="border border-border rounded-lg p-4 space-y-4">
                  <h4 className="font-semibold text-sm border-b border-border pb-1">{lang === "th" ? "ประเภทกลุ่ม & หมวดหมู่หลัก" : "Groups & Categories"}</h4>
                  <div className="grid gap-4 sm:grid-cols-2">
                    {[
                      { key: "group", label: text.group, code: editProduct.group_code, names: editProduct.group_names },
                      { key: "groupsubone", label: text.groupsubone, code: editProduct.groupsubonecode, names: editProduct.groupsubonenames },
                      { key: "groupsubtwo", label: text.groupsubtwo, code: editProduct.groupsubtwocode, names: editProduct.groupsubtwonames },
                      { key: "brand", label: text.brand, code: editProduct.brand_code, names: editProduct.brandnames },
                      { key: "category", label: text.category, code: editProduct.categorycode, names: editProduct.category_names },
                      { key: "class", label: text.class, code: editProduct.classcode, names: editProduct.classnames },
                      { key: "design", label: text.design, code: editProduct.designcode, names: editProduct.designnames },
                      { key: "model", label: text.model, code: editProduct.modelcode, names: editProduct.modelnames },
                      { key: "pattern", label: text.pattern, code: editProduct.patterncode, names: editProduct.patternnames },
                      { key: "grade", label: text.grade, code: editProduct.gradecode, names: editProduct.gradenames },
                    ].map((field) => (
                      <div key={field.key} className="space-y-1">
                        <label className="text-xs text-muted-foreground">{field.label}</label>
                        <div className="flex gap-1.5">
                          <Input readOnly value={field.code ? \`\${field.code} — \${pickName(field.names, lang)}\` : ""} />
                          <Button type="button" variant="outline" aria-label={\`Select \${field.label}\`} onClick={() => openPicker(field.key, field.key)}>...</Button>
                          {field.code && (
                            <Button type="button" variant="ghost" aria-label={\`Clear \${field.label}\`} onClick={() => clearPickerField(field.key)}>
                              <X className="h-4 w-4" />
                            </Button>
                          )}
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              )}`;

// We replace from the start of classification tab to the end of the Grade block
const startIndex = content.indexOf('{/* Tab: classification */}');
const endIndex = content.indexOf('{/* Tab: units */}');

if (startIndex !== -1 && endIndex !== -1) {
  content = content.substring(0, startIndex) + loopContent + '\n\n' + content.substring(endIndex);
  console.log("Classification tab replaced with loop successfully!");
} else {
  console.error("Could not find classification or units tab anchors!");
}

fs.writeFileSync(path, content, 'utf8');
