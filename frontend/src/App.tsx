import { Toaster } from "@/components/ui/toaster";
import { Toaster as Sonner } from "@/components/ui/sonner";
import { TooltipProvider } from "@/components/ui/tooltip";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { BrowserRouter, Routes, Route } from "react-router-dom";
import Index from "./pages/Index";
import NotFound from "./pages/NotFound";

const queryClient = new QueryClient();

const App = () => (
  <QueryClientProvider client={queryClient}>
    <TooltipProvider>
      <Toaster />
      <Sonner />
      <BrowserRouter>
        <Routes>
          <Route path="/" element={<Index />} />
          <Route path="/projects" element={<Index section="projects" />} />
          <Route path="/hub" element={<Index section="modules" />} />
          <Route path="/hub/:moduleId" element={<Index section="modules" />} />

          <Route path="/services" element={<Index section="services" />} />
          <Route path="/entities" element={<Index section="entities" />} />
          <Route path="/architecture" element={<Index section="architecture" />} />
          <Route path="/terminal" element={<Index section="terminal" />} />

          <Route path="/behavior-lab" element={<Index section="behavior-lab" />} />
          <Route path="/components" element={<Index section="components" />} />
          <Route path="/templates" element={<Index section="templates" />} />
          <Route path="/audit" element={<Index section="audit" />} />
          <Route path="/ai" element={<Index section="ai" />} />
          <Route path="/generate" element={<Index section="generate" />} />
          <Route path="/code" element={<Index section="code" />} />
          <Route path="/settings" element={<Index section="settings" />} />

          {/* ADD ALL CUSTOM ROUTES ABOVE THE CATCH-ALL "*" ROUTE */}
          <Route path="*" element={<NotFound />} />
        </Routes>
      </BrowserRouter>
    </TooltipProvider>
  </QueryClientProvider>
);

export default App;
