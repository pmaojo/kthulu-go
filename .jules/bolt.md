## 2024-05-23 - [Parallelizing CMS File I/O]
**Learning:** Next.js Server Components allow async data fetching, which is a perfect opportunity to parallelize file I/O operations that were previously synchronous. In a file-based CMS, reading files one by one (especially with full content) is a major bottleneck during build time.
**Action:** When working with file-based content systems in Next.js, always prefer `fs.promises` and `Promise.all` over `fs.readFileSync` loops, and implement field selection to avoid loading unnecessary large content strings into memory.

## 2025-05-23 - [Redundant CMS File Lookups]
**Learning:** The CMS logic in `getAllDocs` was performing redundant filesystem checks (guessing `.md` vs `/index.md`) for every file it found, even though it already knew the exact file path from `readdir`.
**Action:** When iterating files to generate content lists, always pass the known full path to the reader function to bypass path resolution heuristics.

## 2025-05-23 - [Persisting UI in Next.js App Router]
**Learning:** In Next.js App Router, placing persistent UI elements (like a Sidebar) in `page.tsx` of a dynamic route (e.g., `[[...slug]]`) causes them to be part of the route's payload on every navigation. Moving them to `layout.tsx` ensures they persist, reducing the server response size and client-side reconciliation work.
**Action:** Always verify if shared UI components in catch-all routes should be moved to a `layout.tsx` to optimize navigation performance.

## 2025-05-24 - [EntityParser Map Allocation]
**Learning:** EntityParser was allocating a lookup map in `isBasicType` on every call, causing unnecessary GC pressure during parsing.
**Action:** Always hoist static lookup maps to package-level variables or `var` blocks outside the hot path to ensure zero-allocation lookups.

## 2025-05-24 - [Hoisting Static Maps in Resolver]
**Learning:** The DependencyResolver was allocating static maps (descriptions, categories, conflicts) on every method call. Moving these to package-level variables reduced allocations from ~28/op to ~4/op in hot paths.
**Action:** When working with static lookup data in Go, always hoist it to package-level variables or `init()` blocks to avoid redundant allocations.
