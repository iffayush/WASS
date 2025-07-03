import './globals.css';
import Providers from './providers';

export const metadata = {
  title: 'WASS',
  description: 'Web App Security Scanner',
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body>
        <Providers initialSession={null}>{children}</Providers>
      </body>
    </html>
  );
}
