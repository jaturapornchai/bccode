const fs = require('fs');
const path = 'D:/bccode/frontend/src/app/menu/product-barcode-screen.tsx';
let content = fs.readFileSync(path, 'utf8');

// 1. Declare searchInput and useEffect debounce
const searchStateOld = '  const [search, setSearch] = useState("");';
const searchStateNew = `  const [searchInput, setSearchInput] = useState("");
  const [search, setSearch] = useState("");

  useEffect(() => {
    const handler = setTimeout(() => {
      setSearch(searchInput);
    }, 300);
    return () => clearTimeout(handler);
  }, [searchInput]);`;

if (content.includes(searchStateOld)) {
  content = content.replace(searchStateOld, searchStateNew);
} else {
  console.error("search state declaration not found in barcode screen!");
}

// 2. Update search Input binding in JSX
const inputBindingOld = `                  value={search}
                  onChange={(event) => setSearch(event.target.value)}`;
const inputBindingNew = `                  value={searchInput}
                  onChange={(event) => setSearchInput(event.target.value)}`;

if (content.includes(inputBindingOld)) {
  content = content.replace(inputBindingOld, inputBindingNew);
} else {
  console.error("Search input binding not found in barcode screen!");
}

fs.writeFileSync(path, content, 'utf8');
console.log("Barcode search logic updated successfully!");
