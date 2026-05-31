const fs = require('fs');
const path = 'D:/bccode/frontend/src/app/menu/product-screen.tsx';
let content = fs.readFileSync(path, 'utf8');

// 1. Add barcodeSearchInput state and debounce useEffect
const barcodeSearchStateOld = 'const [barcodeSearch, setBarcodeSearch] = useState("");';
const barcodeSearchStateNew = `const [barcodeSearchInput, setBarcodeSearchInput] = useState("");
  const [barcodeSearch, setBarcodeSearch] = useState("");

  useEffect(() => {
    const handler = setTimeout(() => {
      setBarcodeSearch(barcodeSearchInput);
    }, 300);
    return () => clearTimeout(handler);
  }, [barcodeSearchInput]);`;

if (content.includes(barcodeSearchStateOld)) {
  content = content.replace(barcodeSearchStateOld, barcodeSearchStateNew);
} else {
  console.error("barcodeSearch state declaration not found!");
}

// 2. Update barcode picker button click handler to clear barcodeSearchInput
const barcodePickerTriggerOld = `                      onClick={() => {
                        setBarcodeSearch("");
                        setBarcodeList([]);
                        setShowBarcodePicker(true);
                      }}`;
const barcodePickerTriggerNew = `                      onClick={() => {
                        setBarcodeSearchInput("");
                        setBarcodeSearch("");
                        setBarcodeList([]);
                        setShowBarcodePicker(true);
                      }}`;

if (content.includes(barcodePickerTriggerOld)) {
  content = content.replace(barcodePickerTriggerOld, barcodePickerTriggerNew);
}

// 3. Update main search input bindings
const mainSearchInputOld = `              <Input
                type="search"
                placeholder={text.search}
                className="pl-8"
                value={search}
                onChange={(e) => setSearch(e.target.value)}
              />`;
const mainSearchInputNew = `              <Input
                type="search"
                placeholder={text.search}
                className="pl-8"
                value={searchInput}
                onChange={(e) => setSearchInput(e.target.value)}
              />`;

if (content.includes(mainSearchInputOld)) {
  content = content.replace(mainSearchInputOld, mainSearchInputNew);
} else {
  console.error("Main search Input binding not found!");
}

// 4. Update barcode search input bindings
const barcodeSearchInputOld = `                <Input
                  autoFocus
                  type="search"
                  value={barcodeSearch}
                  onChange={(event) => setBarcodeSearch(event.target.value)}
                  placeholder="ค้นหาด้วยบาร์โค้ด หรือชื่อสินค้า..."
                  className="h-9 pl-9"
                />`;
const barcodeSearchInputNew = `                <Input
                  autoFocus
                  type="search"
                  value={barcodeSearchInput}
                  onChange={(event) => setBarcodeSearchInput(event.target.value)}
                  placeholder="ค้นหาด้วยบาร์โค้ด หรือชื่อสินค้า..."
                  className="h-9 pl-9"
                />`;

if (content.includes(barcodeSearchInputOld)) {
  content = content.replace(barcodeSearchInputOld, barcodeSearchInputNew);
} else {
  console.error("Barcode search Input binding not found!");
}

// 5. Responsive Sidebar Layout Changes
const layoutWrapperOld = `      {/* Main split layout */}
      <div className="flex flex-1 min-h-0">`;
const layoutWrapperNew = `      {/* Main split layout */}
      <div className="flex flex-col md:flex-row flex-1 min-h-0">`;

if (content.includes(layoutWrapperOld)) {
  content = content.replace(layoutWrapperOld, layoutWrapperNew);
}

const leftSidebarOld = `        {/* Left Side: Product List */}
        <div className="w-80 shrink-0 border-r border-border bg-muted/10 flex flex-col min-h-0">`;
const leftSidebarNew = `        {/* Left Side: Product List */}
        <div className={cn("w-full md:w-80 shrink-0 border-b md:border-b-0 md:border-r border-border bg-muted/10 flex flex-col min-h-0", selectedGuid && !editorOpen ? "hidden md:flex" : "flex")}>`;

if (content.includes(leftSidebarOld)) {
  content = content.replace(leftSidebarOld, leftSidebarNew);
}

const rightPaneOld = `        {/* Right Side: Detail or Editor */}
        <div className="flex-1 overflow-y-auto bg-background min-h-0 p-4">`;
const rightPaneNew = `        {/* Right Side: Detail or Editor */}
        <div className="flex-1 overflow-y-auto bg-background min-h-0 p-4">
          {selectedGuid && !editorOpen && (
            <Button
              variant="ghost"
              size="sm"
              className="mb-4 md:hidden flex items-center gap-2"
              onClick={() => setSelectedGuid("")}
            >
              <ChevronLeft className="h-4 w-4" />
              {lang === "th" ? "กลับสู่รายการสินค้า" : "Back to list"}
            </Button>
          )}`;

if (content.includes(rightPaneOld)) {
  content = content.replace(rightPaneOld, rightPaneNew);
}

fs.writeFileSync(path, content, 'utf8');
console.log("Search and Layout logic updated successfully!");
