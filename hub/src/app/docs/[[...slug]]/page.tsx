import { getDocBySlug, REGISTRY_DIR } from '@/lib/cms';
import { notFound } from 'next/navigation';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import { Book, Clock, User, Terminal, Star, Zap, Shield, Cloud, ArrowLeft, ArrowRight } from 'lucide-react';
import Link from 'next/link';

const TOUR_ORDER = [
  '/docs/guide/introduction',
  '/docs/guide/installation',
  '/docs/guide/quick-start',
  '/docs/guide/project-structure',
  '/docs/guide/modules',
  '/docs/guide/deployment',
  '/docs/guide/cli-reference',
];

const TOUR_TITLES: Record<string, string> = {
  '/docs/guide/introduction': 'Introduction',
  '/docs/guide/installation': 'Installation',
  '/docs/guide/quick-start': 'Quick Start Tour',
  '/docs/guide/project-structure': 'Project Structure',
  '/docs/guide/modules': 'Modules',
  '/docs/guide/deployment': 'Deployment',
  '/docs/guide/cli-reference': 'CLI Reference',
};

export default async function DocPage({ params }: { params: { slug: string[] } }) {
  const { slug = [] } = await params;

  const currentPath = `/docs/${slug.join('/')}`;
  const currentIndex = TOUR_ORDER.indexOf(currentPath);
  const prevPage = currentIndex > 0 ? TOUR_ORDER[currentIndex - 1] : null;
  const nextPage = currentIndex >= 0 && currentIndex < TOUR_ORDER.length - 1 ? TOUR_ORDER[currentIndex + 1] : null;

  let doc;
  let isMarketplace = false;

  // Handle Marketplace routes: /docs/marketplace/starters/my-starter
  if (slug.length > 0 && slug[0] === 'marketplace' && slug.length > 2) {
    isMarketplace = true;
    // Strip 'marketplace' and use the rest of the slug
    const marketplaceSlug = slug.slice(1);
    doc = await getDocBySlug(marketplaceSlug, REGISTRY_DIR);
  } else {
    // Normal docs
    doc = await getDocBySlug(slug);
  }

  if (!doc) {
    notFound();
  }

  // Helper to determine if content is sparse (placeholder)
  // Ensure content exists before checking length. If fetched without fields, it should be there.
  const content = doc.content || '';
  const isSparse = content.trim().split('\n').length < 10;

  const iconMap: Record<string, React.ReactNode> = {
    Zap: <Zap size={20} />,
    Shield: <Shield size={20} />,
    Cloud: <Cloud size={20} />,
  };

  const Icon = doc.frontmatter.icon ? iconMap[doc.frontmatter.icon] : null;

  return (
        <article className="max-w-3xl mx-auto">
          <header className="mb-12 border-b border-white/10 pb-12">
            <div className="flex items-center gap-2 text-primary font-mono text-xs uppercase tracking-widest mb-4">
              <Book size={14} /> {isMarketplace ? 'Marketplace' : 'Documentation'} / {slug.length > 0 ? slug[isMarketplace ? 2 : 0] : 'Index'}
            </div>

            <div className="flex items-start justify-between gap-4 mb-6">
              <h1 className="text-3xl md:text-5xl font-bold tracking-tighter kthulu-glow leading-tight break-words">
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
                  <code>kthulu add module {doc.frontmatter.id || (slug.length > 0 ? slug[slug.length-1] : '')}</code>
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
              {content}
            </ReactMarkdown>
          </div>

          <div className="mt-12 pt-12 border-t border-white/10 flex flex-col md:flex-row items-center justify-between gap-6">
             {prevPage ? (
               <Link href={prevPage} className="flex flex-col gap-2 p-6 rounded-xl border border-white/10 hover:border-primary/50 hover:bg-white/5 transition-all w-full md:w-1/2 text-left group">
                 <span className="text-xs text-muted-foreground uppercase tracking-wider flex items-center gap-2">
                   <ArrowLeft size={12} className="group-hover:-translate-x-1 transition-transform" /> Previous
                 </span>
                 <span className="font-bold text-lg text-foreground">{TOUR_TITLES[prevPage]}</span>
               </Link>
             ) : <div className="hidden md:block w-1/2" />}

             {nextPage && (
               <Link href={nextPage} className="flex flex-col gap-2 p-6 rounded-xl border border-white/10 hover:border-primary/50 hover:bg-white/5 transition-all w-full md:w-1/2 text-right group items-end">
                 <span className="text-xs text-muted-foreground uppercase tracking-wider flex items-center gap-2">
                   Next <ArrowRight size={12} className="group-hover:translate-x-1 transition-transform" />
                 </span>
                 <span className="font-bold text-lg text-foreground">{TOUR_TITLES[nextPage]}</span>
               </Link>
             )}
          </div>
        </article>
  );
}
