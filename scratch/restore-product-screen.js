const fs = require('fs');
const path = 'D:/bccode/frontend/src/app/menu/product-screen.tsx';
let content = fs.readFileSync(path, 'utf8');

const target = 'import { CustomSelect } from "@/components/ui/select";';

const insertContent = `
function readAuthSession() {
  if (typeof window === "undefined") return null;
  try {
    const raw = window.localStorage.getItem(workspaceStorageKeys.auth);
    return raw ? JSON.parse(raw) : null;
  } catch {
    return null;
  }
}

function readWorkspaceSession() {
  if (typeof window === "undefined") return null;
  try {
    const raw = window.localStorage.getItem(workspaceStorageKeys.workspace);
    return raw ? JSON.parse(raw) : null;
  } catch {
    return null;
  }
}

export function ProductScreen({ embedded = false, language = "th" }) {
  const lang = normalizeLanguage(language);
  const text = getBarcodeText(lang);
  const { confirm, confirmationDialog } = useConfirmDialog();

  const [auth, setAuth] = useState(null);
  const [workspace, setWorkspace] = useState(null);
  const [items, setItems] = useState([]);
  const [selectedGuid, setSelectedGuid] = useState("");
  const [loading, setLoading] = useState(false);
  const [notice, setNotice] = useState(null);
  const [searchInput, setSearchInput] = useState("");
  const [search, setSearch] = useState("");

  useEffect(() => {
    const handler = setTimeout(() => {
      setSearch(searchInput);
    }, 300);
    return () => clearTimeout(handler);
  }, [searchInput]);

  const [editorOpen, setEditorOpen] = useState(false);
  const [editorMode, setEditorMode] = useState("create");
  const [editProduct, setEditProduct] = useState(null);
  const [saving, setSaving] = useState(false);
  const [productTab, setProductTab] = useState("basic");

  const itemTypes = useMemo(() => [
    { value: 0, label: text.itemTypeStock },
    { value: 1, label: text.itemTypeService },
    { value: 2, label: text.itemTypeSet },
    { value: 3, label: text.itemTypeNotStock },
  ], [text]);

  const vatTypes = useMemo(() => [
    { value: 0, label: text.vatIncluded },
    { value: 1, label: text.vatExcluded },
  ], [text]);

  const materialTypes = useMemo(() => [
    { value: 0, label: text.materialGeneral },
    { value: 1, label: text.materialMaterial },
    { value: 2, label: text.materialSemiFinished },
  ], [text]);

  const sumPointTypes = useMemo(() => [
    { value: "true", label: text.isSumPointYes },
    { value: "false", label: text.isSumPointNo },
  ], [text]);

  const foodTypes = useMemo(() => [
    { value: 0, label: text.foodTypeFood },
    { value: 1, label: text.foodTypeDrink },
    { value: 2, label: text.foodTypeAlcohol },
    { value: 3, label: text.foodTypeOther },
  ], [text]);

  const [pickerOpen, setPickerOpen] = useState(false);
`;

// Replace target, but only once
const index = content.indexOf(target);
if (index === -1) {
  console.error("Target not found!");
  process.exit(1);
}

// Check if it's already restored
if (content.includes("export function ProductScreen")) {
  console.log("Already restored!");
  process.exit(0);
}

const before = content.substring(0, index + target.length);
const after = content.substring(index + target.length);

// We need to clean up any consecutive empty lines or broken state definitions that might cause syntax errors
// In the current file, the line after target is empty, then starts with const [pickerType
// Let's strip the leading whitespace/empty lines from the 'after' string to avoid issues
const cleanAfter = after.replace(/^\s+/, '\n');

fs.writeFileSync(path, before + insertContent + cleanAfter, 'utf8');
console.log("Restored successfully!");
