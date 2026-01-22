import { Sidebar } from '@/components/Navigation';

export default function DocsLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div className="flex">
      <Sidebar />
      <div className="flex-1 md:ml-64 p-6 md:p-12 lg:p-24">
        {children}
      </div>
    </div>
  );
}
