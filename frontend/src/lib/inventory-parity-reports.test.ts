import { describe, it, expect } from "vitest";
import {
  getErpReportConfig,
  isErpReportRoute,
  reportText,
} from "./erp-reports";

describe("Inventory Parity ERP Reports (Champ Parity 5589, 5577, 5654, 5591, 5595, 5610)", () => {
  it("defines max stock alert report with proper columns and metadata", () => {
    const route = "/report/maxstock";
    expect(isErpReportRoute(route)).toBe(true);

    const config = getErpReportConfig(route);
    expect(config).toBeDefined();
    expect(config?.code).toBe("max_stock_report");
    expect(config?.category).toBe("inventory");
    expect(config?.title.th).toContain("ถึงจุดสูงสุด");

    const columnKeys = config?.columns.map((c) => c.key);
    expect(columnKeys).toContain("itemcode");
    expect(columnKeys).toContain("itemname");
    expect(columnKeys).toContain("unitname");
    expect(columnKeys).toContain("maxqty");
    expect(columnKeys).toContain("qty");
    expect(columnKeys).toContain("overqty");

    const overCol = config?.columns.find((c) => c.key === "overqty");
    expect(overCol?.isNumeric).toBe(true);
    expect(overCol?.align).toBe("right");
  });

  it("defines no-movement dead stock report with proper columns and metadata", () => {
    const route = "/report/nomovementstock";
    expect(isErpReportRoute(route)).toBe(true);

    const config = getErpReportConfig(route);
    expect(config).toBeDefined();
    expect(config?.code).toBe("no_movement_stock_report");
    expect(config?.category).toBe("inventory");
    expect(config?.title.th).toContain("ไม่มีการเคลื่อนไหว");

    const columnKeys = config?.columns.map((c) => c.key);
    expect(columnKeys).toContain("itemcode");
    expect(columnKeys).toContain("itemname");
    expect(columnKeys).toContain("unitname");
    expect(columnKeys).toContain("lastmovementdate");
    expect(columnKeys).toContain("qty");
    expect(columnKeys).toContain("totalcost");

    const costCol = config?.columns.find((c) => c.key === "totalcost");
    expect(costCol?.isCurrency).toBe(true);
    expect(costCol?.align).toBe("right");
  });

  it("defines stock count variance report with proper columns and metadata", () => {
    const route = "/report/stockcountvariance";
    expect(isErpReportRoute(route)).toBe(true);

    const config = getErpReportConfig(route);
    expect(config).toBeDefined();
    expect(config?.code).toBe("stock_count_variance_report");
    expect(config?.category).toBe("inventory");
    expect(config?.title.th).toContain("ผลต่างจากการตรวจนับ");

    const columnKeys = config?.columns.map((c) => c.key);
    expect(columnKeys).toContain("itemcode");
    expect(columnKeys).toContain("itemname");
    expect(columnKeys).toContain("systemqty");
    expect(columnKeys).toContain("countedqty");
    expect(columnKeys).toContain("varianceqty");
    expect(columnKeys).toContain("varianceamount");
  });

  it("defines pending receive and pending delivery reports with proper columns and metadata", () => {
    const recRoute = "/report/pendingreceive";
    expect(isErpReportRoute(recRoute)).toBe(true);
    const recConfig = getErpReportConfig(recRoute);
    expect(recConfig?.code).toBe("pending_receive_report");
    expect(recConfig?.columns.map((c) => c.key)).toContain("pono");
    expect(recConfig?.columns.map((c) => c.key)).toContain("pendingqty");

    const delRoute = "/report/pendingdelivery";
    expect(isErpReportRoute(delRoute)).toBe(true);
    const delConfig = getErpReportConfig(delRoute);
    expect(delConfig?.code).toBe("pending_delivery_report");
    expect(delConfig?.columns.map((c) => c.key)).toContain("sono");
    expect(delConfig?.columns.map((c) => c.key)).toContain("pendingqty");
  });

  it("defines serial number movement report with proper columns and metadata", () => {
    const route = "/report/serialmovement";
    expect(isErpReportRoute(route)).toBe(true);

    const config = getErpReportConfig(route);
    expect(config).toBeDefined();
    expect(config?.code).toBe("serial_movement_report");
    expect(config?.category).toBe("inventory");
    expect(config?.title.th).toContain("Serial Number");

    const columnKeys = config?.columns.map((c) => c.key);
    expect(columnKeys).toContain("serialno");
    expect(columnKeys).toContain("itemcode");
    expect(columnKeys).toContain("movementtype");
    expect(columnKeys).toContain("warehouse");
    expect(columnKeys).toContain("status");

    const thTitle = reportText(config!, "title", "th");
    expect(thTitle).toBeTruthy();
  });
});
