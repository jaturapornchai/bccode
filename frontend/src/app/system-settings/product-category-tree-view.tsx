"use client";

import { authFetch } from "@/lib/client-auth-session";
import React, { useCallback, useEffect, useLayoutEffect, useMemo, useRef, useState } from "react";
import {
  ChevronDown,
  ChevronRight,
  ChevronUp,
  Edit3,
  GripVertical,
  Home,
  Loader2,
  Redo2,
  Trash2,
  Undo2,
} from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

// Types matching system-settings
type SettingRecord = Record<string, unknown>;

interface ProductCategoryTreeViewProps {
  auth: { token: string; backendUrl: string } | null;
  workspace: { shop: { holdingcode: string } } | null;
  language: string;
  records: SettingRecord[];
  groupNumber: number | null;
  setGroupNumber: (num: number | null) => void;
  selectedGuid: string;
  setSelectedGuid?: (guid: string) => void;
  searchQuery: string;
  onSelectRecord: (record: SettingRecord) => void | Promise<void>;
  onOpenCreate?: (parentGuid?: string) => void;
  onOpenEdit?: (record: SettingRecord) => void;
  onDeleteRecord?: (record: SettingRecord) => void;
  onRefresh?: () => void;
  saving: boolean;
  loading: boolean;
  readOnly?: boolean;
}

interface CategoryNode {
  guidfixed: string;
  parentguid: string;
  parentguidall: string;
  names: CategoryName[];
  xsorts?: CategoryXSort[];
  groupnumber: number;
  productCount?: number;
}

type CategoryName = { code: string; name: string };
type CategoryXSort = { code: string; xorder: number };
type XSortPayload = { guidfixed: string; code: string; xorder: number };

// Undo/redo: one snapshot = everything needed to replay a move back to a
// prior state via the real backend calls (not a visual-only revert).
type MoveSnapshot = { parentGuid: string; parentGuidAll: string; xsorts: XSortPayload[] };
type MoveRecord = { guid: string; before: MoveSnapshot; after: MoveSnapshot };

type GroupSummary = {
  count: number;
  itemCount: number;
  label: string;
};

const GROUP_PREVIEW_LIMIT = 3;
const TREE_LAYOUT_ANIMATION_MS = 220;
const CATEGORY_LEVEL_STYLES = [
  {
    caret: "text-primary",
    grip: "text-primary/45 group-hover/row:text-primary",
    name: "text-[15px] font-bold text-foreground",
    order: "text-primary",
  },
  {
    caret: "text-sky-600 dark:text-sky-300",
    grip: "text-sky-500/55 group-hover/row:text-sky-600 dark:group-hover/row:text-sky-300",
    name: "font-semibold text-sky-900 dark:text-sky-100",
    order: "text-sky-600 dark:text-sky-300",
  },
  {
    caret: "text-emerald-600 dark:text-emerald-300",
    grip: "text-emerald-500/55 group-hover/row:text-emerald-600 dark:group-hover/row:text-emerald-300",
    name: "font-semibold text-emerald-900 dark:text-emerald-100",
    order: "text-emerald-600 dark:text-emerald-300",
  },
  {
    caret: "text-amber-600 dark:text-amber-300",
    grip: "text-amber-500/60 group-hover/row:text-amber-600 dark:group-hover/row:text-amber-300",
    name: "font-medium text-amber-900 dark:text-amber-100",
    order: "text-amber-600 dark:text-amber-300",
  },
  {
    caret: "text-violet-600 dark:text-violet-300",
    grip: "text-violet-500/55 group-hover/row:text-violet-600 dark:group-hover/row:text-violet-300",
    name: "font-medium text-violet-900 dark:text-violet-100",
    order: "text-violet-600 dark:text-violet-300",
  },
] as const;

interface CategoryTreeNode {
  detail: CategoryNode;
  childCategories: CategoryTreeNode[];
}

type DropPosition = "before" | "after" | "inside";

interface DragState {
  guid: string;
  parentGuid: string;
  rect: {
    left: number;
    top: number;
    width: number;
    height: number;
  };
}

interface PointerDragState extends DragState {
  pointerId: number;
  startX: number;
  startY: number;
  started: boolean;
}

// One-shot geometry cache captured at drag-start (and on scroll) -- see
// captureDragGeometry. Mid-drag hit-testing reads only this, never the DOM.
interface DragGeometryRow {
  guid: string;
  isSelf: boolean;
  top: number;
  bottom: number;
  mid: number;
  reorderAllowed: boolean;
  insideRect: { top: number; bottom: number; left: number; right: number } | null;
}

interface DragGeometry {
  rows: DragGeometryRow[];
  root: { top: number; bottom: number } | null;
}

type ParentOverride = {
  parentGuid: string;
  parentGuidAll: string;
};

const categoryLevelStyle = (level: number) =>
  CATEGORY_LEVEL_STYLES[Math.min(Math.max(level, 0), CATEGORY_LEVEL_STYLES.length - 1)];

const recordGuid = (record: SettingRecord): string =>
  String(record.guidfixed || record.guidfixed || record.guid || "");

const recordParentGuid = (record: SettingRecord): string =>
  String(record.parentguid || record.parentguid || "");

const recordGroupNumber = (record: SettingRecord): number =>
  Number(record.groupnumber ?? record.groupnumber ?? 0);

const recordCodelistCount = (record: SettingRecord): number =>
  new Set(
    (Array.isArray(record.codelist) ? record.codelist : [])
      .filter((item): item is SettingRecord =>
        typeof item === "object" && item !== null && !Array.isArray(item)
      )
      .map((item) => String(item.code ?? "").trim().toUpperCase())
      .filter(Boolean),
  ).size;

const toCategoryNames = (value: unknown): CategoryName[] =>
  Array.isArray(value)
    ? value
        .filter((item): item is SettingRecord => typeof item === "object" && item !== null && !Array.isArray(item))
        .map((item) => ({
          code: String(item.code ?? ""),
          name: String(item.name ?? ""),
        }))
    : [];

const toCategoryXSorts = (value: unknown): CategoryXSort[] =>
  Array.isArray(value)
    ? value
        .filter((item): item is SettingRecord => typeof item === "object" && item !== null && !Array.isArray(item))
        .map((item) => ({
          code: String(item.code ?? ""),
          xorder: Number(item.xorder ?? 0),
        }))
    : [];

