"use client";

import React, { useCallback, useEffect, useMemo, useRef, useState } from "react";
import {
  ChevronDown,
  ChevronRight,
  Edit3,
  GripVertical,
  Loader2,
  FolderPlus,
  Trash2,
} from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

type SettingRecord = Record<string, unknown>;

interface ProductSubgroupTreeViewProps {
  auth: { token: string; backendUrl: string } | null;
  workspace: { shop: { holdingcode: string } } | null;
  language: string;
  records: SettingRecord[];
  selectedGuid: string;
  setSelectedGuid: (guid: string) => void;
  searchQuery: string;
  onOpenCreate: (parentCode?: string) => void;
  onOpenEdit: (record: SettingRecord) => void;
  onDeleteRecord: (record: SettingRecord) => void;
  onRefresh?: () => void;
  saving: boolean;
  loading: boolean;
}

interface SubgroupNode {
  guidfixed: string;
  code: string;
  parentcode: string;
  sortorder: number;
  names: { code: string; name: string }[];
  isdisabled?: boolean;
}

interface TreeNode {
  item: SubgroupNode;
  children: TreeNode[];
  depth: number;
}

type DropPosition = "before" | "after" | "inside";

const LEVEL_STYLES = [
  { name: "text-[15px] font-bold text-foreground", grip: "text-primary/45" },
  { name: "font-semibold text-sky-900 dark:text-sky-100", grip: "text-sky-500/55" },
  { name: "font-semibold text-emerald-900 dark:text-emerald-100", grip: "text-emerald-500/55" },
  { name: "font-medium text-amber-900 dark:text-amber-100", grip: "text-amber-500/60" },
  { name: "font-medium text-violet-900 dark:text-violet-100", grip: "text-violet-500/55" },
] as const;

const levelStyle = (depth: number) =>
  LEVEL_STYLES[Math.min(depth, LEVEL_STYLES.length - 1)];

const pickName = (names: { code: string; name: string }[], lang: string): string => {
  const match = names.find((n) => n.code === lang);
  return match?.name || names[0]?.name || "";
};

function buildTree(items: SubgroupNode[]): TreeNode[] {
  const byParent = new Map<string, SubgroupNode[]>();
  for (const item of items) {
    const key = item.parentcode || "";
    const list = byParent.get(key) || [];
    list.push(item);
    byParent.set(key, list);
  }
  // Sort each level by sortorder
  for (const list of byParent.values()) {
    list.sort((a, b) => a.sortorder - b.sortorder);
  }

  const build = (parentCode: string, depth: number): TreeNode[] => {
    const children = byParent.get(parentCode) || [];
    return children.map((item) => ({
      item,
      children: build(item.code, depth + 1),
      depth,
    }));
  };

  return build("", 0);
}

function flattenTree(nodes: TreeNode[]): TreeNode[] {
  const result: TreeNode[] = [];
  const walk = (list: TreeNode[]) => {
    for (const node of list) {
      result.push(node);
      walk(node.children);
    }
  };
  walk(nodes);
  return result;
}

