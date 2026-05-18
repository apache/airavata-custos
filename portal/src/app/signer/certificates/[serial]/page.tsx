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
