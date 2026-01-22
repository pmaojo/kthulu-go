import Link from "next/link";
import { ArrowRight, ShoppingBag, Terminal, Sparkles, Shield } from "lucide-react";

export default function Home() {
  return (
    <div className="flex flex-col items-center justify-center min-h-[calc(100vh-64px)] px-6 text-center">
      <div className="max-w-4xl space-y-8">
        <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full border border-primary/20 bg-primary/5 text-primary text-xs font-bold tracking-widest uppercase mb-8">
          <Sparkles size={12} /> Version 1.0 is here
        </div>
        
        <h1 className="text-6xl md:text-8xl font-black tracking-tighter leading-[0.9] kthulu-glow">
          BUILD FAST.<br />FAIL NEVER.
        </h1>
        
        <p className="text-xl text-muted-foreground max-w-2xl mx-auto leading-relaxed">
          The high-performance AI-powered CLI framework and ecosystem. 
          Discover modules, read documentation, and join the Kthulu revolution.
        </p>

        <div className="flex flex-wrap items-center justify-center gap-4 pt-8">
          <Link href="/docs/guide/introduction" className="kthulu-btn px-8 py-4 text-lg gap-2">
            Get Started <ArrowRight size={20} />
          </Link>
          <Link href="/marketplace" className="glass border border-white/10 px-8 py-4 rounded font-bold text-lg hover:bg-white/5 transition-all flex items-center gap-2">
            Explore Marketplace <ShoppingBag size={20} />
          </Link>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-6 pt-24">
          <FeatureCard 
            icon={<Terminal size={24} />} 
            title="AI Scaffolding" 
            desc="Natural language to production code in seconds."
          />
          <FeatureCard 
            icon={<Shield size={24} />} 
            title="Self-Healing" 
            desc="Auto-diagnose and fix panics in real-time."
          />
          <FeatureCard 
            icon={<ShoppingBag size={24} />} 
            title="Marketplace" 
            desc="One-click modules for your entire stack."
          />
        </div>
      </div>
    </div>
  );
}

function FeatureCard({ icon, title, desc }: { icon: React.ReactNode, title: string, desc: string }) {
  return (
    <div className="kthulu-card text-left">
      <div className="text-primary mb-4">{icon}</div>
      <h3 className="text-lg font-bold mb-2">{title}</h3>
      <p className="text-sm text-muted-foreground">{desc}</p>
    </div>
  );
}
