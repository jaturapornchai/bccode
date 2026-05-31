const fs = require('fs');
const path = 'D:/bccode/frontend/src/app/menu/product-screen.tsx';
let content = fs.readFileSync(path, 'utf8');

// Replace standard JS signatures with TS signatures
content = content.replace(
  'function readAuthSession() {',
  'function readAuthSession(): AuthSession | null {'
);

content = content.replace(
  'function readWorkspaceSession() {',
  'function readWorkspaceSession(): WorkspaceSession | null {'
);

content = content.replace(
  'export function ProductScreen({ embedded = false, language = "th" }) {',
  'export function ProductScreen({ embedded = false, language = "th" }: { embedded?: boolean; language?: LanguageCode }) {'
);

content = content.replace(
  'const [auth, setAuth] = useState(null);',
  'const [auth, setAuth] = useState<AuthSession | null>(null);'
);

content = content.replace(
  'const [workspace, setWorkspace] = useState(null);',
  'const [workspace, setWorkspace] = useState<WorkspaceSession | null>(null);'
);

content = content.replace(
  'const [items, setItems] = useState([]);',
  'const [items, setItems] = useState<Product[]>([]);'
);

content = content.replace(
  'const [notice, setNotice] = useState(null);',
  'const [notice, setNotice] = useState<{ type: "success" | "error" | "info"; text: string } | null>(null);'
);

content = content.replace(
  'const [editProduct, setEditProduct] = useState(null);',
  'const [editProduct, setEditProduct] = useState<Product | null>(null);'
);

content = content.replace(
  'const [editorMode, setEditorMode] = useState("create");',
  'const [editorMode, setEditorMode] = useState<"create" | "edit">("create");'
);

content = content.replace(
  'const [productTab, setProductTab] = useState("basic");',
  'const [productTab, setProductTab] = useState<"basic" | "classification" | "units" | "bom" | "media" | "restaurant" | "timeforsales" | "business" | "misc">("basic");'
);

fs.writeFileSync(path, content, 'utf8');
console.log("Types added successfully!");
