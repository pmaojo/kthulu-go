import { getMarketplaceItems } from '@/lib/cms';
import Link from 'next/link';
import { Zap, Shield, Cloud, Star, ArrowRight } from 'lucide-react';

export default function MarketplacePage() {
  const { starters, modules, plugins } = getMarketplaceItems();

  return (
    <div className="max-w-7xl mx-auto px-6 py-12">
      <div className="mb-12">
        <h1 className="text-4xl font-bold tracking-tighter mb-4 kthulu-glow">Marketplace</h1>
        <p className="text-muted-foreground text-lg max-w-2xl">
          Extend your Kthulu project with community-built starters, modules, and plugins.
        </p>
      </div>

      <section className="mb-16">
        <div className="flex items-center gap-3 mb-8">
          <div className="w-10 h-10 rounded-lg bg-primary/20 flex items-center justify-center text-primary">
            <Zap size={24} />
          </div>
          <h2 className="text-2xl font-bold tracking-tight">Starters</h2>
        </div>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {starters.map((item) => (
            <MarketplaceCard key={item.slug.join('/')} item={item} />
          ))}
        </div>
      </section>

      <section className="mb-16">
        <div className="flex items-center gap-3 mb-8">
          <div className="w-10 h-10 rounded-lg bg-secondary/20 flex items-center justify-center text-secondary">
            <Shield size={24} />
          </div>
          <h2 className="text-2xl font-bold tracking-tight">Modules</h2>
        </div>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {modules.map((item) => (
            <MarketplaceCard key={item.slug.join('/')} item={item} />
          ))}
        </div>
      </section>

      <section className="mb-16">
        <div className="flex items-center gap-3 mb-8">
          <div className="w-10 h-10 rounded-lg bg-accent/20 flex items-center justify-center text-accent">
            <Cloud size={24} />
          </div>
          <h2 className="text-2xl font-bold tracking-tight">Plugins</h2>
        </div>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
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
    Zap: <Zap size={18} />,
    Shield: <Shield size={18} />,
    Cloud: <Cloud size={18} />,
  };

  const Icon = iconMap[item.frontmatter.icon] || <Zap size={18} />;

  return (
    <div className="kthulu-card group">
      <div className="flex items-start justify-between mb-4">
        <div className="p-2 rounded bg-white/5 text-primary group-hover:bg-primary group-hover:text-primary-foreground transition-all">
          {Icon}
        </div>
        <div className="flex items-center gap-1 text-xs text-muted-foreground bg-white/5 px-2 py-1 rounded">
          <Star size={12} className="text-accent" />
          {item.frontmatter.stars}
        </div>
      </div>
      <h3 className="text-xl font-bold mb-2 group-hover:text-primary transition-colors">{item.frontmatter.title}</h3>
      <p className="text-sm text-muted-foreground mb-6 line-clamp-2">{item.frontmatter.description}</p>
      
      <div className="flex items-center justify-between mt-auto">
        <span className="text-xs font-mono text-muted-foreground">by {item.frontmatter.author}</span>
        <Link 
          href={`/docs/marketplace/${item.slug.slice(1).join('/')}`}
          className="text-xs font-bold flex items-center gap-1 hover:text-primary transition-colors uppercase tracking-widest"
        >
          Details <ArrowRight size={14} />
        </Link>
      </div>
    </div>
  );
}
