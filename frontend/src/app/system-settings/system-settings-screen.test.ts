import { readFileSync } from "node:fs";
import { join } from "node:path";
import { describe, expect, it } from "vitest";

const source = readFileSync(
  join(process.cwd(), "src/app/system-settings/system-settings-screen.tsx"),
  "utf8",
);

describe("SystemSettingsScreen layout contract", () => {
  it("keeps standalone system setting pages full width without hiding horizontal overflow", () => {
    expect(source).toContain(
      '<main className="min-h-dvh w-full max-w-none bg-background p-2 text-foreground sm:p-3">',
    );
    expect(source).not.toContain(
      '<main className="min-h-dvh w-full overflow-x-hidden bg-background p-2 text-foreground sm:p-3">',
    );
  });
});
