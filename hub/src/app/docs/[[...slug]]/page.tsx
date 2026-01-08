import { getDocBySlug } from '@/lib/cms';
import { notFound } from 'next/navigation';
import { Sidebar } from '@/components/Navigation';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import { Book, Clock, User } from 'lucide-react';

export default async function DocPage({ params }: { params: { slug: string[] } }) {
  const { slug } = await params;
  const doc = getDocBySlug(slug);

  if (!doc) {
    notFound();
  }

  return (
    <div className="flex">
      <Sidebar />
      <div className="flex-1 md:ml-64 p-6 md:p-12 lg:p-24">
        <article className="max-w-3xl mx-auto">
          <header className="mb-12 border-b border-white/10 pb-12">
            <div className="flex items-center gap-2 text-primary font-mono text-xs uppercase tracking-widest mb-4">
              <Book size={14} /> Documentation / {slug[0]}
            </div>
            <h1 className="text-5xl font-bold tracking-tighter mb-6 kthulu-glow leading-tight">
              {doc.title}
            </h1>
            <div className="flex flex-wrap items-center gap-6 text-sm text-muted-foreground">
              {doc.description && (
                <p className="text-xl text-foreground/80 w-full mb-4 leading-relaxed">
                  {doc.description}
                </p>
              )}
              {doc.frontmatter.author && (
                <span className="flex items-center gap-2">
                  <User size={14} /> {doc.frontmatter.author}
                </span>
              )}
              <span className="flex items-center gap-2">
                <Clock size={14} /> SSG Generated
              </span>
            </div>
          </header>
          
          <div className="prose-custom">
            <ReactMarkdown remarkPlugins={[remarkGfm]}>
              {doc.content}
            </ReactMarkdown>
          </div>
        </article>
      </div>
    </div>
  );
}
