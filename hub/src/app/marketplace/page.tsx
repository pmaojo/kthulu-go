import { getMarketplaceItems } from '@/lib/cms';
import Link from 'next/link';
import { Zap, Shield, Cloud, Star, ArrowRight } from 'lucide-react';

export default function MarketplacePage() {
  const { starters, modules, plugins } = getMarketplaceItems();

  return (
    <div className="kthulu-container py-12">
      <div className="mb-16 text-center">
        <h1 className="text-5xl md:text-7xl font-black tracking-tighter mb-4 text-gradient kthulu-glow">Marketplace</h1>
        <p className="text-muted-foreground text-xl max-w-2xl mx-auto">
          Scale your Kthulu ecosystem with high-performance starters, modules, and community plugins.
        </p>
      </div>

      <section className="mb-20">
        <div className="flex items-center gap-4 mb-10 border-b border-white/5 pb-4">
          <div className="w-12 h-12 rounded-xl bg-primary/10 flex items-center justify-center text-primary shadow-[0_0_15px_rgba(0,255,102,0.1)]">
            <Zap size={28} />
          </div>
          <div>
            <h2 className="text-3xl font-black tracking-tight uppercase italic">Starters</h2>
            <p className="text-sm text-muted-foreground font-medium uppercase tracking-widest mt-1">Foundation Blueprints</p>
          </div>
        </div>
        <div className="kthulu-grid">
          {starters.map((item) => (
            <MarketplaceCard key={item.slug.join('/')} item={item} />
          ))}
        </div>
      </section>

      <section className="mb-20">
        <div className="flex items-center gap-4 mb-10 border-b border-white/5 pb-4">
          <div className="w-12 h-12 rounded-xl bg-secondary/10 flex items-center justify-center text-secondary shadow-[0_0_15px_rgba(168,85,247,0.1)]">
            <Shield size={28} />
          </div>
          <div>
            <h2 className="text-3xl font-black tracking-tight uppercase italic">Modules</h2>
            <p className="text-sm text-muted-foreground font-medium uppercase tracking-widest mt-1">Foundational Capabilities</p>
          </div>
        </div>
        <div className="kthulu-grid">
          {modules.map((item) => (
            <MarketplaceCard key={item.slug.join('/')} item={item} />
          ))}
        </div>
      </section>

      <section className="mb-20">
        <div className="flex items-center gap-4 mb-10 border-b border-white/5 pb-4">
          <div className="w-12 h-12 rounded-xl bg-accent/10 flex items-center justify-center text-accent shadow-[0_0_15px_rgba(249,115,22,0.1)]">
            <Cloud size={28} />
          </div>
          <div>
            <h2 className="text-3xl font-black tracking-tight uppercase italic">Plugins</h2>
            <p className="text-sm text-muted-foreground font-medium uppercase tracking-widest mt-1">Cloud & Infrastructure</p>
          </div>
        </div>
        <div className="kthulu-grid">
          {plugins.map((item) => (
            <MarketplaceCard key={item.slug.join('/')} item={item} />
          ))}
        </div>
      </section>
    </div>
  );
}

function MarketplaceCard({ item }: { item: any }) {
  const iconMap: Record<string, any> = {
    Zap: <Zap size={22} />,
    Shield: <Shield size={22} />,
    Cloud: <Cloud size={22} />,
  };

  const Icon = iconMap[item.frontmatter.icon] || <Zap size={22} />;

  return (
    <div className="kthulu-card group">
      <div className="flex items-start justify-between mb-8">
        <div className="w-14 h-14 rounded-2xl bg-white/[0.03] border border-white/[0.08] flex items-center justify-center text-primary group-hover:border-primary/40 group-hover:bg-primary/10 transition-all duration-500 shadow-inner">
          {Icon}
        </div>
        <div className="badge flex items-center gap-2">
          <Star size={12} className="fill-primary" />
          <span className="font-mono">{item.frontmatter.stars}</span>
        </div>
      </div>
      
      <h3 className="text-2xl font-black tracking-tight mb-4 group-hover:text-primary transition-colors duration-300 italic uppercase">
        {item.frontmatter.title}
      </h3>
      
      <p className="text-base text-muted-foreground/70 mb-10 line-clamp-3 leading-relaxed font-medium">
        {item.frontmatter.description}
      </p>
      
      <div className="flex items-center justify-between mt-auto pt-8 border-t border-white/[0.05]">
        <div className="flex flex-col">
          <span className="text-[9px] font-black uppercase tracking-[0.3em] text-muted-foreground/30 mb-1">Created By</span>
          <span className="text-[11px] font-bold uppercase tracking-widest text-muted-foreground/80">{item.frontmatter.author}</span>
        </div>
        <Link 
          href={`/docs/marketplace/${item.slug.join('/')}`}
          className="kthulu-btn !py-2 !px-4 !text-[10px] !rounded-lg"
        >
          Details <ArrowRight size={14} className="group-hover:translate-x-1 transition-transform" />
        </Link>
      </div>
    </div>
  );
}