export function ProductSubgroupTreeView({
  auth,
  workspace,
  language,
  records,
  selectedGuid,
  setSelectedGuid,
  searchQuery,
  onOpenCreate,
  onOpenEdit,
  onDeleteRecord,
  onRefresh,
  saving,
  loading,
}: ProductSubgroupTreeViewProps) {
  const [expandedNodes, setExpandedNodes] = useState<Record<string, boolean>>({});
  const [dragGuid, setDragGuid] = useState<string | null>(null);
  const [dropTarget, setDropTarget] = useState<{ code: string; position: DropPosition } | null>(null);
  const [moveError, setMoveError] = useState("");
  const [moveInFlight, setMoveInFlight] = useState(false);
  const treeRef = useRef<HTMLDivElement>(null);

  const items: SubgroupNode[] = useMemo(() => {
    return records
      .map((r) => ({
        guidfixed: String(r.guidfixed || ""),
        code: String(r.code || ""),
        parentcode: String(r.parentcode || ""),
        sortorder: Number(r.sortorder ?? 0),
        names: Array.isArray(r.names)
          ? r.names.map((n: any) => ({ code: String(n.code || ""), name: String(n.name || "") }))
          : [],
        isdisabled: Boolean(r.isdisabled),
      }))
      .filter((item) => {
        if (!searchQuery.trim()) return true;
        const q = searchQuery.toLowerCase();
        const name = pickName(item.names, language).toLowerCase();
        return item.code.toLowerCase().includes(q) || name.includes(q);
      });
  }, [records, searchQuery, language]);

  const tree = useMemo(() => buildTree(items), [items]);
  const flatList = useMemo(() => flattenTree(tree), [tree]);

  // Auto-expand all on first load
  useEffect(() => {
    const all: Record<string, boolean> = {};
    for (const item of items) {
      if (item.code) all[item.code] = true;
    }
    setExpandedNodes(all);
  }, [items.length]); // eslint-disable-line react-hooks/exhaustive-deps

  const toggleExpand = useCallback((code: string) => {
    setExpandedNodes((prev) => ({ ...prev, [code]: !prev[code] }));
  }, []);

  // Save reorder/reparent via atlas update
  const saveItem = useCallback(
    async (item: SubgroupNode, changes: Partial<SubgroupNode>) => {
      if (!auth || !workspace) return;
      setMoveInFlight(true);
      setMoveError("");
      try {
        const updated = { ...item, ...changes, holdingcode: workspace.shop.holdingcode };
        const res = await fetch("/api/system-settings/productsubgroup", {
          method: "PUT",
          headers: {
            "Content-Type": "application/json",
            Authorization: `Bearer ${auth.token}`,
            "x-bc-backend-url": auth.backendUrl,
          },
          body: JSON.stringify({
            ...updated,
            guidfixed: item.guidfixed,
            holdingcode: workspace.shop.holdingcode,
          }),
        });
        if (!res.ok) {
          const data = await res.json().catch(() => ({}));
          throw new Error(data.message || "Failed to save");
        }
        onRefresh?.();
      } catch (err: any) {
        setMoveError(err.message || "Failed to save");
      } finally {
        setMoveInFlight(false);
      }
    },
    [auth, workspace, onRefresh],
  );

  // Handle drop: reorder or reparent
  const handleDrop = useCallback(
    async (targetCode: string, position: DropPosition) => {
      if (!dragGuid) return;
      const dragged = items.find((i) => i.guidfixed === dragGuid);
      const target = items.find((i) => i.code === targetCode);
      if (!dragged || !target || dragged.code === targetCode) {
        setDragGuid(null);
        setDropTarget(null);
        return;
      }

      if (position === "inside") {
        // Reparent: move dragged inside target
        const siblings = items
          .filter((i) => i.parentcode === target.code && i.guidfixed !== dragged.guidfixed)
          .sort((a, b) => a.sortorder - b.sortorder);
        await saveItem(dragged, { parentcode: target.code, sortorder: siblings.length });
      } else {
        // Reorder: place before/after target at same parent level
        const newParent = target.parentcode;
        const siblings = items
          .filter((i) => i.parentcode === newParent && i.guidfixed !== dragged.guidfixed)
          .sort((a, b) => a.sortorder - b.sortorder);
        const targetIdx = siblings.findIndex((i) => i.code === targetCode);
        const insertIdx = position === "before" ? targetIdx : targetIdx + 1;
        siblings.splice(insertIdx, 0, { ...dragged, parentcode: newParent });
        // Save new sortorder for dragged item
        await saveItem(dragged, { parentcode: newParent, sortorder: insertIdx });
        // Update siblings sortorder
        for (let i = 0; i < siblings.length; i++) {
          if (siblings[i].guidfixed !== dragged.guidfixed && siblings[i].sortorder !== i) {
            await saveItem(siblings[i], { sortorder: i });
          }
        }
      }
      setDragGuid(null);
      setDropTarget(null);
    },
    [dragGuid, items, saveItem],
  );

  const renderNode = (node: TreeNode) => {
    const { item, children, depth } = node;
    const isExpanded = expandedNodes[item.code] !== false;
    const hasChildren = children.length > 0;
    const isSelected = selectedGuid === item.guidfixed;
    const style = levelStyle(depth);
    const isDropTarget = dropTarget?.code === item.code;

    return (
      <div key={item.guidfixed}>
        <div
          className={cn(
            "group/row flex items-center gap-1 rounded-md px-2 py-1.5 cursor-pointer transition-colors",
            "hover:bg-muted/60",
            isSelected && "bg-primary/10 ring-1 ring-primary/30",
            isDropTarget && dropTarget?.position === "inside" && "ring-2 ring-primary/50 bg-primary/5",
            dragGuid === item.guidfixed && "opacity-40",
          )}
          style={{ paddingLeft: `${depth * 20 + 8}px` }}
          onClick={() => setSelectedGuid(item.guidfixed)}
          onDragOver={(e) => {
            e.preventDefault();
            const rect = e.currentTarget.getBoundingClientRect();
            const y = e.clientY - rect.top;
            const h = rect.height;
            let pos: DropPosition;
            if (y < h * 0.25) pos = "before";
            else if (y > h * 0.75) pos = "after";
            else pos = "inside";
            setDropTarget({ code: item.code, position: pos });
          }}
          onDragLeave={() => {
            if (dropTarget?.code === item.code) setDropTarget(null);
          }}
          onDrop={(e) => {
            e.preventDefault();
            if (dropTarget) void handleDrop(dropTarget.code, dropTarget.position);
          }}
        >
          {/* Drag handle */}
          <span
            className={cn("shrink-0 cursor-grab active:cursor-grabbing", style.grip)}
            draggable
            onDragStart={(e) => {
              e.dataTransfer.effectAllowed = "move";
              setDragGuid(item.guidfixed);
            }}
            onDragEnd={() => {
              setDragGuid(null);
              setDropTarget(null);
            }}
          >
            <GripVertical className="h-4 w-4" />
          </span>

          {/* Expand/collapse */}
          {hasChildren ? (
            <button
              className="shrink-0 p-0.5 rounded hover:bg-muted"
              onClick={(e) => {
                e.stopPropagation();
                toggleExpand(item.code);
              }}
            >
              {isExpanded ? (
                <ChevronDown className="h-4 w-4 text-muted-foreground" />
              ) : (
                <ChevronRight className="h-4 w-4 text-muted-foreground" />
              )}
            </button>
          ) : (
            <span className="w-5" />
          )}

          {/* Name */}
          <span className={cn("flex-1 truncate", style.name, item.isdisabled && "line-through opacity-50")}>
            {pickName(item.names, language) || item.code}
          </span>

          {/* Code badge */}
          <span className="shrink-0 text-[11px] text-muted-foreground font-mono bg-muted/50 rounded px-1.5 py-0.5">
            {item.code}
          </span>

          {/* Actions */}
          <div className="shrink-0 flex items-center gap-0.5 opacity-0 group-hover/row:opacity-100 transition-opacity">
            <Button
              variant="ghost"
              size="icon"
              className="h-6 w-6"
              title="เพิ่มกลุ่มย่อย"
              onClick={(e) => {
                e.stopPropagation();
                onOpenCreate(item.code);
              }}
            >
              <FolderPlus className="h-3.5 w-3.5" />
            </Button>
            <Button
              variant="ghost"
              size="icon"
              className="h-6 w-6"
              title="แก้ไข"
              onClick={(e) => {
                e.stopPropagation();
                onOpenEdit(records.find((r) => String(r.guidfixed) === item.guidfixed) || {});
              }}
            >
              <Edit3 className="h-3.5 w-3.5" />
            </Button>
            <Button
              variant="ghost"
              size="icon"
              className="h-6 w-6 text-destructive hover:text-destructive"
              title="ลบ"
              onClick={(e) => {
                e.stopPropagation();
                onDeleteRecord(records.find((r) => String(r.guidfixed) === item.guidfixed) || {});
              }}
            >
              <Trash2 className="h-3.5 w-3.5" />
            </Button>
          </div>
        </div>

        {/* Drop indicator line */}
        {isDropTarget && dropTarget?.position !== "inside" && (
          <div
            className="h-0.5 bg-primary rounded-full mx-4"
            style={{ marginLeft: `${depth * 20 + 16}px` }}
          />
        )}

        {/* Children */}
        {hasChildren && isExpanded && (
          <div>{children.map((child) => renderNode(child))}</div>
        )}
      </div>
    );
  };

  return (
    <Card className="h-full flex flex-col min-h-0">
      <CardContent className="flex-1 min-h-0 flex flex-col p-3 gap-2">
        {/* Header */}
        <div className="flex items-center justify-between gap-2">
          <h3 className="text-sm font-semibold text-foreground">
            กลุ่มย่อยสินค้า
            <span className="ml-2 text-xs font-normal text-muted-foreground">
              {items.length} รายการ
            </span>
          </h3>
          <div className="flex items-center gap-1">
            {moveInFlight && <Loader2 className="h-4 w-4 animate-spin text-primary" />}
            <Button
              variant="outline"
              size="sm"
              className="h-7 gap-1"
              onClick={() => onOpenCreate(undefined)}
            >
              <FolderPlus className="h-3.5 w-3.5" />
              เพิ่ม
            </Button>
          </div>
        </div>

        {/* Error */}
        {moveError && (
          <p className="text-xs text-destructive bg-destructive/10 rounded px-2 py-1">{moveError}</p>
        )}

        {/* Tree */}
        <div
          ref={treeRef}
          className="flex-1 min-h-0 overflow-y-auto rounded-md border border-border/50 p-1"
          onDragOver={(e) => {
            // Allow drop on empty area = move to root
            if (dragGuid && !dropTarget) {
              e.preventDefault();
            }
          }}
          onDrop={(e) => {
            // Drop on empty area = move to root end
            if (dragGuid && !dropTarget) {
              e.preventDefault();
              const dragged = items.find((i) => i.guidfixed === dragGuid);
              if (dragged) {
                const rootItems = items.filter((i) => !i.parentcode).sort((a, b) => a.sortorder - b.sortorder);
                void saveItem(dragged, { parentcode: "", sortorder: rootItems.length });
              }
              setDragGuid(null);
            }
          }}
        >
          {loading ? (
            <div className="flex items-center justify-center py-8">
              <Loader2 className="h-6 w-6 animate-spin text-primary" />
            </div>
          ) : tree.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-8 text-muted-foreground gap-2">
              <FolderPlus className="h-8 w-8 opacity-40" />
              <p className="text-sm">ยังไม่มีกลุ่มย่อย — กด "เพิ่ม" เพื่อสร้าง</p>
            </div>
          ) : (
            tree.map((node) => renderNode(node))
          )}
        </div>
      </CardContent>
    </Card>
  );
}
