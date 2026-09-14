import { GeneralLedgerScreen } from "../general-ledger-screen";

export default async function GeneralLedgerPage({ params }: { params: Promise<{ glPath: string[] }> }) {
  const { glPath } = await params;
  return <GeneralLedgerScreen route={`/gl/${glPath.join("/")}`} />;
}