export function ProductCategoryTreeView({
  auth,
  workspace,
  language,
  records,
  groupNumber,
  setGroupNumber,
  selectedGuid,
  setSelectedGuid,
  searchQuery,
  onSelectRecord,
  onOpenCreate,
  onOpenEdit,
  onDeleteRecord,
  onRefresh,
  saving,
  loading,
  readOnly = false,
}: ProductCategoryTreeViewProps) {
  const [expandedNodes, setExpandedNodes] = useState<Record<string, boolean>>({});
  const [dragState, setDragState] = useState<DragState | null>(null);
  const [dropTarget, setDropTarget] = useState<{ guid: string; position: DropPosition } | null>(null);
  const [orderOverrides, setOrderOverrides] = useState<Record<string, number>>({});
  const [parentOverrides, setParentOverrides] = useState<Record<string, ParentOverride>>({});
  const [reorderError, setReorderError] = useState<string>("");
  const [rootDropActive, setRootDropActive] = useState(false);
  // Undo/redo stack for sibling-reorder and reparent moves (drag or button).
  // index === -1 means "nothing to undo". Session-local, reset on reload or
  // whenever `records`/`groupNumber` changes (same effect that already clears
  // the other optimistic override state below).
  const [moveHistory, setMoveHistory] = useState<{ stack: MoveRecord[]; index: number }>({ stack: [], index: -1 });
  // One shared in-flight flag for every move-persisting call (button/drag
  // move AND undo/redo apply) so they can't race each other.
  const [isMoveInFlight, setIsMoveInFlight] = useState(false);
  const [arrivalHighlight, setArrivalHighlight] = useState<{ guid: string; nonce: number } | null>(null);
  const treeListRef = useRef<HTMLDivElement | null>(null);
  const pointerDragRef = useRef<PointerDragState | null>(null);
  const dropTargetRef = useRef<{ guid: string; position: DropPosition } | null>(null);
  const rootDropActiveRef = useRef(false);
  const dragWindowListenersRef = useRef<{
    move: (event: PointerEvent) => void;
    up: (event: PointerEvent) => void;
    cancel: (event: PointerEvent) => void;
  } | null>(null);
  const ignoreNextClickRef = useRef(false);
  const pendingLayoutRectsRef = useRef<Map<string, DOMRect> | null>(null);
  const layoutAnimationFrameRef = useRef<number | null>(null);
  const arrivalTimerRef = useRef<number | null>(null);
  // rAF-coalesced pointer tracking: only the ghost transform + drop-target
  // hit test run per animation frame, never per raw pointermove.
  const latestPointerRef = useRef<{ clientX: number; clientY: number } | null>(null);
  const dragRafRef = useRef<number | null>(null);
  const dragGhostRef = useRef<HTMLDivElement | null>(null);
  const dragGeometryRef = useRef<DragGeometry | null>(null);

  const setDropTargetState = (next: { guid: string; position: DropPosition } | null) => {
    dropTargetRef.current = next;
    setDropTarget(next);
  };

  const setRootDropActiveState = (next: boolean) => {
    rootDropActiveRef.current = next;
    setRootDropActive(next);
  };

  const captureTreeLayout = useCallback(() => {
    const root = treeListRef.current;
    if (!root) return;

    const rects = new Map<string, DOMRect>();
    root.querySelectorAll<HTMLElement>("[data-category-row-guid]").forEach((element) => {
      const guid = element.dataset.categoryRowGuid;
      if (guid) rects.set(guid, element.getBoundingClientRect());
    });
    pendingLayoutRectsRef.current = rects;
  }, []);

  const findCategoryRowElement = useCallback((guid: string): HTMLElement | null => {
    const root = treeListRef.current;
    if (!root) return null;

    return Array.from(root.querySelectorAll<HTMLElement>("[data-category-row-guid]"))
      .find((element) => element.dataset.categoryRowGuid === guid) ?? null;
  }, []);

  const markCategoryArrived = useCallback((guid: string) => {
    setArrivalHighlight({ guid, nonce: Date.now() });
  }, []);

  const playPendingTreeLayoutAnimation = useCallback(() => {
    const previousRects = pendingLayoutRectsRef.current;
    const root = treeListRef.current;
    if (!previousRects || !root) return;

    pendingLayoutRectsRef.current = null;
    if (layoutAnimationFrameRef.current !== null) {
      window.cancelAnimationFrame(layoutAnimationFrameRef.current);
    }

    layoutAnimationFrameRef.current = window.requestAnimationFrame(() => {
      layoutAnimationFrameRef.current = null;
      if (window.matchMedia("(prefers-reduced-motion: reduce)").matches) return;

      root.querySelectorAll<HTMLElement>("[data-category-row-guid]").forEach((element) => {
        const guid = element.dataset.categoryRowGuid;
        if (!guid) return;

        const currentRect = element.getBoundingClientRect();
        const previousRect = previousRects.get(guid);
        if (!previousRect) {
          element.animate(
            [
              { opacity: 0, transform: "translateY(-6px)" },
              { opacity: 1, transform: "translateY(0)" },
            ],
            {
              duration: TREE_LAYOUT_ANIMATION_MS,
              easing: "cubic-bezier(0.2, 0.8, 0.2, 1)",
            }
          );
          return;
        }

        const deltaX = previousRect.left - currentRect.left;
        const deltaY = previousRect.top - currentRect.top;
        if (Math.abs(deltaX) < 1 && Math.abs(deltaY) < 1) return;

        element.animate(
          [
            { transform: `translate(${deltaX}px, ${deltaY}px)` },
            { transform: "translate(0, 0)" },
          ],
          {
            duration: TREE_LAYOUT_ANIMATION_MS,
            easing: "cubic-bezier(0.2, 0.8, 0.2, 1)",
          }
        );
      });
    });
  }, []);

  useEffect(() => {
    setOrderOverrides({});
    setParentOverrides({});
    setArrivalHighlight(null);
    setDragState(null);
    dropTargetRef.current = null;
    rootDropActiveRef.current = false;
    setDropTarget(null);
    setRootDropActive(false);
    pointerDragRef.current = null;
    dragGeometryRef.current = null;
    latestPointerRef.current = null;
    if (dragRafRef.current !== null) {
      window.cancelAnimationFrame(dragRafRef.current);
      dragRafRef.current = null;
    }
    ignoreNextClickRef.current = false;
    setMoveHistory({ stack: [], index: -1 });
    setIsMoveInFlight(false);
  }, [records, groupNumber]);

  useEffect(() => {
    return () => {
      if (layoutAnimationFrameRef.current !== null) {
        window.cancelAnimationFrame(layoutAnimationFrameRef.current);
      }
      if (arrivalTimerRef.current !== null) {
        window.clearTimeout(arrivalTimerRef.current);
      }
      if (dragRafRef.current !== null) {
        window.cancelAnimationFrame(dragRafRef.current);
      }
      if (dragWindowListenersRef.current) {
        window.removeEventListener("pointermove", dragWindowListenersRef.current.move);
        window.removeEventListener("pointerup", dragWindowListenersRef.current.up);
        window.removeEventListener("pointercancel", dragWindowListenersRef.current.cancel);
        dragWindowListenersRef.current = null;
      }
    };
  }, []);

  useEffect(() => {
    if (!arrivalHighlight || typeof window === "undefined") return;

    if (arrivalTimerRef.current !== null) {
      window.clearTimeout(arrivalTimerRef.current);
    }

    const animationFrame = window.requestAnimationFrame(() => {
      const element = findCategoryRowElement(arrivalHighlight.guid);
      if (!element) return;

      const reduceMotion = window.matchMedia("(prefers-reduced-motion: reduce)").matches;
      const container = treeListRef.current;
      if (container) {
        const containerRect = container.getBoundingClientRect();
        const rowRect = element.getBoundingClientRect();
        const rowOutsideView =
          rowRect.top < containerRect.top + 8 ||
          rowRect.bottom > containerRect.bottom - 8;
        if (rowOutsideView) {
          element.scrollIntoView({
            block: "center",
            behavior: reduceMotion ? "auto" : "smooth",
          });
        }
      }

      if (!reduceMotion) {
        element.animate(
          [
            {
              backgroundColor: "rgba(16, 185, 129, 0.06)",
              boxShadow: "inset 0 0 0 0 rgba(16, 185, 129, 0), 0 0 0 0 rgba(16, 185, 129, 0)",
              transform: "scale(1)",
            },
            {
              backgroundColor: "rgba(16, 185, 129, 0.18)",
              boxShadow: "inset 0 0 0 2px rgba(16, 185, 129, 0.55), 0 0 0 7px rgba(16, 185, 129, 0.14)",
              transform: "scale(1.006)",
            },
            {
              backgroundColor: "rgba(16, 185, 129, 0.04)",
              boxShadow: "inset 0 0 0 0 rgba(16, 185, 129, 0), 0 0 0 0 rgba(16, 185, 129, 0)",
              transform: "scale(1)",
            },
          ],
          {
            duration: 900,
            easing: "cubic-bezier(0.2, 0.8, 0.2, 1)",
          }
        );
      }
    });

    arrivalTimerRef.current = window.setTimeout(() => {
      setArrivalHighlight((current) =>
        current?.guid === arrivalHighlight.guid && current.nonce === arrivalHighlight.nonce
          ? null
          : current
      );
    }, 1500);

    return () => window.cancelAnimationFrame(animationFrame);
  }, [arrivalHighlight, findCategoryRowElement]);

  // Helper to extract display name in correct language
  const getDisplayName = useCallback((names: CategoryName[]): string => {
    if (!Array.isArray(names)) return "";
    const match = names.find((n) => n.code === language);
    if (match?.name) return match.name;
    const thMatch = names.find((n) => n.code === "th");
    if (thMatch?.name) return thMatch.name;
    const enMatch = names.find((n) => n.code === "en");
    if (enMatch?.name) return enMatch.name;
    return names[0]?.name || "";
  }, [language]);

  // Convert generic SettingRecord to CategoryNode
  const typedCategories = useMemo<CategoryNode[]>(() => {
    return records.map((r) => {
      const guid = recordGuid(r);
      const orderOverride = orderOverrides[guid];
      const parentOverride = parentOverrides[guid];
      return {
        guidfixed: guid,
        parentguid: parentOverride?.parentGuid ?? recordParentGuid(r),
        parentguidall: parentOverride?.parentGuidAll ?? String(r.parentguidall || ""),
        names: toCategoryNames(r.names),
        xsorts: orderOverride ? [{ code: "X", xorder: orderOverride }] : toCategoryXSorts(r.xsorts),
        groupnumber: recordGroupNumber(r),
        productCount: recordCodelistCount(r),
      };
    });
  }, [records, orderOverrides, parentOverrides]);

  // Build tree from flat categories
  const treeRoots = useMemo<CategoryTreeNode[]>(() => {
    // Sort items by xorder first
    const sorted = [...typedCategories].sort((a, b) => {
      const orderA = a.xsorts?.[0]?.xorder ?? 0;
      const orderB = b.xsorts?.[0]?.xorder ?? 0;
      return orderA - orderB;
    });

    const nodeMap = new Map<string, CategoryTreeNode>();
    const roots: CategoryTreeNode[] = [];

    // Create node wrappers
    for (const item of sorted) {
      nodeMap.set(item.guidfixed, {
        detail: item,
        childCategories: [],
      });
    }

    // Build parent-child relationships
    for (const node of nodeMap.values()) {
      const parentGuid = node.detail.parentguid;
      if (parentGuid && nodeMap.has(parentGuid)) {
        nodeMap.get(parentGuid)!.childCategories.push(node);
      } else {
        roots.push(node);
      }
    }

    const needle = searchQuery.trim().toLowerCase();
    if (!needle) return roots;

    const keepMatchesWithAncestors = (nodes: CategoryTreeNode[]): CategoryTreeNode[] =>
      nodes.flatMap((node) => {
        const childCategories = keepMatchesWithAncestors(node.childCategories);
        const matches = getDisplayName(node.detail.names).toLowerCase().includes(needle);
        return matches || childCategories.length > 0
          ? [{ ...node, childCategories }]
          : [];
      });

    return keepMatchesWithAncestors(roots);
  }, [typedCategories, searchQuery, getDisplayName]);

  useLayoutEffect(() => {
    playPendingTreeLayoutAnimation();
  }, [expandedNodes, treeRoots, playPendingTreeLayoutAnimation]);

  const categoryTreeLookups = useMemo(() => {
    const nodeByGuid = new Map<string, CategoryTreeNode>();
    const siblingsByGuid = new Map<string, CategoryTreeNode[]>();

    const visit = (nodes: CategoryTreeNode[]) => {
      for (const node of nodes) {
        nodeByGuid.set(node.detail.guidfixed, node);
        siblingsByGuid.set(node.detail.guidfixed, nodes);
        visit(node.childCategories);
      }
    };

    visit(treeRoots);
    return { nodeByGuid, siblingsByGuid };
  }, [treeRoots]);
  const categoryTreeLookupsRef = useRef(categoryTreeLookups);

  useEffect(() => {
    categoryTreeLookupsRef.current = categoryTreeLookups;
  }, [categoryTreeLookups]);

  const canDropOnTarget = (
    activeDrag: DragState,
    node: CategoryTreeNode,
    _siblings: CategoryTreeNode[],
    position: DropPosition
  ) => {
    if (activeDrag.guid === node.detail.guidfixed) return false;
    const targetParentChain = node.detail.parentguidall
      .split(",")
      .map((item) => item.trim())
      .filter(Boolean);
    const canMoveAsChild =
      position === "inside" &&
      !targetParentChain.includes(activeDrag.guid);
    const canReorderSibling =
      position !== "inside" && !targetParentChain.includes(activeDrag.guid);
    return canReorderSibling || canMoveAsChild;
  };

  // One-shot geometry snapshot for the whole drag -- captured once when the
  // drag starts and again on scroll (see the effect below), never per
  // pointermove. Everything mid-drag hit-tests against this cache only, so
  // dragging never triggers a DOM read. The dragged row's own rect is kept
  // (isSelf) purely to detect "pointer hasn't left its own footprint yet"
  // without falling through to a false nearest-row match.
  const captureDragGeometry = (activeDrag: DragState): DragGeometry => {
    const root = treeListRef.current;
    if (!root) return { rows: [], root: null };

    const rows: DragGeometryRow[] = [];
    root.querySelectorAll<HTMLElement>("[data-category-row-guid]").forEach((element) => {
      const guid = element.dataset.categoryRowGuid;
      if (!guid) return;
      const isSelf = guid === activeDrag.guid;
      const node = categoryTreeLookups.nodeByGuid.get(guid);
      const rect = element.getBoundingClientRect();
      const reorderAllowed = !isSelf && !!node && canDropOnTarget(activeDrag, node, [], "before");
      const insideEl = isSelf ? null : element.querySelector<HTMLElement>("[data-category-inside-drop-guid]");
      let insideRect: DragGeometryRow["insideRect"] = null;
      if (insideEl) {
        const r = insideEl.getBoundingClientRect();
        insideRect = { top: r.top - 3, bottom: r.bottom + 3, left: r.left - 3, right: r.right + 3 };
      }
      rows.push({ guid, isSelf, top: rect.top, bottom: rect.bottom, mid: rect.top + rect.height / 2, reorderAllowed, insideRect });
    });

    const rootEl = root.querySelector<HTMLElement>("[data-category-root-drop='true']");
    const rootRect = rootEl ? rootEl.getBoundingClientRect() : null;
    return { rows, root: rootRect ? { top: rootRect.top, bottom: rootRect.bottom } : null };
  };

  // Snapshot geometry exactly once when a drag starts (dragState guid
  // none -> set), and re-snapshot on scroll so a mid-drag scroll can't stale
  // the cache. Runs as a layout effect so it reads the DOM after the row
  // re-render that shows the "Drop child" chips, but before paint.
  //
  // This hook (and canDropOnTarget/captureDragGeometry above) MUST stay
  // before the `groupNumber === null` early return below -- moving it into
  // the tree-view portion of the render (like the rest of the drag/move
  // logic) would make its hook call conditional on groupNumber, which
  // violates the Rules of Hooks (React error #310) the moment a user picks
  // a group for the first time in a session.
  useLayoutEffect(() => {
    if (!dragState) {
      dragGeometryRef.current = null;
      return;
    }
    const activeDrag = dragState;
    dragGeometryRef.current = captureDragGeometry(activeDrag);
    const root = treeListRef.current;
    if (!root) return;
    const onScroll = () => {
      dragGeometryRef.current = captureDragGeometry(activeDrag);
    };
    root.addEventListener("scroll", onScroll, { passive: true });
    return () => root.removeEventListener("scroll", onScroll);
  }, [dragState?.guid]); // eslint-disable-line react-hooks/exhaustive-deps

  const groupSummaries = useMemo<Record<number, GroupSummary>>(() => {
    const summaries: Record<number, GroupSummary> = {};
    for (const r of records) {
      const gn = recordGroupNumber(r);
      if (gn >= 1 && gn <= 20) {
        if (!summaries[gn]) {
          summaries[gn] = { count: 0, itemCount: 0, label: "" };
        }
        summaries[gn].count += 1;
        summaries[gn].itemCount += recordCodelistCount(r);
      }
    }
    for (const gn of Object.keys(summaries).map(Number)) {
      const rootNames = records
        .filter((record) => recordGroupNumber(record) === gn && !recordParentGuid(record))
        .sort((a, b) => {
          const aSort = toCategoryXSorts(a.xsorts)[0]?.xorder ?? 0;
          const bSort = toCategoryXSorts(b.xsorts)[0]?.xorder ?? 0;
          return aSort - bSort;
        })
        .map((record) => getDisplayName(toCategoryNames(record.names)))
        .filter(Boolean);
      const preview = rootNames.slice(0, GROUP_PREVIEW_LIMIT).join(" / ");
      const remaining = Math.max(0, rootNames.length - GROUP_PREVIEW_LIMIT);
      summaries[gn].label = remaining > 0
        ? language === "th"
          ? `${preview} และอีก ${remaining} หมวดหลัก`
          : `${preview} and ${remaining} more root categories`
        : preview;
    }
    return summaries;
  }, [records, getDisplayName, language]);

  const configuredGroupCount = Object.values(groupSummaries).filter(
    (summary) => summary.count > 0,
  ).length;

  // Group selector: compact 20 group buttons
  if (groupNumber === null) {
    return (
      <div className="grid w-full gap-3 p-2 sm:p-3" data-testid="product-category-group-selector">
        <header className="flex flex-wrap items-end justify-between gap-2 border-b border-border pb-2">
          <div className="grid gap-0.5">
            <h1 className="text-base font-bold tracking-tight text-foreground">
              {language === "th" ? "เลือกชุดหมวดสินค้า" : "Select Category Set"}
            </h1>
            <p className="text-xs text-muted-foreground">
              {language === "th"
                ? "แต่ละชุดใช้จัดหมวดให้เหมาะกับหน้าจอขายหรือช่องทางใช้งาน"
                : "Each set organizes categories for a sales screen or usage channel."}
            </p>
          </div>
          <span className="rounded-full border border-primary/20 bg-primary/5 px-2.5 py-1 text-xs font-semibold text-primary">
            {language === "th"
              ? `${configuredGroupCount} ชุดมีข้อมูล · ${20 - configuredGroupCount} ชุดว่าง`
              : `${configuredGroupCount} configured · ${20 - configuredGroupCount} empty`}
          </span>
        </header>

        {loading ? (
          <div className="flex min-h-32 items-center justify-center gap-2 rounded-xl border border-border bg-card p-6 text-sm text-muted-foreground">
            <Loader2 className="size-5 animate-spin text-primary" />
            {language === "th" ? "กำลังโหลดข้อมูล..." : "Loading..."}
          </div>
        ) : (
          <main className="grid w-full grid-cols-1 gap-2.5 sm:grid-cols-2 lg:grid-cols-4 xl:grid-cols-5" data-testid="product-category-group-grid">
            {Array.from({ length: 20 }, (_, i) => i + 1).map((num) => {
              const summary = groupSummaries[num] ?? { count: 0, itemCount: 0, label: "" };
              const hasData = summary.count > 0;
              const groupLabel = summary.label || (language === "th" ? "ยังไม่มีหมวดสินค้า" : "No categories yet");
              const countLabel = language === "th"
                ? `${summary.count} หมวด / ${summary.itemCount} สินค้า`
                : `${summary.count} categories / ${summary.itemCount} items`;
              return (
                <button
                  type="button"
                  key={num}
                  className={cn(
                    "group flex min-h-14 w-full items-center gap-2 rounded-xl border border-border bg-card px-2.5 py-2 text-left shadow-sm transition-[background-color,border-color,box-shadow,transform] duration-200 hover:-translate-y-0.5 hover:border-primary/50 hover:bg-primary/5 hover:shadow-md focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/40",
                    hasData && "border-primary/25 bg-primary/[0.04]",
                  )}
                  onClick={() => setGroupNumber(num)}
                >
                  <span className={cn(
                    "grid size-9 shrink-0 place-items-center rounded-lg text-lg font-black transition-colors",
                    hasData
                      ? "bg-primary/10 text-primary"
                      : "bg-muted text-foreground group-hover:bg-primary/10 group-hover:text-primary",
                  )}>
                    {num}
                  </span>
                  <span className="grid min-w-0 flex-1 gap-0.5">
                    <span
                      className="line-clamp-2 break-words text-xs font-semibold leading-snug text-foreground"
                      title={groupLabel}
                    >
                      {groupLabel}
                    </span>
                    <span className={cn(
                      "w-fit rounded-full px-2 py-0.5 text-xs font-bold leading-none",
                      hasData
                        ? "bg-primary/10 text-primary"
                        : "bg-muted text-muted-foreground",
                    )}>
                      {countLabel}
                    </span>
                  </span>
                </button>
              );
            })}
          </main>
        )}
      </div>
    );
  }

  // Toggle node expansion
  const toggleExpand = (guid: string, e: React.MouseEvent) => {
    e.stopPropagation();
    captureTreeLayout();
    setExpandedNodes((prev) => ({
      ...prev,
      [guid]: !prev[guid],
    }));
  };

  const getResponseErrorMessage = async (response: Response, fallback: string) => {
    try {
      const contentType = response.headers.get("content-type") ?? "";
      if (contentType.includes("application/json")) {
        const payload = await response.json();
        if (payload && typeof payload === "object") {
          const record = payload as Record<string, unknown>;
          const nestedError = record.error && typeof record.error === "object"
            ? (record.error as Record<string, unknown>)
            : null;
          const message = record.message ?? nestedError?.message ?? nestedError?.error ?? record.error;
          if (typeof message === "string" && message.trim()) return message;
        }
      }
      const message = await response.text();
      if (!message.trim()) return fallback;
      try {
        const payload = JSON.parse(message) as unknown;
        if (payload && typeof payload === "object") {
          const record = payload as Record<string, unknown>;
          const parsedMessage = record.message ?? record.error;
          if (typeof parsedMessage === "string" && parsedMessage.trim()) return parsedMessage;
        }
      } catch {
        // keep original text below
      }
      return message;
    } catch {
      return fallback;
    }
  };

  const validateXSortPayload = (items: XSortPayload[]): XSortPayload[] => {
    const seen = new Set<string>();
    for (const item of items) {
      if (!item.guidfixed.trim()) {
        throw new Error(language === "th" ? "ข้อมูลลำดับไม่ครบ: guidfixed ว่าง" : "Invalid order payload: empty guidfixed");
      }
      if (seen.has(item.guidfixed)) {
        throw new Error(language === "th" ? "ข้อมูลลำดับซ้ำ: guidfixed ซ้ำ" : "Invalid order payload: duplicated guidfixed");
      }
      if (!Number.isFinite(item.xorder) || item.xorder < 1) {
        throw new Error(language === "th" ? "ข้อมูลลำดับไม่ถูกต้อง: xorder ต้องมากกว่า 0" : "Invalid order payload: xorder must be greater than 0");
      }
      seen.add(item.guidfixed);
    }
    return items;
  };

  const saveXSorts = async (updateList: XSortPayload[]) => {
    if (!auth || !workspace || updateList.length === 0) return;
    const payload = validateXSortPayload(updateList);
    const response = await authFetch(
      `/api/system-settings/productcategorygroupselectscreen/xsort?holdingcode=${encodeURIComponent(workspace.shop.holdingcode)}`,
      {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
          "x-bc-backend-url": auth.backendUrl,
          Authorization: `Bearer ${auth.token}`,
        },
        body: JSON.stringify(payload),
      }
    );
    if (!response.ok) {
      const message = await getResponseErrorMessage(response, "Failed to save category order");
      throw new Error(message);
    }
  };

  const saveCategoryRecord = async (guid: string, payload: SettingRecord) => {
    if (!auth || !workspace) return;
    const response = await authFetch(
      `/api/system-settings/productcategorygroupselectscreen/${encodeURIComponent(guid)}?holdingcode=${encodeURIComponent(workspace.shop.holdingcode)}`,
      {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
          "x-bc-backend-url": auth.backendUrl,
          Authorization: `Bearer ${auth.token}`,
        },
        body: JSON.stringify(payload),
      }
    );
    if (!response.ok) {
      const message = await getResponseErrorMessage(response, "Failed to update category");
      throw new Error(message);
    }
  };

  const buildParentGuidAll = (parentGuid: string): string => {
    if (!parentGuid) return "";
    const parent = typedCategories.find((item) => item.guidfixed === parentGuid);
    if (!parent) return parentGuid;
    return parent.parentguidall ? `${parent.parentguidall},${parentGuid}` : parentGuid;
  };

  const sortedChildrenOf = (parentGuid: string): CategoryNode[] =>
    typedCategories
      .filter((item) => item.parentguid === parentGuid)
      .sort((a, b) => (a.xsorts?.[0]?.xorder ?? 0) - (b.xsorts?.[0]?.xorder ?? 0));

  const normalizeSiblingOrders = (items: Array<CategoryTreeNode | CategoryNode>): XSortPayload[] =>
    items.map((item, index) => {
      const detail = "detail" in item ? item.detail : item;
      return {
        guidfixed: detail.guidfixed,
        code: "X",
        xorder: index + 1,
      };
    });

  const uniqueXSortPayload = (items: XSortPayload[]): XSortPayload[] =>
    Array.from(new Map(items.map((item) => [item.guidfixed, item])).values());

  // New move after an undo clears the redo tail (standard undo/redo semantics).
  const pushMoveRecord = (record: MoveRecord) => {
    setMoveHistory((prev) => ({
      stack: [...prev.stack.slice(0, prev.index + 1), record],
      index: prev.index + 1,
    }));
  };

  const categoryErrorText = (prefixTh: string, prefixEn: string, err: unknown): string =>
    language === "th"
      ? `${prefixTh}: ${err instanceof Error ? err.message : "unknown error"}`
      : `${prefixEn}: ${err instanceof Error ? err.message : "unknown error"}`;

  const recordWithOptimisticOverrides = (record: SettingRecord): SettingRecord => {
    const guid = recordGuid(record);
    const parentOverride = parentOverrides[guid];
    const orderOverride = orderOverrides[guid];
    if (!parentOverride && !orderOverride) return record;

    return {
      ...record,
      ...(parentOverride
        ? {
            parentguid: parentOverride.parentGuid,
            parentguidall: parentOverride.parentGuidAll,
          }
        : {}),
      ...(orderOverride ? { xsorts: [{ code: "X", xorder: orderOverride }] } : {}),
    };
  };

  const reorderCategory = async (
    draggedGuid: string,
    targetGuid: string,
    position: DropPosition,
    siblings: CategoryTreeNode[]
  ) => {
    if (!auth || !workspace) return;
    if (position === "inside") return;
    if (draggedGuid === targetGuid) return;
    if (isMoveInFlight) return;

    const draggedRecord = records.find((record) => recordGuid(record) === draggedGuid);
    const draggedCategory = typedCategories.find((item) => item.guidfixed === draggedGuid);
    const targetCategory = typedCategories.find((item) => item.guidfixed === targetGuid);
    if (!draggedRecord || !draggedCategory || !targetCategory) return;

    const targetParentGuid = targetCategory.parentguid || "";
    const newParentGuidAll = buildParentGuidAll(targetParentGuid);
    const oldParentGuid = draggedCategory.parentguid || "";
    const movingAcrossParents = oldParentGuid !== targetParentGuid;
    const targetParentChain = targetCategory.parentguidall
      .split(",")
      .map((item) => item.trim())
      .filter(Boolean);
    if (targetParentChain.includes(draggedGuid)) {
      setReorderError(
        language === "th"
          ? "ย้ายไม่ได้: ไม่สามารถย้ายหมวดไปไว้ใต้หมวดย่อยของตัวเอง"
          : "Move failed: category cannot be moved inside its own child branch."
      );
      return;
    }

    const currentTargetSiblings = siblings.length
      ? siblings.map((item) => item.detail)
      : sortedChildrenOf(targetParentGuid);
    const withoutDragged = currentTargetSiblings.filter((item) => item.guidfixed !== draggedGuid);
    const targetIndexAfterRemoval = withoutDragged.findIndex((item) => item.guidfixed === targetGuid);
    if (targetIndexAfterRemoval === -1) return;
    const insertIndex = targetIndexAfterRemoval + (position === "after" ? 1 : 0);
    const reorderedSiblings = [...withoutDragged];
    const movedCategory: CategoryNode = {
      ...draggedCategory,
      parentguid: targetParentGuid,
      parentguidall: newParentGuidAll,
    };
    reorderedSiblings.splice(insertIndex, 0, movedCategory);

    const sameOrder = currentTargetSiblings.every((item, index) => item.guidfixed === reorderedSiblings[index]?.guidfixed);
    if (!movingAcrossParents && sameOrder) return;

    const targetSiblingPayload = normalizeSiblingOrders(reorderedSiblings);
    const oldSiblingPayload = movingAcrossParents
      ? normalizeSiblingOrders(sortedChildrenOf(oldParentGuid).filter((item) => item.guidfixed !== draggedGuid))
      : [];
    const updateList = uniqueXSortPayload([...oldSiblingPayload, ...targetSiblingPayload]);
    const draggedOrder = targetSiblingPayload.find((item) => item.guidfixed === draggedGuid)?.xorder ?? 1;
    const payload: SettingRecord = {
      ...draggedRecord,
      parentguid: targetParentGuid,
      parentguidall: newParentGuidAll,
      groupnumber: groupNumber ?? recordGroupNumber(draggedRecord),
      xsorts: [{ code: "X", xorder: draggedOrder }],
    };

    const beforeXSorts = uniqueXSortPayload([
      ...(movingAcrossParents ? normalizeSiblingOrders(sortedChildrenOf(oldParentGuid)) : []),
      ...normalizeSiblingOrders(currentTargetSiblings),
    ]);
    const moveRecord: MoveRecord = {
      guid: draggedGuid,
      before: { parentGuid: oldParentGuid, parentGuidAll: draggedCategory.parentguidall, xsorts: beforeXSorts },
      after: { parentGuid: targetParentGuid, parentGuidAll: newParentGuidAll, xsorts: updateList },
    };

    captureTreeLayout();
    setReorderError("");
    if (targetParentGuid) {
      setExpandedNodes((prev) => ({ ...prev, [targetParentGuid]: true }));
    }
    setParentOverrides((prev) => ({
      ...prev,
      [draggedGuid]: { parentGuid: targetParentGuid, parentGuidAll: newParentGuidAll },
    }));
    setOrderOverrides((prev) => ({
      ...prev,
      ...Object.fromEntries(updateList.map((item) => [item.guidfixed, item.xorder])),
    }));

    setIsMoveInFlight(true);
    try {
      if (movingAcrossParents) {
        await saveCategoryRecord(draggedGuid, payload);
      }
      await saveXSorts(updateList);
      markCategoryArrived(draggedGuid);
      pushMoveRecord(moveRecord);
    } catch (err) {
      setReorderError(categoryErrorText("ย้ายหรือบันทึกลำดับไม่สำเร็จ", "Move or reorder failed", err));
      onRefresh?.();
    } finally {
      setIsMoveInFlight(false);
    }
  };

  const moveCategoryAsChild = async (draggedGuid: string, targetGuid: string) => {
    if (!auth || !workspace) return;
    if (draggedGuid === targetGuid) return;
    if (isMoveInFlight) return;

    const draggedRecord = records.find((record) => recordGuid(record) === draggedGuid);
    const draggedCategory = typedCategories.find((item) => item.guidfixed === draggedGuid);
    const targetCategory = typedCategories.find((item) => item.guidfixed === targetGuid);
    if (!draggedRecord || !draggedCategory || !targetCategory) return;

    const targetParentChain = targetCategory.parentguidall
      .split(",")
      .map((item) => item.trim())
      .filter(Boolean);
    if (targetParentChain.includes(draggedGuid)) {
      setReorderError(
        language === "th"
          ? "ย้ายไม่ได้: ไม่สามารถย้ายหมวดไปไว้ใต้หมวดย่อยของตัวเอง"
          : "Move failed: category cannot be moved under its own child."
      );
      return;
    }

    const newParentGuidAll = buildParentGuidAll(targetGuid);
    const nextChildren = sortedChildrenOf(targetGuid).filter((item) => item.guidfixed !== draggedGuid);
    const nextOrder = nextChildren.length + 1;
    const oldParentGuid = draggedCategory.parentguid || "";
    const oldSiblings = sortedChildrenOf(oldParentGuid).filter((item) => item.guidfixed !== draggedGuid);
    const payload: SettingRecord = {
      ...draggedRecord,
      parentguid: targetGuid,
      parentguidall: newParentGuidAll,
      groupnumber: groupNumber ?? recordGroupNumber(draggedRecord),
      xsorts: [{ code: "X", xorder: nextOrder }],
    };

    const beforeXSorts = uniqueXSortPayload([
      ...normalizeSiblingOrders(sortedChildrenOf(oldParentGuid)),
      ...normalizeSiblingOrders(sortedChildrenOf(targetGuid)),
    ]);
    const afterXSorts = uniqueXSortPayload([
      ...normalizeSiblingOrders(oldSiblings),
      ...normalizeSiblingOrders([...nextChildren, { ...draggedCategory, parentguid: targetGuid, parentguidall: newParentGuidAll }]),
    ]);
    const moveRecord: MoveRecord = {
      guid: draggedGuid,
      before: { parentGuid: oldParentGuid, parentGuidAll: draggedCategory.parentguidall, xsorts: beforeXSorts },
      after: { parentGuid: targetGuid, parentGuidAll: newParentGuidAll, xsorts: afterXSorts },
    };

    captureTreeLayout();
    setReorderError("");
    setExpandedNodes((prev) => ({ ...prev, [targetGuid]: true }));
    setParentOverrides((prev) => ({
      ...prev,
      [draggedGuid]: { parentGuid: targetGuid, parentGuidAll: newParentGuidAll },
    }));
    setOrderOverrides((prev) => ({
      ...prev,
      [draggedGuid]: nextOrder,
      ...Object.fromEntries(normalizeSiblingOrders(oldSiblings).map((item) => [item.guidfixed, item.xorder])),
    }));

    setIsMoveInFlight(true);
    try {
      await saveCategoryRecord(draggedGuid, payload);
      await saveXSorts(afterXSorts);
      markCategoryArrived(draggedGuid);
      pushMoveRecord(moveRecord);
    } catch (err) {
      setReorderError(categoryErrorText("ย้ายหมวดสินค้าไม่สำเร็จ", "Move category failed", err));
      onRefresh?.();
    } finally {
      setIsMoveInFlight(false);
    }
  };

  const moveCategoryToRoot = async (draggedGuid: string) => {
    if (!auth || !workspace) return;
    if (isMoveInFlight) return;

    const draggedRecord = records.find((record) => recordGuid(record) === draggedGuid);
    const draggedCategory = typedCategories.find((item) => item.guidfixed === draggedGuid);
    if (!draggedRecord || !draggedCategory) return;

    const oldParentGuid = draggedCategory.parentguid || "";
    const oldSiblings = sortedChildrenOf(oldParentGuid).filter((item) => item.guidfixed !== draggedGuid);
    const rootSiblings = sortedChildrenOf("").filter((item) => item.guidfixed !== draggedGuid);
    const nextRootOrder = rootSiblings.length + 1;
    const nextRootItems: CategoryNode[] = [
      ...rootSiblings,
      { ...draggedCategory, parentguid: "", parentguidall: "", xsorts: [{ code: "X", xorder: nextRootOrder }] },
    ];
    const rootOrderPayload = normalizeSiblingOrders(nextRootItems);
    const oldSiblingPayload = oldParentGuid ? normalizeSiblingOrders(oldSiblings) : [];
    const payload: SettingRecord = {
      ...draggedRecord,
      parentguid: "",
      parentguidall: "",
      groupnumber: groupNumber ?? recordGroupNumber(draggedRecord),
      xsorts: [{ code: "X", xorder: nextRootOrder }],
    };

    const afterXSorts = uniqueXSortPayload([...oldSiblingPayload, ...rootOrderPayload]);
    const beforeXSorts = uniqueXSortPayload([
      ...normalizeSiblingOrders(sortedChildrenOf(oldParentGuid)),
      ...normalizeSiblingOrders(sortedChildrenOf("")),
    ]);
    const moveRecord: MoveRecord = {
      guid: draggedGuid,
      before: { parentGuid: oldParentGuid, parentGuidAll: draggedCategory.parentguidall, xsorts: beforeXSorts },
      after: { parentGuid: "", parentGuidAll: "", xsorts: afterXSorts },
    };

    captureTreeLayout();
    setReorderError("");
    setParentOverrides((prev) => ({
      ...prev,
      [draggedGuid]: { parentGuid: "", parentGuidAll: "" },
    }));
    setOrderOverrides((prev) => ({
      ...prev,
      ...Object.fromEntries(afterXSorts.map((item) => [item.guidfixed, item.xorder])),
    }));

    setIsMoveInFlight(true);
    try {
      await saveCategoryRecord(draggedGuid, payload);
      await saveXSorts(afterXSorts);
      markCategoryArrived(draggedGuid);
      pushMoveRecord(moveRecord);
    } catch (err) {
      setReorderError(categoryErrorText("ย้ายหมวดสินค้าเป็นหมวดหลักไม่สำเร็จ", "Move category to root failed", err));
      onRefresh?.();
    } finally {
      setIsMoveInFlight(false);
    }
  };

  // Shared undo/redo apply: re-runs the real persistence calls (same shape
  // as the move functions above) against a captured snapshot, so undo/redo
  // genuinely restores server state instead of only the local UI.
  const applyMoveSnapshot = async (guid: string, snapshot: MoveSnapshot): Promise<boolean> => {
    if (!auth || !workspace || isMoveInFlight) return false;
    const draggedRecord = records.find((record) => recordGuid(record) === guid);
    const draggedCategory = typedCategories.find((item) => item.guidfixed === guid);
    if (!draggedRecord || !draggedCategory) return false;

    const parentChanged = draggedCategory.parentguid !== snapshot.parentGuid;
    const ownOrder = snapshot.xsorts.find((item) => item.guidfixed === guid)?.xorder ?? 1;
    const payload: SettingRecord = {
      ...draggedRecord,
      parentguid: snapshot.parentGuid,
      parentguidall: snapshot.parentGuidAll,
      groupnumber: groupNumber ?? recordGroupNumber(draggedRecord),
      xsorts: [{ code: "X", xorder: ownOrder }],
    };

    captureTreeLayout();
    setReorderError("");
    if (snapshot.parentGuid) {
      setExpandedNodes((prev) => ({ ...prev, [snapshot.parentGuid]: true }));
    }
    setParentOverrides((prev) => ({
      ...prev,
      [guid]: { parentGuid: snapshot.parentGuid, parentGuidAll: snapshot.parentGuidAll },
    }));
    setOrderOverrides((prev) => ({
      ...prev,
      ...Object.fromEntries(snapshot.xsorts.map((item) => [item.guidfixed, item.xorder])),
    }));

    setIsMoveInFlight(true);
    try {
      if (parentChanged) {
        await saveCategoryRecord(guid, payload);
      }
      await saveXSorts(snapshot.xsorts);
      markCategoryArrived(guid);
      return true;
    } catch (err) {
      setReorderError(categoryErrorText("เลิกทำ/ทำซ้ำไม่สำเร็จ", "Undo/redo failed", err));
      onRefresh?.();
      return false;
    } finally {
      setIsMoveInFlight(false);
    }
  };

  const canUseHistoryControls = !readOnly && !saving && !loading && !isMoveInFlight;

  const undoMove = async () => {
    if (!canUseHistoryControls || moveHistory.index < 0) return;
    const record = moveHistory.stack[moveHistory.index];
    const ok = await applyMoveSnapshot(record.guid, record.before);
    if (ok) setMoveHistory((prev) => ({ ...prev, index: prev.index - 1 }));
  };

  const redoMove = async () => {
    if (!canUseHistoryControls || moveHistory.index >= moveHistory.stack.length - 1) return;
    const record = moveHistory.stack[moveHistory.index + 1];
    const ok = await applyMoveSnapshot(record.guid, record.after);
    if (ok) setMoveHistory((prev) => ({ ...prev, index: prev.index + 1 }));
  };

  const canDragRows = !readOnly && !saving && !loading && !searchQuery.trim();

  const isInteractiveDragTarget = (target: EventTarget | null): boolean =>
    target instanceof Element &&
    Boolean(target.closest("button,a,input,textarea,select,[role='button']"));

  // Resolves the drop target from the cached geometry only (see above) --
  // pure number comparisons, zero DOM reads. The "Drop child" chip's cached
  // rect (+-3px forgiveness) is the dedicated reparent target; everywhere
  // else on a row is a plain top/bottom-half reorder split (half-open
  // interval so a boundary Y never double-matches two rows).
  const resolveDragPointerTarget = (clientX: number, clientY: number) => {
    const geometry = dragGeometryRef.current;
    if (!geometry) return null;

    if (geometry.root && clientY >= geometry.root.top && clientY <= geometry.root.bottom) {
      return { type: "root" as const };
    }

    for (const row of geometry.rows) {
      if (
        row.insideRect &&
        clientX >= row.insideRect.left && clientX <= row.insideRect.right &&
        clientY >= row.insideRect.top && clientY <= row.insideRect.bottom
      ) {
        return { type: "row" as const, guid: row.guid, position: "inside" as DropPosition };
      }
    }

    let containing: DragGeometryRow | null = null;
    let nearest: { row: DragGeometryRow; distance: number } | null = null;
    for (const row of geometry.rows) {
      if (clientY >= row.top && clientY < row.bottom) {
        containing = row;
        break;
      }
      const distance = clientY < row.top ? row.top - clientY : clientY - row.bottom;
      if (!nearest || distance < nearest.distance) nearest = { row, distance };
    }

    // Pointer still squarely over the dragged row's own footprint (or a row
    // that can't legally accept this drop) -- no candidate, and no falling
    // through to nearest, or a tiny in-place drag would silently reparent to
    // whichever row happens to be closest anywhere in the tree.
    const hit = containing ?? nearest?.row ?? null;
    if (!hit || hit.isSelf || !hit.reorderAllowed) return null;
    const position: DropPosition = clientY < hit.mid ? "before" : "after";
    return { type: "row" as const, guid: hit.guid, position };
  };

  const cleanupWindowDragListeners = () => {
    const listeners = dragWindowListenersRef.current;
    if (!listeners || typeof window === "undefined") return;

    window.removeEventListener("pointermove", listeners.move);
    window.removeEventListener("pointerup", listeners.up);
    window.removeEventListener("pointercancel", listeners.cancel);
    dragWindowListenersRef.current = null;
  };

  const cancelPendingDragFrame = () => {
    if (dragRafRef.current !== null) {
      window.cancelAnimationFrame(dragRafRef.current);
      dragRafRef.current = null;
    }
  };

  // rAF-coalesced per-frame work: at most one drop-target hit test and one
  // ghost transform write per animation frame, no matter how many raw
  // pointermove events arrived. The ghost follows the cursor by writing
  // directly to its DOM node (see dragGhostRef) -- never through React
  // state, so following the cursor never triggers a re-render.
  const processPendingDragMove = () => {
    dragRafRef.current = null;
    const activeDrag = pointerDragRef.current;
    const pointer = latestPointerRef.current;
    if (!activeDrag || !pointer) return;

    if (dragGhostRef.current) {
      const offsetY = pointer.clientY - activeDrag.startY;
      dragGhostRef.current.style.transform = `translateY(${offsetY}px) scale(1.006)`;
    }

    const candidate = resolveDragPointerTarget(pointer.clientX, pointer.clientY);
    if (candidate?.type === "root") {
      setRootDropActiveState(true);
      setDropTargetState(null);
      return;
    }
    setRootDropActiveState(false);
    if (candidate?.type === "row") {
      setDropTargetState({ guid: candidate.guid, position: candidate.position });
      return;
    }
    setDropTargetState(null);
  };

  const queueDragFrame = () => {
    if (dragRafRef.current !== null) return;
    dragRafRef.current = window.requestAnimationFrame(processPendingDragMove);
  };

  const handleRawPointerMove = (
    pointerId: number,
    clientX: number,
    clientY: number,
    preventDefault?: () => void
  ) => {
    const activeDrag = pointerDragRef.current;
    if (!activeDrag || activeDrag.pointerId !== pointerId) return;

    latestPointerRef.current = { clientX, clientY };

    if (!activeDrag.started) {
      const distance = Math.hypot(clientX - activeDrag.startX, clientY - activeDrag.startY);
      if (distance <= 5) return;
      const startedDrag: PointerDragState = { ...activeDrag, started: true };
      pointerDragRef.current = startedDrag;
      preventDefault?.();
      // Synchronous baseline snapshot *before* the dragState commit: rows
      // always exist in the DOM, so reorder hit-testing is correct from the
      // very first frame. The "Drop child" chips and root-drop zone only
      // render once dragState goes truthy, so this pass can't see them yet
      // -- the layout effect below re-snapshots right after that commit to
      // add them. Without this synchronous pass, a drop that lands before
      // React has re-rendered (state updates from a raw window listener
      // aren't flushed synchronously) would hit-test against a still-null
      // geometry cache and silently no-op the whole drag.
      dragGeometryRef.current = captureDragGeometry(startedDrag);
      setDragState({ guid: startedDrag.guid, parentGuid: startedDrag.parentGuid, rect: startedDrag.rect });
      queueDragFrame();
      return;
    }

    preventDefault?.();
    queueDragFrame();
  };

  const finishPointerDrag = (
    pointerId: number,
    clientX: number,
    clientY: number,
    preventDefault?: () => void,
    stopPropagation?: () => void
  ) => {
    const activeDrag = pointerDragRef.current;
    if (!activeDrag || activeDrag.pointerId !== pointerId) return;

    pointerDragRef.current = null;
    cleanupWindowDragListeners();
    cancelPendingDragFrame();
    if (!activeDrag.started) return;

    preventDefault?.();
    stopPropagation?.();
    ignoreNextClickRef.current = true;
    window.setTimeout(() => {
      ignoreNextClickRef.current = false;
    }, 250);

    const finalCandidate = resolveDragPointerTarget(clientX, clientY);
    const lookups = categoryTreeLookupsRef.current;
    const rootActive = rootDropActiveRef.current || finalCandidate?.type === "root";
    const target =
      finalCandidate?.type === "row"
        ? { guid: finalCandidate.guid, position: finalCandidate.position }
        : dropTargetRef.current;
    const finalSiblings = target ? lookups.siblingsByGuid.get(target.guid) || [] : [];
    setDragState(null);
    setDropTargetState(null);
    setRootDropActiveState(false);
    dragGeometryRef.current = null;

    if (rootActive) {
      void moveCategoryToRoot(activeDrag.guid);
      return;
    }

    if (!target) return;
    const targetNode = lookups.nodeByGuid.get(target.guid);
    if (!targetNode) return;

    if (target.position === "inside") {
      void moveCategoryAsChild(activeDrag.guid, target.guid);
    } else {
      void reorderCategory(activeDrag.guid, target.guid, target.position, finalSiblings);
    }
  };

  const cancelPointerDrag = () => {
    cleanupWindowDragListeners();
    cancelPendingDragFrame();
    pointerDragRef.current = null;
    setDragState(null);
    setDropTargetState(null);
    setRootDropActiveState(false);
    dragGeometryRef.current = null;
  };

  const attachWindowDragListeners = (pointerId: number) => {
    if (typeof window === "undefined") return;
    cleanupWindowDragListeners();

    const move = (event: PointerEvent) => {
      if (event.pointerId !== pointerId) return;
      handleRawPointerMove(event.pointerId, event.clientX, event.clientY, () => event.preventDefault());
    };
    const up = (event: PointerEvent) => {
      if (event.pointerId !== pointerId) return;
      finishPointerDrag(event.pointerId, event.clientX, event.clientY, () => event.preventDefault());
    };
    const cancel = (event: PointerEvent) => {
      if (event.pointerId !== pointerId) return;
      cancelPointerDrag();
    };

    dragWindowListenersRef.current = { move, up, cancel };
    window.addEventListener("pointermove", move, { passive: false });
    window.addEventListener("pointerup", up);
    window.addEventListener("pointercancel", cancel);
  };

  const handlePointerDown = (
    event: React.PointerEvent<HTMLDivElement>,
    node: CategoryTreeNode
  ) => {
    const touchStartedOnGrip = event.target instanceof Element &&
      Boolean(event.target.closest("[data-category-drag-handle]"));
    if (
      !canDragRows ||
      event.button !== 0 ||
      isInteractiveDragTarget(event.target) ||
      (event.pointerType === "touch" && !touchStartedOnGrip)
    ) return;

    const rect = event.currentTarget.getBoundingClientRect();
    event.currentTarget.setPointerCapture(event.pointerId);
    pointerDragRef.current = {
      guid: node.detail.guidfixed,
      parentGuid: node.detail.parentguid || "",
      rect: {
        left: rect.left,
        top: rect.top,
        width: rect.width,
        height: rect.height,
      },
      pointerId: event.pointerId,
      startX: event.clientX,
      startY: event.clientY,
      started: false,
    };
    setReorderError("");
    attachWindowDragListeners(event.pointerId);
  };

  // Deliberately no row-level onPointerMove/onPointerUp/onPointerCancel:
  // setPointerCapture (above) retargets delivery, but the event still bubbles
  // through the normal DOM tree up to window -- so a row-level handler here
  // would double-fire alongside the window listeners in
  // attachWindowDragListeners for every physical pointer move. The window
  // listeners alone are sufficient and are the single source of truth.

  const renderDragOverlay = () => {
    if (!dragState) return null;

    const node = categoryTreeLookups.nodeByGuid.get(dragState.guid);
    const detail = node?.detail ?? typedCategories.find((item) => item.guidfixed === dragState.guid);
    if (!detail) return null;

    const childCount = node?.childCategories.length ?? 0;
    const order = detail.xsorts?.[0]?.xorder ?? 0;
    const displayName = getDisplayName(detail.names);

    return (
      <div
        ref={dragGhostRef}
        className="pointer-events-none fixed z-[9999] flex items-center rounded-xl border border-dashed border-primary/55 bg-transparent px-2 py-1.5 text-foreground shadow-[0_0_0_1px_rgba(8,145,178,0.08),0_8px_20px_rgba(8,145,178,0.08)] ring-1 ring-primary/10"
        style={{
          left: dragState.rect.left,
          top: dragState.rect.top,
          width: dragState.rect.width,
          minHeight: dragState.rect.height,
          transform: "scale(1.006)",
          transformOrigin: "left center",
        }}
      >
        <div
          className="flex min-w-0 items-center gap-2 rounded-lg border border-primary/25 bg-background/75 px-2 py-1 shadow-sm backdrop-blur-[1px]"
          style={{ maxWidth: "min(520px, calc(100% - 1rem))" }}
        >
          <GripVertical className="size-4 shrink-0 text-primary" />
          <span className="min-w-[1.25rem] text-sm font-semibold text-slate-400">
            {order}
          </span>
          <span className="min-w-0 flex-1 whitespace-normal break-words text-sm font-semibold">
            {displayName}
          </span>
          {childCount > 0 ? (
            <span className="shrink-0 rounded-full border border-primary/20 bg-primary/10 px-2 py-0.5 text-[11px] font-semibold leading-5 text-primary">
              {language === "th"
                ? `ลูก ${childCount}`
                : `${childCount} ${childCount === 1 ? "child" : "children"}`}
            </span>
          ) : null}
          <span className="shrink-0 rounded-full border border-sky-200 dark:border-sky-800 bg-sky-50 dark:bg-sky-950/40 px-2 py-0.5 text-[11px] font-semibold leading-5 text-sky-700 dark:text-sky-300">
            {language === "th"
              ? `สินค้า ${detail.productCount ?? 0}`
              : `${detail.productCount ?? 0} ${detail.productCount === 1 ? "product" : "products"}`}
          </span>
          <span className="shrink-0 rounded-full bg-primary/90 px-2 py-0.5 text-[11px] font-semibold leading-5 text-primary-foreground">
            {language === "th" ? "กำลังย้าย" : "Moving"}
          </span>
        </div>
      </div>
    );
  };

  // Recursive Tree Node Renderer
  const renderTreeNodes = (nodes: CategoryTreeNode[], level = 0): React.ReactNode => {
    return (
      <div className="grid">
        {nodes.map((node, index) => {
          const childCount = node.childCategories.length;
          const hasChildren = childCount > 0;
          const isExpanded = Boolean(searchQuery.trim()) || (expandedNodes[node.detail.guidfixed] ?? false);
          const isSelected = selectedGuid === node.detail.guidfixed;
          const isDragging = dragState?.guid === node.detail.guidfixed;
          const isArrivalHighlighted = arrivalHighlight?.guid === node.detail.guidfixed;
          const activeDropTarget = dropTarget?.guid === node.detail.guidfixed ? dropTarget.position : null;
          const canAcceptChildDrop = dragState
            ? canDropOnTarget(dragState, node, [], "inside")
            : false;

          const order = node.detail.xsorts?.[0]?.xorder ?? (index + 1);
          const displayName = getDisplayName(node.detail.names);
          const levelStyle = categoryLevelStyle(level);
          return (
            <div key={node.detail.guidfixed} className="grid w-full">
              <div
                className={cn(
                  "group/row relative flex items-center justify-between border-b border-border/40 px-2 py-1.5 transition-[background-color,border-color,box-shadow,opacity,transform] duration-200 ease-out focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-primary/40",
                  canDragRows ? "cursor-grab select-none touch-pan-y active:cursor-grabbing" : "cursor-pointer",
                  isSelected
                    ? "bg-primary/5 text-primary"
                    : "hover:bg-muted/40",
                  isDragging &&
                    "pointer-events-none bg-primary/5 opacity-25 ring-1 ring-dashed ring-primary/30",
                  isArrivalHighlighted &&
                    "z-10 bg-emerald-50/80 ring-2 ring-emerald-400/60 dark:bg-emerald-950/35",
                  activeDropTarget && "bg-primary/10 shadow-[0_0_0_1px_var(--primary)]",
                  activeDropTarget === "inside" && "ring-2 ring-primary/30"
                )}
                data-category-row-guid={node.detail.guidfixed}
                aria-expanded={hasChildren ? isExpanded : undefined}
                aria-level={level + 1}
                aria-selected={isSelected}
                role="treeitem"
                tabIndex={0}
                style={{
                  paddingLeft: `${Math.min(level, 5) * 16 + 10}px`,
                }}
                onPointerDown={(event) => handlePointerDown(event, node)}
                onClick={(event) => {
                  if (ignoreNextClickRef.current) {
                    ignoreNextClickRef.current = false;
                    event.preventDefault();
                    event.stopPropagation();
                    return;
                  }
                  const orig = records.find(
                    (r) => recordGuid(r) === node.detail.guidfixed
                  );
                  if (orig) void onSelectRecord(recordWithOptimisticOverrides(orig));
                }}
                onKeyDown={(event) => {
                  if (
                    event.currentTarget !== event.target ||
                    (event.key !== "Enter" && event.key !== " ")
                  ) return;
                  event.preventDefault();
                  const orig = records.find(
                    (record) => recordGuid(record) === node.detail.guidfixed,
                  );
                  if (orig) void onSelectRecord(recordWithOptimisticOverrides(orig));
                }}
              >
                {isSelected ? (
                  <span className="pointer-events-none absolute bottom-0 left-0 top-0 w-1 bg-primary" />
                ) : null}
                {activeDropTarget === "before" ? (
                  <span className="pointer-events-none absolute left-2 right-2 top-0 z-30 flex -translate-y-1/2 items-center">
                    <span className="h-1 flex-1 rounded-full bg-primary shadow-[0_0_0_2px_rgba(255,255,255,0.65)]" />
                    <span className="ml-2 rounded-full bg-primary px-2 py-0.5 text-[11px] font-semibold leading-5 text-white shadow-sm">
                      {language === "th" ? "วางก่อน" : "Drop before"}
                    </span>
                  </span>
                ) : null}
                {activeDropTarget === "after" ? (
                  <span className="pointer-events-none absolute bottom-0 left-2 right-2 z-30 flex translate-y-1/2 items-center">
                    <span className="h-1 flex-1 rounded-full bg-primary shadow-[0_0_0_2px_rgba(255,255,255,0.65)]" />
                    <span className="ml-2 rounded-full bg-primary px-2 py-0.5 text-[11px] font-semibold leading-5 text-white shadow-sm">
                      {language === "th" ? "วางหลัง" : "Drop after"}
                    </span>
                  </span>
                ) : null}
                {activeDropTarget === "inside" ? (
                  <span className="pointer-events-none absolute inset-x-2 top-1/2 z-30 flex -translate-y-1/2 items-center justify-center">
                    <span className="rounded-full border border-emerald-700 bg-emerald-600 px-3 py-1 text-[11px] font-semibold leading-5 text-white shadow-md">
                      {language === "th" ? "วางเข้าเป็นลูกของหมวดนี้" : "Drop inside this category"}
                    </span>
                  </span>
                ) : null}
                <div className="flex min-w-0 flex-1 flex-wrap items-center gap-x-2 gap-y-1 pr-2">
                  <GripVertical
                    className={cn("size-4 shrink-0 touch-none transition-colors", levelStyle.grip)}
                    data-category-drag-handle
                  />
                  {/* Caret Toggle */}
                  <button
                    type="button"
                    className={cn(
                      "size-6 flex items-center justify-center rounded hover:bg-muted text-muted-foreground",
                      !hasChildren && "invisible"
                    )}
                    onClick={(e) => toggleExpand(node.detail.guidfixed, e)}
                  >
                    {isExpanded ? (
                      <ChevronDown className={cn("size-4", levelStyle.caret)} />
                    ) : (
                      <ChevronRight className={cn("size-4", levelStyle.caret)} />
                    )}
                  </button>

                  <span className={cn("min-w-[1.25rem] text-sm font-bold", levelStyle.order)}>
                    {order}
                  </span>

                  <span className={cn("min-w-0 flex-1 basis-40 whitespace-normal break-words text-sm", levelStyle.name)}>
                    {displayName}
                  </span>
                  {hasChildren ? (
                    <button
                      type="button"
                      className="shrink-0 rounded-full border border-primary/20 bg-primary/10 px-2 py-0.5 text-[11px] font-semibold leading-5 text-primary transition-colors hover:border-primary/40 hover:bg-primary/15 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/40"
                      aria-label={
                        language === "th"
                          ? `${isExpanded ? "ซ่อน" : "แสดง"}หมวดย่อย ${childCount} รายการ`
                          : `${isExpanded ? "Hide" : "Show"} ${childCount} child ${childCount === 1 ? "category" : "categories"}`
                      }
                      title={
                        language === "th"
                          ? `${isExpanded ? "ซ่อน" : "แสดง"}หมวดย่อย`
                          : `${isExpanded ? "Hide" : "Show"} children`
                      }
                      onClick={(e) => toggleExpand(node.detail.guidfixed, e)}
                    >
                      {language === "th"
                        ? `หมวดย่อย ${childCount}`
                        : `${childCount} ${childCount === 1 ? "child" : "children"}`}
                    </button>
                  ) : null}
                  <span className="shrink-0 rounded-full border border-border bg-muted/60 px-2 py-0.5 text-[11px] font-semibold leading-5 text-muted-foreground">
                    {language === "th"
                      ? `สินค้า ${node.detail.productCount ?? 0}`
                      : `${node.detail.productCount ?? 0} ${node.detail.productCount === 1 ? "product" : "products"}`}
                  </span>
                  {isArrivalHighlighted ? (
                    <span
                      key={arrivalHighlight?.nonce ?? node.detail.guidfixed}
                      className="shrink-0 animate-pulse rounded-full bg-emerald-600 px-2 py-0.5 text-[11px] font-semibold leading-5 text-white shadow-sm"
                    >
                      {language === "th" ? "ย้ายมาแล้ว" : "Moved here"}
                    </span>
                  ) : null}
                  {canAcceptChildDrop ? (
                    <span
                      className={cn(
                        "shrink-0 rounded-full border border-dashed border-emerald-500 bg-emerald-50 px-2.5 py-1 text-[11px] font-semibold leading-5 text-emerald-700 shadow-sm transition-[background-color,border-color,color,box-shadow,transform] duration-150 dark:bg-emerald-950/30 dark:text-emerald-300",
                        activeDropTarget === "inside" &&
                          "scale-110 border-emerald-700 bg-emerald-600 text-white shadow-md dark:bg-emerald-500 dark:text-white"
                      )}
                      aria-label={language === "th" ? "วางเป็นหมวดย่อย" : "Drop as child category"}
                      data-category-inside-drop-guid={node.detail.guidfixed}
                    >
                      {language === "th" ? "วางเป็นลูก" : "Drop child"}
                    </span>
                  ) : null}
                </div>

                {/* Row actions */}
                {!readOnly && (
                  <div
                    className={cn(
                      "flex shrink-0 items-center gap-0.5 transition-opacity group-hover/row:opacity-100 group-focus-within/row:opacity-100",
                      isSelected ? "opacity-100" : "opacity-45",
                    )}
                  >
                    <Button
                      type="button"
                      variant="ghost"
                      size="icon"
                      className="size-7 rounded-full text-muted-foreground hover:text-foreground hover:bg-muted disabled:opacity-30"
                      aria-label={language === "th" ? "ย้ายขึ้น" : "Move up"}
                      disabled={index === 0 || !canDragRows || isMoveInFlight}
                      onClick={(e) => {
                        e.stopPropagation();
                        if (index === 0) return;
                        void reorderCategory(node.detail.guidfixed, nodes[index - 1].detail.guidfixed, "before", nodes);
                      }}
                    >
                      <ChevronUp className="size-3.5" />
                    </Button>
                    <Button
                      type="button"
                      variant="ghost"
                      size="icon"
                      className="size-7 rounded-full text-muted-foreground hover:text-foreground hover:bg-muted disabled:opacity-30"
                      aria-label={language === "th" ? "ย้ายลง" : "Move down"}
                      disabled={index === nodes.length - 1 || !canDragRows || isMoveInFlight}
                      onClick={(e) => {
                        e.stopPropagation();
                        if (index === nodes.length - 1) return;
                        void reorderCategory(node.detail.guidfixed, nodes[index + 1].detail.guidfixed, "after", nodes);
                      }}
                    >
                      <ChevronDown className="size-3.5" />
                    </Button>
                    {onOpenEdit && (
                      <Button
                        type="button"
                        variant="ghost"
                        size="icon"
                        className="size-7 rounded-full text-blue-600 hover:text-blue-700 hover:bg-blue-50 dark:hover:bg-blue-950/40"
                        aria-label={language === "th" ? "แก้ไข" : "Edit"}
                        title={language === "th" ? "แก้ไขหมวดนี้" : "Edit"}
                        onClick={(e) => {
                          e.stopPropagation();
                          setSelectedGuid?.(node.detail.guidfixed);
                          const orig = records.find(
                            (r) => recordGuid(r) === node.detail.guidfixed
                          );
                          if (orig) onOpenEdit(recordWithOptimisticOverrides(orig));
                        }}
                      >
                        <Edit3 className="size-3.5" />
                      </Button>
                    )}
                    {onDeleteRecord && (
                      <Button
                        type="button"
                        variant="ghost"
                        size="icon"
                        className="size-7 rounded-full text-destructive hover:bg-destructive/10"
                        aria-label={language === "th" ? "ลบ" : "Delete"}
                        title={language === "th" ? "ลบหมวดนี้" : "Delete"}
                        onClick={(e) => {
                          e.stopPropagation();
                          const orig = records.find(
                            (r) => recordGuid(r) === node.detail.guidfixed
                          );
                          if (orig) onDeleteRecord(orig);
                        }}
                      >
                        <Trash2 className="size-3.5" />
                      </Button>
                    )}
                  </div>
                )}
              </div>

              {/* Recursive Children rendering */}
              {hasChildren && isExpanded && (
                <div className="w-full">
                  {renderTreeNodes(node.childCategories, level + 1)}
                </div>
              )}
            </div>
          );
        })}
      </div>
    );
  };

  return (
    <div className="grid h-full min-h-0 w-full gap-2" data-testid="product-category-tree-pane">
      {/* Categories Tree list */}
      <Card className="flex h-full min-h-0 flex-col overflow-hidden border-border bg-card shadow-sm">
        <CardContent className="flex min-h-0 flex-1 flex-col p-0">
          {!readOnly ? (
            <div className="flex min-w-0 items-center justify-between gap-2 border-b border-border/40 bg-muted/20 px-2.5 py-1.5">
              <div className="flex min-w-0 items-baseline gap-2">
                <span className="text-xs font-bold text-foreground">
                  {language === "th" ? "โครงสร้างหมวด" : "Category tree"}
                </span>
                <span className="shrink-0 text-[11px] text-muted-foreground">
                  {language === "th" ? `${records.length} หมวด` : `${records.length} categories`}
                </span>
              </div>
              <div className="flex shrink-0 items-center gap-0.5">
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  className="h-7 gap-1 px-2 text-xs disabled:opacity-40"
                  disabled={!canUseHistoryControls || moveHistory.index < 0}
                  onClick={() => void undoMove()}
                >
                  <Undo2 className="size-3.5" />
                  {language === "th" ? "เลิกทำ" : "Undo"}
                </Button>
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  className="h-7 gap-1 px-2 text-xs disabled:opacity-40"
                  disabled={!canUseHistoryControls || moveHistory.index >= moveHistory.stack.length - 1}
                  onClick={() => void redoMove()}
                >
                  <Redo2 className="size-3.5" />
                  {language === "th" ? "ทำซ้ำ" : "Redo"}
                </Button>
              </div>
            </div>
          ) : null}
          {loading ? (
            <div className="flex h-full min-h-24 items-center justify-center gap-2 p-6 text-sm text-muted-foreground">
              <Loader2 className="animate-spin size-5" />
              {language === "th" ? "กำลังโหลดข้อมูล..." : "Loading..."}
            </div>
          ) : treeRoots.length === 0 ? (
            <div className="flex h-full min-h-24 flex-col items-center justify-center gap-1.5 p-4 text-center text-sm text-muted-foreground">
              <span className="font-medium">
                {searchQuery.trim()
                  ? language === "th"
                    ? "ไม่พบหมวดที่ค้นหา"
                    : "No matching categories"
                  : language === "th"
                    ? "ยังไม่มีหมวดสินค้า"
                    : "No categories yet"}
              </span>
              <span className="text-xs text-muted-foreground max-w-xs">
                {searchQuery.trim()
                  ? language === "th"
                    ? "ลองใช้คำค้นที่สั้นลง หรือตรวจสอบการสะกดอีกครั้ง"
                    : "Try a shorter search or check the spelling."
                  : readOnly
                  ? language === "th"
                    ? "กลุ่มนี้ยังไม่มีหมวดสินค้า กรุณาไปสร้างหมวดสินค้าที่หน้าจอ 'จัดหมวดสินค้า' ก่อน"
                    : "This group has no categories yet. Please create categories on the 'Product Categories' screen first."
                  : language === "th"
                    ? "คุณสามารถกดปุ่ม 'เพิ่มหมวดสินค้า' ด้านบนเพื่อสร้างข้อมูลใหม่ได้"
                    : "You can click 'Add Category' above to start adding categories."}
              </span>
            </div>
          ) : (
            <div
              ref={treeListRef}
              className="flex min-h-0 flex-1 flex-col overflow-y-auto overscroll-contain"
              role="tree"
            >
              {reorderError ? (
                <div className="border-b border-destructive/30 bg-destructive/10 px-3 py-2 text-sm font-medium text-destructive">
                  {reorderError}
                </div>
              ) : null}
              {searchQuery.trim() ? (
                <div className="border-b border-border/40 bg-muted/40 px-3 py-2 text-xs text-muted-foreground">
                  {language === "th"
                    ? "ระหว่างค้นหา ระบบจะปิดการจัดลำดับชั่วคราว"
                    : "Reordering is temporarily disabled while searching."}
                </div>
              ) : null}
              {renderTreeNodes(treeRoots)}
              {dragState ? (
                <div
                  aria-label={language === "th" ? "ย้ายเป็นหมวดหลัก" : "Move to root category"}
                  className={cn(
                    "sticky bottom-0 z-20 mt-auto flex h-10 w-full items-center justify-center gap-2 border-t border-emerald-600 bg-emerald-500 text-sm font-semibold text-white transition-[background-color,box-shadow,transform] duration-150",
                    rootDropActive && "scale-[0.995] bg-emerald-600 shadow-inner",
                  )}
                  data-category-root-drop="true"
                  title={language === "th" ? "ย้ายเป็นหมวดหลัก" : "Move to root category"}
                >
                  <Home className="size-6" />
                  <span>
                    {language === "th" ? "วางเป็นหมวดหลัก" : "Drop as root"}
                  </span>
                </div>
              ) : null}
            </div>
          )}
        </CardContent>
      </Card>
      {renderDragOverlay()}
    </div>
  );
}
