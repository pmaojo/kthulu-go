## 2024-05-23 - [Parallelizing CMS File I/O]
**Learning:** Next.js Server Components allow async data fetching, which is a perfect opportunity to parallelize file I/O operations that were previously synchronous. In a file-based CMS, reading files one by one (especially with full content) is a major bottleneck during build time.
**Action:** When working with file-based content systems in Next.js, always prefer `fs.promises` and `Promise.all` over `fs.readFileSync` loops, and implement field selection to avoid loading unnecessary large content strings into memory.
