import { encodeEan13 } from "@/lib/product-barcode/utils";

export function Ean13Barcode({
  className,
  value,
}: {
  className?: string;
  value: string;
}) {
  const modules = encodeEan13(value);
  if (!modules) return null;

  return (
    <svg
      aria-label={`EAN-13 ${value}`}
      className={className}
      preserveAspectRatio="xMidYMid meet"
      role="img"
      shapeRendering="crispEdges"
      viewBox="0 0 115 62"
    >
      <rect fill="white" height="62" width="115" />
      {Array.from(modules, (module, index) =>
        module === "1" ? (
          <rect fill="black" height="46" key={index} width="1" x={index + 10} y="2" />
        ) : null,
      )}
      <text
        fill="black"
        fontFamily="Arial, sans-serif"
        fontSize="9"
        letterSpacing="1.5"
        textAnchor="middle"
        x="57.5"
        y="59"
      >
        {value}
      </text>
    </svg>
  );
}
