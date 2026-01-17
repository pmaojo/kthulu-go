## 2024-05-23 - [Parallelizing CMS File I/O]
**Learning:** Next.js Server Components allow async data fetching, which is a perfect opportunity to parallelize file I/O operations that were previously synchronous. In a file-based CMS, reading files one by one (especially with full content) is a major bottleneck during build time.
**Action:** When working with file-based content systems in Next.js, always prefer `fs.promises` and `Promise.all` over `fs.readFileSync` loops, and implement field selection to avoid loading unnecessary large content strings into memory.

## 2025-05-23 - [Redundant CMS File Lookups]
**Learning:** The CMS logic in `getAllDocs` was performing redundant filesystem checks (guessing `.md` vs `/index.md`) for every file it found, even though it already knew the exact file path from `readdir`.
**Action:** When iterating files to generate content lists, always pass the known full path to the reader function to bypass path resolution heuristics.
