const fs = require('fs');
const path = 'D:/bccode/frontend/src/app/menu/product-screen.tsx';
let content = fs.readFileSync(path, 'utf8');

// 1. Fix the basic info card at view mode details
const brokenCardOld = `                <Card>
                  <CardHeader className="pb-2">

                  </CardContent>
                </Card>`;

const brokenCardNew = `                <Card>
                  <CardHeader className="pb-2">
                    <CardTitle className="text-sm font-semibold">{lang === "th" ? "ข้อมูลเบื้องต้น" : "Basic Info"}</CardTitle>
                  </CardHeader>
                  <CardContent className="space-y-2 text-sm">
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">{text.itemType}:</span>
                      <span className="font-semibold">{itemTypes.find((t) => t.value === selectedProduct.item_type)?.label ?? selectedProduct.item_type}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">{text.vatType}:</span>
                      <span className="font-semibold">{vatTypes.find((t) => t.value === selectedProduct.vat_type)?.label ?? selectedProduct.vat_type}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">{text.materialType}:</span>
                      <span className="font-semibold">{materialTypes.find((t) => t.value === selectedProduct.materialtype)?.label ?? selectedProduct.materialtype ?? "-"}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">{text.isSumPoint}:</span>
                      <span className="font-semibold">{selectedProduct.issumpoint ? text.isSumPointYes : text.isSumPointNo}</span>
                    </div>
                  </CardContent>
                </Card>`;

if (content.includes(brokenCardOld)) {
  content = content.replace(brokenCardOld, brokenCardNew);
} else {
  // If formatting/newlines are slightly different, use a regex or check index-based replace
  // Let's do a substring replace if target matches
  console.log("brokenCardOld target not matched by exact string, trying fallback...");
  const searchPattern = /<Card>\s*<CardHeader className="pb-2">\s*<\/CardContent>\s*<\/Card>/;
  content = content.replace(searchPattern, brokenCardNew);
}

// 2. Fix keyof LocalText issue
content = content.replace('pickerType as keyof LocalText', 'pickerType as keyof typeof text');

// 3. Inject text and foodTypes inside TabProductRestaurant
const tabProductRestaurantOld = `function TabProductRestaurant({
  value,
  onChange,
  auth,
  language,
  shopLanguages,
}: {
  value: Product;
  onChange: ProductStateAction;
  auth: AuthSession | null;
  language: string;
  shopLanguages: string[];
}) {
  const updR = useCallback(`;

const tabProductRestaurantNew = `function TabProductRestaurant({
  value,
  onChange,
  auth,
  language,
  shopLanguages,
}: {
  value: Product;
  onChange: ProductStateAction;
  auth: AuthSession | null;
  language: string;
  shopLanguages: string[];
}) {
  const text = getBarcodeText(language);
  const foodTypes = [
    { value: 0, label: text.foodTypeFood },
    { value: 1, label: text.foodTypeDrink },
    { value: 2, label: text.foodTypeAlcohol },
    { value: 3, label: text.foodTypeOther },
  ];

  const updR = useCallback(`;

if (content.includes(tabProductRestaurantOld)) {
  content = content.replace(tabProductRestaurantOld, tabProductRestaurantNew);
} else {
  console.error("TabProductRestaurant signature not found!");
}

fs.writeFileSync(path, content, 'utf8');
console.log("Card and Restaurant tab logic updated successfully!");
