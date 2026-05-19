"use client";

import {
  type ColumnDef,
  type ColumnFiltersState,
  type SortingState,
  type VisibilityState,
  flexRender,
  getCoreRowModel,
  getFilteredRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  useReactTable,
} from "@tanstack/react-table";
import { ArrowUpDown, Eye, MoreHorizontal, Search } from "lucide-react";
import { useMemo, useState } from "react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Input } from "@/components/ui/input";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { backendText, type BackendLanguageDictionary } from "@/lib/backend-language";
import type { MenuItem } from "@/lib/menu-data";
import { cn } from "@/lib/utils";
import { MenuRouteIcon } from "./menu-icon";
import type { ErpMenuRow } from "./menu-dashboard-data";

type MenuDataTableProps = {
  data: ErpMenuRow[];
  dictionary: BackendLanguageDictionary;
  globalSearch: string;
  onGlobalSearchChange: (value: string) => void;
  onOpen: (item: MenuItem) => void;
};

export function MenuDataTable({ data, dictionary, globalSearch, onGlobalSearchChange, onOpen }: MenuDataTableProps) {
  const [sorting, setSorting] = useState<SortingState>([{ id: "activity", desc: true }]);
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([]);
  const [columnVisibility, setColumnVisibility] = useState<VisibilityState>({
    route: false,
  });

  const columns = useMemo<ColumnDef<ErpMenuRow>[]>(
    () => [
      {
        accessorKey: "name",
        header: ({ column }) => (
          <Button variant="ghost" size="sm" onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}>
            {backendText(dictionary, "menu")}
            <ArrowUpDown className="h-3.5 w-3.5" />
          </Button>
        ),
        cell: ({ row }) => (
          <div className="flex min-w-60 items-center gap-3">
            <span className="grid h-9 w-9 shrink-0 place-items-center rounded-2xl bg-primary/10 text-primary">
              <MenuRouteIcon item={row.original.item} size={17} />
            </span>
            <div className="min-w-0">
              <div className="truncate font-semibold">{row.original.name}</div>
              <div className="truncate text-xs text-muted-foreground">{row.original.route}</div>
            </div>
          </div>
        ),
      },
      {
        accessorKey: "module",
        header: backendText(dictionary, "module"),
        cell: ({ row }) => <Badge variant="outline">{row.original.module}</Badge>,
      },
      {
        accessorKey: "group",
        header: backendText(dictionary, "group"),
        cell: ({ row }) => <span className="line-clamp-1 text-sm text-muted-foreground">{row.original.group}</span>,
      },
      {
        accessorKey: "status",
        header: backendText(dictionary, "status"),
        cell: ({ row }) => {
          const status = row.original.status;
          return (
            <Badge variant={status === "ready" ? "success" : status === "migration" ? "warning" : "secondary"}>
              {status === "ready" ? backendText(dictionary, "status_ready") : status === "migration" ? backendText(dictionary, "status_migration") : backendText(dictionary, "status_planned")}
            </Badge>
          );
        },
      },
      {
        accessorKey: "activity",
        header: ({ column }) => (
          <Button variant="ghost" size="sm" onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}>
            {backendText(dictionary, "activity")}
            <ArrowUpDown className="h-3.5 w-3.5" />
          </Button>
        ),
        cell: ({ row }) => <span className="font-mono text-sm">{row.original.activity}</span>,
      },
      {
        accessorKey: "updatedAt",
        header: backendText(dictionary, "updated_at"),
      },
      {
        accessorKey: "route",
        header: backendText(dictionary, "route"),
      },
      {
        id: "actions",
        enableHiding: false,
        cell: ({ row }) => (
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="icon" aria-label={`${backendText(dictionary, "commands")} ${row.original.name}`}>
                <MoreHorizontal className="h-4 w-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem onClick={() => onOpen(row.original.item)}>
                <Eye className="h-4 w-4" />
                {backendText(dictionary, "open_as_tab")}
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        ),
      },
    ],
    [dictionary, onOpen],
  );

  const table = useReactTable({
    data,
    columns,
    state: {
      sorting,
      columnFilters,
      columnVisibility,
      globalFilter: globalSearch,
    },
    onSortingChange: setSorting,
    onColumnFiltersChange: setColumnFilters,
    onColumnVisibilityChange: setColumnVisibility,
    onGlobalFilterChange: onGlobalSearchChange,
    getCoreRowModel: getCoreRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    initialState: {
      pagination: {
        pageSize: 8,
      },
    },
  });

  return (
    <div className="grid min-w-0 grid-cols-[minmax(0,1fr)] gap-3">
      <div className="flex min-w-0 flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
        <label className="relative w-full min-w-0 sm:max-w-sm">
          <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input className="pl-9" placeholder={backendText(dictionary, "search_menu_group_route")} value={globalSearch} onChange={(event) => onGlobalSearchChange(event.target.value)} />
        </label>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="outline">{backendText(dictionary, "columns")}</Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            {table
              .getAllColumns()
              .filter((column) => column.getCanHide())
              .map((column) => (
                <DropdownMenuCheckboxItem
                  key={column.id}
                  checked={column.getIsVisible()}
                  onCheckedChange={(value) => column.toggleVisibility(!!value)}
                >
                  {column.id}
                </DropdownMenuCheckboxItem>
              ))}
          </DropdownMenuContent>
        </DropdownMenu>
      </div>

      <div className="grid gap-2 md:hidden">
        {table.getRowModel().rows.length ? (
          table.getRowModel().rows.map((row) => (
            <button
              className="grid w-full grid-cols-[auto_minmax(0,1fr)] gap-3 rounded-2xl border border-border bg-card p-3 text-left shadow-sm"
              key={row.id}
              onClick={() => onOpen(row.original.item)}
              type="button"
            >
              <span className="grid h-10 w-10 place-items-center rounded-2xl bg-primary/10 text-primary">
                <MenuRouteIcon item={row.original.item} size={18} />
              </span>
              <span className="min-w-0">
                <span className="block truncate text-sm font-semibold">{row.original.name}</span>
                <span className="block truncate text-xs text-muted-foreground">{row.original.route}</span>
                <span className="mt-2 flex flex-wrap gap-1.5">
                  <Badge variant="outline">{row.original.module}</Badge>
                  <Badge variant={row.original.status === "ready" ? "success" : row.original.status === "migration" ? "warning" : "secondary"}>
                    {row.original.status === "ready" ? backendText(dictionary, "status_ready") : row.original.status === "migration" ? backendText(dictionary, "status_migration") : backendText(dictionary, "status_planned")}
                  </Badge>
                  <Badge variant="secondary">{backendText(dictionary, "activity")} {row.original.activity}</Badge>
                </span>
              </span>
            </button>
          ))
        ) : (
          <div className="rounded-2xl border border-border bg-card p-4 text-center text-sm text-muted-foreground">{backendText(dictionary, "no_data_by_filter")}</div>
        )}
      </div>

      <div className="hidden min-w-0 overflow-hidden rounded-2xl border border-border bg-card md:block">
        <Table>
          <TableHeader>
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow key={headerGroup.id}>
                {headerGroup.headers.map((header) => (
                  <TableHead key={header.id} className={cn(header.column.id === "actions" && "w-12")}>
                    {header.isPlaceholder ? null : flexRender(header.column.columnDef.header, header.getContext())}
                  </TableHead>
                ))}
              </TableRow>
            ))}
          </TableHeader>
          <TableBody>
            {table.getRowModel().rows.length ? (
              table.getRowModel().rows.map((row) => (
                <TableRow key={row.id}>
                  {row.getVisibleCells().map((cell) => (
                    <TableCell key={cell.id}>{flexRender(cell.column.columnDef.cell, cell.getContext())}</TableCell>
                  ))}
                </TableRow>
              ))
            ) : (
              <TableRow>
                <TableCell colSpan={columns.length} className="h-28 text-center text-muted-foreground">
                  {backendText(dictionary, "no_data_by_filter")}
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>

      <div className="flex min-w-0 flex-col gap-2 text-sm text-muted-foreground sm:flex-row sm:items-center sm:justify-between">
        <span>
          {backendText(dictionary, "displaying")} {table.getRowModel().rows.length} {backendText(dictionary, "from_count")} {table.getFilteredRowModel().rows.length} {backendText(dictionary, "items")}
        </span>
        <div className="flex gap-2">
          <Button variant="outline" size="sm" onClick={() => table.previousPage()} disabled={!table.getCanPreviousPage()}>
            {backendText(dictionary, "previous")}
          </Button>
          <Button variant="outline" size="sm" onClick={() => table.nextPage()} disabled={!table.getCanNextPage()}>
            {backendText(dictionary, "next")}
          </Button>
        </div>
      </div>
    </div>
  );
}
