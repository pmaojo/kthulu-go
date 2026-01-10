import { getDocBySlug, REGISTRY_DIR } from '@/lib/cms';
import { notFound } from 'next/navigation';
import { Sidebar } from '@/components/Navigation';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import { Book, Clock, User, Terminal, Star, Zap, Shield, Cloud } from 'lucide-react';

export default async function DocPage({ params }: { params: { slug: string[] } }) {
  const { slug } = await params;

  let doc;
  let isMarketplace = false;

  // Handle Marketplace routes: /docs/marketplace/starters/my-starter
  if (slug[0] === 'marketplace' && slug.length > 2) {
    isMarketplace = true;
    // Strip 'marketplace' and use the rest of the slug
    const marketplaceSlug = slug.slice(1);
    doc = getDocBySlug(marketplaceSlug, REGISTRY_DIR);
  } else {
    // Normal docs
    doc = getDocBySlug(slug);
  }

  if (!doc) {
    notFound();
  }

  // Helper to determine if content is sparse (placeholder)
  const isSparse = doc.content.trim().split('\n').length < 10;

  const iconMap: Record<string, any> = {
    Zap: <Zap size={20} />,
    Shield: <Shield size={20} />,
    Cloud: <Cloud size={20} />,
  };

  const Icon = doc.frontmatter.icon ? iconMap[doc.frontmatter.icon] : null;

  return (
    <div className="flex">
      <Sidebar />
      <div className="flex-1 md:ml-64 p-6 md:p-12 lg:p-24">
        <article className="max-w-3xl mx-auto">
          <header className="mb-12 border-b border-white/10 pb-12">
            <div className="flex items-center gap-2 text-primary font-mono text-xs uppercase tracking-widest mb-4">
              <Book size={14} /> {isMarketplace ? 'Marketplace' : 'Documentation'} / {slug[isMarketplace ? 2 : 0]}
            </div>

            <div className="flex items-start justify-between gap-4 mb-6">
              <h1 className="text-5xl font-bold tracking-tighter kthulu-glow leading-tight">
                {doc.title}
              </h1>
              {Icon && (
                <div className="p-3 bg-white/5 rounded-xl border border-white/10 text-primary hidden md:block">
                  {Icon}
                </div>
              )}
            </div>

            <div className="flex flex-wrap items-center gap-6 text-sm text-muted-foreground">
              {doc.description && (
                <p className="text-xl text-foreground/80 w-full mb-6 leading-relaxed border-l-2 border-primary/50 pl-4">
                  {doc.description}
                </p>
              )}

              <div className="flex items-center gap-6 w-full">
                 {doc.frontmatter.author && (
                  <span className="flex items-center gap-2 bg-white/5 px-3 py-1 rounded-full border border-white/10">
                    <User size={14} /> {doc.frontmatter.author}
                  </span>
                )}
                {doc.frontmatter.stars && (
                   <span className="flex items-center gap-2 bg-white/5 px-3 py-1 rounded-full border border-white/10 text-yellow-500">
                    <Star size={14} fill="currentColor" /> {doc.frontmatter.stars}
                  </span>
                )}
                <span className="flex items-center gap-2 ml-auto opacity-50">
                  <Clock size={14} /> SSG Generated
                </span>
              </div>
            </div>
          </header>
          
          {/* Enhanced content for sparse/generated modules */}
          {isMarketplace && isSparse && (
            <div className="mb-12 space-y-8">
              <div className="bg-white/[0.02] border border-white/10 rounded-xl p-6">
                <h3 className="text-lg font-bold mb-4 flex items-center gap-2">
                  <Terminal size={18} className="text-primary" /> Installation
                </h3>
                <div className="bg-black/50 rounded-lg p-4 font-mono text-sm text-muted-foreground border border-white/5 flex items-center justify-between group">
                  <code>kthulu add module {doc.frontmatter.id || slug[slug.length-1]}</code>
                </div>
              </div>

              <div className="grid md:grid-cols-2 gap-6">
                <div className="bg-white/[0.02] border border-white/10 rounded-xl p-6">
                   <h3 className="text-lg font-bold mb-2">Features</h3>
                   <ul className="list-disc list-inside space-y-2 text-muted-foreground text-sm">
                     <li>Production-ready {doc.frontmatter.type || 'module'} architecture</li>
                     <li>Seamless Kthulu CLI integration</li>
                     <li>Fully typed Go/TypeScript implementation</li>
                   </ul>
                </div>
                 <div className="bg-white/[0.02] border border-white/10 rounded-xl p-6">
                   <h3 className="text-lg font-bold mb-2">Requirements</h3>
                   <ul className="list-disc list-inside space-y-2 text-muted-foreground text-sm">
                     <li>Kthulu CLI v1.0.0+</li>
                     <li>Go 1.21+</li>
                     <li>Docker (Optional)</li>
                   </ul>
                </div>
              </div>
            </div>
          )}

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
