// Route entry for /signer/certificates/[serial]. Resolves the dynamic
// segment server-side and hands the value to the detail client component.
import { CertificateDetailPage } from "../../CertificateDetailPage";

type Props = {
  params: Promise<{
    serial: string;
  }>;
};

export default async function CertificatePage({ params }: Props) {
  const { serial } = await params;
  return <CertificateDetailPage serial={serial} />;
}
